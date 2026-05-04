package dresources

import (
	"context"

	"github.com/databricks/cli/libs/utils"
	"github.com/databricks/cli/ucm/config/resources"
	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/catalog"
)

type ResourceConnection struct {
	client *databricks.WorkspaceClient
}

func (*ResourceConnection) New(client *databricks.WorkspaceClient) *ResourceConnection {
	return &ResourceConnection{client: client}
}

func (*ResourceConnection) PrepareState(input *resources.Connection) *catalog.CreateConnection {
	return &input.CreateConnection
}

func (*ResourceConnection) RemapState(info *catalog.ConnectionInfo) *catalog.CreateConnection {
	return &catalog.CreateConnection{
		Comment:         info.Comment,
		ConnectionType:  info.ConnectionType,
		Name:            info.Name,
		Options:         info.Options,
		Properties:      info.Properties,
		ReadOnly:        info.ReadOnly,
		ForceSendFields: utils.FilterFields[catalog.CreateConnection](info.ForceSendFields),
	}
}

func (r *ResourceConnection) DoRead(ctx context.Context, id string) (*catalog.ConnectionInfo, error) {
	return r.client.Connections.GetByName(ctx, id)
}

func (r *ResourceConnection) DoCreate(ctx context.Context, config *catalog.CreateConnection) (string, *catalog.ConnectionInfo, error) {
	response, err := r.client.Connections.Create(ctx, *config)
	if err != nil || response == nil {
		return "", nil, err
	}
	return response.Name, response, nil
}

func (r *ResourceConnection) DoUpdate(ctx context.Context, id string, config *catalog.CreateConnection, _ *PlanEntry) (*catalog.ConnectionInfo, error) {
	updateRequest := catalog.UpdateConnection{
		Name:            id,
		Options:         config.Options,
		Owner:           "",
		ForceSendFields: utils.FilterFields[catalog.UpdateConnection](config.ForceSendFields, "Owner"),
	}

	return r.client.Connections.Update(ctx, updateRequest)
}

func (r *ResourceConnection) DoDelete(ctx context.Context, id string) error {
	return r.client.Connections.DeleteByName(ctx, id)
}

// DoUpdateWithID renames a connection in place via the UpdateConnection
// API's NewName field. Triggered by update_id_on_changes on the name
// field — without this method the planner would unnecessarily recreate
// the connection just to rename it.
func (r *ResourceConnection) DoUpdateWithID(ctx context.Context, id string, config *catalog.CreateConnection) (string, *catalog.ConnectionInfo, error) {
	updateRequest := catalog.UpdateConnection{
		Name:            id,
		Options:         config.Options,
		Owner:           "",
		ForceSendFields: utils.FilterFields[catalog.UpdateConnection](config.ForceSendFields, "Owner"),
	}

	if config.Name != id {
		updateRequest.NewName = config.Name
	}

	response, err := r.client.Connections.Update(ctx, updateRequest)
	if err != nil || response == nil {
		return "", nil, err
	}

	return response.Name, response, nil
}
