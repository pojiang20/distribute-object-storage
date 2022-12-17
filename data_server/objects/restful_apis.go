package objects

import (
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	STORAGE_ROOT = "STORAGE_ROOT"
)

func put(w http.ResponseWriter, r *http.Request) {
	f, err := os.Create(os.Getenv(STORAGE_ROOT) + "/objects/" +
		utils.GetObjectName(r.URL.EscapedPath()))
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
		utils.GetObjectName(r.URL.EscapedPath()))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}
