package utils

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"go-apevolo/payloads/response"
	"strconv"
)

func GetVerifyErr(err error) (actionError *response.ActionError) {
	mapError := make(map[string]string)
	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	if !ok {
		mapError[strconv.FormatInt(int64(GenerateID()), 10)] = err.Error()
		actionError = &response.ActionError{
			Errors: mapError,
		}
		return actionError
	}
	for _, validationErr := range validationErrors {
		mapError[validationErr.Field()] = validationErr.Error()
	}
	actionError = &response.ActionError{
		Errors: mapError,
	}
	return actionError
}

// VerifyData
// @description: 验证数据模型
// @param: s
// @return: error
func VerifyData(s interface{}) error {
	validate := validator.New()
	err := validate.Struct(s)
	return err
}
