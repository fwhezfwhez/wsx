package wsx

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/gorilla/websocket"
	"net/http"
	"runtime/debug"
)

func ListenAndServe(port string, mux *Mux, pool *PoolV2) error {

	Mode = DEBUG
	mux.PanicOnExistRouter()


	var ud = &websocket.Upgrader{
		// 解决跨域问题
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	http.HandleFunc("/kf", func(w http.ResponseWriter, r *http.Request) {
		conn, e := ud.Upgrade(w, r, w.Header())
		if e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			w.WriteHeader(500)
			w.Write([]byte(e.Error()))
			return
		}
		ctx := NewContext(conn)

		// ctx.SpyingOnHeartbeat()

		fmt.Println("请求进入:", conn.RemoteAddr())

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
				handleStream(stream, conn, ctx, mux)
			} else {
				go handleStream(stream, conn, ctx, mux)
			}
		}
	})

	fmt.Println("ws begin to listen", port)
	e := http.ListenAndServe(port, nil)
	if e != nil {
		return errorx.Wrap(e)
	}
	return nil
}

func handleStream(stream []byte, conn *websocket.Conn, ctx *Context, mux *Mux) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(fmt.Sprintf("recover from: %v, \n %s", e, debug.Stack()))
			return
		}
	}()

	ctxSession := ctx.Clone()
	ctxSession.Stream = stream

	HandleMiddleware(&ctxSession, *mux)
}
