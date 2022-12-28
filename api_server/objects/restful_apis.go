package objects

import (
	"github.com/pojiang20/distribute-object-storage/src/es"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

func put(w http.ResponseWriter, r *http.Request) {
	hash := utils.GetHashFromHeader(r.Header)
	if hash == "" {
		log.Println("missing object hash in digest header")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	size := utils.GetSizeFromHeader(r.Header)
	respCode, err := storeObject(r.Body, hash, size)
	if err != nil {
		log.Println(err)
		w.WriteHeader(respCode)
	}
	if respCode != http.StatusOK {
		w.WriteHeader(respCode)
		return
	}
	name := utils.GetObjectName(r.URL.EscapedPath())
	err = es.AddVersion(name, size, hash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	name := utils.GetObjectName(r.URL.EscapedPath())
	versionId := r.URL.Query().Get("version")
	version := 0
	var err error
	if len(versionId) != 0 {
		version, err = strconv.Atoi(versionId)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	meta, err := es.GetMetadata(name, version)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if meta.Hash == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	stream, err := GetStream(url.PathEscape(meta.Hash), meta.Size)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err = io.Copy(w, stream)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	stream.Close()
}

func del(w http.ResponseWriter, r *http.Request) {
	name := utils.GetObjectName(r.URL.EscapedPath())
	version, err := es.SearchLatestVersion(name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = es.PutMetadata(name, version.Version+1, 0, "")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
