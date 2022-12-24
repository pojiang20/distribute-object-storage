package temp

import (
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"net/http"
	"os"
	"path"
)

func del(w http.ResponseWriter, r *http.Request) {
	uuid := utils.GetObjectName(r.URL.EscapedPath())
	infoFile := path.Join(os.Getenv("STORAGE_ROOT"), "temp", uuid)
	datFile := infoFile + ".dat"
	os.Remove(infoFile)
	os.Remove(datFile)
}
