package wsx

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/gorilla/websocket"
	"net/http"
	"runtime/debug"
)

// ws begin to listen on
func listenAndServe(relPath string, port string, wsx *Wsx) error {

	// Mode = DEBUG
	wsx.mux.PanicOnExistRouter()

	var ud = &websocket.Upgrader{
		// cross-area
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	http.HandleFunc(relPath, func(w http.ResponseWriter, r *http.Request) {
		conn, e := ud.Upgrade(w, r, w.Header())
		if e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			w.WriteHeader(500)
			w.Write([]byte(e.Error()))
			return
		}
		ctx := NewContext(conn)
		defer conn.Close()

		sessionId, _ := NewWrapConn(conn)
		ctx.SetSessionID(sessionId)

		// 心跳机制
		if wsx.enableHeartbeat == true {
			ctx.SpyingOnHeartbeatWithArgs(wsx.heartBeatInterval)
		}

		// 连接关闭收尾机制
		if wsx.onClose != nil {
			ctx.SetOnClose(func() error {
				wsx.onClose(ctx)
				return nil
			})
		}

		Debugf("[%s]请求进入:", conn.RemoteAddr())

		onConnectMessageExampleMsgid, e := Pack(10000, nil, H{"message": "welcome, this is an example of message 10000"})

		if e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			return
		}

		onConnectMessageExampleURL, e := Pack(0, H{
			"Router-Type":       "URL_PATTERN",
			"URL-Pattern-Value": "/example-of-url-pattern/",
		}, H{"message": "welcome, this is an example of url parttern /example-of-url-pattern/"})

		conn.WriteMessage(websocket.BinaryMessage, onConnectMessageExampleMsgid)
		conn.WriteMessage(websocket.BinaryMessage, onConnectMessageExampleURL)

		for {
			_, co, e := conn.NextReader()
			if e != nil {
				fmt.Println(errorx.Wrap(e).Error())
				break
			}

			block, e := UnpackToBlockFromReader(co)

			messageID, e := MessageIDOf(block)
			if e != nil {
				fmt.Println(errorx.Wrap(e).Error())
				break
			}

			ctxCopy := ctx.Clone()
			ctxCopy.Stream = block
			if IsSerial(messageID) {
				handleStream(block, conn, &ctxCopy, wsx.mux)
			} else {
				go handleStream(block, conn, &ctxCopy, wsx.mux)
			}
		}
	})

	Infof("ws begin to listen %s", port)
	e := http.ListenAndServe(port, nil)
	if e != nil {
		return errorx.Wrap(e)
	}
	return nil
}

// handle stream via middleware
func handleStream(stream []byte, conn *websocket.Conn, ctx *Context, mux *Mux) {
	defer func() {
		if e := recover(); e != nil {
			Fatalf("recover from: %v, \n %s", e, debug.Stack())
			return
		}
	}()

	ctxSession := ctx.Clone()
	ctxSession.Stream = stream

	HandleMiddleware(&ctxSession, *mux)
}
