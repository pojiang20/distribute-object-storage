package objects

import (
	"fmt"
	"github.com/pojiang20/distribute-object-storage/api_server/heartbeat"
	"github.com/pojiang20/distribute-object-storage/api_server/locate"
	"github.com/pojiang20/distribute-object-storage/src/object_stream"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"io"
	"net/http"
	"net/url"
)

func storeObject(reader io.Reader, hash string, size int64) (int, error) {
	escapedHash := url.PathEscape(hash)
	//如果对象内容数据已经存在，则直接返回。否则暂存对象并校验
	if locate.Exist(escapedHash) {
		return http.StatusOK, nil
	}
	stream, err := putStream(escapedHash, size)
	if err != nil {
		return http.StatusServiceUnavailable, err
	}

	//读取的同时进行数据写入
	//stream会调用实现了的write方法进行数据写入
	r := io.TeeReader(reader, stream)
	actualHash := utils.CalculateHash(r)
	if actualHash != hash {
		stream.Coommit(false)
		err = fmt.Errorf("Error: object hash value is not match, actualHash=[%s], expectedHash=[%s]\n", actualHash, hash)
		return http.StatusBadRequest, err
	}
	stream.Coommit(true)
	return http.StatusOK, nil
}

func putStream(hash string, size int64) (*object_stream.TempPutStream, error) {
	server := heartbeat.ChooseRandomDataServer()
	if server == "" {
		return nil, fmt.Errorf("cannot find any dataServer")
	}
	return object_stream.NewTempPutStream(server, hash, size)
}

func getStream(object string) (io.Reader, error) {
	server := locate.Locate(object)
	if server == "" {
		return nil, fmt.Errorf("object %s locate fail", object)
	}
	return NewGetStream(server, object)
}
