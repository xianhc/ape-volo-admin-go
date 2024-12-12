package job

import (
	"reflect"
	"runtime"
)

// FuncType 定义函数类型
type FuncType func()

// GlobalFuncs 全局函数映射
var GlobalFuncs = map[string]FuncType{}

func RegisterTaskFuncs() {
	// 使用反射遍历当前包中所有的函数
	funcs := []FuncType{
		PrintTimeJob,
		SendEmailJob,
	}

	// 将函数名及对应的函数添加到 GlobalFuncs 中
	for _, fn := range funcs {
		name := getFunctionName(fn)
		GlobalFuncs[name] = fn
	}
}

// 获取函数名
func getFunctionName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}
