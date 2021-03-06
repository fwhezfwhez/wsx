package wsx

import (
	"encoding/json"
	"github.com/fwhezfwhez/errorx"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

const ABORT = 3000

type Context struct {
	// 全局值: 一个连接一旦建立，那么所有的基于该连接的请求，都会share以下字段
	Conn *websocket.Conn
	l    *sync.RWMutex

	PerConnectionContext *sync.Map     // 连接域内的map
	username             *string       // 连接所属的username
	heartbeatChan        chan struct{} // 连接所属的心跳chan
	onClose              func() error  // close时的回调
	chanel               *string       // 连接标记的渠道，用于多端登录
	sessionID            *string       // 该连接的sessionID,在wrapConn时生成
	state                *int          // 登录状态, 1在线，2离线，3忙碌

	// 临时值: 每次到达一个请求，都会Clone一个Context，复用了它的全局值，以下值都会重置
	handlers          []func(c *Context)
	Stream            []byte
	PerRequestContext *sync.Map
	offset            int
	contentType       string
	urlPattern        string
}

func NewContext(conn *websocket.Conn) *Context {
	var u string
	var chanel string
	var sessionID string
	var state int
	return &Context{
		Conn:                 conn,
		l:                    &sync.RWMutex{},
		PerConnectionContext: &sync.Map{},
		username:             &u,
		chanel:               &chanel,
		sessionID:            &sessionID,
		state:                &state,

		heartbeatChan:     make(chan struct{}, 5),
		PerRequestContext: &sync.Map{},
		handlers:          make([]func(c *Context), 0, 10),
		offset:            -1,
	}
}

// Bind reads from stream, not from reader. It means this api can call anywhere without times limit.
// This is the promotion comparing to gin.Context.Bind()
func (c *Context) Bind(dest interface{}) error {
	body, e := BodyBytesOf(c.Stream)
	if e != nil {
		return errorx.Wrap(e)
	}

	// 客户端如果没指定content-type会默认json
	if c.contentType == "" {
		return errorx.NewFromString("content-type not found")
	}
	switch c.contentType {
	case CONTENT_TYPE_JSON:
		return json.Unmarshal(body, dest)
	default:
		return errorx.NewFromStringf("content-type %s not found", c.contentType)
	}
	return nil
}

func (c *Context) JSON(messageID int, v interface{}) error {
	buf, e := json.Marshal(v)
	if e != nil {
		return errorx.Wrap(e)
	}
	res, e := PackWithMarshallerAndBody(Message{
		MessageID: int32(messageID),
		Header:    nil,
	}, buf)
	if e != nil {
		return errorx.Wrap(e)
	}
	func() {
		c.l.Lock()
		defer c.l.Unlock()
		c.Conn.WriteMessage(websocket.BinaryMessage, res)
	}()
	return nil
}

func (c *Context) JSONUrlPattern(v interface{}) error {
	return c.jsonUrlPattern(0, c.urlPattern, v)
}

func (c *Context) JSONSetUrlPattern(urlPattern string, v interface{}) error {
	return c.jsonUrlPattern(0, urlPattern, v)
}

func (c *Context) jsonUrlPattern(messageID int, urlPattern string, v interface{}) error {
	var vh H
	switch v.(type) {
	case H:
		vh = v.(H)
		if urlPattern != "" {
			vh["url_pattern"] = urlPattern
			v = vh
		}

	}

	buf, e := json.Marshal(v)
	if e != nil {
		return errorx.Wrap(e)
	}
	res, e := PackWithMarshallerAndBody(Message{
		MessageID: int32(messageID),
		Header: map[string]interface{}{
			HEADER_ROUTER_KEY:            HEADER_ROUTER_TYPE_URL_PATTERN,
			HEADER_URL_PATTERN_VALUE_KEY: urlPattern,
		},
	}, buf)
	if e != nil {
		return errorx.Wrap(e)
	}
	func() {
		c.l.Lock()
		defer c.l.Unlock()
		c.Conn.WriteMessage(websocket.BinaryMessage, res)
	}()
	return nil
}

func (c *Context) WriteMessage(buf []byte) error {
	c.l.Lock()
	defer c.l.Unlock()
	return c.Conn.WriteMessage(websocket.BinaryMessage, buf)
}

func (c *Context) Clone() Context {
	/*

	// 全局值: 一个连接一旦建立，那么所有的基于该连接的请求，都会share以下字段
	Conn   *websocket.Conn
	l *sync.RWMutex

	PerConnectionContext *sync.Map
	username *string

	// 临时值: 每次到达一个请求，都会Clone一个Context，复用了它的全局值，以下值都会重置
	handlers []func(c *Context)
	Stream []byte
	PerRequestContext    *sync.Map
	offset int
	contentType string
	urlPattern string
	*/
	return Context{
		// 复用全局值
		Conn:                 c.Conn,
		l:                    c.l,
		PerConnectionContext: c.PerConnectionContext,
		username:             c.username,
		heartbeatChan:        c.heartbeatChan,
		onClose:              c.onClose,
		chanel:               c.chanel,
		sessionID:            c.sessionID,
		state:                c.state,

		// 重置临时值
		PerRequestContext: &sync.Map{},
		offset:            -1,
		contentType:       "",
		urlPattern:        "",
		Stream:            nil,
		handlers:          nil,
	}
}

func (c *Context) Next() {
	c.offset ++
	s := len(c.handlers)
	for ; c.offset < s; c.offset++ {
		if !c.isAbort() {
			c.handlers[c.offset](c)
		} else {
			return
		}
	}
}

func (c *Context) isAbort() bool {
	if c.offset >= ABORT {
		return true
	}
	return false
}
func (c *Context) Abort() {
	c.offset = ABORT
}

func (c *Context) Reset() {
	c.PerRequestContext = &sync.Map{}
	c.offset = -1
	if c.handlers == nil {
		c.handlers = make([]func(*Context), 0, 10)
		return
	}
	c.handlers = c.handlers[:0]
	c.contentType = ""
	c.urlPattern = ""
}

func (c *Context) GetUrlPattern() string {
	return c.urlPattern
}

func (c *Context) SetUsername(username string) {
	*(c.username) = username
}
func (c *Context) Username() string {

	if c.username == nil {
		return ""
	}
	return *c.username
}

func (c *Context) GetUsername() string {
	return c.Username()
}

func (c *Context) SetChanel(chanel string) {
	*(c.chanel) = chanel
}

func (c *Context) GetChanel() string {
	if c.chanel == nil {
		return ""
	}
	return *c.chanel
}

func (c *Context) SetSessionID(sessionID string) {
	*(c.sessionID) = sessionID
}

func (c *Context) GetSessionID() string {
	if c.sessionID == nil {
		return ""
	}
	return *c.sessionID
}

func (c *Context) RecvHeartbeat() {
	select {
	case <-time.After(15 * time.Second):
		// Printf("heartbeat chan is locked: \n %s", debug.Stack())
	case c.heartbeatChan <- struct{}{}:
		// Printf("%s收到心跳", c.Username())
	}
}

func (c *Context) SpyingOnHeartbeat() {
	go func() {
	L:
		for {
			select {
			case <-time.After(10 * time.Second):
				c.Close()
				Debugf("%s未收到心跳，自动关闭", c.Username())
				break L
			case <-c.heartbeatChan:
				// do nothing
				Debugf("%s收到心跳，自动续约", c.Username())
			}
		}
	}()
}

func (c *Context) Close() {

	func() {
		c.l.Lock()
		defer c.l.Unlock()

		c.Conn.Close()
	}()

	if c.onClose != nil {
		c.onClose()
	}

}

func (c *Context) SetOnClose(f func() error) {
	c.onClose = f
}

func (c *Context) SetOnlineState() {
	*c.state = 1
}
func (c *Context) SetOfflineState() {
	*c.state = 2
}
func (c *Context) SetBusyState() {
	*c.state = 3
}

func (c *Context) GetState() int {
	if c.state == nil {
		return 0
	}
	return *c.state
}
