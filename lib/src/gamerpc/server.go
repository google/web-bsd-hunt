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
package gamerpc

import(
	"fmt"
	"time"
	"net"
	"net/http"
	"net/rpc"

	grpc "github.com/gorilla/rpc"
	gjson "github.com/gorilla/rpc/json"
)

type GameServer struct {
	host		string
	port		string
	rpcType		int
	keepaliveDelay	time.Duration
	eventc		chan interface{}
	listener	net.Listener
	gserver		*grpc.Server
}

func (gs *GameServer) info(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "host %s port %s rpctype %d\n", gs.host, gs.port, gs.rpcType)
}

func (gs *GameServer) emptyjs(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "console.log(\"yay Loaded empty.js from game server!\")\n")
}

func (gs *GameServer) keepalive(eventc chan interface{}) {
	var seq uint64
	for {
		eventc <- &KeepaliveRequest{Seq: seq}
		seq++

		time.Sleep(gs.keepaliveDelay)
	}
}

func (gs *GameServer) start() {
	http.HandleFunc("/info", gs.info)
	http.HandleFunc("/empty.js", gs.emptyjs)

	go gs.keepalive(gs.eventc)

	err := http.Serve(gs.listener, nil)

	event := &HTTPServerExit{err: err}
	gs.eventc <- event
}

func NewGameServer(host string, port string, rpcTypeStr string, service interface{}, keepaliveDelay time.Duration, eventc chan interface{}) (*GameServer, error) {
	var err error
	var rpcType int

	rpcType, err = StringToRpcType(rpcTypeStr)
	if err != nil {
		return nil, err
	}

	gs := &GameServer{
		host:		host,
		port:		port,
		rpcType:	rpcType,
		keepaliveDelay:	keepaliveDelay,
		eventc:		eventc,
		listener:	nil,
		gserver:	nil,
	}

	switch(rpcType) {
	case GR_NETRPC:
		err = rpc.Register(service)
		if err != nil {
			return nil, err
		}
		rpc.HandleHTTP()
	case GR_JSONRPC:
		s := grpc.NewServer()
		s.RegisterCodec(gjson.NewCodec(), "application/json")
		s.RegisterCodec(gjson.NewCodec(), "text/plain")
		s.RegisterService(service, "")
		http.Handle("/jsonrpc", s)
		gs.gserver = s
	default:
		return nil, fmt.Errorf("unhandled rpc type '%s' (%d)", rpcTypeStr, rpcType)
	}

	gs.listener, err = net.Listen("tcp", host + ":" + port)
	if err != nil {
		return nil, err
	}

	go gs.start()

	return gs, nil
}
