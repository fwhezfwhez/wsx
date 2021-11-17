package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/wsx"
	ws2 "github.com/gorilla/websocket"
	"net/url"
)

func main() {

	for i := 0; i < 1; i++ {
		go gorilla()
	}

	select {

	}
}

func gorilla() {
	u := url.URL{Scheme: "ws", Host: "localhost:8111", Path: "/kf-ws"}
	fmt.Println("connecting to ", u.String())

	c, _, e := ws2.DefaultDialer.Dial(u.String(), nil)
	if e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		return
	}

	go func() {
		for {
			_, message, e := c.ReadMessage()
			if e != nil {
				fmt.Println(errorx.Wrap(e).Error())
				return
			}

			body, e := wsx.BodyBytesOf(message)
			if e != nil {
				fmt.Println(errorx.Wrap(e).Error())
				return
			}
			// fmt.Println("header:", header)
			fmt.Println("body:", string(body))
		}
	}()

	buf, e := wsx.Pack(0, wsx.H{
		"Router-Type":       "URL_PATTERN",
		"URL-Pattern-Value": "/login/",
	}, wsx.H{
		"username": "fengtao",
		"chanel":   "vx",
	})
	if e != nil {
		panic(e)
	}

	c.WriteMessage(ws2.BinaryMessage, buf)
}
