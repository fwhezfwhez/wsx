package main

import (
	"flag"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/wsx"
	ws2 "github.com/gorilla/websocket"
	"net/url"
)

func main() {
	gorilla()
}

func gorilla() {
	var addr = flag.String("addr", "localhost:8111", "http service address")
	flag.Parse()
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/kf-ws"}
	fmt.Println("connecting to ", u.String())

	c, _, e := ws2.DefaultDialer.Dial(u.String(), nil)
	if e != nil {
		fmt.Println(errorx.Wrap(e).Error())
		return
	}
	defer c.Close()

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
			fmt.Println("body:", string(body))
		}
	}()

	buf, e := wsx.Pack(0, wsx.H{
		"Router-Type":       "URL_PATTERN",
		"URL-Pattern-Value": "/login/",
	}, wsx.H{
		"username": "fengtao",
		"password": "qq",
	})
	if e != nil {
		panic(e)
	}

	c.WriteMessage(ws2.BinaryMessage, buf)

	select {}
}
