package temp

import (
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"io"
	"log"
	"net/http"
	"os"
)

func get(w http.ResponseWriter, r *http.Request) {
	uuid := utils.GetObjectName(r.URL.EscapedPath())
	f, err := os.Open(os.Getenv("STORAGE_ROOT") + "/temp/" + uuid + ".dat")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}
