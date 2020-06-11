package wsx

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"runtime"
)

// Mux在websocket监听开始以前，都可以预设路由，中间件和一些配置，是可写的。
// 而一旦websocket 发起了监听，mux是只读的
type Mux struct {
	globalMiddlewares []func(c *Context)

	messageIDMux  map[int][]func(c *Context)
	urlPatternMux map[string][]func(c *Context)
	readOnly      bool

	messageIDRouteInfo map[int]Route
	urlRouteInfo       map[string]Route

	panicOnExistRouting bool
}

func NewMux() *Mux {
	return &Mux{
		messageIDMux:      make(map[int][]func(c *Context)),
		urlPatternMux:     make(map[string][]func(c *Context)),
		globalMiddlewares: make([]func(c *Context), 0, 10),

		messageIDRouteInfo: make(map[int]Route),
		urlRouteInfo:       make(map[string]Route),
	}
}

// 基于messageID添加路由
func (m *Mux) AddMessageIDHandler(messageID int, handlers ... func(c *Context)) error {
	if m.readOnly == false {
		_, file, line, _ := runtime.Caller(1)
		routeInfo := Route{
			Whereis:   []string{fmt.Sprintf("%s:%d", file, line)},
			MessageId: messageID,
		}

		h, ok := m.messageIDMux[messageID]
		if ok {
			if m.panicOnExistRouting {
				panic(fmt.Errorf("handler conflicts on the same messageID: \n%s\nThe existed route-info is at:\n%s", routeInfo.Whereis[0], m.messageIDRouteInfo[messageID].Location()))
			}
			m.messageIDMux[messageID] = append(h, handlers...)
		} else {
			m.messageIDMux[messageID] = handlers
		}

		r, exist := m.messageIDRouteInfo[messageID]
		if !exist {
			m.messageIDRouteInfo[messageID] = routeInfo
		} else {
			m.messageIDRouteInfo[messageID] = r.Merge(routeInfo)
		}

		return nil
	} else {
		return errorx.NewFromString("mux is only writable before mux.LockWrite()")
	}
}

// 基于url-pattern添加路由
func (m *Mux) AddURLPatternHandler(urlPattern string, handlers ... func(c *Context)) error {
	if m.readOnly == false {
		_, file, line, _ := runtime.Caller(1)
		routeInfo := Route{
			Whereis:   []string{fmt.Sprintf("%s:%d", file, line)},
			URLPattern: urlPattern,
		}


		h, ok := m.urlPatternMux[urlPattern]
		if ok {
			if m.panicOnExistRouting {
				panic(fmt.Errorf("handler conflicts on the same url-pattern: \n%s\nThe existed route-info is at:\n%s", routeInfo.Whereis[0], m.urlRouteInfo[urlPattern].Location()))
			}

			m.urlPatternMux[urlPattern] = append(h, handlers...)
		} else {
			m.urlPatternMux[urlPattern] = handlers
		}

		r, exist := m.urlRouteInfo[urlPattern]
		if !exist {
			m.urlRouteInfo[urlPattern] = routeInfo
		} else {
			m.urlRouteInfo[urlPattern] = r.Merge(routeInfo)
		}
		return nil
	} else {
		return errorx.NewFromString("mux is only writable before mux.LockWrite()")
	}
}

// 全局中间件
func (m *Mux) UseGlobal(handlers ... func(c *Context)) error {
	if m.readOnly == false {
		m.globalMiddlewares = append(m.globalMiddlewares, handlers...)
		return nil
	} else {
		return errorx.NewFromString("mux is only writable before mux.LockWrite()")
	}
}

// MessageID和URL路由在添加时，如果已存在，则会panic。
func (m *Mux) PanicOnExistRouter() error {
	if m.readOnly == false {
		m.panicOnExistRouting = true
		return nil
	} else {
		return errorx.NewFromString("mux is only writable before mux.LockWrite()")
	}
}

// 锁定后，无法再添加路由
func (m *Mux) LockWrite() {
	m.readOnly = true
}
