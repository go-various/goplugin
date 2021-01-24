package proto

import (
	"encoding/json"
	"errors"
	"github.com/go-various/goplugin/logical"
)

const (
	ErrTypeUnknown uint32 = iota
	ErrTypeCodedError
)

func ProtoErrToErr(e *ProtoError) error {
	if e == nil {
		return nil
	}

	var err error
	switch e.ErrType {
	case ErrTypeUnknown:
		err = errors.New(e.ErrMsg)
	case ErrTypeCodedError:
		err = logical.NewCodedError(int(e.ErrCode), e.ErrMsg)
	}

	return err
}

func ErrToString(e error) string {
	if e == nil {
		return ""
	}

	return e.Error()
}
func ErrToProtoErr(e error) *ProtoError {
	if e == nil {
		return nil
	}
	pbErr := &ProtoError{
		ErrMsg:  e.Error(),
	}

	switch e.(type) {
	case logical.CodedError:
		pbErr.ErrType = ErrTypeCodedError
	default:
		pbErr.ErrType = ErrTypeUnknown
	}

	return pbErr
}

func LogicalRequestToProtoRequest(r *logical.Request) (*Request, error) {

	if r == nil {
		return nil, errors.New("request is null")
	}

	headers := map[string]*Header{}
	for k, v := range r.Headers {
		headers[k] = &Header{Header: v}
	}
	authBytes, err := json.Marshal(r.Authorized)
	if nil != err {
		return nil, err
	}
	return &Request{
		Id:         r.ID,
		Operation:  r.Operation,
		Namespace:  r.Namespace,
		Data:       r.Data,
		Token:      r.Token,
		Authorized: authBytes,
		Headers:    headers,
	}, nil
}

func ProtoRequestToLogicalRequest(r *Request) (*logical.Request, error) {
	if r == nil {
		return nil, nil
	}

	var headers map[string][]string
	if len(r.Headers) > 0 {
		headers = make(map[string][]string, len(r.Headers))
		for k, v := range r.Headers {
			headers[k] = v.Header
		}
	}
	var authorized logical.Authorized
	if err := json.Unmarshal(r.Authorized, &authorized); err != nil {
		return nil, err
	}
	return &logical.Request{
		ID:         r.Id,
		Operation:  r.Operation,
		Namespace:  r.Namespace,
		Token:      r.Token,
		Authorized: &authorized,
		Data:       r.Data,
		Headers:    headers,
		Connection: ProtoConnectionToLogicalConnection(r.Connection),
	}, nil
}

func LogicalConnectionToProtoConnection(c *logical.Connection) *Connection {
	if c == nil {
		return nil
	}

	return &Connection{
		RemoteAddr: c.RemoteAddr,
	}
}

func ProtoConnectionToLogicalConnection(c *Connection) *logical.Connection {
	if c == nil {
		return nil
	}

	return &logical.Connection{
		RemoteAddr: c.RemoteAddr,
	}
}

func LogicalResponseToProtoResponse(r *logical.Response) (*HandlerResponse, error) {
	if r == nil {
		return nil, nil
	}

	buf, err := json.Marshal(r.Content)
	if err != nil {
		return nil, err
	}

	headers := map[string]*Header{}
	for k, v := range r.Headers {
		headers[k] = &Header{Header: v}
	}

	return &HandlerResponse{
		ResultCode: r.ResultCode,
		ResultMsg:  r.ResultMsg,
		Data:       string(buf),
		Headers:    headers,
	}, nil
}
func ProtoResponseToLogicalResponse(r *HandlerResponse) (*logical.Response, error) {
	if r == nil {
		return nil, nil
	}

	data := logical.Content{}
	err := json.Unmarshal([]byte(r.Data), &data)
	if err != nil {
		return nil, err
	}

	var headers map[string][]string
	if len(r.Headers) > 0 {
		headers = make(map[string][]string, len(r.Headers))
		for k, v := range r.Headers {
			headers[k] = v.Header
		}
	}

	return &logical.Response{
		ResultCode: r.ResultCode,
		ResultMsg:  r.ResultMsg,
		Content:    &data,
		Headers:    headers,
	}, nil
}

func protoFieldToLogicalField(fields []*Field) []*logical.Field {
	if nil == fields || len(fields) == 0 {
		return []*logical.Field{}
	}
	var outFields []*logical.Field
	for _, field := range fields {
		of := logical.Field{
			Field:      field.Field,
			Name:       field.Name,
			Kind:       field.Kind,
			Required:   field.Required,
			Deprecated: field.Deprecated,
			IsList:     field.IsList,
			Example:    field.Example,
		}
		if field.Reference != nil {
			of.Reference = protoFieldToLogicalField(field.Reference)
		}
		outFields = append(outFields, &of)
	}
	return outFields
}
func ProtoNamespaceSchemasToLigicalNamespaceSchemas(ns *SchemaRequestReply) *logical.SchemaReply {
	var schemas logical.Namespaces
	for _, schema := range ns.NamespaceSchemas {
		operations := map[string]*logical.Operation{}
		for key, opt := range schema.Operations {
			operations[key] = &logical.Operation{
				Description: opt.Description,
				Authorized:  opt.Authorized,
				Deprecated:  opt.Deprecated,

				Input:  protoFieldToLogicalField(opt.Input),
				Output: protoFieldToLogicalField(opt.Output),
			}
		}

		sc := logical.Namespace{
			Namespace:   schema.Namespace,
			Description: schema.Description,
			Operations:  operations,
		}
		schemas = append(schemas, &sc)
	}
	response := &logical.SchemaReply{
		Namespaces: schemas,
	}
	return response
}

func logicalFieldToProtoField(fields []*logical.Field) []*Field {
	if nil == fields || len(fields) == 0 {
		return []*Field{}
	}
	var outFields []*Field
	for _, field := range fields {
		of := Field{
			Field:      field.Field,
			Name:       field.Name,
			Kind:       field.Kind,
			Required:   field.Required,
			Deprecated: field.Deprecated,
			Example:    field.Example,
			IsList:     field.IsList,
		}
		if field.Reference != nil {
			of.Reference = logicalFieldToProtoField(field.Reference)
		}
		outFields = append(outFields, &of)
	}
	return outFields
}

func LogicalNamespaceSchemasToProtoNamespaceSchemas(ns *logical.SchemaReply) *SchemaRequestReply {
	var schemas []*NamespaceSchema
	for _, schema := range ns.Namespaces {
		operations := map[string]*Schema{}
		for key, opt := range schema.Operations {
			operations[string(key)] = &Schema{
				Description: opt.Description,
				Authorized:  opt.Authorized,
				Deprecated:  opt.Deprecated,
				Input:       logicalFieldToProtoField(opt.Input),
				Output:      logicalFieldToProtoField(opt.Output),
			}
		}

		sc := NamespaceSchema{
			Namespace:   schema.Namespace,
			Description: schema.Description,
			Operations:  operations,
		}
		schemas = append(schemas, &sc)
	}
	response := &SchemaRequestReply{
		NamespaceSchemas: schemas,
	}
	return response
}
