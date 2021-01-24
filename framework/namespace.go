package framework

import (
	"github.com/go-various/goplugin/logical"
	"strings"
)

// NamespaceAppend is a helper for appending lists of paths into a single
// list.
func NamespaceAppend(paths ...[]*Namespace) []*Namespace {
	result := make([]*Namespace, 0, 10)
	for _, ps := range paths {
		result = append(result, ps...)
	}
	return result
}

// Namespace is a single path that the grpc-backend responds to.
type Namespace struct {
	Pattern        string
	Description    string
	Operations     map[string]OperationHandler
	ExistenceCheck ExistenceFunc
	Deprecated     bool
}

// OperationHandler defines and describes a specific operation handler.
type OperationHandler interface {
	Handler() OperationFunc
	Properties() OperationProperties
}

// OperationProperties describes an operation for documentation, help text,
// and other clients. A Summary should always be provided, whereas other
// fields can be populated as needed.
type OperationProperties struct {
	Description string
	Authorized  bool
	Deprecated  bool
	Input       *logical.Attribute `json:"-"`
	Output      *logical.Attribute `json:"-"`
	Errors      logical.Errors      `json:"errors"`
}

type Response struct {
	Description string // summary of the the response and should always be provided
	MediaType   string // media type of the response, defaulting to "application/json" if empty
}

// NamespaceOperation is a concrete implementation of OperationHandler.
type NamespaceOperation struct {
	Callback    OperationFunc
	Description string
	Authorized  bool
	Deprecated  bool
	Input       *logical.Attribute
	Output      *logical.Attribute
	Errors      logical.Errors `json:"errors"`
}

func (p *NamespaceOperation) Handler() OperationFunc {
	return p.Callback
}

func (p *NamespaceOperation) Properties() OperationProperties {
	return OperationProperties{
		Description: strings.TrimSpace(p.Description),
		Deprecated:  p.Deprecated,
		Authorized:  p.Authorized,
		Input:       p.Input,
		Output:      p.Output,
		Errors:      p.Errors,
	}
}
