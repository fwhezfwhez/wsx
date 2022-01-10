package wsx

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"runtime/debug"
	"time"
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

		// 注入连接实例唯一id
		sessionId, _ := NewWrapConn(conn)
		ctx.SetSessionID(sessionId)

		// 注入连接实例所在内网ip和该端口
		ctx.SetHostPort(wsx.getHostPort())

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
				fmt.Printf("%s read reader %s err: %s \n", time.Now().Format("2006-01-02 15:04:05"), ctx.GetSessionID(), errorx.Wrap(e).Error())
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

// 本机内网ip
func GetLocalIP(innerIP string) string {

	if innerIP != "" {
		return innerIP
	}

	localIps, err := getLocalIpList()
	if err != nil {
		fmt.Printf("get local ip failed,err: " + err.Error())
		panic(err)
	}
	if len(localIps) == 0 {
		innerIP = "127.0.0.1"
	} else {
		innerIP = localIps[0]
	}

	return innerIP
}

func getLocalIpList() ([]string, error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return nil, err
	}

	var localipList []string
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localipList = append(localipList, ipnet.IP.To4().String())
			}

		}
	}
	return localipList, nil
}
