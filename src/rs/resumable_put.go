package rs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pojiang20/distribute-object-storage/src/object_stream"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"io"
	"log"
	"net/http"
)

// 封装对象的相关信息
type resumableToken struct {
	ObjectName string //对象名
	Size       int64  //总数据大小
	Hash       string
	Servers    []string //数据分片存储的数据服务节点
	UUIDS      []string //每个节点上对应的临时对象的uuid
}

type RSResumablePutStream struct {
	*RSPutStream
	*resumableToken
}

// 反序列化token，创建对象
func NewRSResumablePutStream(dataServers []string, name string, size int64, hash string) (*RSResumablePutStream, error) {
	//获取对象临时数据的各个写入流
	putStream, err := NewRSPutStream(dataServers, hash, size)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	// 获取各个数据节点上的uuid，用于创建token对象
	uuids := make([]string, ALL_SHARDS)
	for i := 0; i < ALL_SHARDS; i++ {
		uuids[i] = putStream.writers[i].(*object_stream.TempPutStream).UUID
	}
	token := &resumableToken{
		ObjectName: name,
		Size:       size,
		Hash:       hash,
		Servers:    dataServers,
		UUIDS:      uuids,
	}
	return &RSResumablePutStream{
		RSPutStream:    putStream,
		resumableToken: token,
	}, nil
}

func (s *RSResumablePutStream) ToToken() string {
	b, _ := json.Marshal(s)
	return base64.StdEncoding.EncodeToString(b)
}

// 向数据服务节点发送head请求，获取已上传数据的大小
func (s *RSResumablePutStream) CurrentSize() int64 {
	// 向dataServer发送head请求，获取上传的临时对象大小
	resp, err := http.Head(fmt.Sprintf("http://%s/temp/%s", s.Servers[0], s.UUIDS[0]))
	if err != nil {
		log.Println(err)
		return NOT_FOUND
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("Error: http request [method: HEAD] status code:", resp.StatusCode)
		return NOT_FOUND
	}
	size := utils.GetSizeFromHeader(resp.Header) * DATA_SHARDS
	//TODO 什么情况会大于？为什么可以直接返回对象大小
	if size > s.Size {
		size = s.Size
	}
	return size
}

func NewRSResumablePutStreamFromToken(token string) (*RSResumablePutStream, error) {
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var info resumableToken
	err = json.Unmarshal(b, &info)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//根据信息初始化编码器
	writers := make([]io.Writer, ALL_SHARDS)
	for i := 0; i < ALL_SHARDS; i++ {
		writers[i] = &object_stream.TempPutStream{
			Server: info.Servers[i],
			UUID:   info.UUIDS[i],
		}
	}
	enc := NewRSEncoder(writers)
	return &RSResumablePutStream{
		RSPutStream:    &RSPutStream{enc},
		resumableToken: &info,
	}, nil
}
