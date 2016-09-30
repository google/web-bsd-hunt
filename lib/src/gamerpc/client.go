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
	"log"
	"fmt"
	"net/http"
	"net/url"

	"apputils"
)

type GameClient struct {
	URL		*url.URL
	RpcType		int
	RProxyOptions	uint64
}

func NewGameClient(ustr string, rpcTypeStr string, rpOptions uint64) (*GameClient, error) {
	var u *url.URL
	u, err := url.Parse(ustr)
	if err != nil {
		return nil, err
	}

	log.Printf("NewGameClient: URL %s\n", u.String())

	rpcType, err := StringToRpcType(rpcTypeStr)
	if err != nil {
		return nil, err
	}

	switch rpcType {
	default:
		return nil, fmt.Errorf("rpctype %d unsupported", rpcType)
	case GR_NETRPC:
		return nil, fmt.Errorf("netrpc deprecated")
	case GR_JSONRPC:
	}

	gc := &GameClient{
		URL:		u,
		RpcType:	rpcType,
		RProxyOptions:	rpOptions,
	}

	return gc, nil
}

func (gc *GameClient) rpc(r *http.Request, service string, method string, request interface{}, reply interface{}) error {
	switch gc.RpcType {
	case GR_NETRPC:
		return fmt.Errorf("GR_NETRPC not supported")
	case GR_JSONRPC:
		return apputils.HttpJsonRpc(r, gc.URL, service + ".J" + method, gc.RProxyOptions, request, &reply)
	}
	return fmt.Errorf("uhnandled rpctype %d", gc.RpcType)
}

func (gc *GameClient) Message(r *http.Request, req *MessageRequest) (*MessageReply, error) {
	var reply MessageReply

	err := gc.rpc(r, "HuntDaemon", "Message", req, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, err
}

func (gc *GameClient) Quit(r *http.Request, req *QuitRequest) (*QuitReply, error) {
	var reply QuitReply

	err := gc.rpc(r, "HuntDaemon", "Quit", req, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, err
}

func (gc *GameClient) Join(r *http.Request, req *JoinRequest) (*JoinReply, error) {
	var reply JoinReply

	err := gc.rpc(r, "HuntDaemon", "Join", req, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

func (gc *GameClient) GameData(r *http.Request, req *GameDataRequest) (*GameDataReply, error) {
	var reply GameDataReply

	err := gc.rpc(r, "HuntDaemon", "GameData", req, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, err
}

func (gc *GameClient) Input(r *http.Request, req *InputRequest) (*InputReply, error) {
	var reply InputReply

	err := gc.rpc(r, "HuntDaemon", "Input", req, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, err
}

func (gc *GameClient) Stats(r *http.Request, req *StatsRequest) (*StatsReply, error) {
	var reply StatsReply

	err := gc.rpc(r, "HuntDaemon", "Stats", req, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}

func (gc *GameClient) Ping(r *http.Request, req *PingRequest) (*PingReply, error) {
	var reply PingReply

	err := gc.rpc(r, "HuntDaemon", "Ping", req, &reply)
	if err != nil {
		return nil, err
	}

	return &reply, nil
}
