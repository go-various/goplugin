package goplugin

import (
	"context"
	"github.com/go-various/goplugin/logical"
	"github.com/go-various/goplugin/proto"
	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

var _ proto.BackendServer = (*backendGRPCPluginServer)(nil)

type backendGRPCPluginServer struct {
	broker         *plugin.GRPCBroker
	backend        logical.Backend
	factory        logical.Factory
	brokeredClient *grpc.ClientConn
	logger         log.Logger
}

func (b *backendGRPCPluginServer) mustEmbedUnimplementedBackendServer() {
	panic("implement me")
}

func (b *backendGRPCPluginServer) Name(ctx context.Context, empty *proto.Empty) (*proto.NameReply, error) {
	name := b.backend.Name()
	return &proto.NameReply{Name: name}, nil
}

func (b *backendGRPCPluginServer) SchemaRequest(ctx context.Context, empty *proto.Empty) (*proto.SchemaRequestReply, error) {
	resp, err := b.backend.SchemaRequest(ctx)
	if err != nil {
		return nil, err
	}
	return proto.LogicalNamespaceSchemasToProtoNamespaceSchemas(resp), nil
}

func (b *backendGRPCPluginServer) Setup(ctx context.Context, args *proto.SetupArgs) (*proto.SetupReply, error) {
	// Dial for storage
	brokeredClient, err := b.broker.Dial(args.BrokerId)
	if err != nil {
		return &proto.SetupReply{}, err
	}

	b.brokeredClient = brokeredClient
	consul := newGRPCConsulClient(brokeredClient)
	config := &logical.BackendConfig{
		ConsulView:  consul,
		Logger:      b.logger,
		Config:      args.Config,
		BackendUUID: args.BackendUUID,
	}

	// Call the underlying grpc-backend factory after shims have been created
	// to set b.grpc-backend
	backend, err := b.factory(ctx, config)
	if err != nil {
		return &proto.SetupReply{
			Err: proto.ErrToString(err),
		}, nil
	}
	b.backend = backend
	return &proto.SetupReply{}, nil
}

type params struct {
	data []byte
}

func (p params) Encode() []byte {
	return p.data
}

func (p params) Decode(out interface{}) error {
	return nil
}

func (b *backendGRPCPluginServer) Initialize(ctx context.Context, args *proto.InitializeArgs) (*proto.InitializeReply, error) {

	req := &logical.InitializationRequest{
		Params: params{data: args.Params},
	}

	respErr := b.backend.Initialize(ctx, req)

	return &proto.InitializeReply{
		Err: proto.ErrToProtoErr(respErr),
	}, nil
}

func (b *backendGRPCPluginServer) HandleRequest(ctx context.Context, args *proto.HandleRequestArgs) (*proto.HandleRequestReply, error) {

	logicalReq, err := proto.ProtoRequestToLogicalRequest(args.Request)
	if err != nil {
		return &proto.HandleRequestReply{}, err
	}
	resp, respErr := b.backend.HandleRequest(ctx, logicalReq)
	if respErr != nil {
		return nil, respErr
	}

	pbResp, err := proto.LogicalResponseToProtoResponse(resp)
	if err != nil {
		return &proto.HandleRequestReply{}, err
	}

	return &proto.HandleRequestReply{
		Response: pbResp,
		Err:      proto.ErrToProtoErr(respErr),
	}, nil
}

func (b *backendGRPCPluginServer) Cleanup(ctx context.Context, empty *proto.Empty) (*proto.Empty, error) {
	b.backend.Cleanup(ctx)
	// Close rpc clients
	b.brokeredClient.Close()
	return &proto.Empty{}, nil
}

func (b *backendGRPCPluginServer) Type(ctx context.Context, _ *proto.Empty) (*proto.TypeReply, error) {
	return &proto.TypeReply{
		Type: uint32(b.backend.Type()),
	}, nil
}

func (b backendGRPCPluginServer) Version(ctx context.Context, _ *proto.Empty) (*proto.VersionReply, error) {
	version := b.backend.Version(ctx)

	return &proto.VersionReply{Version: version}, nil
}
