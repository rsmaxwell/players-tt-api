package model

import (
	"fmt"

	"github.com/rsmaxwell/players-tt-api/internal/codeerror"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
	"github.com/rsmaxwell/players-tt-api/internal/utils"

	"golang.org/x/crypto/bcrypt"
	validator "gopkg.in/go-playground/validator.v9"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
)

// Registration type
type Registration struct {
	FirstName string `json:"firstname" validate:"required,min=3,max=20"`
	LastName  string `json:"lastname" validate:"required,min=3,max=20"`
	Knownas   string `json:"knownas" validate:"required,min=2,max=20"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"max=20"`
	Password  string `json:"password" validate:"required,min=8,max=30"`
}

var (
	functionToPerson               = debug.NewFunction(pkg, "ToPerson")
	functionNewRegistrationFromMap = debug.NewFunction(pkg, "NewRegistrationFromMap")
)

var (
	validate = validator.New()
	trans    ut.Translator
)

func init() {
	english := en.New()
	uni := ut.New(english, english)
	trans, _ = uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, trans)
}

// NewRegistration initialises a Registration object
func NewRegistration(firstname string, lastname string, knownas string, email string, phone string, password string) *Registration {
	r := new(Registration)
	r.FirstName = firstname
	r.LastName = lastname
	r.Knownas = knownas
	r.Email = email
	r.Phone = phone
	r.Password = password
	return r
}

func NewRegistrationFromMap(data *map[string]interface{}) (*Registration, error) {
	f := functionNewRegistrationFromMap
	f.DebugVerbose("")

	firstname, err := utils.GetStringFromMap("firstname", data)
	if err != nil {
		return nil, err
	}

	lastname, err := utils.GetStringFromMap("lastname", data)
	if err != nil {
		return nil, err
	}

	email, err := utils.GetStringFromMap("email", data)
	if err != nil {
		return nil, err
	}

	knownas, err := utils.GetStringFromMap("knownas", data)
	if err != nil {
		return nil, err
	}

	phone, err := utils.GetStringFromMap("phone", data)
	if err != nil {
		return nil, err
	}

	password, err := utils.GetStringFromMap("password", data)
	if err != nil {
		return nil, err
	}

	person := NewRegistration(firstname, lastname, knownas, email, phone, password)
	return person, nil
}

// ToPerson converts a Registration into a person
func (r *Registration) ToPerson() (*FullPerson, error) {
	f := functionToPerson

	err := validate.Struct(r)
	if err != nil {
		rawMessage := fmt.Sprintf("validation failed for [%s]: %s", r.Email, err.Error())
		f.DebugVerbose(rawMessage)

		errs := translateError(err, trans)
		message := errs[0].Error()
		f.DebugVerbose(message)

		return nil, codeerror.NewBadRequest(message)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.MinCost)
	if err != nil {
		message := "Could not generate password hash"
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, err
	}
	p := NewPerson(r.FirstName, r.LastName, r.Knownas, r.Email, r.Phone, hash)

	return p, nil
}

func translateError(err error, trans ut.Translator) (errs []error) {
	if err == nil {
		return nil
	}
	validatorErrs := err.(validator.ValidationErrors)
	for _, e := range validatorErrs {
		translatedErr := fmt.Errorf(e.Translate(trans))
		errs = append(errs, translatedErr)
	}
	return errs
}
