package main

import (
	"github.com/pojiang20/distribute-object-storage/data_server/heartbeat"
	"github.com/pojiang20/distribute-object-storage/data_server/locate"
	"github.com/pojiang20/distribute-object-storage/objects"
	"log"
	"net/http"
	"os"
)

const (
	listenAddress  = "LISTEN_ADDRESS"
	ObjectsPattern = "/objects/"
)

func main() {
	go heartbeat.StartHeartbeat()
	go locate.ListenLocate()
	http.HandleFunc(ObjectsPattern, objects.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv(listenAddress), nil))
}
