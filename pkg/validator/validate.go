package validation

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
)

var ErrInvalidAttribute = errors.New("Invalid Attribute")

type Valider struct {
	checker *validator.Validate
	trans   ut.Translator
}

func NewValider() *Valider {
	return &Valider{
		checker: validator.New(),
	}
}

func (v *Valider) RegisterValidator() error {
	if err := v.registerTranslation(); err != nil {
		return err
	}
	if err := v.registerValidations(); err != nil {
		return err
	}

	return nil
}

func (v Valider) ValidateStruct(s interface{}) ([]string, error) {
	var strErrs []string
	err := v.checker.Struct(s)
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return strErrs, err
	}

	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			strErrs = append(strErrs, e.Translate(v.trans))
		}
	}
	return strErrs, nil
}

func (v Valider) registerValidations() error {
	if err := v.checker.RegisterValidation("pwd", pwd); err != nil {
		return err
	}

	if err := v.checker.RegisterValidation("phone", phone); err != nil {
		return err
	}

	if err := v.checker.RegisterValidation("ref", ref); err != nil {
		return err
	}

	if err := v.checker.RegisterValidation("rfe", rfe); err != nil {
		return err
	}

	v.checker.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return nil
}

func (v *Valider) registerTranslation() error {
	translator := en.New()
	uni := ut.New(translator, translator)

	trans, found := uni.GetTranslator("en")
	if !found {
		return errors.New("translator not found")
	}

	v.trans = trans

	if err := en_translations.RegisterDefaultTranslations(v.checker, trans); err != nil {
		return err
	}

	if err := v.checker.RegisterTranslation("pwd", trans, func(ut ut.Translator) error {
		return ut.Add("pwd", "{0} is not strong enough", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("pwd", fe.Field())
		return t
	}); err != nil {
		return err
	}

	if err := v.checker.RegisterTranslation("rfe", trans, func(ut ut.Translator) error {
		return ut.Add("rfe", "{0} is needed with this fields", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("rfe", fe.Field())
		return t
	}); err != nil {
		return err
	}

	if err := v.checker.RegisterTranslation("ref", trans, func(ut ut.Translator) error {
		return ut.Add("ref", "{0} do not match with ref pattern [a-z | A-Z | 0-9] only", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("ref", fe.Field())
		return t
	}); err != nil {
		return err
	}

	if err := v.checker.RegisterTranslation("phone", trans, func(ut ut.Translator) error {
		return ut.Add("phone", "{0} do not match with phone pattern", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("phone", fe.Field())
		return t
	}); err != nil {
		return err
	}

	return nil
}

// TODO
func pwd(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) > 6
}

// TODO
func ref(fl validator.FieldLevel) bool {
	for _, l := range fl.Field().String() {
		if !unicode.IsLetter(l) && !unicode.IsNumber(l) {
			return false
		}
	}
	return true
}

// TODO
func phone(fl validator.FieldLevel) bool {
	return true
}

func rfe(fl validator.FieldLevel) bool {
	param := strings.Split(fl.Param(), `:`)
	paramField := param[0]
	paramValue := param[1]

	if paramField == `` {
		return true
	}

	// param field reflect.Value.
	var paramFieldValue reflect.Value

	if fl.Parent().Kind() == reflect.Ptr {
		paramFieldValue = fl.Parent().Elem().FieldByName(paramField)
	} else {
		paramFieldValue = fl.Parent().FieldByName(paramField)
	}

	if isEq(paramFieldValue, paramValue) == false {
		return true
	}

	return hasValue(fl)
}

func hasValue(fl validator.FieldLevel) bool {
	return requireCheckFieldKind(fl, "")
}

func requireCheckFieldKind(fl validator.FieldLevel, param string) bool {
	field := fl.Field()
	if len(param) > 0 {
		if fl.Parent().Kind() == reflect.Ptr {
			field = fl.Parent().Elem().FieldByName(param)
		} else {
			field = fl.Parent().FieldByName(param)
		}
	}
	switch field.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !field.IsNil()
	default:
		_, _, nullable := fl.ExtractType(field)
		if nullable && field.Interface() != nil {
			return true
		}
		return field.IsValid() && field.Interface() != reflect.Zero(field.Type()).Interface()
	}
}

func isEq(field reflect.Value, value string) bool {
	switch field.Kind() {

	case reflect.String:
		return field.String() == value

	case reflect.Slice, reflect.Map, reflect.Array:
		p := asInt(value)

		return int64(field.Len()) == p

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p := asInt(value)

		return field.Int() == p

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p := asUint(value)

		return field.Uint() == p

	case reflect.Float32, reflect.Float64:
		p := asFloat(value)

		return field.Float() == p
	}

	// panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	return false
}

func asInt(param string) int64 {

	i, err := strconv.ParseInt(param, 0, 64)
	panicIf(err)

	return i
}

func asUint(param string) uint64 {

	i, err := strconv.ParseUint(param, 0, 64)
	panicIf(err)

	return i
}

func asFloat(param string) float64 {

	i, err := strconv.ParseFloat(param, 64)
	panicIf(err)

	return i
}

func panicIf(err error) {
	if err != nil {
		panic(err.Error())
	}
}
