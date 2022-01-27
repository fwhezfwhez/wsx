package wsx

import (
	"fmt"
	"github.com/fwhezfwhez/cmap"
	"github.com/fwhezfwhez/errorx"
	"reflect"
	"time"
)

// 用户池
// 用户池v1只存放*wsx.Context
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
func (p *Pool) Online(username string, c *Context) error {
	//if p.beforeOnline != nil {
	//	p.beforeOnline(c)
	//}
	cCopy := c.Clone()

	oldContextI, ok := p.pool.Get(username)
	if ok {
		oldContext, canTransfer := oldContextI.(*Context)
		if !canTransfer {
			return errorx.NewFromStringf("wrong type assertion of wsx.Pool, requires *wsx.Context but got %s", reflect.TypeOf(oldContextI).Name())
		}

		// 如果旧连接和新连接，是同一条，则返回。不是同一条，则将旧的那条关闭。
		if oldContext.GetSessionID() != c.GetSessionID() {

			Debuglnf("recv_event online_duplicated_sessionid old_session_id %s new_session_id %s username %s", oldContext.GetSessionID(), c.GetSessionID(),c.GetUsername())


			fmt.Printf("%s triger context conflict, new %s old %s, old has been closed\n", time.Now().Format("2006-01-02 15:04:05"), c.GetSessionID(), oldContext.GetSessionID())
			oldContext.Close()
		} else {
			return nil
		}
	}

	p.pool.Set(username, &cCopy)
	// *(c.username) = username

	if p.afterOnline != nil {
		p.afterOnline(c)
	}
	return nil
}

// 上线指定渠道
func (p *Pool) OnlineWithChanel(chanel string, username string, c *Context) error {
	var key = GetChanelUsername(chanel, username)

	cCopy := c.Clone()

	oldContextI, ok := p.pool.Get(key)
	if ok {
		oldContext, canTransfer := oldContextI.(*Context)
		if !canTransfer {
			return errorx.NewFromStringf("wrong type assertion of wsx.Pool, requires *wsx.Context but got %s", reflect.TypeOf(oldContextI).Name())
		}

		oldContext.Close()
	}

	p.pool.Set(key, &cCopy)
	// *(c.username) = username

	if p.afterOnline != nil {
		p.afterOnline(c)
	}
	return nil
}

// 离线指定渠道
func (p *Pool) OfflineWithChanel(chanel string, username string, c *Context) {
	c, isOnline := p.IsOnlineWithChanel(chanel, username)
	if !isOnline || c == nil {
		return
	}

	if p.beforeOffline != nil {
		p.beforeOffline(c)
	}
	p.pool.Delete(GetChanelUsername(chanel, username))
	c.l.Lock()
	c.Conn.Close()
	c.l.Unlock()

	if p.afterOffline != nil {
		p.afterOffline(c)
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

func (p *Pool) OfflineCtx(ctx *Context) {

	username := ctx.GetUsername()
	if username == "" {
		return
	}

	c, isOnline := p.IsOnline(username)
	if !isOnline || c == nil {
		return
	}

	// 不是同一个连接，则不处理。
	// 场景: 同一个用户连上两个连接，第二个连接会杀死第一个连接。但是在第一个连接的心跳超时位置，会按照ctx来离线，如果不作sessionid幂等，则会杀死新连接
	if ctx.GetSessionID() != c.GetSessionID() {
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

// p.IsOnline("fengtao")
func (p *Pool) IsOnlineWithChanel(chanel string, username string) (*Context, bool) {
	key := GetChanelUsername(chanel, username)
	ctx, exist := p.pool.Get(key)
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
	fmt.Printf("%s send buf to username %s buf %v \n", time.Now().Format("2006-01-02 15:04:05"), username, buf)

	return errorx.Wrap(ctx.WriteMessage(buf))
}

// p.Send("fwhez", "/user/", wsx.H{"message": "welcome"})
func (p *Pool) Send(username string, urlPattern string, v interface{}) error {
	return p.CommonSend(username, 0, *HURLPattern(urlPattern), v, JSON)
}
