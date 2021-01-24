package logical

import (
	"encoding/json"
	"errors"
)

const AuthTokenName = "x-auth-token"
const Authorization = "authorization"

var ErrAuthMethodRequired = errors.New("Token method required")
var ErrAuthMethodNotFound = errors.New("Token method not found")
var ErrAuthorizationTokenRequired = errors.New("Token token required")
var ErrAuthorizationTokenInvalid = errors.New("Token token invalid")

//验证信息
type Authorized struct {
	ID        interface{} `json:"id" name:"账户ID"`
	Token     string      `json:"token" name:"authorization token"`
	Principal Principal   `json:"principal" name:"账户凭证(用户信息)"`
}
type Principal map[string]interface{}

func NewAuthorized(id interface{}, token string, principal Principal) Authorized {
	return Authorized{ID: id, Token: token, Principal: principal}
}

func (a Authorized) Encode() ([]byte, error) {
	return json.Marshal(a)
}

func (a Authorized) GetPrincipal() Principal {
	return a.Principal
}

func (a Authorized) SetPrincipal(in Principal) {
	a.Principal = in
}

func init() {
}
