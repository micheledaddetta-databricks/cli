package phases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/databricks/cli/libs/log"
	"github.com/databricks/cli/libs/logdiag"
	"github.com/databricks/cli/libs/structs/structdiff"
	"github.com/databricks/cli/ucm"
	"github.com/databricks/cli/ucm/config"
	"github.com/databricks/cli/ucm/direct/dresources"
	"github.com/databricks/cli/ucm/direct/dstate"
	"github.com/databricks/databricks-sdk-go/apierr"
)

// FieldDrift records a single drift finding for one field of one resource.
// State is what ucm's recorded state has; Live is what the SDK read back.
// Values are rendered via fmt.Sprintf("%v", ...) at display time so nested
// maps/slices survive the JSON/text output paths without a custom marshaller.
type FieldDrift struct {
	Field string `json:"field"`
	State any    `json:"state"`
	Live  any    `json:"live"`
}

// ResourceDrift bundles all field-level drift for a single state entry.
// Key is the ucm plan key (e.g. "resources.catalogs.sales").
type ResourceDrift struct {
	Key    string       `json:"key"`
	Fields []FieldDrift `json:"fields"`
}

// Report is the full drift result returned by Drift.
type Report struct {
	Drift []ResourceDrift `json:"drift"`
}

// HasDrift reports whether the report contains any drift findings.
func (r *Report) HasDrift() bool { return len(r.Drift) > 0 }

// Drift opens the direct-engine state for the current target and compares
// every recorded resource field-by-field against the live UC read served by
// the resource adapter's DoRead. The resulting *Report is returned for the
// caller to render. Errors are reported via logdiag; on error Drift returns nil.
//
// Drift always routes through the direct SDK client regardless of the
// configured engine — reading live UC state is an SDK-level operation and
// the terraform engine has no native concept of it. The terraform-engine
// path still needs a reader for its own recorded state; until that lands,
// drift on a terraform-engine target reports nothing (the direct state file
// is absent) and the verb's long help calls out the limitation.
//
// Grants are skipped: the UC Grants API returns an authoritative set
// per-securable which doesn't map cleanly onto ucm's per-key grant state.
func Drift(ctx context.Context, u *ucm.Ucm, _ Options) *Report {
	log.Info(ctx, "Phase: drift")

	client, err := u.WorkspaceClientE()
	if err != nil {
		logdiag.LogError(ctx, fmt.Errorf("resolve workspace client: %w", err))
		return nil
	}

	adapters, err := dresources.InitAll(client)
	if err != nil {
		logdiag.LogError(ctx, fmt.Errorf("init resource adapters: %w", err))
		return nil
	}

	var db dstate.DeploymentState
	if err := db.Open(DirectStatePath(u)); err != nil {
		logdiag.LogError(ctx, fmt.Errorf("open direct state: %w", err))
		return nil
	}

	keys := make([]string, 0, len(db.Data.State))
	for key := range db.Data.State {
		// Skip grants — the UC Grants API surface doesn't map cleanly onto
		// per-key state; this mirrors the legacy ComputeDrift's decision.
		if strings.Contains(key, ".grants") {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	r := &Report{}
	for _, key := range keys {
		entry := db.Data.State[key]

		group := config.GetResourceTypeFromKey(key)
		adapter, ok := adapters[group]
		if !ok {
			logdiag.LogError(ctx, fmt.Errorf("drift %s: unknown resource type %q", key, group))
			return nil
		}

		fields, err := computeResourceDrift(ctx, adapter, entry)
		if err != nil {
			logdiag.LogError(ctx, fmt.Errorf("drift %s: %w", key, err))
			return nil
		}

		if len(fields) > 0 {
			r.Drift = append(r.Drift, ResourceDrift{Key: key, Fields: fields})
		}
	}

	return r
}

// computeResourceDrift reads the live state via adapter.DoRead and structdiffs
// it against the saved state. Treats a missing remote (404) as a single
// "_exists: true→false" drift entry so out-of-band deletes surface without
// returning the underlying not-found error.
func computeResourceDrift(ctx context.Context, adapter *dresources.Adapter, entry dstate.ResourceEntry) ([]FieldDrift, error) {
	savedState, err := parseSavedState(adapter.StateType(), entry.State)
	if err != nil {
		return nil, fmt.Errorf("interpreting saved state: %w", err)
	}

	liveRemote, err := adapter.DoRead(ctx, entry.ID)
	if err != nil {
		if isResourceGone(err) {
			return []FieldDrift{{Field: "_exists", State: true, Live: false}}, nil
		}
		return nil, fmt.Errorf("reading id=%q: %w", entry.ID, err)
	}

	live, err := adapter.RemapState(liveRemote)
	if err != nil {
		return nil, fmt.Errorf("remapping live state: %w", err)
	}

	changes, err := structdiff.GetStructDiff(savedState, live, adapter.KeyedSlices())
	if err != nil {
		return nil, fmt.Errorf("diffing live state: %w", err)
	}

	fields := make([]FieldDrift, 0, len(changes))
	for _, ch := range changes {
		fields = append(fields, FieldDrift{
			Field: ch.Path.String(),
			State: ch.Old,
			Live:  ch.New,
		})
	}
	return fields, nil
}

// parseSavedState decodes a state-JSON blob into the adapter's saved-state
// Go type. Mirrors ucm/direct.parseState (unexported); kept here so phases
// doesn't need to expand the direct package's surface for a single caller.
func parseSavedState(destType reflect.Type, raw json.RawMessage) (any, error) {
	destPtr := reflect.New(destType).Interface()
	if err := json.Unmarshal(raw, destPtr); err != nil {
		return nil, fmt.Errorf("unmarshalling into %s: %w", destType, err)
	}
	return reflect.ValueOf(destPtr).Elem().Interface(), nil
}

// isResourceGone returns true when err signals the remote object no longer
// exists. Mirrors ucm/direct.isResourceGone (unexported).
func isResourceGone(err error) bool {
	return errors.Is(err, apierr.ErrResourceDoesNotExist) || errors.Is(err, apierr.ErrNotFound)
}
