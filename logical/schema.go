package logical

import (
	"reflect"
	"strings"
)

type EmptySchema struct{}
type Errors map[int]string


type SchemaReply struct {
	Namespaces Namespaces `json:"namespace_schemas"`
}

type Namespaces []*Namespace
type Namespace struct {
	Namespace   string                `json:"namespace"`
	Description string                `json:"description"`
	Operations  map[string]*Operation `json:"operations"`
}

//Operation属性
type Operation struct {
	Description string   `json:"description"`
	Authorized  bool     `json:"authorized"`
	Deprecated  bool     `json:"deprecated"`
	Input       []*Field `json:"input,omitempty"`
	Output      []*Field `json:"output,omitempty"`
	Errors      Errors   `json:"errors,omitempty"`
}
//接口属性列
type Field struct {
	Field      string   `json:"field"`
	Name       string   `json:"name"`
	Kind       string   `json:"kind"`
	Required   bool     `json:"required"`
	Deprecated bool     `json:"deprecated"`
	Reference  []*Field `json:"reference"`
	Example    string   `json:"example"`
	IsList     bool     `json:"is_list"`
}

//列属性猜解
type Attribute struct {
	Type reflect.Type
}
func (s *Attribute) Fields() ([]*Field, error) {
	return getFields(s.Type), nil
}

func getType(t reflect.Type) reflect.Type {
	switch t.Kind() {
	case reflect.Ptr:
		return t.Elem()
	case reflect.Struct:
		return t
	case reflect.Slice:
		fallthrough
	case reflect.Map:
		return getType(t.Elem())
	default:
		return t
	}
}

func getFields(Type reflect.Type) []*Field {
	defer func() {
		recover()
	}()
	var fields []*Field
	for i := 0; i < Type.NumField(); i++ {
		field := new(Field)
		f := Type.Field(i)
		isList := f.Type.Kind() == reflect.Slice
		realType := getType(f.Type)
		kindString := realType.Kind().String()
		//fmt.Println(f.Name, realType.Kind(), realType.Kind().String(), realType.Name())
		if realType.Kind() == reflect.Struct && realType.Name() != "Time" {
			field.Reference = getFields(realType)
		}
		if realType.Name() == "Time" {
			kindString = "datetime"
		}

		fValue := f.Tag.Get("json")
		if fValue == "" {
			fValue = f.Name
		}
		fName := f.Tag.Get("name")
		if fName == "" {
			fName = f.Name
		}

		example := f.Tag.Get("example")
		validate := f.Tag.Get("validate")
		required := validate != "" && strings.Contains(strings.ToLower(validate), "required")

		field.Name = fName
		field.Field = fValue
		field.Required = required
		field.Kind = kindString
		field.Deprecated = f.Tag.Get("deprecated") != ""
		field.Example = example
		field.IsList = isList
		fields = append(fields, field)
	}
	return fields
}

