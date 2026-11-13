package response

import (
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/xianhc/ape-volo-admin-go/internal/utils"
)

// GetVerifyErr 把校验错误整理成 ActionError。
//
// 放在 response 而不是 utils：它产出的是响应结构，
// 留在工具包里会让 utils 反向依赖传输对象。
func GetVerifyErr(err error) (actionError *ActionError) {
	mapError := make(map[string]string)
	var validationErrors validator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	if !ok {
		mapError[strconv.FormatInt(int64(utils.GenerateID()), 10)] = err.Error()
		actionError = &ActionError{
			Errors: mapError,
		}
		return actionError
	}
	for _, validationErr := range validationErrors {
		mapError[validationErr.Field()] = validationErr.Error()
	}
	actionError = &ActionError{
		Errors: mapError,
	}
	return actionError
}
