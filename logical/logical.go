package logical

import (
	"context"
	"github.com/go-various/goplugin/pluginutil"
	log "github.com/hashicorp/go-hclog"
)

const BackendExecuteName = "backend.plugin"

// BackendType is the type of grpc-backend that is being implemented
type BackendType uint32

// The these are the types of backends that can be derived from
// logical.Backend
const (
	TypeUnknown BackendType = 0 // This is also the zero-value for BackendType
	TypeBuiltin BackendType = 1
	TypeLogical BackendType = 2
)

// BackendConfig is provided to the factory to initialize the grpc-backend
type BackendConfig struct {
	LookupUtil    pluginutil.PluginLookupUtil
	ConsulView    Consul
	Logger        log.Logger
	BackendUUID   string
	Config        map[string]string
}

type Factory func(context.Context, *BackendConfig) (Backend, error)

type Backend interface {
	Setup(context.Context, *BackendConfig) error
	Initialize(context.Context, *InitializationRequest) error
	SchemaRequest(context.Context) (*SchemaReply, error)
	HandleRequest(context.Context, *Request) (*Response, error)
	Cleanup(context.Context)
	Logger() log.Logger
	Type() BackendType
	Version(ctx context.Context) string
	Name()string
}
