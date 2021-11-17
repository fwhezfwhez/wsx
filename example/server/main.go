package main

import (
	"fmt"
	"log"

	"github.com/fwhezfwhez/wsx"
)

var pool = wsx.NewPoolV2(nil)

func init() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
}
func main() {
	go ws()
	select {}
}

func ws() {
	// mux
	r:= wsx.NewWsx("/kf-ws")
	r.Any("/login/", func(c *wsx.Context) {
		type Param struct {
			Username string `json:"username"`
			Chanel   string `json:"chanel"`
		}
		var param Param
		if e := c.Bind(&param); e != nil {
			panic(e)
		}
		sessionID, wrapC := wsx.NewWrapConn(c.Conn)

		c.SetSessionID(sessionID)
		c.SetUsername(param.Username)
		pool.Online(param.Username, param.Chanel, wrapC)
		c.JSONUrlPattern(wsx.H{"message": "登录成功"})
	})

	r.UseGlobal(func(c *wsx.Context) {
		fmt.Printf("当前用户: %s\n", c.Username())
		fmt.Printf("当前路由：%s\n", c.GetUrlPattern())
	})


	r.ListenAndServe(":8111")
}
