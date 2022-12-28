package objects

import (
	"fmt"
	"github.com/pojiang20/distribute-object-storage/api_server/heartbeat"
	"github.com/pojiang20/distribute-object-storage/api_server/locate"
	"github.com/pojiang20/distribute-object-storage/src/rs"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"io"
	"log"
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
		stream.Commit(false)
		err = fmt.Errorf("Error: object hash value is not match, actualHash=[%s], expectedHash=[%s]\n", actualHash, hash)
		return http.StatusBadRequest, err
	}
	stream.Commit(true)
	return http.StatusOK, nil
}

func putStream(hash string, size int64) (*rs.RSPutStream, error) {
	servers := heartbeat.ChooseServers(rs.ALL_SHARDS, nil)
	if len(servers) != rs.ALL_SHARDS {
		return nil, fmt.Errorf("apiServer Error: cannot find enough dataServers\n")
	}
	log.Printf("apiServer INFO: Choose random data servers to save object %s: %v\n", hash, servers)
	return rs.NewRSPutStream(servers, hash, size)
}

func GetStream(objectName string, size int64) (*rs.RSGetStream, error) {
	locateInfo := locate.Locate(objectName)
	if len(locateInfo) < rs.DATA_SHARDS {
		return nil, fmt.Errorf("Error: object %s locate failed, the data shards located is not enough: %v\n",
			objectName, locateInfo)
	}
	//若是获取的对象分片不足，说明需要修复
	dataServers := make([]string, 0)
	if len(locateInfo) < rs.ALL_SHARDS {
		log.Printf("INFO: some of shards need to repair\n")
		dataServers = heartbeat.ChooseServers(rs.ALL_SHARDS-len(locateInfo), locateInfo)
	}
	return rs.NewRSGetStream(locateInfo, dataServers, objectName, size)
}
