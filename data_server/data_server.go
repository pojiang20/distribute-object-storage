package main

import (
	"github.com/pojiang20/distribute-object-storage/data_server/heartbeat"
	"github.com/pojiang20/distribute-object-storage/data_server/locate"
	"github.com/pojiang20/distribute-object-storage/data_server/objects"
	"github.com/pojiang20/distribute-object-storage/data_server/temp"
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	go heartbeat.StartHeartbeat()
	go locate.ListenLocate()
	go temp.CleanEvery12hour()
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/temp/", temp.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}
