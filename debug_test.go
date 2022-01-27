package wsx

import "testing"

func TestDebuglnf(t *testing.T) {
	Debuglnf("recv eventname %s username %s arg %v", "recv_heart_beat", "fwhezfwhez", jsonline([]int{0,1,2}))
}
