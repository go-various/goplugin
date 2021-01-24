package logical

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/copystructure"
	"reflect"
	"strings"
)

var ErrInvalidData = errors.New("invalid request data")

type Request struct {
	ID         string              `json:"id" structs:"id" mapstructure:"id"`
	Operation  string              `json:"operation" structs:"operation" mapstructure:"operation"`
	Namespace  string              `json:"namespace" structs:"path" mapstructure:"namespace"`
	Data       []byte              `json:"map" structs:"data" mapstructure:"data"`
	Authorized *Authorized         `json:"authorized" structs:"authorized" mapstructure:"authorized"`
	Token      string              `json:"token" structs:"token" mapstructure:"token"`
	Headers    map[string][]string `json:"headers" structs:"headers" mapstructure:"headers"`
	Connection *Connection         `json:"connection" structs:"connection" mapstructure:"connection"`
}

func (r *Request) GetData() []byte {
	return r.Data
}

func (r *Request) JSONDecode(out interface{}) error {
	if r.Data == nil {
		return ErrInvalidData
	}
	err := json.Unmarshal(r.Data, out)
	if err != nil {
		return err
	}
	return r.Validate(out)
}

func (r *Request) XMLDecode(out interface{}) error {
	if r.Data == nil {
		return ErrInvalidData
	}
	err := xml.Unmarshal(r.Data, out)
	if err != nil {
		return err
	}

	return r.Validate(out)
}
func (r *Request) Validate(in interface{}) error {
	err := validator.New().Struct(in)
	var fs []string
	if err != nil {
		switch ferr := err.(type) {
		case validator.ValidationErrors:
			for _, fieldError := range ferr {
				fs = append(fs, getTag(in, fieldError.StructField()))
			}
		default:
			fs = append(fs, err.Error())
		}
		return fmt.Errorf("missing params: %s", strings.Join(fs, ", "))
	}
	return nil
}

func getTag(in interface{}, name string) string {
	obj := reflect.TypeOf(in)
	var field reflect.StructField
	switch obj.Kind() {
	case reflect.Ptr:
		f, ok := obj.Elem().FieldByName(name)
		if ok {
			field = f
		}
	case reflect.Struct:
		f, ok := obj.FieldByName(name)
		if ok {
			field = f
		}
	default:
		return ""
	}
	if field.Tag.Get("json") != "" {
		return field.Tag.Get("json")
	} else {
		return field.Name
	}
}

func (r *Request) GetAuthorized() *Authorized {
	return r.Authorized
}
func (r *Request) SetAuthorized(a *Authorized) {
	r.Authorized = a
}

// Clone returns a deep copy of the request by using copystructure
func (r *Request) Clone() (*Request, error) {
	cpy, err := copystructure.Copy(r)
	if err != nil {
		return nil, err
	}
	return cpy.(*Request), nil
}

func (r *Request) GoString() string {
	return fmt.Sprintf("*%#v", *r)
}
func (r *Request) String() string {
	value, _ := json.Marshal(r)
	return string(value)
}

// InitializationRequest stores the parameters and context of an Initialize()
// call being made to a logical.Backend.
type InitializationRequest struct {
	Params Encodable
}

// Encodable defines types that can be encoded into a byte slice.
type Encodable interface {
	Encode() []byte
	Decode(out interface{}) error
}
