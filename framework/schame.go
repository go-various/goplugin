package framework

import (
	"fmt"
	"github.com/go-various/goplugin/logical"
)

func (b *Backend) initSchemaOnce() {
	_ = b.initSchema()
}

func (b *Backend) initSchema() error {
	schemas := logical.Namespaces{}
	for _, ns := range b.Namespaces {
		if ns.Description == "" {
			return fmt.Errorf("namespace[%s] description required", ns.Pattern)
		}
		namespace := logical.Namespace{
			Namespace:   ns.Pattern,
			Description: ns.Description,
			Operations:  make(map[string]*logical.Operation),
		}
		for opt, handler := range ns.Operations {
			properties := handler.Properties()
			if properties.Description == "" {
				return descriptionError(ns.Pattern, opt)
			}
			input, err := properties.Input.Fields()
			if err != nil {
				return schemaError(ns.Pattern, opt, err)
			}
			output, err := properties.Output.Fields()
			if err != nil {
				return schemaError(ns.Pattern, opt, err)
			}
			schema := &logical.Operation{
				Description: properties.Description,
				Authorized:  properties.Authorized,
				Deprecated:  properties.Deprecated,
				Input:       input,
				Output:      output,
				Errors:      properties.Errors,
			}
			namespace.Operations[opt] = schema
		}
		schemas = append(schemas, &namespace)
	}
	b.entities = schemas
	return nil
}

func schemaError(pattern string, operation string, err error) error {
	return fmt.Errorf("namespace[%s] operation[%s] %s", pattern, operation, err)
}
func descriptionError(pattern string, operation string) error {
	return fmt.Errorf("namespace[%s] operation[%s] Description required", pattern, operation)
}
