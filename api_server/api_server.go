package main

import (
	"github.com/pojiang20/distribute-object-storage/api_server/heartbeat"
	"github.com/pojiang20/distribute-object-storage/api_server/locate"
	"github.com/pojiang20/distribute-object-storage/api_server/objects"
	"github.com/pojiang20/distribute-object-storage/api_server/versions"
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	go heartbeat.ListenHeartbeat()
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/locate/", locate.Handler)
	http.HandleFunc("/versions/", versions.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}
