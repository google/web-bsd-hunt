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


const(
	ServerVersion	= 0xFFFFFFFF

	C_PLAYER	= 0	// response: game play tcp port
	C_MONITOR	= 1	// response: like C_PLAYER, but no response if 0 players
	C_MESSAGE	= 2	// response: number of players currently in the game
	C_SCORES	= 3	// response: statistics tcp port

	Q_CLOAK		= 1	// enter: cloaked
	Q_FLY		= 2	// enter: flying
	Q_SCAN		= 3	// enter: scanning
)

type ConnectEvent struct {
	Token	int
	Err	error
}

type MessageArgs struct {
	Message	string
	Token	int
}

type QuitRequest struct {
	Token		int
	PlayerID	string
}

type QuitReply struct {
	Token	int
}

type JoinRequest struct {
	Uid		uint32
	Name		string
	Team		string	// must be "0" .. "9" or " "
	EnterStatus	uint32	// any of the Q_* consts
	Ttyname		string
	ConnectMode	uint32	// must be C_MESSAGE, C_PLAYER, or C_MONITOR

	Token		int	// passed back to client in reply
}

type JoinReply struct {
	Token		int
	PlayerID	string	// uniquely identifies a joined player
}

type StatsRequest struct {
	Token	int
}

type StatsReply struct {
	Token	int
	Stats	string
}

type PingRequest struct {
	Token	int
	Seq	int
}

type PingReply struct {
	Token	int
	Seq	int
}

type MessageRequest struct {
	Token		int
	Join		JoinRequest
	Message		string
}

type MessageReply struct {
	Token	int
}

type GameDataRequest struct {
	Token		int
	PlayerID	string
}

type GameDataReply struct {
	Token		int
	Timeout		bool
	TimeoutError	string
	Data		[]uint32
}

type InputRequest struct {
	Token		int
	PlayerID	string
	Keys		string
}

type InputReply struct {
	Token		int
	Timeout		bool
	TimeoutError	string
}

type KeepaliveRequest struct {
	Seq	uint64
}

type KeepaliveReply struct {
	Seq	uint64
}

type HTTPServerExit struct {
	err	error
}
