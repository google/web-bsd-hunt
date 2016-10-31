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
package frontend

import(
	"log"
	"fmt"
	"os"
	"io"
	"time"
	"strings"
	"net/http"
	"encoding/json"

	"apputils"
	"gamerpc"

	"github.com/tadhunt/httputils"
	"github.com/gorilla/mux"
)

const(
	KeepAliveTimeout	= 30 * time.Second
)

var gameURLStr string
var rpcTypeStr string
var keepAliveTopic string
var rpOptions = apputils.RProxyOptions(os.Getenv("SERVER_FRONTEND_RPROXY_OPTIONS"))
var staticGameClient *gamerpc.GameClient

func NewGameHandler(handler func(game *gamerpc.GameClient, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		instance, found := vars["instance"]
		if !found {
			err := fmt.Errorf("missing instance: %s", r.URL.String())
			apputils.InternalServerError(w, r, "missing instance", err)
			return
		}

		urlstr := strings.Replace(gameURLStr, "{{instance}}", instance, -1)

		game, err := FindGameInstance(r, urlstr)
		if err != nil {
			err = fmt.Errorf("urlstr=%s %v", err)
			apputils.InternalServerError(w, r, "no such instance", err)
			return
		}

		handler(game, w, r)
	}
}

func setupHandlers() {
	r := mux.NewRouter()

	r.HandleFunc("/api/v1/keepalive",		keepaliveHandler)
	r.HandleFunc("/api/v1/instances",		instancesHandler)
	r.HandleFunc("/api/v1/stats",			allStatsHandler)
	r.HandleFunc("/api/v1/info/{instance}",		NewGameHandler(infoHandler))
	r.HandleFunc("/api/v1/join/{instance}",		NewGameHandler(joinHandler))
	r.HandleFunc("/api/v1/stats/{instance}",	NewGameHandler(statsHandler))
	r.HandleFunc("/api/v1/message/{instance}",	NewGameHandler(messageHandler))
	r.HandleFunc("/api/v1/quit/{instance}",		NewGameHandler(quitHandler))
	r.HandleFunc("/api/v1/gamedata/{instance}",	NewGameHandler(gameDataHandler))
	r.HandleFunc("/api/v1/input/{instance}",	NewGameHandler(inputHandler))
	r.HandleFunc("/api/v1/ping/{instance}",		NewGameHandler(pingHandler))

	http.Handle("/", r)
}

func init() {
	gameURLStr = os.Getenv("SERVER_GAME_URL")
	if gameURLStr == "" {
		log.Fatalf("environment variable SERVER_GAME_URL not set")
	}

	rpcTypeStr = os.Getenv("SERVER_GAME_RPC")
	if rpcTypeStr == "" {
		log.Fatalf("environment variable SERVER_GAME_RPC not set")
	}

	keepAliveTopic = os.Getenv("SERVER_KEEPALIVE_TOPIC")
	if rpcTypeStr == "" {
		log.Fatalf("SERVER_KEEPALIVE_TOPIC not set")
	}

	standalone := os.Getenv("SERVER_FRONTEND_STANDALONE") == "yes"

	if !strings.Contains(gameURLStr, "{{instance}}") {
		var err error
		staticGameClient, err = gamerpc.NewGameClient(gameURLStr, rpcTypeStr, rpOptions)
		if err != nil {
			log.Fatalf("failed to create static client: %v", err)
		}
	}

	log.Printf("Frontend PID:   %d", os.Getpid())
	log.Printf("Game Server:    %s", gameURLStr)
	log.Printf("Standalone:     %v", standalone)
	log.Printf("KeepAliveTopic: %s", keepAliveTopic)

	setupHandlers()

	if standalone {
		log.Fatalf("standalone mode no longer supported")
	}
}

/*
 * Side effect: may modify team to meet the backend huntd protocol requirements
 */
func CheckJoin(request *gamerpc.JoinRequest) error {
	if request.Uid == 0 {
		request.Uid = 31337
	}
	if request.Name == "" {
		return fmt.Errorf("missing Name")
	}

	switch request.Team {
	case "none":
		request.Team = " "
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		break
	default:
		return fmt.Errorf("bad Team")
	}

	switch request.EnterStatus {
	case gamerpc.Q_CLOAK:
	case gamerpc.Q_FLY:
	case gamerpc.Q_SCAN:
	default:
		return fmt.Errorf("bad EnterStatus")
	}

	if request.Ttyname == "" {
		request.Ttyname = "/dev/tty-web"
	}

	switch request.ConnectMode {
	case gamerpc.C_PLAYER:
	case gamerpc.C_MONITOR:
	case gamerpc.C_MESSAGE:
	default:
		return fmt.Errorf("bad ConnectMode")
	}

	return nil
}


func DecodeJoin(r io.Reader) (*gamerpc.JoinRequest, error) {
	dec := json.NewDecoder(r)

	var request gamerpc.JoinRequest
	err := dec.Decode(&request)
	if err != nil {
		return nil, err
	}

	err = CheckJoin(&request)
	if err != nil {
		return nil, err
	}

	return &request, nil
}

func joinHandler(game *gamerpc.GameClient, w http.ResponseWriter, r *http.Request) {
	var err error

	err = httputils.RequestAcceptsJSON(r)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, "client does not accept application/json", err)
		return
	}

	var request *gamerpc.JoinRequest
	request, err = DecodeJoin(r.Body)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, err.Error(), err)
		return
	}

	var reply *gamerpc.JoinReply
	reply, err = game.Join(r, request)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(reply)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}
}

func DecodeMessage(r io.Reader) (*gamerpc.MessageRequest, error) {
	dec := json.NewDecoder(r)

	var request *gamerpc.MessageRequest
	err := dec.Decode(&request)
	if err != nil {
		return nil, err
	}

	request.Join.ConnectMode = gamerpc.C_MESSAGE

	err = CheckJoin(&request.Join)
	if err != nil {
		return nil, err
	}

	if request.Message == "" {
		return nil, fmt.Errorf("missing Message")
	}

	return request, nil

}

func DecodeQuit(r io.Reader) (*gamerpc.QuitRequest, error) {
	dec := json.NewDecoder(r)

	var request *gamerpc.QuitRequest
	err := dec.Decode(&request)
	if err != nil {
		return nil, err
	}

	if request.PlayerID == "" {
		return nil, fmt.Errorf("missing PlayerID")
	}

	return request, nil
}

func quitHandler(game *gamerpc.GameClient, w http.ResponseWriter, r *http.Request) {
	err := httputils.RequestAcceptsJSON(r)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, "client does not accept application/json", err)
		return
	}
	err = gamerpc.ContentTypeIsJSON(r.Header)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, err.Error(), err)
		return
	}

	var request *gamerpc.QuitRequest
	request, err = DecodeQuit(r.Body)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, err.Error(), err)
		return
	}

	var reply *gamerpc.QuitReply
	reply, err = game.Quit(r, request)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(reply)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}
}

func DecodeInput(r io.Reader) (*gamerpc.InputRequest, error) {
	dec := json.NewDecoder(r)

	var request *gamerpc.InputRequest
	err := dec.Decode(&request)
	if err != nil {
		return nil, err
	}

	if request.PlayerID == "" {
		return nil, fmt.Errorf("missing PlayerID")
	}
	if request.Keys == "" {
		return nil, fmt.Errorf("missing Keys")
	}

	return request, nil
}

func inputHandler(game *gamerpc.GameClient, w http.ResponseWriter, r *http.Request) {
	err := httputils.RequestAcceptsJSON(r)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, "client does not accept application/json", err)
		return
	}
	err = gamerpc.ContentTypeIsJSON(r.Header)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, err.Error(), err)
		return
	}

	var request *gamerpc.InputRequest
	request, err = DecodeInput(r.Body)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, err.Error(), err)
		return
	}

	var reply *gamerpc.InputReply
	reply, err = game.Input(r, request)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(reply)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}
}

func DecodePing(r io.Reader) (*gamerpc.PingRequest, error) {
	dec := json.NewDecoder(r)

	var request *gamerpc.PingRequest
	err := dec.Decode(&request)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func pingHandler(game *gamerpc.GameClient, w http.ResponseWriter, r *http.Request) {
	err := httputils.RequestAcceptsJSON(r)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, "client does not accept application/json", err)
		return
	}
	err = gamerpc.ContentTypeIsJSON(r.Header)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, err.Error(), err)
		return
	}

	var request *gamerpc.PingRequest
	request, err = DecodePing(r.Body)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, err.Error(), err)
		return
	}

	var reply *gamerpc.PingReply
	reply, err = game.Ping(r, request)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(reply)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}
}

func messageHandler(game *gamerpc.GameClient, w http.ResponseWriter, r *http.Request) {
	err := httputils.RequestAcceptsJSON(r)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, "client does not accept application/json", err)
		return
	}
	err = gamerpc.ContentTypeIsJSON(r.Header)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, err.Error(), err)
		return
	}

	var request *gamerpc.MessageRequest
	request, err = DecodeMessage(r.Body)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, err.Error(), err)
		return
	}

	var reply *gamerpc.MessageReply
	reply, err = game.Message(r, request)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(reply)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}
}

func DecodeGameData(r io.Reader) (*gamerpc.GameDataRequest, error) {
	dec := json.NewDecoder(r)

	var request *gamerpc.GameDataRequest
	err := dec.Decode(&request)
	if err != nil {
		return nil, err
	}

	if request.PlayerID == "" {
		return nil, fmt.Errorf("missing PlayerID")
	}

	return request, nil
}

func gameDataHandler(game *gamerpc.GameClient, w http.ResponseWriter, r *http.Request) {
	err := httputils.RequestAcceptsJSON(r)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, "client does not accept application/json", err)
		return
	}
	err = gamerpc.ContentTypeIsJSON(r.Header)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, err.Error(), err)
		return
	}

	var request *gamerpc.GameDataRequest
	request, err = DecodeGameData(r.Body)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, err.Error(), err)
		return
	}

	var reply *gamerpc.GameDataReply
	reply, err = game.GameData(r, request)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}

	if reply.Timeout {
		apputils.Error(w, r, http.StatusRequestTimeout, "Timeout waiting for game data", fmt.Errorf("%s", reply.TimeoutError))
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(reply)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}
}

func DecodeKeepalive(r io.Reader) (*gamerpc.KeepaliveRequest, error) {
	dec := json.NewDecoder(r)

	var request *gamerpc.KeepaliveRequest
	err := dec.Decode(&request)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func processKeepAlives(w http.ResponseWriter, r *http.Request) (int, error) {
	apputils.LogRequestHeaders(r, "keepalive")

	keepalive, err := apputils.NewKeepAlive(r, keepAliveTopic, KeepAliveTimeout)
	if err != nil {
		return 0, err
	}

	n := 0

	err = keepalive.Pull(func (km *apputils.KeepAliveMessage) error {
		apputils.Log(r, fmt.Sprintf("processKeepAlives: Msg[%d]: %s", n, km.String()))

		_, err := UpdateGameInstance(r, km.Instance, km.URL.String())
		if err != nil {
			return fmt.Errorf("processKeepAlives: Msg[%d]: UpdateGameInstance: %s", n, err)
		}

		n++
		return nil
	})

	if err != nil {
		return n, err
	}

	return n, nil
}

func keepaliveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	n, err := processKeepAlives(w, r)
	if err != nil {
		fmt.Fprintf(w, "ERROR: processKeepAlives: %v\n", err)
	}
	fmt.Fprintf(w, "INFO: Processed %d keepalives\n", n)

	n, err = ReapGameInstances(r)
	if err != nil {
		fmt.Fprintf(w, "ERROR: ReapGameInstances: %v\n", err)
	}
	fmt.Fprintf(w, "INFO: Reaped %d instances\n", n)
}

type InstancesReply struct {
	InstanceIDs	[]string
}

func instancesHandler(w http.ResponseWriter, r *http.Request) {
	err := httputils.RequestAcceptsJSON(r)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, "client does not accept application/json", err)
		return
	}

	var instances []*GameInstance
	instances, err = GameInstances(r)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}

	var reply = &InstancesReply{}

	for _, instance := range instances {
		reply.InstanceIDs = append(reply.InstanceIDs, instance.InstanceID)
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(reply)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}
}

func statsHandler(game *gamerpc.GameClient, w http.ResponseWriter, r *http.Request) {
	err := httputils.RequestAcceptsJSON(r)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, "client does not accept application/json", err)
		return
	}

	request := &gamerpc.StatsRequest {
		Token:		123,
	}

	var reply *gamerpc.StatsReply
	reply, err = game.Stats(r, request)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(reply)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}
}

type InstanceStatsReply struct {
	InstanceID	string
	Stats		string
}

type AllStatsReply struct {
	AllStats	[]*InstanceStatsReply
}

func allStatsHandler(w http.ResponseWriter, r *http.Request) {
	err := httputils.RequestAcceptsJSON(r)
	if err != nil {
		apputils.Error(w, r, http.StatusBadRequest, "client does not accept application/json", err)
		return
	}

	var instances []*GameInstance

	instances, err = GameInstances(r)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}

	request := &gamerpc.StatsRequest {
		Token:		123,
	}

	reply := &AllStatsReply{}

	for _, instance := range instances {
		game, err := FindGameInstance(r, instance.URL)
		if err != nil {
			apputils.Log(r, fmt.Sprintf("allStatsHandler: Ignore error: FindGameInstance %s: %v", instance.URL, err))
			continue
		}

		var stats *gamerpc.StatsReply
		stats, err = game.Stats(r, request)
		if err != nil {
			apputils.Log(r, fmt.Sprintf("allStatsHandler: Ignore error: game.Stats %s: %v", instance.URL, err))
			continue
		}

		ireply := &InstanceStatsReply {
			InstanceID:	instance.InstanceID,
			Stats:		stats.Stats,
		}

		reply.AllStats = append(reply.AllStats, ireply)
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(reply)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}
}

func infoHandler(game *gamerpc.GameClient, w http.ResponseWriter, r *http.Request) {
	cfg := &apputils.RProxyConfig{
		Copy:		[]string{"Cookie", "Accept", "Content-Type"},
		Add:		[]*apputils.Header{
					&apputils.Header{Key: "REDACTED", Value: "REDACTED"},
				},
		Options:	rpOptions,
	}

	u := *game.URL
	u.Path = "/info"

	buf, err := apputils.RProxy(r, cfg, r.Method, &u, nil)
	if err != nil {
		apputils.InternalServerError(w, r, "request failed due to internal error", err)
		return
	}

	_, err = w.Write(buf)
	if err != nil {
		apputils.InternalServerError(w, r, err.Error(), err)
		return
	}
}
