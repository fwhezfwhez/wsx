package wsx

import (
	"github.com/gorilla/websocket"
	"strconv"
	"sync"
	"time"
)

// 官方的ws连接不支持并发读写，所以需要wrap一层锁
type WrapConn struct {
	SessionId string
	conn      *websocket.Conn
	l         *sync.RWMutex
}

func NewWrapConn(con *websocket.Conn) (string, *WrapConn) {
	sessionID := MD5(strconv.FormatInt(time.Now().UnixNano(), 10))
	return sessionID, &WrapConn{
		SessionId: sessionID,
		conn:      con,
		l:         &sync.RWMutex{},
	}
}

func (wc *WrapConn) Write(buf []byte) error {
	wc.l.Lock()
	defer wc.l.Unlock()
	return wc.conn.WriteMessage(websocket.BinaryMessage, buf)
}

func (wc *WrapConn) Close() error {
	wc.l.Lock()
	defer wc.l.Unlock()
	return wc.conn.Close()
}
