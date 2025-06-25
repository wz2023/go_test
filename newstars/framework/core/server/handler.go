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

package server

import (
	"fmt"
	"net"
	component2 "newstars/framework/core/component"
	"newstars/framework/core/internal/codec"
	"newstars/framework/core/internal/message"
	"newstars/framework/core/internal/packet"
	"newstars/framework/core/session"
	"reflect"
	"time"

	"newstars/framework/glog"
)

// Unhandled message buffer size
const packetBacklog = 1024
const funcBacklog = 1 << 8

var (
	// handler service singleton
	handler = newHandlerService()

	hbd []byte // heartbeat packet data
)

func init() {
	var err error
	hbd, err = codec.Encode(packet.Heartbeat, nil)
	if err != nil {
		panic(err)
	}
}

type (
	handlerService struct {
		services       map[string]*component2.Service // all registered service
		handlers       map[string]*component2.Handler // all handler method
		chLocalProcess chan unhandledMessage          // packets that process locally
		chCloseSession chan *session.Session          // closed session
		chFunction     chan func()
	}

	unhandledMessage struct {
		handler reflect.Method
		args    []reflect.Value
	}
)

func newHandlerService() *handlerService {
	h := &handlerService{
		services:       make(map[string]*component2.Service),
		handlers:       make(map[string]*component2.Handler),
		chLocalProcess: make(chan unhandledMessage, packetBacklog),
		chCloseSession: make(chan *session.Session, packetBacklog),
		chFunction:     make(chan func(), funcBacklog),
	}

	return h
}

// call handler with protected
func pcall(method reflect.Method, args []reflect.Value) {
	defer func() {
		if err := recover(); err != nil {
			glog.SErrorf("nano/dispatch: %v %v", err, stack())
		}
	}()

	if r := method.Func.Call(args); len(r) > 0 {
		if err := r[0].Interface(); err != nil {
			glog.SInfof(err.(error).Error())
		}
	}
}

func pinvoke(fn func()) {
	defer func() {
		if err := recover(); err != nil {
			glog.SErrorf(fmt.Sprintf("nano/invoke: %v %v", err, stack()))
		}
	}()

	fn()
}

func onSessionClosed(s *session.Session) {
	defer func() {
		if err := recover(); err != nil {
			glog.SErrorf("nano/onSessionClosed: %v %v", err, stack())
		}
	}()

	env.muCallbacks.RLock()
	defer env.muCallbacks.RUnlock()

	if len(env.callbacks) < 1 {
		return
	}

	for _, fn := range env.callbacks {
		fn(s)
	}
}

// dispatch message to corresponding logic handler
func (h *handlerService) dispatch() {
	// close chLocalProcess & chCloseSession when application quit
	defer func() {
		close(h.chLocalProcess)
		close(h.chCloseSession)
		globalTicker.Stop()
	}()

	// handle packet that sent to chLocalProcess
	for {
		select {
		case m := <-h.chLocalProcess: // logic dispatch
			pcall(m.handler, m.args)

		case s := <-h.chCloseSession: // session closed callback
			onSessionClosed(s)

		case fn := <-h.chFunction:
			pinvoke(fn)

		case <-globalTicker.C: // execute cron task
			cron()

		case t := <-timerManager.chCreatedTimer: // new timers
			timerManager.timers[t.id] = t

		case id := <-timerManager.chClosingTimer: // closing timers
			delete(timerManager.timers, id)

		case <-env.die: // application quit signal
			return
		}
	}
}

func (h *handlerService) register(c component2.Component) error {
	s := &component2.Service{
		Type:     reflect.TypeOf(c),
		Receiver: reflect.ValueOf(c),
	}
	s.Name = reflect.Indirect(s.Receiver).Type().Name()

	if _, ok := h.services[s.Name]; ok {
		return fmt.Errorf("handler: service already defined: %s", s.Name)
	}

	if err := s.ExtractHandler(); err != nil {
		return err
	}

	// register all handlers
	h.services[s.Name] = s
	for name, method := range s.Methods {
		h.handlers[name] = method
	}
	return nil
}

func (h *handlerService) handle(conn net.Conn) {
	// create a client agent and startup write gorontine
	agent := newAgent(conn)

	// startup write goroutine
	go agent.write()

	if env.debug {
		glog.SInfof("New session established: %s", agent.String())
	}

	// guarantee agent related resource be destroyed
	defer func() {
		agent.Close()
		if env.debug {
			glog.SInfof("Session read goroutine exit, SessionID=%d, UID=%d", agent.session.ID(), agent.session.UID())
		}
	}()

	// read loop
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			glog.SInfof("Read message error: %s, session will be closed immediately", err.Error())
			return
		}

		// TODO(warning): decoder use slice for performance, packet data should be copy before next Decode
		packets, err := agent.decoder.Decode(buf[:n])
		if err != nil {
			glog.SInfof(err.Error())
			return
		}

		if len(packets) < 1 {
			continue
		}

		// process all packet
		for i := range packets {
			if err := h.processPacket(agent, packets[i]); err != nil {
				glog.SInfof(err.Error())
				return
			}
		}
	}
}

func (h *handlerService) processPacket(agent *agent, p *packet.Packet) error {
	switch p.Type {
	case packet.Data:
		// if agent.status() < statusWorking {
		// 	return fmt.Errorf("Receive data on socket which not yet ACK, session will be closed immediately, remote=%s",
		// 		agent.conn.RemoteAddr().String())
		// }

		msg, err := message.Decode2(p.Data)
		if err != nil {
			return err
		}
		h.processMessage(agent, msg)

	case packet.Heartbeat:
		// expected
	}

	agent.lastAt = time.Now().Unix()
	return nil
}

func (h *handlerService) processMessage(agent *agent, msg *message.Message) {
	switch msg.Type {
	case message.Request:
		agent.session.LastRID = msg.ID
	case message.Notify:
		agent.session.LastRID = 0
	}

	handler, ok := h.handlers[msg.Route]
	if !ok {
		glog.SInfof("nano/handler: %s not found(forgot registered?)", msg.Route)
		return
	}

	var data interface{}
	if handler.IsRawArg {
		data = msg.Data
	} else {
		data = reflect.New(handler.Type.Elem()).Interface()
		err := serializer.Unmarshal(msg.Data, data)
		if err != nil {
			glog.SInfof("deserialize error:%s", err.Error())
			return
		}
	}

	if env.debug {
		glog.SInfof("UID=%s, Message={%s}, Data=%+v", agent.session.UID(), msg.String(), data)
	}

	args := []reflect.Value{handler.Receiver, agent.srv, reflect.ValueOf(data), reflect.ValueOf(msg.ID)}
	h.chLocalProcess <- unhandledMessage{handler.Method, args}
}

// DumpServices outputs all registered services
func (h *handlerService) DumpServices() {
	for name := range h.handlers {
		glog.SInfof("registered service %s", name)
	}
}
