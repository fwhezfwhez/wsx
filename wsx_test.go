package wsx

import (
	"fmt"
	"testing"
)

func TestUserConn(t *testing.T) {
	srv := NewWsx("/kf")


	if e := srv.ListenAndServe(":8181"); e != nil {
		fmt.Println(e.Error())
		return
	}
}
