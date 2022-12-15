package objects

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	STORAGE_ROOT = "STORAGE_ROOT"
)

func put(w http.ResponseWriter, r *http.Request) {
	f, err := os.Create(os.Getenv(STORAGE_ROOT) + "/objects/" +
		strings.Split(r.URL.EscapedPath(), "/")[2])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(f, r.Body)
}

func get(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(os.Getenv(STORAGE_ROOT) + "/objects/" +
		strings.Split(r.URL.EscapedPath(), "/")[2])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}
