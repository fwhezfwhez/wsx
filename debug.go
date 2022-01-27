package wsx

import (
	"encoding/json"
	"fmt"
	"time"
)

var debugging bool

// set wsx to debug mod
func SetDebug() {
	debugging = true
}

func Debuglnf(f string, args ... interface{}) {
	if debugging == true {
		fmt.Printf(fmt.Sprintf("%s %s\n", time.Now().Format("2006-01-02 15:04:05"), f), args ...)
	}
}

func jsonline(i interface{}) string {
	bf, _ := json.Marshal(i)
	return string(bf)
}
