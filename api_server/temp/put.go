package temp

import (
	"github.com/pojiang20/distribute-object-storage/api_server/locate"
	"github.com/pojiang20/distribute-object-storage/src/es"
	"github.com/pojiang20/distribute-object-storage/src/rs"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"io"
	"log"
	"net/http"
	"net/url"
)

func put(w http.ResponseWriter, r *http.Request) {
	token := utils.GetObjectName(r.URL.EscapedPath())
	//解析token构建对象
	stream, err := rs.NewRSResumablePutStreamFromToken(token)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// 校验请求中的offset与实际数据节点的大小是否一致
	offset, _ := utils.GetOffsetFromHeader(r.Header)
	uploadedSize := stream.CurrentSize()
	if uploadedSize < 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if uploadedSize != offset {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	buff := make([]byte, rs.BLOCK_SIZE)
	for {
		readSize, err := io.ReadFull(r.Body, buff)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		uploadedSize = uploadedSize + int64(readSize)
		if uploadedSize > stream.Size {
			stream.Commit(false)
			log.Println("apiServer Error: the object data to be uploaded is mismatch with the uploaded data")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if readSize != rs.BLOCK_SIZE && uploadedSize < stream.Size {
			return
		}
		stream.Write(buff)
		if uploadedSize == stream.Size {
			stream.Flush()
			getStream, err := rs.NewRSResumableGetStream(stream.Servers, stream.UUIDS, stream.Size)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			actualHash := url.PathEscape(utils.CalculateHash(getStream))
			if actualHash != stream.Hash {
				log.Println("apiServer Error: the actual uploaded data's hash is mismatch ")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if locate.Exist(stream.Hash) {
				stream.Commit(false)
			} else {
				stream.Commit(true)
			}
			//在元数据服务中添加对象元数据
			err = es.AddVersion(stream.ObjectName, stream.Size, stream.Hash)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}
}
