package wsx

import (
	"crypto/md5"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"strings"
)

// handle c and mux
// c has received stream and mux has routers
func HandleMiddleware(c *Context, mux Mux) error {
	if c.handlers == nil {
		c.handlers = make([]func(c *Context), 0, 10)
	}

	if len(mux.globalMiddlewares) != 0 {
		c.handlers = append(c.handlers, mux.globalMiddlewares...)
	}

	if len(c.Stream) == 0 {
		return errorx.NewFromString("c.Stream is nil")
	}
	messageID, e := MessageIDOf(c.Stream)
	if e != nil {
		return errorx.Wrap(e)
	}
	header, e := HeaderOf(c.Stream)
	if e != nil {
		return errorx.Wrap(e)
	}

	var handlers []func(c *Context)
	var exist bool

	if len(header) > 0 {
		v, ok := header[HEADER_ROUTER_KEY]
		if ok {
			vstring, ok2 := v.(string)
			if !ok2 {
				return errorx.NewFromStringf("bad type of header.%s, requires string but got %v", HEADER_ROUTER_KEY, v)
			}
			switch vstring {
			case HEADER_ROUTER_TYPE_MESSAGEID:
				handlers, exist = mux.messageIDMux[int(messageID)]
				if !exist {
					return errorx.NewFromStringf("not found handler for messageID '%d'", messageID)
				}
			case HEADER_ROUTER_TYPE_URL_PATTERN:
				urlPatternI, ok3 := header[HEADER_URL_PATTERN_VALUE_KEY]
				if !ok3 {
					return errorx.NewFromStringf("detected route type '%s=%s', but urlPattern '%s' not set yet",
						HEADER_ROUTER_KEY, vstring, HEADER_URL_PATTERN_VALUE_KEY)
				}
				urlPattern, ok4 := urlPatternI.(string)
				if !ok4 {
					return errorx.NewFromStringf("%s requires string type", HEADER_URL_PATTERN_VALUE_KEY)
				}

				handlers, exist = mux.urlPatternMux[urlPattern]
				if !exist {
					return errorx.NewFromStringf("not found handler for messageID '%d'", messageID)
				}
				c.urlPattern = urlPattern
			}
		} else {
			handlers, ok = mux.messageIDMux[int(messageID)]
			if !ok {
				return errorx.NewFromStringf("not found handler for messageID '%d'", messageID)
			}
		}

		contentType, exist, e := headerGetString(header, HEADER_CONTENT_TYPE_KEY)
		if e != nil {
			return errorx.Wrap(e)
		}
		if !exist || v == CONTENT_TYPE_JSON {
			c.contentType = CONTENT_TYPE_JSON
		} else {
			c.contentType = contentType
		}
	}

	if len(handlers) > 0 {
		c.handlers = append(c.handlers, handlers ...)
	}

	if len(c.handlers) > 0 {
		c.Next()
	}
	c.Reset()

	return nil
}

// get key-value from a header
func headerGetString(header map[string]interface{}, key string) (string, bool, error) {
	var exist bool
	var value string
	var valueI interface{}
	if len(header) == 0 {
		return "", false, nil
	}
	valueI, exist = header[key]

	if !exist {
		return "", exist, nil
	}

	var canConvert bool
	value, canConvert = valueI.(string)
	if !canConvert {
		return "", exist, errorx.NewFromStringf("key '%s'exist but is not a string type", key)
	}
	return value, exist, nil
}

// whether messageID means serial.
// A stream marked serial means  other requests after this serial message will wait until this stream handled
func IsSerial(messageID int32) bool {
	return messageID == SERIAL
}

//
func GetChanelUsername(chanel string, username string) string {
	return fmt.Sprintf("%s:%s", chanel, username)
}

func MD5(rawMsg string) string {
	data := []byte(rawMsg)
	has := md5.Sum(data)
	md5str1 := fmt.Sprintf("%x", has)
	return strings.ToUpper(md5str1)
}
