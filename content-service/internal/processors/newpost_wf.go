package processors

import (
	processorsDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	"go.temporal.io/sdk/workflow"
)

func NewPostWorkflow(ctx workflow.Context, input *processorsDomain.NewPostWorkflowInput) error {
	panic("unimplemented")
}
