// Copyright 2022 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middlewares

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/xgfone/go-apiserver/tcp"
)

type tcpHandler struct{ buf *bytes.Buffer }

func (h tcpHandler) OnConnection(net.Conn)      { h.buf.WriteString("pass") }
func (h tcpHandler) OnServerExit(err error)     {}
func (h tcpHandler) OnShutdown(context.Context) {}

type tcpAddr struct{ addr string }

func (a tcpAddr) Network() string { return "tcp" }
func (a tcpAddr) String() string  { return a.addr }

type tcpConn struct{ remoteAddr string }

func (c tcpConn) Read(b []byte) (n int, err error)         { return }
func (c tcpConn) Write(b []byte) (n int, err error)        { return }
func (c tcpConn) Close() (err error)                       { return }
func (c tcpConn) LocalAddr() (addr net.Addr)               { return }
func (c tcpConn) RemoteAddr() net.Addr                     { return tcpAddr{c.remoteAddr} }
func (c tcpConn) SetDeadline(t time.Time) (err error)      { return }
func (c tcpConn) SetReadDeadline(t time.Time) (err error)  { return }
func (c tcpConn) SetWriteDeadline(t time.Time) (err error) { return }

func TestIPWhitelist(t *testing.T) {
	mw, err := IPWhitelist(0, "10.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(nil)
	handler := mw.Handler(tcpHandler{buf}).(tcp.Handler)
	handler.OnConnection(tcpConn{"10.1.2.3:1234"})
	if s := buf.String(); s != "pass" {
		t.Errorf("expect '%s', but got '%s'", "pass", s)
	}

	buf.Reset()
	handler.OnConnection(tcpConn{"1.2.3.4:1234"})
	if s := buf.String(); s != "" {
		t.Errorf("expect '%s', but got '%s'", "", s)
	}
}
