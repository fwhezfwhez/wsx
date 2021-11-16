package wsx

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/gorilla/websocket"
	"net/http"
	"runtime/debug"
)

// ws begin to listen on
func listenAndServe(relPath string,port string, wsx *Wsx) error {

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

		// ctx.SpyingOnHeartbeat()

		Debugf("[%s]请求进入:", conn.RemoteAddr())

		defer conn.Close()

		// 当粘包时，res存放多余的粘包块，用于给下一次读取补齐
		var res []byte
		for {
			var raw []byte
			_, message, e := conn.ReadMessage()
			if e != nil {
				fmt.Println(errorx.Wrap(e).Error())
				break
			}
			fmt.Println("message:", message)
			if len(res) != 0 {
				raw = append(res, message ...)
				res = nil
			} else {
				raw = message
			}
			fmt.Println("raw", raw)
			stream, e := FirstBlockOfBytes(raw)
			if e != nil {
				fmt.Println(errorx.Wrap(e).Error())
				break
			}

			res = raw[len(stream):]

			fmt.Println(MessageIDOf(stream))

			messageID, e := MessageIDOf(stream)
			if e != nil {
				fmt.Println(errorx.Wrap(e).Error())
				break
			}

			if IsSerial(messageID) {
				handleStream(stream, conn, ctx, wsx.mux)
			} else {
				go handleStream(stream, conn, ctx, wsx.mux)
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