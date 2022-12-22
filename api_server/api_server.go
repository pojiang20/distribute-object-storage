package main

import (
	"github.com/pojiang20/distribute-object-storage/api_server/heartbeat"
	"github.com/pojiang20/distribute-object-storage/api_server/locate"
	"github.com/pojiang20/distribute-object-storage/api_server/objects"
	"log"
	"net/http"
	"os"
)

func main() {
	go heartbeat.ListenHeartbeat()
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/locate/", locate.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}
