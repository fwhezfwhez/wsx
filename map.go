package wsx

import (
	"sync"
	"time"
)

type MapI interface {
	Set(key string, value interface{})
	SetEx(key string, value interface{}, expireSeconds int)
	Get(key string) (interface{}, bool)
	Delete(key string)
}

// 使用cmap.Map实现的Map
// github.com/fwhezfwhez/cmap.Map

// 使用sync.Map实现的Map
type GoMap struct {
	// 存放数据
	m sync.Map
	// 存放数据对应的时效
	tm sync.Map
}

func NewGoMap() *GoMap{
	return &GoMap{}
}

func (gm *GoMap) Set(key string, value interface{}) {
	gm.SetEx(key, value, -1)
}

func (gm *GoMap) SetEx(key string, value interface{}, expireSeconds int) {
	gm.m.Store(key, value)
	if expireSeconds > 0 {
		gm.tm.Store(key, time.Now().Add(time.Duration(expireSeconds) * time.Second).UnixNano())
	}
}
func (gm *GoMap) Get(key string) (interface{}, bool) {
	v, exist := gm.m.Load(key)
	if !exist {
		return nil, false
	}
	TimeLimit, ok := gm.tm.Load(key)

	if !ok {
		return v, true
	}
	if TimeLimit.(int64) < time.Now().UnixNano() {
		gm.Delete(key)
		return nil, false
	}

	return v, true

}
func (gm *GoMap) Delete(key string) {
	gm.m.Delete(key)
	gm.tm.Delete(key)
}
