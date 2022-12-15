package main

import (
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
	http.HandleFunc(ObjectsPattern, objects.Handler)
	log.Print(http.ListenAndServe(os.Getenv(listenAddress), nil))
}
