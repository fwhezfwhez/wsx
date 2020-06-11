package wsx

import (
	"encoding/json"
	"github.com/fwhezfwhez/errorx"
	"github.com/gorilla/websocket"
	"sync"
)

const ABORT = 3000

type H map[string]interface{}
type Context struct {
	// 全局值: 一个连接一旦建立，那么所有的基于该连接的请求，都会share以下字段
	Conn *websocket.Conn
	l    *sync.RWMutex

	PerConnectionContext *sync.Map
	username             *string

	// 临时值: 每次到达一个请求，都会Clone一个Context，复用了它的全局值，以下值都会重置
	handlers          []func(c *Context)
	Stream            []byte
	PerRequestContext *sync.Map
	offset            int
	contentType       string
	urlPattern        string
}

func NewContext(conn *websocket.Conn) *Context {
	return &Context{
		Conn:                 conn,
		l:                    &sync.RWMutex{},
		PerConnectionContext: &sync.Map{},
		PerRequestContext:    &sync.Map{},
		handlers:             make([]func(c *Context), 0, 10),
		offset:               -1,
	}
}

func (c *Context) Bind(dest interface{}) error {
	body, e := BodyBytesOf(c.Stream)
	if e != nil {
		return errorx.Wrap(e)
	}

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
	buf, e := json.Marshal(v)
	if e != nil {
		return errorx.Wrap(e)
	}
	res, e := PackWithMarshallerAndBody(Message{
		MessageID: int32(0),
		Header: map[string]interface{}{
			HEADER_ROUTER_KEY: HEADER_ROUTER_TYPE_URL_PATTERN,
			HEADER_URL_PATTERN_VALUE_KEY: c.urlPattern,
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

func (c *Context) jsonUrlPattern(messageID int, urlPattern string, v interface{}) error {
	buf, e := json.Marshal(v)
	if e != nil {
		return errorx.Wrap(e)
	}
	res, e := PackWithMarshallerAndBody(Message{
		MessageID: int32(messageID),
		Header: map[string]interface{}{
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
func( c*Context) Abort() {
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