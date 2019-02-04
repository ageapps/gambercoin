package http_server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"github.com/ageapps/gambercoin/pkg/utils"
	"github.com/google/uuid"
)

// Health message
func Health(w http.ResponseWriter, r *http.Request) {
	// A very simple health check.
	status := "HEALTHY"
	send(&w, &status)
}

// GetMessages func
func GetMessages(w http.ResponseWriter, r *http.Request) {
	name, ok := getNameFromRequest(r)
	if !ok {
		sendError(&w, errors.New("Error: no peer requested for messages"))
		return
	}
	send(&w, getNodeMessages(name))
}

// GetPrivateMessages func
func GetPrivateMessages(w http.ResponseWriter, r *http.Request) {
	name, ok := getNameFromRequest(r)
	if !ok {
		sendError(&w, errors.New("Error: no peer requested for private messages"))
		return
	}
	send(&w, getNodePrivateMessages(name))
}

// GetRoutes func
func GetRoutes(w http.ResponseWriter, r *http.Request) {
	name, ok := getNameFromRequest(r)
	if !ok {
		sendError(&w, errors.New("Error: no peer requested for routes"))
		return
	}
	send(&w, getNodeRoutes(name))
}

// GetNodes func
func GetNodes(w http.ResponseWriter, r *http.Request) {
	name, ok := getNameFromRequest(r)
	if !ok {
		sendError(&w, errors.New("Error: no peer requested for nodes"))
		return
	}
	send(&w, getNodePeers(name))
}

// PostMessage func
func PostMessage(w http.ResponseWriter, r *http.Request) {
	params := *readBody(&w, r)
	name, ok := params["name"].(string)
	if !ok {
		sendError(&w, errors.New("Error: no peer requested for new message"))
		return
	}
	msg, ok := params["msg"].(string)
	if !ok || !sendMessage(name, msg) {
		sendError(&w, errors.New("Error while sending new message"))
		return
	}
	send(&w, getNodeMessages(name))
}

// PostPrivateMessage func
func PostPrivateMessage(w http.ResponseWriter, r *http.Request) {
	params := *readBody(&w, r)
	name, ok := params["name"].(string)
	if !ok {
		sendError(&w, errors.New("Error: no peer requested for private message"))
		return
	}
	dest, ok := params["destination"].(string)
	if !ok {
		sendError(&w, errors.New("Error: no destination requested"))
		return
	}
	msg, ok := params["msg"].(string)
	if !ok || !sendPrivateMessage(name, dest, msg) {
		sendError(&w, errors.New("Error while sending new message"))
		return
	}
	send(&w, getNodeMessages(name))
}

// PostNode func
func PostNode(w http.ResponseWriter, r *http.Request) {
	params := *readBody(&w, r)
	name, ok := params["name"].(string)
	if !ok {
		sendError(&w, errors.New("Error: no peer requested"))
		return
	}
	peer, ok := params["node"].(string)
	if !ok || !addPeer(name, peer) {
		sendError(&w, errors.New("Error while adding new peer"))
		return
	}
	send(&w, getNodePeers(name))
}

func PostTransaction(w http.ResponseWriter, r *http.Request) {
	params := *readBody(&w, r)
	name, ok := params["name"].(string)
	if !ok {
		sendError(&w, errors.New("Error: no name requested"))
		return
	}
	in, ok := params["in"].(string)
	if !ok {
		sendError(&w, errors.New("Error: no in requested"))
		return
	}
	out, ok := params["out"].(string)
	if !ok {
		sendError(&w, errors.New("Error: no out requested"))
		return
	}
	amount, ok := params["amount"].(int)
	if !ok {
		sendError(&w, errors.New("Error: no amount requested"))
		return
	}
	send(&w, sendTransaction(name, in, out, amount))
}

// GetID func
func GetID(w http.ResponseWriter, r *http.Request) {
	name, ok := getNameFromRequest(r)
	if !ok {
		sendError(&w, errors.New("Error: no peer requested"))
		return
	}

	send(&w, getStatusResponse(name))
}

// GetID func
func GetBalance(w http.ResponseWriter, r *http.Request) {
	name, ok := getNameFromRequest(r)
	if !ok {
		sendError(&w, errors.New("Error: no peer requested"))
		return
	}
	hashStr, ok := getHashFromRequest(r)
	if !ok {
		sendError(&w, errors.New("Error: no hash requested"))
		return
	}
	hash, err := utils.GetHash(hashStr)
	if err != nil {
		sendError(&w, errors.New("Error: bad hash conversion"))
		return
	}
	send(&w, geetHashBalance(name, hash))
}

// Delete node
func Delete(w http.ResponseWriter, r *http.Request) {
	params := *readBody(&w, r)
	name, ok := params["name"].(string)
	if !ok {
		sendError(&w, errors.New("Error: no peer requested"))
		return
	}
	deleteNode(name)
	sendOk(&w)
}

// Start node
func Start(w http.ResponseWriter, r *http.Request) {

	body := readBody(&w, r)
	if body == nil {
		sendError(&w, errors.New("Not enough parameters"))
		return
	}
	params := *body
	name, found := params["name"].(string)
	if !found {
		uuid, err := uuid.NewRandom()
		if err != nil {
			log.Fatal(err)
		}
		name = uuid.String()
	}

	var peers = utils.PeerAddresses{}
	var gossipAddr = utils.PeerAddress{}

	if address, ok := params["address"]; ok {
		if err := gossipAddr.Set(address.(string)); err != nil {
			sendError(&w, err)
			return
		}
	}
	if peerParams, ok := params["peers"]; ok {
		if err := peers.Set(peerParams.(string)); err != nil {
			sendError(&w, err)
			return
		}
	}
	nodeName := startNode(name, gossipAddr.String(), &peers)
	if nodeName == "" {
		sendError(&w, errors.New("Error starting node"))
		return
	}
	send(&w, getStatusResponse(nodeName))
}

func send(w *http.ResponseWriter, v interface{}) {
	if reflect.TypeOf(v) != reflect.TypeOf(int(0)) {
		if reflect.ValueOf(v).IsNil() {
			sendError(w, errors.New("Error sending response"))
			return
		}
	}
	(*w).Header().Set("Content-Type", "application/json; charset=UTF-8")
	(*w).WriteHeader(http.StatusOK)
	if err := json.NewEncoder(*w).Encode(v); err != nil {
		panic(err)
	}
}
func sendOk(w *http.ResponseWriter) {
	(*w).WriteHeader(http.StatusOK)
	fmt.Fprintf(*w, "OK")
}
func sendError(w *http.ResponseWriter, msg error) {
	(*w).WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(*w, "There was an error processing the request: %v\n", msg.Error())
	fmt.Printf("There was an error processing the request: %v\n", msg.Error())
}

func readBody(w *http.ResponseWriter, r *http.Request) *map[string]interface{} {
	var params map[string]interface{}
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		sendError(w, err)
		return nil
	}
	if err := r.Body.Close(); err != nil {
		sendError(w, err)
		return nil
	}
	if err := json.Unmarshal(body, &params); err != nil {
		sendError(w, err)
		return nil
	}
	return &params
}

func getNameFromRequest(r *http.Request) (string, bool) {
	name, ok := r.URL.Query()["name"]
	if !ok || len(name[0]) < 1 {
		return "", false
	}
	return name[0], true
}
func getHashFromRequest(r *http.Request) (string, bool) {
	name, ok := r.URL.Query()["hash"]
	if !ok || len(name[0]) < 1 {
		return "", false
	}
	return name[0], true
}
