package main

import (
	"fmt"
	"time"

	"log"

	"wsx"
)

var pool = wsx.NewPoolV2(nil)

func init() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
}
func main() {
	go ws()

	go func() {
		time.Sleep(30 * time.Second)

		uc, ok, e := pool.IsOnline("fengtao")
		fmt.Println("isOnline:", uc, ok, e)

		uc.JSONUrlPattern("/hehe", wsx.H{
			"message": "测试推送次数",
		})
	}()
	select {}
}

func ws() {
	// mux
	mux := wsx.NewMux()
	mux.AddURLPatternHandler("/login/", func(c *wsx.Context) {
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

	mux.UseGlobal(func(c *wsx.Context) {
		fmt.Printf("当前用户: %s\n", c.Username())
		fmt.Printf("当前路由：%s\n", c.GetUrlPattern())
	})

	mux.LockWrite()

	wsx.ListenAndServe(":8111", mux, pool)
}
