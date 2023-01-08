package temp

import (
	"fmt"
	"github.com/pojiang20/distribute-object-storage/src/rs"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"log"
	"net/http"
)

// 获取已经上传过的数据大小
func head(w http.ResponseWriter, r *http.Request) {
	token := utils.GetObjectName(r.URL.EscapedPath())
	stream, err := rs.NewRSResumablePutStreamFromToken(token)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	uploadedSize := stream.CurrentSize()
	if uploadedSize < 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("content-length", fmt.Sprintf("%d", uploadedSize))
}
