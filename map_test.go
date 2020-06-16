package wsx

import (
	"github.com/fwhezfwhez/cmap"
	"testing"
	"time"
)

func TestGMap(t *testing.T) {
	var m MapI
	m = NewGoMap()

	for i:=0;i<=10000;i ++ {
		go func() {
		  m.SetEx("ft", "123", 15)
		}()
		go func() {
			m.Set("ft", "123")
		}()
		go func() {
			m.Delete("ft")
		}()
	}
	time.Sleep(10* time.Second)
}

func TestCMap(t *testing.T) {
	var m MapI
	m = cmap.NewMap()

	for i:=0;i<=10000;i ++ {
		go func() {
			m.SetEx("ft", "123", 15)
		}()
		go func() {
			m.Set("ft", "123")
		}()
		go func() {
			m.Delete("ft")
		}()
	}
	time.Sleep(10* time.Second)
}