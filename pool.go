package wsx

import (
	"fmt"
	"github.com/fwhezfwhez/cmap"
	"github.com/fwhezfwhez/errorx"
)

// 用户池
type Pool struct {
	pool MapI

	// beforeOnline func(c *Context)
	afterOnline func(c *Context)

	beforeOffline func(c *Context)
	afterOffline  func(c *Context)
}

// 初始化用户池
// p := wsx.NewPool(wsx.NewGoMap())
func NewPool(mi MapI) *Pool {
	if mi == nil {
		mi = cmap.NewMap()
	}
	return &Pool{
		pool: mi,
	}
}

//func (p *Pool) SetBeforeOnline(f func(c *Context)) {
//	p.beforeOnline = f
//}

// Callback of after online
func (p *Pool) SetAfterOnline(f func(c *Context)) {
	p.afterOnline = f
}

// Callback of before online
func (p *Pool) SetBeforeOffline(f func(c *Context)) {
	p.beforeOffline = f
}

// Callback of after offline
func (p *Pool) SetAfterOffline(f func(c *Context)) {
	p.afterOffline = f
}

// 上线
// p.Online("fengtao")
func (p *Pool) Online(username string, c *Context) {
	//if p.beforeOnline != nil {
	//	p.beforeOnline(c)
	//}
	cCopy := c.Clone()
	p.pool.Set(username, &cCopy)
	// *(c.username) = username

	if p.afterOnline != nil {
		p.afterOnline(c)
	}
}

// p.Offline("fengtao")
func (p *Pool) Offline(username string) {
	c, isOnline := p.IsOnline(username)
	if !isOnline || c == nil {
		return
	}

	if p.beforeOffline != nil {
		p.beforeOffline(c)
	}
	p.pool.Delete(username)
	c.l.Lock()
	c.Conn.Close()
	c.l.Unlock()

	if p.afterOffline != nil {
		p.afterOffline(c)
	}
}

// p.IsOnline("fengtao")
func (p *Pool) IsOnline(username string) (*Context, bool) {
	ctx, exist := p.pool.Get(username)
	if !exist {
		return nil, false
	}
	return ctx.(*Context), true
}

var (
	ErrNotOnline = fmt.Errorf("user not online yet")
)

// 公用发送消息模版
func (p *Pool) CommonSend(username string, messageID int, header H, v interface{}, marshaller Marshaller) error {
	ctx, online := p.IsOnline(username)
	if !online {
		return ErrNotOnline
	}
	if ctx == nil {
		return errorx.NewFromStringf("username '%s' is online but ctx is nil", username)
	}

	buf, e := PackWithMarshaller(Message{
		MessageID: int32(messageID),
		Header:    header,
		Body:      v,
	}, marshaller)
	if e != nil {
		return errorx.Wrap(e)
	}

	return errorx.Wrap(ctx.WriteMessage(buf))
}

// p.Send("fwhez", "/user/", wsx.H{"message": "welcome"})
func (p *Pool) Send(username string, urlPattern string, v interface{}) error {
	return p.CommonSend(username, 0, *HURLPattern(urlPattern), v, JSON)
}
