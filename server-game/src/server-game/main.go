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
package main

import(
	"encoding/binary"
	"fmt"
	"flag"
	"time"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"os"
	"net/url"

	"byteutils"
	"gamerpc"
	"netutils"
	"loggy"
	"apputils"

	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/compute/metadata"
)

const(
	HuntdTimeout		= 1*1000 * time.Millisecond	// don't wait longer than this for any huntd I/O to complete
	KeepAliveTimeout	= 1*10000 * time.Millisecond	// send a keepalive every 10 seconds
)

type Player struct {
	ID		string
	serverVersion	uint32
	gameAddr	*net.TCPAddr
	gameConn	*netutils.TimeoutTCPConn
	joinRequest	gamerpc.JoinRequest
}

type HuntDaemon struct {
	WellKnownPort	string
	StatisticsAddr	string
	GamePlayAddr	string

	wkAddr		*net.UDPAddr
	wkConn		*netutils.TimeoutUDPConn

	gameAddr	*net.TCPAddr
	statsAddr	*net.TCPAddr

	Players		map[string]*Player
}

var logger = loggy.MustNewLoggerFromString(
		[]string{
			"LOG_STARTUP",
			"LOG_EVENT",
			"LOG_RPC",
			"LOG_HUNTD_CONNECT",
			"LOG_PLAYER_API",
			"LOG_KEEPALIVE",
		},
		os.Getenv("SERVER_GAME_OPTIONS"))

var LOG_STARTUP		= logger.MustLevel("LOG_STARTUP")
var LOG_EVENT		= logger.MustLevel("LOG_EVENT")
var LOG_RPC		= logger.MustLevel("LOG_RPC")
var LOG_HUNTD_CONNECT	= logger.MustLevel("LOG_HUNTD_CONNECT")
var LOG_PLAYER_API	= logger.MustLevel("LOG_PLAYER_API")
var LOG_KEEPALIVE	= logger.MustLevel("LOG_KEEPALIVE")

func NewHuntDaemon(host string, wkport string) (*HuntDaemon, error) {
	logger.Log(LOG_HUNTD_CONNECT, "Contacting huntd @ %s ...", wkport)

	huntd := &HuntDaemon{
		WellKnownPort:	wkport,
		Players:	make(map[string]*Player),
	}

	var err error

	huntd.wkAddr, err = net.ResolveUDPAddr("udp", host + ":" + huntd.WellKnownPort)
	if err != nil {
		return nil, err
	}

	var conn *net.UDPConn
	conn, err = net.DialUDP("udp", nil, huntd.wkAddr)
	if err != nil {
		return nil, err
	}

	huntd.wkConn = netutils.NewTimeoutUDPConn(conn, HuntdTimeout)

	logger.Log(LOG_HUNTD_CONNECT, "Requesting gameplay port")
	gPort, wkpAddr, err := huntd.wkRequest(gamerpc.C_PLAYER)
	if err != nil {
		huntd.wkConn.Close()
		return nil, err
	}
	logger.Log(LOG_HUNTD_CONNECT, "gameplay port is %d on host %s", gPort, ReplacePort(wkpAddr.String(), ""))

	logger.Log(LOG_HUNTD_CONNECT, "Requesting stats port")
	sPort, statsAddr, err := huntd.wkRequest(gamerpc.C_SCORES)
	if err != nil {
		huntd.wkConn.Close()
		return nil, err
	}
	logger.Log(LOG_HUNTD_CONNECT, "statistics port is %d on host %s", sPort, ReplacePort(statsAddr.String(), ""))

	gpstr := ReplacePort(wkpAddr.String(), fmt.Sprintf("%d", gPort))
	huntd.gameAddr, err = net.ResolveTCPAddr("tcp", gpstr)
	if err != nil {
		huntd.wkConn.Close()
		return nil, err
	}

	ststr := ReplacePort(statsAddr.String(), fmt.Sprintf("%d", sPort))
	huntd.statsAddr, err = net.ResolveTCPAddr("tcp", ststr)
	if err != nil {
		huntd.wkConn.Close()
		return nil, err
	}

	return huntd, nil
}

func (huntd *HuntDaemon) wkRequest(op uint16) (uint16, *net.UDPAddr, error) {
	var err error

	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, op)

	var n int
	n, err = huntd.wkConn.Write(b)
	if err != nil {
		return 0, nil, err
	}
	if n != len(b) {
		return 0, nil, fmt.Errorf("short write: wrote %d expected %d", n, len(b))
	}

	rxBuf := make([]byte, 2)
	n, fromAddr, err := huntd.wkConn.ReadFromUDP(rxBuf)
	if err != nil {
		return 0, nil, err
	}
	if n != len(rxBuf) {
		return 0, nil, fmt.Errorf("short read: read %d expected %d", n, len(rxBuf))
	}

	return binary.BigEndian.Uint16(rxBuf), fromAddr, nil
}

func (huntd *HuntDaemon) player(id string) (*Player, error) {
	player, found := huntd.Players[id]
	if !found {
		return nil, fmt.Errorf("%s: no such player", id)
	}

	return player, nil
}


func (huntd *HuntDaemon) newPlayer() (*Player, error) {
	player := &Player{
		ID:		uuid.NewV4().String(),
		gameAddr:	huntd.gameAddr,
		gameConn:	nil,
	}

	logger.Log(LOG_PLAYER_API, "Created new player %s", player.ID)

	return player, nil
}

func readServerVersion(tc *netutils.TimeoutTCPConn) (uint32, error) {
	verbuf := make([]byte, 4)

	n, err := tc.Read(verbuf)
	if err != nil {
		return 0, err
	}
	if n != 4 {
		return 0, fmt.Errorf("short read: got %d expected 4", n)
	}
	serverVersion := binary.BigEndian.Uint32(verbuf)
	if serverVersion != gamerpc.ServerVersion {
		return 0, fmt.Errorf("unexpected server version %x", serverVersion)
	}

	return serverVersion, nil
}

/*
 * The join protocol is somewhat messy because according to the spec (README.protocol),
 * Paraphrasing:
 *	C->S: Join Message
 *	S->C: Server Version
 *
 * In actual fact, the server sends the version before reading anything (on debian),
 * which becomes problematic to adhere to given the following observation:
 *
 * If Connect Mode is C_MESSAGE, the protocol is specified as follows:
 *	C->S: Join Message
 *	S->C: Server Version
 *	C->S: Message (up to 1024 bytes)
 *	C:    hangup
 *
 * There is a subtle issue which makes message receipt by the server
 * somewhat hit or miss.  Immediately after reading the last byte of 
 * the Join message, the server places the socket into nonblocking mode
 * attempts to read the message, and hangs up as soon as it gets a 0 byte read.
 * Thus, depending on the vagaries of process scheduling, network latency, and
 * the phase of the moon, message delivery may or may not work.
 *
 * The solution is to read the server version before sending anything, and then
 * send a single buffer containing the join message and the message.
 * While I don't think it's guaranteed that all of the bytes will be delivered and available
 * to the client, in most cases everything fits into a single packet, so in practice
 * this should be fine.
 *
 * Additionally, the debian 'bsdgames' package version 2.17-21 actually has a bug
 * in the server, where it reads the last attribute of the Join message as an 8 byte
 * u_long when the spec declares it to be a 4 byte value.  This causes
 * the first 4 bytes of the message to get lost, and for the mode value to be
 * filled with a bogus number, which luckily enough gets ignored due to passing
 * it through a call to ntohl().  As it turns out, the real hunt client also exhibits 
 * this bug.  I investigated the NetBSD version, and this bug does not exist there.
 *
 * As if this weren't enough, the bsdgames-osx github version has a different
 * join implementation, which doesn't send the server version until
 * after the join message is received.
 */
func (p *Player) Join(joinRequest *gamerpc.JoinRequest, mstr string) error {
	logger.Log(LOG_PLAYER_API, "Player %s: join huntd gameplay server @ %s\n", p.ID, p.gameAddr)

	var xgc *net.TCPConn
	var tgc *netutils.TimeoutTCPConn
	var err error
	var n int

	xgc, err = net.DialTCP("tcp", nil, p.gameAddr)
	if err != nil {
		return err
	}

	tgc = netutils.NewTimeoutTCPConn(xgc, HuntdTimeout)

	var serverVersion uint32

	if !VERSION_AFTER_JOIN {
		// debian
		serverVersion, err = readServerVersion(tgc)
		if err != nil {
			tgc.Close()
			return err
		}
	}

	msgb := []byte(mstr)

	modesize := 4
	if DEBIAN_ULONG_BUG {
		modesize = 8
	}

	msg := make([]byte, 4+20+1+4+20+modesize+len(msgb))
	i := 0

	binary.BigEndian.PutUint32(msg[i:i+4], joinRequest.Uid)
	i += 4

	err = byteutils.StrToBytes(msg[i:i+20], joinRequest.Name, 20, 0)
	if err != nil {
		tgc.Close()
		return err
	}
	i += 20

	msg[i] = joinRequest.Team[0]
	i += 1

	binary.BigEndian.PutUint32(msg[i:i+4], joinRequest.EnterStatus)
	i += 4

	err = byteutils.StrToBytes(msg[i:i+20], joinRequest.Ttyname, 20, 0)
	if err != nil {
		tgc.Close()
		return err
	}
	i += 20

	binary.BigEndian.PutUint32(msg[i:i+4], joinRequest.ConnectMode)
	i += 4
	if DEBIAN_ULONG_BUG {
		binary.BigEndian.PutUint32(msg[i:i+4], 0xfeedface)
		i += 4
	}

	n = copy(msg[i:], msgb)
	i += len(msgb)

	if i != len(msg) {
		panic(fmt.Sprintf("packing error %d expected %d", i, len(msg)))
	}

	n, err = tgc.Write(msg)
	if err != nil {
		tgc.Close()
		return err
	}
	if n != len(msg) {
		tgc.Close()
		return fmt.Errorf("short write: wrote %d expected %d", n, len(msg))
	}

	if VERSION_AFTER_JOIN {
		// OSX (and probably Dragonfly BSD)
		serverVersion, err = readServerVersion(tgc)
		if err != nil {
			tgc.Close()
			return err
		}
	}

	p.gameConn = tgc
	p.serverVersion = serverVersion
	p.joinRequest = *joinRequest

	return nil
}

// Note: may block
func (p *Player) GameData() (error, []byte, error) {
	logger.Log(LOG_PLAYER_API, "Player %s: Request GameData", p.ID)

	buf := make([]byte, 2048)

	n, err := p.gameConn.Read(buf)
	logger.Log(LOG_PLAYER_API, "Player %s: Read GameData: n %d err %v", p.ID, n, err)
	if err != nil {
		nerr, isNetErr := err.(net.Error)
		if isNetErr && nerr.Timeout() {
			return err, nil, nil
		}
		return nil, nil, err
	}

	if n < 1 {
		return nil, nil, fmt.Errorf("short read: got %d expected >= 1", n)
	}

	return nil, buf[:n], nil
}

// Note: may block
func (p *Player) Input(keys string) (error, error) {
	logger.Log(LOG_PLAYER_API, "Player %s: Send Input {%v}", p.ID, keys)

	buf := []byte(keys)
	var err error

	var n int
	n, err = p.gameConn.Write(buf)
	if err != nil {
		nerr, isNetErr := err.(net.Error)
		if isNetErr && nerr.Timeout() {
			return err, nil
		}
		return nil, err
	}

	if n != len(buf) {
		return nil, fmt.Errorf("short write: wrote %d expected %d", n, len(buf))
	}

	return nil, nil
}

func (p *Player) Close() error {
	logger.Log(LOG_PLAYER_API, "Player %s: Close\n", p.ID) 

	if p.gameConn == nil {
		return nil
	}
	return p.gameConn.Close()
}

func (huntd *HuntDaemon) Stats(req *gamerpc.StatsRequest, reply *gamerpc.StatsReply) error {
	logger.Log(LOG_RPC, "Contacting huntd stats @ %s\n", huntd.statsAddr)

	c, err := net.DialTCP("tcp", nil, huntd.statsAddr)
	if err != nil {
		return err
	}
	defer c.Close()

	stats, err := ioutil.ReadAll(c)
	if err != nil {
		return err
	}

	reply.Token = req.Token
	reply.Stats = string(stats)

	return nil
}

func (huntd *HuntDaemon) JStats(r *http.Request, req *gamerpc.StatsRequest, reply *gamerpc.StatsReply) error {
	return huntd.Stats(req, reply)
}

func (huntd *HuntDaemon) Message(req *gamerpc.MessageRequest, reply *gamerpc.MessageReply) error {
	logger.Log(LOG_RPC, "Message %s\n", req.Message)

	player, err := huntd.newPlayer()
	if err != nil {
		return err
	}
	defer player.Close()

	err = player.Join(&req.Join, req.Message)
	if err != nil {
		return err
	}

	reply.Token = req.Token

	return nil
}

func (huntd *HuntDaemon) JMessage(r *http.Request, req *gamerpc.MessageRequest, reply *gamerpc.MessageReply) error {
	return huntd.Message(req, reply)
}

func (huntd *HuntDaemon) Join(req *gamerpc.JoinRequest, reply *gamerpc.JoinReply) error {
	logger.Log(LOG_RPC, "Join %s\n", req.Name)

	player, err := huntd.newPlayer()
	if err != nil {
		return err
	}

	err = player.Join(req, "")
	if err != nil {
		return err
	}

	huntd.Players[player.ID] = player

	reply.Token = req.Token
	reply.PlayerID = player.ID

	return nil
}

func (huntd *HuntDaemon) JJoin(r *http.Request, req *gamerpc.JoinRequest, reply *gamerpc.JoinReply) error {
	return huntd.Join(req, reply)
}

func (huntd *HuntDaemon) Quit(req *gamerpc.QuitRequest, reply *gamerpc.QuitReply) error {
	logger.Log(LOG_RPC, "Quit %s\n", req.PlayerID)

	player, found := huntd.Players[req.PlayerID]
	if found {
		delete(huntd.Players, req.PlayerID)
		player.Close()
	}

	reply.Token = req.Token

	return nil
}

func (huntd *HuntDaemon) JQuit(r *http.Request, req *gamerpc.QuitRequest, reply *gamerpc.QuitReply) error {
	return huntd.Quit(req, reply)
}

func (huntd *HuntDaemon) GameData(req *gamerpc.GameDataRequest, reply *gamerpc.GameDataReply) error {
	logger.Log(LOG_RPC, "GameData %s\n", req.PlayerID)

	player, err := huntd.player(req.PlayerID)
	if err != nil {
		return err
	}

	var timeoutErr error
	var data []byte
	timeoutErr, data, err = player.GameData()
	if err != nil {
		return err
	}

	if timeoutErr != nil {
		reply.Timeout = true
		reply.TimeoutError = timeoutErr.Error()
	}

	//
	// (yuck) we repack the 8 bit data into uint32 values because
	// otherwise the handling in javascript out on the client side
	// becomes incredibly complex
	//
	reply.Data = make([]uint32, len(data))
	for i, v := range data {
		reply.Data[i] = uint32(v)
	}
		
	reply.Token = req.Token

	return nil
}

func (huntd *HuntDaemon) JGameData(r *http.Request, req *gamerpc.GameDataRequest, reply *gamerpc.GameDataReply) error {
	return huntd.GameData(req, reply)
}

func (huntd *HuntDaemon) Input(req *gamerpc.InputRequest, reply *gamerpc.InputReply) error {
	logger.Log(LOG_RPC, "Input %s Keys %s\n", req.PlayerID, req.Keys)

	player, err := huntd.player(req.PlayerID)
	if err != nil {
		return err
	}

	var timeoutErr error
	timeoutErr, err = player.Input(req.Keys)
	if err != nil {
		return err
	}

	reply.Timeout = false
	if timeoutErr != nil {
		reply.Timeout = true
		reply.TimeoutError = timeoutErr.Error()
	}

	reply.Token = req.Token

	return nil
}

func (huntd *HuntDaemon) JInput(r *http.Request, req *gamerpc.InputRequest, reply *gamerpc.InputReply) error {
	return huntd.Input(req, reply)
}

func (huntd *HuntDaemon) Ping(req *gamerpc.PingRequest, reply *gamerpc.PingReply) error {
	logger.Log(LOG_RPC, "Ping Token %d Seq %d", req.Token, req.Seq)

	reply.Token = req.Token
	reply.Seq = req.Seq
	return nil
}

func (huntd *HuntDaemon) JPing(r *http.Request, req *gamerpc.PingRequest, reply *gamerpc.PingReply) error {
	return huntd.Ping(req, reply)
}

// it's ok for port to be empty, which simply strips it off
func ReplacePort(addr string, port string) string {
	if port != "" {
		port = ":" + port
	}

	i := strings.LastIndex(addr, ":")
	if i < 0 {
		return addr + port
	}

	return addr[:i] + port
}

type KeepAlive struct {
	Topic		*pubsub.Topic
	URL		*url.URL
	Hostname	string
	Instance	string
	MsgData		[]byte
}

func (keepAlive *KeepAlive) KeepAlive(seq uint64) error {
	msg := &pubsub.Message{
		Attributes: map[string]string{
			"hostname":	keepAlive.Hostname,
			"instance":	keepAlive.Instance,
			"seq":		fmt.Sprintf("%d", seq),
		},
		Data: keepAlive.MsgData,
	}

	msgIDs, err := keepAlive.Topic.Publish(context.Background(), msg)
	if err != nil {
		return err
	}

	logger.Log(LOG_KEEPALIVE, "Published seqid %d msgid %v", seq, msgIDs)

	return nil
}

func NewKeepAlive(topic string, gameURL string) (*KeepAlive, error) {
	if topic == "" {
		return nil, fmt.Errorf("missing keepalive topic") 
	}

	if gameURL == "" {
		return nil, fmt.Errorf("missing gameurl")
	}

	instance, err := apputils.GAEBackendInstance()
	if err != nil {
		return nil, err
	}

	ustr := strings.Replace(gameURL, "{{instance}}", instance, -1)
	u, err := url.Parse(ustr)
	if err != nil {
		return nil, err
	}

	project, err := apputils.ProjectID()
	if err != nil {
		return nil, err
	}

	hostname, err := metadata.Hostname()
	if err != nil {
		return nil, err
	}

	client, err := pubsub.NewClient(context.Background(), project)
	if err != nil {
		return nil, err
	}

	t, err := client.CreateTopic(context.Background(), topic)
	if err != nil {
		s := fmt.Sprintf("%v", err)
		if !strings.Contains(s, "Resource already exists in the project") {
			return nil, err
		}

		t = client.Topic(topic)
	}

	keepalive :=  &KeepAlive{
		Topic:		t,
		URL:		u,
		Hostname:	hostname,
		Instance:	instance,
		MsgData:	[]byte(u.String()),
	}

	return keepalive, nil
}

func main() {
	var listenHost string
	var listenPort string
	var huntdHost string
	var huntdPort string
	var rpcType string
	var err error
	var huntd *HuntDaemon
	var server *gamerpc.GameServer
	eventc := make(chan interface{})

	flag.StringVar(&listenHost, "server-host", "", "host to receive frontend rpcs on")
	flag.StringVar(&listenPort, "server-port", "", "port to receive frontend rpcs on")
	flag.StringVar(&huntdHost,  "huntd-well-known-host", "localhost", "'well known' hostname/address huntd listens on")
	flag.StringVar(&huntdPort,  "huntd-well-known-port", "", "UDP 'well known' port huntd listens on")
	flag.StringVar(&rpcType,    "rpc-type", "netrpc", "'netrpc' for golang stdlib or 'jsonrpc' for jsonrpc")

	flag.Parse()

	if listenHost == "" {
		logger.Fatalf("-server-host required")
	}

	if listenPort == "" {
		logger.Fatalf("-server-port required")
	}

	if huntdPort == "" {
		logger.Fatalf("-huntd-wellknown-port required")
	}

	ShowBugs()

	huntd, err = NewHuntDaemon(huntdHost, huntdPort)
	if err != nil {
		logger.Fatalf("NewHuntDaemon: %v", err)
	}

	logger.Log(LOG_STARTUP, "huntd: %v\n", huntd)

	server, err = gamerpc.NewGameServer(listenHost, listenPort, rpcType, huntd, KeepAliveTimeout, eventc)
	if err != nil {
		logger.Fatalf("NewGameServer: %v", err)
	}
	logger.Log(LOG_STARTUP, "server: %v\n", server)

	var keepalive *KeepAlive
	topic := os.Getenv("SERVER_KEEPALIVE_TOPIC");
	if topic != "" {
		keepalive, err = NewKeepAlive(os.Getenv("SERVER_KEEPALIVE_TOPIC"), os.Getenv("SERVER_GAME_URL"))
		if err != nil {
			logger.Fatalf("NewKeepAlive: %v", err)
		}
	}
	logger.Log(LOG_STARTUP, "keepalive: %v\n", keepalive)

	for {
		event := <- eventc
		logger.Log(LOG_EVENT, "event %T %v\n", event, event)

		if event == nil {
			logger.Fatalf("nil event!")
		}

		switch t := event.(type) {
		default:
			logger.Log(LOG_EVENT, "unhandled event type %T", t)
		case *gamerpc.KeepaliveRequest:
			if keepalive == nil {
				logger.Log(LOG_KEEPALIVE, "KeepaliveRequest ignored: %v", t)
				break
			}
			err := keepalive.KeepAlive(t.Seq)
			if err != nil {
				logger.Log(LOG_KEEPALIVE, "Keepalive failed: %v", err)
			}
		}
	}
}
