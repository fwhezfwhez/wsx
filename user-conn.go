package wsx

import (
	"github.com/fwhezfwhez/errorx"
	"sync"
)

// 用户连接
type UserConn struct {
	conns    map[string]*WrapConn
	l        *sync.RWMutex
	Username string
}

func NewUserConn(username string) *UserConn {
	return &UserConn{
		conns:    make(map[string]*WrapConn, 10),
		l:        &sync.RWMutex{},
		Username: username,
	}
}

// Write to all connections
func (c *UserConn) Write(buf []byte) error {
	c.l.Lock()
	defer c.l.Unlock()
	var er error
	for k, _ := range c.conns {
		if e := c.conns[k].Write(buf); e != nil {
			er = errorx.GroupErrors(errorx.Wrap(e))
			continue
		}
	}
	return er
}

// Close all connections
func (c *UserConn) Close() error {
	c.l.Lock()
	defer c.l.Unlock()
	var er error
	for k, _ := range c.conns {
		if e := c.conns[k].Close(); e != nil {
			er = errorx.GroupErrors(errorx.Wrap(e))
			continue
		}
	}
	return er
}

// 关闭某个渠道的连接
func (c *UserConn) CloseChanel(chanel string) (int) {
	c.l.Lock()
	defer c.l.Unlock()
	con, ok := c.conns[chanel]
	if !ok {
		return len(c.conns)
	}
	delete(c.conns, chanel)
	con.Close()
	return len(c.conns)
}

// 获取某个渠道的连接
func (c *UserConn) GetWrapConn(chanel string) (*WrapConn, bool) {
	c.l.RLock()
	defer c.l.RLock()

	wc, ok := c.conns[chanel]

	if !ok {
		return nil, false
	}
	return wc, true
}

func (c *UserConn) AddChanelConn(chanel string, wrapConn *WrapConn) {
	c.l.RLock()
	old, exist := c.conns[chanel]
	c.l.RUnlock()

	if exist {
		old.Close()
	}
	c.l.Lock()
	c.conns[chanel] = wrapConn
	c.l.Unlock()
}
