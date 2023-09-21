package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

var authToken = os.Getenv("AUTH_TOKEN")
var egressAddr = &net.TCPAddr{IP: net.ParseIP(os.Getenv("EGRESS_ADDRESS"))}

type PingResponse struct {
	Reachable bool  `json:"reachable"`
	Ping      int64 `json:"ping"`
}

func isAuthorized(req *http.Request) bool {

	token, hasToken := req.Header["Authorization"]

	if !hasToken {
		return false
	}

	if len(token) != 1 {
		return false
	}

	return token[0] == authToken
}

func checkPing(ip, port string) PingResponse {
	dialer := net.Dialer{
		Timeout:   time.Second * 4,
		LocalAddr: egressAddr,
	}
	currentTime := time.Now()

	dial, err := dialer.Dial("tcp4", fmt.Sprintf("%s:%s", ip, port))
	if err != nil {
		return PingResponse{}
	}

	defer dial.Close()

	return PingResponse{
		Reachable: true,
		Ping:      time.Now().Sub(currentTime).Milliseconds(),
	}
}

func handlePing(writer http.ResponseWriter, req *http.Request) {
	if !isAuthorized(req) {
		writer.WriteHeader(403)
		return
	}

	query := req.URL.Query()

	ip, ipExists := query["ip"]
	port, portExists := query["port"]

	if !ipExists || !portExists || len(ip) != 1 || len(port) != 1 {
		writer.WriteHeader(400)
		return
	}

	result := checkPing(ip[0], port[0])
	data, err := json.Marshal(result)

	if err != nil {
		writer.WriteHeader(500)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, string(data))
}

func main() {
	http.HandleFunc("/ping", handlePing)

	if err := http.ListenAndServe("0.0.0.0:8040", nil); err != nil {
		fmt.Println(err)
	}
}
