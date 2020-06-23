package wsx

import (
	"github.com/fwhezfwhez/cmap"
	"github.com/fwhezfwhez/errorx"
	"reflect"
)

// 用户池
type PoolV2 struct {
	pool MapI

	// beforeOnline func(c *Context)
	afterOnline func(c *Context)

	beforeOffline func(c *Context)
	afterOffline  func(c *Context)
}

// 初始化用户池
// p := wsx.NewPool(wsx.NewGoMap())
func NewPoolV2(mi MapI) *PoolV2 {
	if mi == nil {
		mi = cmap.NewMap()
	}
	return &PoolV2{
		pool: mi,
	}
}

//func (p *Pool) SetBeforeOnline(f func(c *Context)) {
//	p.beforeOnline = f
//}

// Callback of after online
func (p *PoolV2) SetAfterOnline(f func(c *Context)) {
	p.afterOnline = f
}

// Callback of before online
func (p *PoolV2) SetBeforeOffline(f func(c *Context)) {
	p.beforeOffline = f
}

// Callback of after offline
func (p *PoolV2) SetAfterOffline(f func(c *Context)) {
	p.afterOffline = f
}

// 上线
// p.Online("fengtao")
func (p *PoolV2) Online(username string, chanel string, wrapConn *WrapConn) error {
	if chanel == "" {
		chanel = "default"
	}

	var uc *UserConn

	oldUserConnI, exist := p.pool.Get(username)
	if !exist {
		uc = NewUserConn(username)
		uc.AddChanelConn(chanel, wrapConn)
		return nil
	} else {
		var cantransfer bool
		uc, cantransfer = oldUserConnI.(*UserConn)
		if !cantransfer {
			return errorx.NewFromStringf("poolV2 require value typed '*wsx.UserConn' but get '%s'", reflect.TypeOf(oldUserConnI).Name())
		}
		uc.AddChanelConn(chanel, wrapConn)
		p.pool.Set(username, uc)
		return nil
	}
	return nil
}

// p.Offline("fengtao")
func (p *PoolV2) Offline(chanel string, username string) error {
	if chanel == "" {
		chanel = "default"
	}

	ucI, exist := p.pool.Get(username)
	if !exist {
		return errorx.NewFromStringf("not found username '%s' userConn", username)
	}

	uc, cantransfer := ucI.(*UserConn)
	if !cantransfer {
		return errorx.NewFromStringf("wsx.PoolV2 requires value typed '*wsx.UserConn', but got '%s'", reflect.TypeOf(ucI).Name())
	}

	n := uc.CloseChanel(chanel)

	if n <= 0 {
		p.pool.Delete(username)
	}
	return nil
}

// p.IsOnline("fengtao")
func (p *PoolV2) IsOnline(username string) (*UserConn, bool, error) {
	ucI, exist := p.pool.Get(username)
	if !exist {
		return nil, false, nil
	}
	uc, ok := ucI.(*UserConn)
	if !ok {
		return nil, false, errorx.NewFromStringf("wsx.PoolV2 requires value typed '*wsx.UserConn', but get '%s'", reflect.TypeOf(ucI).Name())
	}
	return uc, true, nil
}

// 公用发送消息模版
func (p *PoolV2) CommonSend(username string, messageID int, header H, v interface{}, marshaller Marshaller) error {
	uc, online, e := p.IsOnline(username)
	if e != nil {
		return errorx.Wrap(e)
	}
	if !online {
		return ErrNotOnline
	}
	if uc == nil {
		return errorx.NewFromStringf("username '%s' is online but uc is nil", username)
	}

	buf, e := PackWithMarshaller(Message{
		MessageID: int32(messageID),
		Header:    header,
		Body:      v,
	}, marshaller)
	if e != nil {
		return errorx.Wrap(e)
	}

	return errorx.Wrap(uc.Write(buf))
}

// p.Send("fwhez", "/user/", wsx.H{"message": "welcome"})
func (p *PoolV2) Send(username string, urlPattern string, v interface{}) error {
	return p.CommonSend(username, 0, *HURLPattern(urlPattern), v, JSON)
}
