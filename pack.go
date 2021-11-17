package wsx

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"io"
)

func Pack(messageID int, header map[string]interface{}, body interface{}) ([]byte, error) {
	buf, e := PackWithMarshaller(Message{
		MessageID: int32(messageID),
		Header:    header,
		Body:      body,
	}, JsonMarshaller{})
	if e != nil {
		return nil, errorx.Wrap(e)
	}

	return buf, nil
}

func FirstBlockOfBytes(buffer []byte) ([]byte, error) {
	if len(buffer) < 16 {
		return nil, errorx.NewFromStringf("require buffer length more than 16 but got %d", len(buffer))
	}
	var length = binary.BigEndian.Uint32(buffer[0:4])
	if len(buffer) < 4+int(length) {
		return nil, errorx.NewFromStringf("require buffer length more than %d but got %d", 4+int(length), len(buffer))
	}
	return buffer[:4+int(length)], nil
}

func PackWithMarshallerAndBody(message Message, body []byte) ([]byte, error) {
	var e error
	var lengthBuf = make([]byte, 4)
	var messageIDBuf = make([]byte, 4)
	binary.BigEndian.PutUint32(messageIDBuf, uint32(message.MessageID))
	var headerLengthBuf = make([]byte, 4)
	var bodyLengthBuf = make([]byte, 4)
	var headerBuf []byte
	var bodyBuf []byte
	headerBuf, e = json.Marshal(message.Header)
	if e != nil {
		return nil, e
	}
	binary.BigEndian.PutUint32(headerLengthBuf, uint32(len(headerBuf)))
	bodyBuf = body
	binary.BigEndian.PutUint32(bodyLengthBuf, uint32(len(bodyBuf)))
	var content = make([]byte, 0, 1024)

	content = append(content, messageIDBuf...)
	content = append(content, headerLengthBuf...)
	content = append(content, bodyLengthBuf...)
	content = append(content, headerBuf...)
	content = append(content, bodyBuf...)

	binary.BigEndian.PutUint32(lengthBuf, uint32(len(content)))

	var packet = make([]byte, 0, 1024)

	packet = append(packet, lengthBuf...)
	packet = append(packet, content...)
	return packet, nil
}

func BodyBytesOf(stream []byte) ([]byte, error) {
	headerLen, e := HeaderLengthOf(stream)
	if e != nil {
		return nil, e
	}
	bodyLen, e := BodyLengthOf(stream)
	if e != nil {
		return nil, e
	}
	if len(stream) < 16+int(headerLen)+int(bodyLen) {
		return nil, errors.New(fmt.Sprintf("stream lenth should be bigger than %d", 16+int(headerLen)+int(bodyLen)))
	}
	body := stream[16+headerLen : 16+headerLen+bodyLen]
	return body, nil
}

// Body length of a stream received
func BodyLengthOf(stream []byte) (int32, error) {
	if len(stream) < 16 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than %d", 16))
	}
	bodyLength := binary.BigEndian.Uint32(stream[12:16])
	return int32(bodyLength), nil
}

// Header length of a stream received
func HeaderLengthOf(stream []byte) (int32, error) {
	if len(stream) < 12 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 12"))
	}
	headerLength := binary.BigEndian.Uint32(stream[8:12])
	return int32(headerLength), nil
}

// PackWithMarshaller will encode message into blocks of length,messageID,headerLength,header,bodyLength,body.
// Users don't need to know how pack serializes itself if users use UnpackPWithMarshaller.
//
// If users want to use this protocol across languages, here are the protocol details:
// (they are ordered as list)
// [0 0 0 24 0 0 0 1 0 0 0 6 0 0 0 6 2 1 19 18 13 11 11 3 1 23 12 132]
// header: [0 0 0 24]
// mesageID: [0 0 0 1]
// headerLength, bodyLength [0 0 0 6]
// header: [2 1 19 18 13 11]
// body: [11 3 1 23 12 132]
// [4]byte -- length             fixed_size,binary big endian encode
// [4]byte -- messageID          fixed_size,binary big endian encode
// [4]byte -- headerLength       fixed_size,binary big endian encode
// [4]byte -- bodyLength         fixed_size,binary big endian encode
// []byte -- header              marshal by json
// []byte -- body                marshal by marshaller/
func PackWithMarshaller(message Message, marshaller Marshaller) ([]byte, error) {
	if marshaller == nil {
		marshaller = JsonMarshaller{}
	}
	var e error
	var lengthBuf = make([]byte, 4)
	var messageIDBuf = make([]byte, 4)
	binary.BigEndian.PutUint32(messageIDBuf, uint32(message.MessageID))
	var headerLengthBuf = make([]byte, 4)
	var bodyLengthBuf = make([]byte, 4)
	var headerBuf []byte
	var bodyBuf []byte
	headerBuf, e = json.Marshal(message.Header)
	if e != nil {
		return nil, e
	}
	binary.BigEndian.PutUint32(headerLengthBuf, uint32(len(headerBuf)))
	if message.Body != nil {
		bodyBuf, e = marshaller.Marshal(message.Body)
		if e != nil {
			return nil, e
		}
	}

	binary.BigEndian.PutUint32(bodyLengthBuf, uint32(len(bodyBuf)))
	var content = make([]byte, 0, 1024)

	content = append(content, messageIDBuf...)
	content = append(content, headerLengthBuf...)
	content = append(content, bodyLengthBuf...)
	content = append(content, headerBuf...)
	content = append(content, bodyBuf...)

	binary.BigEndian.PutUint32(lengthBuf, uint32(len(content)))

	var packet = make([]byte, 0, 1024)

	packet = append(packet, lengthBuf...)
	packet = append(packet, content...)
	return packet, nil
}

// messageID of a stream.
// Use this to choose which struct for unpacking.
func MessageIDOf(stream []byte) (int32, error) {
	if len(stream) < 8 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 8"))
	}
	messageID := binary.BigEndian.Uint32(stream[4:8])
	return int32(messageID), nil
}

// header of a block
func HeaderOf(stream []byte) (map[string]interface{}, error) {
	var header map[string]interface{}
	headerBytes, e := HeaderBytesOf(stream)
	if e != nil {
		return nil, errorx.Wrap(e)
	}
	e = json.Unmarshal(headerBytes, &header)
	if e != nil {
		return nil, errorx.Wrap(e)
	}
	return header, nil
}

// Header bytes of a block
func HeaderBytesOf(stream []byte) ([]byte, error) {
	headerLen, e := HeaderLengthOf(stream)
	if e != nil {
		return nil, e
	}
	if len(stream) < 16+int(headerLen) {
		return nil, errors.New(fmt.Sprintf("stream lenth should be bigger than %d", 16+int(headerLen)))
	}
	header := stream[16 : 16+headerLen]
	return header, nil
}

func UnpackToBlockFromReader(reader io.Reader) ([]byte, error) {
	if reader == nil {
		return nil, errors.New("reader is nil")
	}
	var info = make([]byte, 4, 4)
	if e := readUntil(reader, info); e != nil {
		if e == io.EOF {
			return nil, e
		}
		return nil, errorx.Wrap(e)
	}

	length, e := LengthOf(info)
	if e != nil {
		return nil, e
	}
	fmt.Println(length)
	var content = make([]byte, length, length)
	if e := readUntil(reader, content); e != nil {
		if e == io.EOF {
			return nil, e
		}
		return nil, errorx.Wrap(e)
	}

	return append(info, content ...), nil
}

func readUntil(reader io.Reader, buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	var offset int
	for {
		n, e := reader.Read(buf[offset:])
		if e != nil {
			if e == io.EOF {
				return e
			}
			return errorx.Wrap(e)
		}
		offset += n
		if offset >= len(buf) {
			break
		}
	}
	return nil
}

// Length of the stream starting validly.
// Length doesn't include length flag itself, it refers to a valid message length after it.
func LengthOf(stream []byte) (int32, error) {
	if len(stream) < 4 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 4"))
	}
	length := binary.BigEndian.Uint32(stream[0:4])
	return int32(length), nil
}
