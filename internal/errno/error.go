package errno

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func IsNotEffect(err error) bool {
	return err.Error() == "no rows affected"
}

func Is(err error, target ErrorPair) bool {
	if err == nil {
		return false
	}

	if pair, ok := err.(*ParamsError); ok {
		return pair.code == target.Code
	}

	return false
	
}

type ParamsError struct {
	code   uint32
	msg    string
	params []interface{}
}

func NewParamsError(errorPair ErrorPair, params ...interface{}) *ParamsError {
	return &ParamsError{
		code:   errorPair.Code,
		msg:    errorPair.Msg,
		params: params,
	}
}

func DefaultParamsError(errString string, params ...interface{}) *ParamsError {
	return &ParamsError{
		code:   0,
		msg:    errString,
		params: params,
	}
}
func DefaultParamsErrorWithCode(errString string, errCode uint32, params ...interface{}) *ParamsError {
	return &ParamsError{
		code:   errCode,
		msg:    errString,
		params: params,
	}
}

func (pe *ParamsError) Error() string {
	return formatString(pe.msg, pe.params)
}

func (pe *ParamsError) Code() uint32 {
	return pe.code
}

func formatString(format string, params []interface{}) string {
	args, i := make([]string, len(params)*2), 0
	for k, v := range params {
		args[i] = fmt.Sprintf("{%d}", k)
		args[i+1] = fmt.Sprint(v)
		i += 2
	}
	return strings.NewReplacer(args...).Replace(format)
}
