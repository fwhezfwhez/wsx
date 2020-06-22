package wsx

var Mode string

const (
	DEBUG   = "debug"
	RELEASE = "release"
)

const (
	// messageID=SERIAL时，该消息会被串行化处理
	SERIAL = 2020
)

var ForbidenServiceMessageIDs = []int32{
	SERIAL,
}
