package wsx

import (
	"encoding/json"
	"fmt"
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
func (uc *UserConn) Write(buf []byte) error {
	uc.l.Lock()
	defer uc.l.Unlock()

	var er error
	for k, _ := range uc.conns {
		fmt.Println(errorx.NewFromStringf("detail-KEY:\n key: %s value: %p", k, uc.conns[k].conn))
		if e := uc.conns[k].Write(buf); e != nil {
			er = errorx.GroupErrors(errorx.Wrap(e))
			continue
		}
	}
	return er
}

func (uc *UserConn) WriteChanel(chanel string, buf []byte) error {
	uc.l.Lock()
	defer uc.l.Unlock()

	con, ok := uc.conns[chanel]
	if !ok {
		return nil
	}
	return con.Write(buf)
}

// Close all connections
func (uc *UserConn) Close() error {
	uc.l.Lock()
	defer uc.l.Unlock()
	var er error
	for k, _ := range uc.conns {
		if e := uc.conns[k].Close(); e != nil {
			er = errorx.GroupErrors(errorx.Wrap(e))
			continue
		}
	}
	return er
}

// 关闭某个渠道的连接
func (uc *UserConn) CloseChanel(chanel string) (int) {
	uc.l.Lock()
	defer uc.l.Unlock()
	con, ok := uc.conns[chanel]
	if !ok {
		return len(uc.conns)
	}

	con.Close()
	delete(uc.conns, chanel)
	return len(uc.conns)
}

func (uc *UserConn) CloseChanelWithSessionID(chanel string, sessionId string) (int) {
	uc.l.Lock()
	defer uc.l.Unlock()
	con, ok := uc.conns[chanel]
	if !ok {
		return len(uc.conns)
	}

	if con.SessionId == sessionId {
		con.Close()
		delete(uc.conns, chanel)
	}

	return len(uc.conns)
}

// 获取某个渠道的连接
func (uc *UserConn) GetWrapConn(chanel string) (*WrapConn, bool) {
	uc.l.RLock()
	defer uc.l.RLock()

	wc, ok := uc.conns[chanel]

	if !ok {
		return nil, false
	}
	return wc, true
}

func (uc *UserConn) AddChanelConn(chanel string, wrapConn *WrapConn) {
	uc.l.RLock()
	old, exist := uc.conns[chanel]
	uc.l.RUnlock()

	if exist {
		Infof("username %s has more than 1 connection on channel %s, and has close old conn", uc.Username, chanel)
		old.Close()
	}
	uc.l.Lock()
	uc.conns[chanel] = wrapConn
	uc.l.Unlock()
}

func (uc *UserConn) JSONUrlPattern(urlPattern string, v interface{}) error {
	buf, e := json.Marshal(v)
	if e != nil {
		return errorx.Wrap(e)
	}
	res, e := PackWithMarshallerAndBody(Message{
		MessageID: int32(0),
		Header: map[string]interface{}{
			HEADER_ROUTER_KEY:            HEADER_ROUTER_TYPE_URL_PATTERN,
			HEADER_URL_PATTERN_VALUE_KEY: urlPattern,
		},
	}, buf)
	if e != nil {
		return errorx.Wrap(e)
	}

	uc.Write(res)
	return nil
}

func (uc *UserConn) JSONUrlPatternWithChanel(chanel string, urlPattern string, v interface{}) error {
	buf, e := json.Marshal(v)
	if e != nil {
		return errorx.Wrap(e)
	}
	res, e := PackWithMarshallerAndBody(Message{
		MessageID: int32(0),
		Header: map[string]interface{}{
			HEADER_ROUTER_KEY:            HEADER_ROUTER_TYPE_URL_PATTERN,
			HEADER_URL_PATTERN_VALUE_KEY: urlPattern,
		},
	}, buf)
	if e != nil {
		return errorx.Wrap(e)
	}

	uc.WriteChanel(chanel, res)
	return nil
}
