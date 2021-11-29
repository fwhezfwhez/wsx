package wsx

import (
	"context"
	"fmt"
	"time"
)

// refer to wsx.routeType
const (
	RouteTypeMessageID  = HEADER_ROUTER_TYPE_MESSAGEID
	RouteTypeUrlPattern = HEADER_ROUTER_TYPE_URL_PATTERN
	RouteTypeAuto       = "AUTO"
)

// wsx Object
type Wsx struct {
	ctx context.Context

	// http Upgrade to websocket on this path
	relPath string

	// Adding messageID routes and url-pattern routes
	mux *Mux

	// value ranges in ["AUTO", "MESSAGE_ID","URL_PATTERN"]
	// This value serves for wsx.Any(routeKey interface{}, handlers ... func(c *wsx.Context))
	// `wsxSrv.Any(routeKey, handlers...)`
	// Its default routeType is AUTO. In this case, routeKey can be int or string.
	// If routeType is MESSAGE_ID, routeKey should be int.
	// If routeType is URL_PATTERN, routeKey should be string.
	routeType string

	// heartbeat module.
	// Enabled by `wsxSrv.EnableHeartbeat(20 * time.Second)`
	// whether enable heartbeat.
	enableHeartbeat bool
	// when enableHeartbeat = true, client should keep sending heartbeat in this interval
	heartBeatInterval time.Duration

	// as soon as ctx.Close() is called, onClose will be called
	onClose func(c *Context)

	// lister on
	port string
}

// NewWsxObject
func NewWsx(relPath string) *Wsx {
	m := NewMux()
	m.PanicOnExistRouter()
	return &Wsx{
		relPath:   relPath,
		mux:       NewMux(),
		routeType: RouteTypeAuto,
	}
}

// Op is just an empty struct to help exec config chain, like:
// srv.Config(
//     srv.SetRouteType(wsx.RouteTypeAuto),
//     srv.EnableHeartbeat(15 * time.Second),
//     ...
// )
type Op struct{}

// Config is designed to wrap all configurations.
// However, config chain is not necessary.
// Besides, wsx object newed by NewWsx() has its default config worked well.
func (wsx *Wsx) Config(ops ... Op) {
}

// Set wsxSrv route type.
// routeType only allows wsx.RouteTypeUrlPattern, wsx.RouteTypeAuto, wsx.RouterTypeMessageID
func (wsx *Wsx) SetRouteType(routeType string) Op {
	wsx.routeType = routeType
	return Op{}
}

// Enable heartbeat and set its interval.
// If client not send heartbeat buffer in this interval, server side will close the connection.
func (wsx *Wsx) EnableHeartbeat(interval time.Duration) Op {
	wsx.enableHeartbeat = true
	wsx.heartBeatInterval = interval
	return Op{}
}

// Websocket server will run at port
func (wsx *Wsx) ListenAndServe(port string) error {
	wsx.port = port
	return listenAndServe(wsx.relPath, port, wsx)
}

// Do same as wsx.ListenAndServe.
func (wsx *Wsx) Run(port string) error {
	return wsx.ListenAndServe(port)
}

func (wsx *Wsx) Any(urlPattern string, f ... func(c *Context)) error {
	return wsx.mux.AddURLPatternHandler(urlPattern, f...)
}

func (wsx *Wsx) UseGlobal(f ... func(c *Context)) error {
	return wsx.mux.UseGlobal(f...)
}

func (wsx *Wsx) OnCtxClose(f func(c *Context)) {
	wsx.onClose = f
}

func (wsx *Wsx) getHostPort() string {
	innerIP := GetLocalIP("")
	return fmt.Sprintf("%s%s", innerIP, wsx.port)
}
