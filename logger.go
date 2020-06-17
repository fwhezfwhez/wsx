package wsx

import (
	"fmt"
	"runtime"
)

func Printf(f string, v ...interface{}) {
	_, file, l, _ := runtime.Caller(1)
	if Mode == DEBUG {
		fmt.Println(fmt.Sprintf("%s:%d ", file, l) + fmt.Sprintf(f, v...))
	}
}
