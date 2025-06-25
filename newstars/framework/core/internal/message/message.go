// Copyright (c) newstars Author. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package message

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strings"
)

// Type represents the type of message, which could be Request/Notify/Response/Push
type Type byte

// Message types
const (
	Request  Type = 0x02
	Notify        = 0x03
	Response      = 0x01
	Push          = 0x04
)

const (
	msgHeadLength  = 0x05
	msgHead2Length = 0x06
)

var types = map[Type]string{
	Request:  "C",
	Notify:   "N",
	Response: "S",
	Push:     "P",
}

var (
	routes = make(map[string]uint16) // route map to code
	codes  = make(map[uint16]string) // code map to route
)

// Errors that could be occurred in message codec
var (
	ErrWrongMessageType  = errors.New("wrong message type")
	ErrInvalidMessage    = errors.New("invalid message")
	ErrRouteInfoNotFound = errors.New("route info not found in dictionary")
)

// Message represents a unmarshaled message or a message which to be marshaled
type Message struct {
	Type  Type   // message type
	ID    uint   // unique id, zero while notify mode
	Route string // route for locating service
	Data  []byte // payload
}

// New returns a new message instance
func New() *Message {
	return &Message{}
}

// String, implementation of fmt.Stringer interface
func (m *Message) String() string {
	return fmt.Sprintf("Type: %s, ID: %d, Route: %s, BodyLength: %d",
		types[m.Type],
		m.ID,
		m.Route,
		len(m.Data))
}

// Encode marshals message to binary format.
func (m *Message) Encode() ([]byte, error) {
	return Encode(m)
}

// Encode2 marshals message to binary format.
func (m *Message) Encode2() ([]byte, error) {
	return Encode2(m)
}

func routable(t Type) bool {
	return t == Request || t == Notify || t == Push
}

func invalidType(t Type) bool {
	return t < Response || t > Push

}

// Encode marshals message to binary format.
func Encode(m *Message) ([]byte, error) {
	if invalidType(m.Type) {
		return nil, ErrWrongMessageType
	}

	buf := make([]byte, 0)
	buf = append(buf, byte(m.ID))

	var (
		key1 string
		key2 uint8
		key3 uint8
		key4 uint16
	)
	//S3010001
	fmt.Sscanf(m.Route, "%1s%1d%2x%4x", &key1, &key2, &key3, &key4)
	tmp := byte(m.Type)<<4 + key2
	buf = append(buf, tmp)
	buf = append(buf, key3)
	buf = append(buf, byte((key4>>8)&0xFF))
	buf = append(buf, byte(key4&0xFF))
	buf = append(buf, m.Data...)
	return buf, nil
}

// Encode2 marshals message to binary format.
func Encode2(m *Message) ([]byte, error) {
	if invalidType(m.Type) {
		return nil, ErrWrongMessageType
	}

	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(m.ID))

	var (
		key1 string
		key2 uint8
		key3 uint8
		key4 uint16
	)
	//S3010001
	fmt.Sscanf(m.Route, "%1s%1d%2x%4x", &key1, &key2, &key3, &key4)
	tmp := byte(m.Type)<<4 + key2
	buf = append(buf, tmp)
	buf = append(buf, key3)
	buf = append(buf, byte((key4>>8)&0xFF))
	buf = append(buf, byte(key4&0xFF))
	buf = append(buf, m.Data...)
	return buf, nil
}

// Decode unmarshal the bytes slice to a message
func Decode(data []byte) (*Message, error) {
	if len(data) < msgHeadLength {
		return nil, ErrInvalidMessage
	}
	m := New()
	m.ID = uint(data[0])
	flag := data[1]
	m.Type = Type(flag >> 4)
	flag = flag & 0x0F

	m.Route = fmt.Sprintf("%s%d%02X%04X", types[m.Type], flag, data[2], binary.BigEndian.Uint16(data[3:5]))
	m.Data = data[5:]
	return m, nil
}

// Decode2 unmarshal the bytes slice to a message
func Decode2(data []byte) (*Message, error) {
	if len(data) < msgHead2Length {
		return nil, ErrInvalidMessage
	}
	m := New()
	m.ID = uint(binary.BigEndian.Uint16(data))
	flag := data[2]
	m.Type = Type(flag >> 4)
	flag = flag & 0x0F

	m.Route = fmt.Sprintf("%s%d%02X%04X", types[m.Type], flag, data[3], binary.BigEndian.Uint16(data[4:6]))
	m.Data = data[6:]
	return m, nil
}

// SetDictionary set routes map which be used to compress route.
// TODO(warning): set dictionary in runtime would be a dangerous operation!!!!!!
func SetDictionary(dict map[string]uint16) {
	for route, code := range dict {
		r := strings.TrimSpace(route)

		// duplication check
		if _, ok := routes[r]; ok {
			log.Printf("duplicated route(route: %s, code: %d)\n", r, code)
		}

		if _, ok := codes[code]; ok {
			log.Printf("duplicated route(route: %s, code: %d)\n", r, code)
		}

		// update map, using last value when key duplicated
		routes[r] = code
		codes[code] = r
	}
}
