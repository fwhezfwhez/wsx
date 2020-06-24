package wsx

import (
	"testing"
	"time"
)

func TestUserConn(t *testing.T) {
	pool := NewPoolV2(nil)
	go func() {
		listenAndServe(":8111", pool)
	}()

	time.Sleep(2 * time.Second)



}
