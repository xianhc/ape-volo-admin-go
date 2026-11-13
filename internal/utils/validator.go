package utils

import "github.com/go-playground/validator/v10"

var requestValidator = validator.New()

// VerifyData 按 validate 标签校验请求字段。
//
// 校验器在进程内复用，缓存类型规则；运行期间不再修改校验规则。
func VerifyData(s any) error {
	return requestValidator.Struct(s)
}
