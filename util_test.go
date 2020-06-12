package wsx

import (
	"fmt"
	"testing"
)

func TestPack(t *testing.T) {
	buf, e:= Pack(3,H{}, H{
		"name": "BMW",
		"value": 1,
	})
	if e!=nil {
		panic(e)
	}
	fmt.Println(buf)
}