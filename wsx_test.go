package wsx

import (
	"testing"
)

func TestUserConn(t *testing.T) {
	srv := NewWsx("/kf")
	srv.Config(
		srv.SetRouteType(RouteTypeAuto),
	)

}
