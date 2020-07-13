package wsx

import (
	"fmt"
	"runtime"
)

var (
	greenBg      = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	whiteBg      = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellowBg     = string([]byte{27, 91, 57, 48, 59, 52, 51, 109})
	redBg        = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blueBg       = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magentaBg    = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyanBg       = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	green        = string([]byte{27, 91, 51, 50, 109})
	white        = string([]byte{27, 91, 51, 55, 109})
	yellow       = string([]byte{27, 91, 51, 51, 109})
	red          = string([]byte{27, 91, 51, 49, 109})
	blue         = string([]byte{27, 91, 51, 52, 109})
	magenta      = string([]byte{27, 91, 51, 53, 109})
	cyan         = string([]byte{27, 91, 51, 54, 109})
		reset        = string([]byte{27, 91, 48, 109})
	disableColor = false
)

// Debugf will print content where called when MODE="debug"
func Debugf(f string, v ...interface{}) {
	if Mode == DEBUG {
		_, file, l, _ := runtime.Caller(1)
		fmt.Println(fmt.Sprintf("|%s|%s| %s:%d ",
			fmt.Sprintf("%s%s%s", magenta, "wsx", reset),
			fmt.Sprintf("%s%s%s", magenta, "debug", reset), file, l) + fmt.Sprintf(f, v...))
	}
}

// Printf will print content where called
func Infof(f string, v ...interface{}) {

	_, file, l, _ := runtime.Caller(1)
	fmt.Println(fmt.Sprintf("|%s|%s| %s:%d ",
		fmt.Sprintf("%s%s%s", cyan, "wsx", reset),
		fmt.Sprintf("%s%s%s", cyan, "info", reset), file, l) + fmt.Sprintf(f, v...))

}

// Printf will print content where called
func Fatalf(f string, v ...interface{}) {
	_, file, l, _ := runtime.Caller(1)
	fmt.Println(fmt.Sprintf("|%s|%s| %s:%d ",
		fmt.Sprintf("%s%s%s", redBg, "wsx", reset),
		fmt.Sprintf("%s%s%s", redBg, "fatal", reset), file, l) + fmt.Sprintf(f, v...))
}

// Tracef will print content where called when context.username == username
func Tracef(c *Context, username string, f string, v ... interface{}) {
	if c.GetUsername() == username {
		_, file, l, _ := runtime.Caller(1)
		fmt.Println(fmt.Sprintf("|%s|%s| %s:%d ",
			fmt.Sprintf("%s%s%s", yellowBg, "wsx", reset),
			fmt.Sprintf("%s%s%s", yellowBg, "trace", reset), file, l) + fmt.Sprintf(f, v...))
	}
}
