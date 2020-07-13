package wsx

import (
	"testing"
)

func TestLoggerColor(t *testing.T) {
	Mode = DEBUG
	Infof("nil point reference")

	Debugf("recv a heartbeat")

	var username = "wsx"
	Tracef(&Context{username: &username}, username, "用户报名成功")

	Fatalf("conflict router")
}
