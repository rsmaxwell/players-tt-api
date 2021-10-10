package model

import (
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/internal/utils"
)

// Registration type
type Signin struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=30"`
}

var (
	functionNewSigninFromMap = debug.NewFunction(pkg, "NewSigninFromMap")
)

// NewRegistration initialises a Signin object
func NewSignin(username string, password string) *Signin {
	r := new(Signin)
	r.Username = username
	r.Password = password
	return r
}

func NewSigninFromMap(data *map[string]interface{}) (*Signin, error) {
	f := functionNewSigninFromMap
	f.DebugVerbose("")

	username, err := utils.GetStringFromMap("username", data)
	if err != nil {
		return nil, err
	}

	password, err := utils.GetStringFromMap("password", data)
	if err != nil {
		return nil, err
	}

	signin := NewSignin(username, password)
	return signin, nil
}
