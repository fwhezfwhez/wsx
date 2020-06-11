package wsx

const(
	// message 如果带有 header["Router-Type"]，则可以根据它的值，选择按照messageID处理或者url方式处理
	// 如果header["Router-Type"] = "MESSAGE_ID",或者没有这个字段,则会从messageID路由中找到handler
	// 如果header["Router-Type"] = "URL_PATTERN",则会从urlPattern路由中找到handler

	// 当header["ROUTER_TYPE"] = "URL_PATTERN"时,则将会以header["URL_PATTERN_VALUE"]的值，用来转发消息处理
	// 实例:
	// header:
	// {
	//     "Router-Type": "URL_PATTERN",
	//     "URL-Pattern-Value": "/user/user-info/"
	// }
	HEADER_ROUTER_KEY = "Router-Type"
	HEADER_ROUTER_TYPE_MESSAGEID = "MESSAGE_ID"
	HEADER_ROUTER_TYPE_URL_PATTERN = "URL_PATTERN"
	HEADER_URL_PATTERN_VALUE_KEY = "URL-Pattern-Value"
)

const (
	// 协议Body会按照声明的CONTENT-TYPE进行解析.
	// 实例:
	// header:
	// {
	//     "Content-Type": "JSON",
	// }
	HEADER_CONTENT_TYPE_KEY = "Content-Type"
)

const(
	CONTENT_TYPE_JSON = "JSON"
)
// Message contains the necessary parts of tcpx protocol
// MessagID is defining a message routing flag.
// Header is an attachment of a message.
// Body is the message itself, it should be raw message not serialized yet, like "hello", not []byte("hello")
type Message struct {
	MessageID int32                  `json:"message_id"`
	Header    map[string]interface{} `json:"header"`
	Body      interface{}            `json:"body"`
}

// Get value of message's header whose key is 'key'
// Get and Set don't have lock to ensure concurrently safe, which means
// if you should never operate the header in multiple goroutines, it's better to design a context yourself per request
// rather than straightly use message.Header.
func (msg Message) Get(key string) interface{} {
	if msg.Header == nil {
		return nil
	}
	return msg.Header[key]
}

// Get and Set don't have lock to ensure concurrently safe, which means
// if you should never operate the header in multiple goroutines, it's better to design a context yourself per request
// rather than straightly use message.Header.
func (msg *Message) Set(k string, v interface{}) {
	msg.Header[k] = v
}
