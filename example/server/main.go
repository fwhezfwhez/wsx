package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"runtime/debug"
	"wsx"
)

func init() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
}
func main() {
	go ws()
	select {}
}

func ws() {
	mux := wsx.NewMux()
	// mux.PanicOnExistRouter()

	mux.UseGlobal(func(c *wsx.Context) {
		fmt.Println("我是全局中间件")
		c.Next()
		fmt.Println("请求完毕")
	})
	// 登录
	mux.AddURLPatternHandler("/user/user-info/login/", func(c *wsx.Context) {
		fmt.Println("我是登录中间件")
		c.Next()
		fmt.Println("登录中间件执行完毕")
	}, func(c *wsx.Context) {
		type UserInfo struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		var ui UserInfo
		if e := c.Bind(&ui); e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			c.JSONUrlPattern(wsx.H{
				"tip":           "参数异常",
				"tip_id":        "1",
				"debug_message": errorx.Wrap(e).Error(),
			})
			return
		}

		fmt.Println(ui)

		c.JSONUrlPattern(wsx.H{
			"tip":    "登录成功",
			"tip_id": "0",
		})

	})

	// 拉取用户信息
	mux.AddURLPatternHandler("/user/user-info/list-users/", func(c *wsx.Context) {
		type Param struct {
			GameId int `json:"game_id"`
		}
		var param Param
		if e:=c.Bind(&param); e!=nil{
			fmt.Println(errorx.Wrap(e).Error())
			return
		}
		fmt.Println(param.GameId)
		c.JSONUrlPattern(wsx.H{
			"tip":    "登录成功",
			"tip_id": "0",
			"users":  []int{1, 2, 3, 4, 5},
		})
	})
	mux.LockWrite()

	var ud = &websocket.Upgrader{
		// 解决跨域问题
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	onConnectMessage, e := wsx.Pack(2, nil, wsx.H{"message": "welcome"})
	if e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		return
	}
	fmt.Println(onConnectMessage)

	http.HandleFunc("/kf", func(w http.ResponseWriter, r *http.Request) {
		conn, e := ud.Upgrade(w, r, w.Header())
		if e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			w.WriteHeader(500)
			w.Write([]byte(e.Error()))
			return
		}
		ctx := wsx.NewContext(conn)

		fmt.Println("请求进入:", conn.RemoteAddr())

		//if e := conn.WriteMessage(websocket.BinaryMessage, onConnectMessage); e != nil {
		//	fmt.Println(errorx.Wrap(e).Error())
		//	return
		//}

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
			//fmt.Println("message:", message)

			if len(res) != 0 {
				raw = append(res, message ...)
				res = nil
			} else {
				raw = message
			}
			//fmt.Println("raw:", raw)
			stream, e := wsx.FirstBlockOfBytes(raw)
			if e != nil {
				fmt.Println(errorx.Wrap(e).Error())
				break
			}

			res = raw[len(stream):]

			go func(stream []byte, conn *websocket.Conn) {
				defer func() {
					if e := recover(); e != nil {
						fmt.Println(fmt.Sprintf("recover from: %v", e))
						fmt.Println(string(debug.Stack()))
						return
					}
				}()

				ctxSession := ctx.Clone()
				ctxSession.Stream = stream

				wsx.HandleMiddleware(&ctxSession, *mux)
			}(stream, conn)
		}
	})
	fmt.Println("begin to listen")
	e = http.ListenAndServe(":8080", nil)
	if e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		return
	}
}
