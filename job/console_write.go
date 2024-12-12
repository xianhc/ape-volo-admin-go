package job

import (
	"fmt"
	"go-apevolo/utils/ext"
)

func PrintTimeJob() {
	fmt.Println("当前时间：", ext.GetCurrentTime())
}
