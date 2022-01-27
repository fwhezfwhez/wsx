package wsx

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestPool(t *testing.T) {
	var raw = "0 0 2 6 0 0 0 0 0 0 0 74 0 0 1 176 123 34 82 111 117 116 101 114 45 84 121 112 101 34 58 34 85 82 76 95 80 65 84 84 69 82 78 34 44 34 85 82 76 45 80 97 116 116 101 114 110 45 86 97 108 117 101 34 58 34 47 114 101 99 118 45 109 101 115 115 97 103 101 45 102 114 111 109 45 107 102 47 34 125 123 34 99 114 101 97 116 101 100 95 97 116 34 58 34 50 48 50 49 45 49 50 45 49 48 84 49 55 58 53 48 58 51 50 46 56 56 57 55 50 57 43 48 56 58 48 48 34 44 34 103 97 109 101 95 105 100 34 58 54 54 44 34 104 101 97 100 101 114 95 117 114 108 34 58 34 104 116 116 112 58 47 47 119 120 46 113 108 111 103 111 46 99 110 47 109 109 111 112 101 110 47 70 122 105 98 84 83 65 74 100 121 120 118 85 78 98 56 117 102 120 117 75 108 113 120 65 85 87 119 86 88 98 71 116 108 97 88 108 87 74 100 78 72 50 101 98 77 55 113 111 118 106 98 65 71 49 102 107 89 98 85 48 67 102 105 98 83 105 97 68 111 68 109 84 81 88 106 84 65 90 74 71 120 118 109 48 101 83 82 109 82 115 52 55 115 77 102 66 52 98 47 48 34 44 34 105 100 34 58 49 44 34 105 109 97 103 101 95 117 114 108 34 58 34 34 44 34 107 102 95 104 101 97 100 101 114 95 117 114 108 34 58 34 34 44 34 107 102 95 110 105 99 107 110 97 109 101 34 58 34 34 44 34 107 102 95 122 111 110 115 116 95 105 100 34 58 48 44 34 109 101 115 115 97 103 101 95 116 121 112 101 34 58 34 116 101 120 116 34 44 34 110 105 99 107 110 97 109 101 34 58 34 98 101 110 98 101 110 122 104 111 117 34 44 34 114 97 119 34 58 110 117 108 108 44 34 115 101 110 100 101 114 95 116 121 112 101 34 58 49 44 34 116 101 120 116 34 58 34 228 189 160 229 165 189 34 44 34 117 112 100 97 116 101 100 95 97 116 34 58 34 50 48 50 49 45 49 50 45 49 48 84 49 55 58 53 48 58 51 50 46 56 56 57 55 50 57 43 48 56 58 48 48 34 44 34 117 115 101 114 95 105 100 34 58 49 48 48 56 54 125"

	var bufstr = strings.Split(raw, " ")

	var buf = make([]byte, 0, 10)

	for _, v := range bufstr {
		bi, e := strconv.Atoi(v)
		if e != nil {
			panic(e)
		}
		buf = append(buf, byte(bi))
	}

	fmt.Println(buf)

	totalLen, e := LengthOf(buf)
	if e != nil {
		panic(e)
	}

	fmt.Println("总长", totalLen)

	headerLen, e := HeaderLengthOf(buf)
	if e != nil {
		panic(e)
	}

	fmt.Println("头长", headerLen)

	bodylen, e := BodyLengthOf(buf)
	if e != nil {
		panic(e)
	}
	fmt.Println("body长", bodylen)

	header, e := HeaderOf(buf)
	if e != nil {
		panic(e)
	}

	fmt.Println("header:", header)

	body, e := BodyBytesOf(buf)
	if e != nil {
		panic(e)
	}

	fmt.Println("body:", string(body))

}

func TestExample(t *testing.T) {
	buf, e := Pack(0, H{
		"Router-Type":       "URL_PATTERN",
		"URL-Pattern-Value": "/example-of-url-pattern/",
	}, H{"message": "welcome, this is an example of url parttern /example-of-url-pattern/"})
	if e != nil {
		panic(e)
	}
	fmt.Println(buf)

	totalLen, e := LengthOf(buf)
	if e != nil {
		panic(e)
	}

	fmt.Println("总长", totalLen)

	headerLen, e := HeaderLengthOf(buf)
	if e != nil {
		panic(e)
	}

	fmt.Println("头长", headerLen)

	bodylen, e := BodyLengthOf(buf)
	if e != nil {
		panic(e)
	}
	fmt.Println("body长", bodylen)

	header, e := HeaderOf(buf)
	if e != nil {
		panic(e)
	}

	fmt.Println("header:", header)

	body, e := BodyBytesOf(buf)
	if e != nil {
		panic(e)
	}

	fmt.Println("body:", string(body))
}
