package logical

import (
	"context"
	"github.com/go-various/goplugin/framework"
	"github.com/go-various/goplugin/logical"
	"github.com/go-various/helper/jsonutil"
)

//获取插件的操作名称
func (m *Transport) getOperation(backend logical.Backend, request *logical.Request) (
	*logical.Operation, error) {

	result, err := backend.SchemaRequest(context.Background())
	if nil != err {
		return nil, err
	}

	var schema *logical.Namespace
	for _, n := range result.Namespaces {
		if n.Namespace == request.Namespace {
			schema = n
		}
	}
	if nil == schema {
		return nil, framework.ErrNamespaceNotExists
	}
	for operation, properties := range schema.Operations {
		if operation == request.Operation {
			return properties, nil
		}
	}
	return nil, framework.ErrOperationNotExists
}
func (m *Transport) authorization(backend logical.Backend, request *logical.Request) (authResp *logical.Response, err error) {
	defer func() {
		if err != nil {
			m.Logger.Error("authorization", "request", request, "Err", err)
		}
	}()
	var operation *logical.Operation
	operation, err = m.getOperation(backend, request)
	if nil != err {
		return nil, err
	}
	if !operation.Authorized {
		return &logical.Response{
			ResultCode: 0,
			ResultMsg:  "",
			Content:    &logical.Content{Data: &logical.Authorized{}},
		}, nil
	}
	authBackend, has := m.PluginManager.GetBackend(m.authMethod.Backend)
	if !has {
		err = logical.ErrAuthMethodNotFound
		return nil, err
	}

	authReq := logical.Request{
		Operation:  m.authMethod.Operation,
		Namespace:  m.authMethod.Namespace,
		Token:      request.Token,
		Data:       request.Data,
		Connection: request.Connection,
	}
	authResp, err = authBackend.HandleRequest(m.Ctx, &authReq)
	if nil != err {
		return nil, err
	}
	if m.Logger.IsTrace() {
		m.Logger.Trace("auth reply", "reply", jsonutil.EncodeToString(authResp))
	}
	return authResp, nil
}
