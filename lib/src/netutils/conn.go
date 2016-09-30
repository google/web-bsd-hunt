// Copyright 2016 The Web BSD Hunt Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
////////////////////////////////////////////////////////////////////////////////
//
// TODO: High-level file comment.
package netutils

import(
	"net"
	"time"
)

type TimeoutTCPConn struct {
	conn	*net.TCPConn
	timeout	time.Duration
}

func NewTimeoutTCPConn(conn *net.TCPConn, timeout time.Duration) *TimeoutTCPConn {
	return &TimeoutTCPConn{conn: conn, timeout: timeout}
}

func (toc *TimeoutTCPConn) Read(buf []byte) (int, error) {
	err := toc.conn.SetReadDeadline(time.Now().Add(toc.timeout))
	if err != nil {
		return 0, err
	}

	return toc.conn.Read(buf)
}

func (toc *TimeoutTCPConn) Write(buf []byte) (int, error) {
	err := toc.conn.SetWriteDeadline(time.Now().Add(toc.timeout))
	if err != nil {
		return 0, err
	}

	return toc.conn.Write(buf)
}

func (toc *TimeoutTCPConn) Close() error {
	return toc.conn.Close()
}

type TimeoutUDPConn struct {
	conn	*net.UDPConn
	timeout	time.Duration
}

func NewTimeoutUDPConn(conn *net.UDPConn, timeout time.Duration) *TimeoutUDPConn {
	return &TimeoutUDPConn{conn: conn, timeout: timeout}
}

func (toc *TimeoutUDPConn) Read(buf []byte) (int, error) {
	err := toc.conn.SetReadDeadline(time.Now().Add(toc.timeout))
	if err != nil {
		return 0, err
	}

	return toc.conn.Read(buf)
}

func (toc *TimeoutUDPConn) Write(buf []byte) (int, error) {
	err := toc.conn.SetWriteDeadline(time.Now().Add(toc.timeout))
	if err != nil {
		return 0, err
	}

	return toc.conn.Write(buf)
}

func (toc *TimeoutUDPConn) Close() error {
	return toc.conn.Close()
}

func (toc *TimeoutUDPConn) ReadFromUDP(buf []byte) (int, *net.UDPAddr, error) {
	err := toc.conn.SetReadDeadline(time.Now().Add(toc.timeout))
	if err != nil {
		return 0, nil, err
	}

	return toc.conn.ReadFromUDP(buf)
}
