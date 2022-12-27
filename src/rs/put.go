package rs

import (
	"fmt"
	"github.com/pojiang20/distribute-object-storage/src/object_stream"
	"io"
)

// 组合rsEncoder，在此基础上实现将对象分片数据写入到数据节点中
type RSPutStream struct {
	*rsEncoder
}

// 封装TempPutStream，根据提供的数据节点IP:PORT将对象的分片数据信息保存在节点的缓存中，等待上传分片数据时与时机的分片数据进行哈希校验
// 写入流的文件格式为：["对象哈希"+"."+"分片编号"]
func NewRSPutStream(dataServers []string, objectHash string, objectSize int64) (rsPutStream *RSPutStream, err error) {
	if len(dataServers) != ALL_SHARDS {
		return nil, fmt.Errorf("Error: dataServer number is %d not equal %d\n", len(dataServers), ALL_SHARDS)
	}
	//向上取整，计算出每个分片的大小
	perShardSize := (objectSize + DATA_SHARDS - 1) / DATA_SHARDS
	//创建分片信息写入流
	writers := make([]io.Writer, ALL_SHARDS)
	for i := 0; i < len(writers); i++ {
		writers[i], err = object_stream.NewTempPutStream(dataServers[i], fmt.Sprintf("%s.%d", objectHash, i), perShardSize)
		if err != nil {
			return nil, err
		}
	}
	rsEnc := NewRSEncoder(writers)
	rsPutStream = &RSPutStream{rsEnc}
	return rsPutStream, nil
}

// 转正判断
// temp下的数据文件的名字格式由["对象哈希"+"."+"分片ID"]，变为["对象哈希"+"."+"分片ID"+"."+"分片哈希"]
func (rsPutStream *RSPutStream) Commit(positive bool) {
	rsPutStream.Flush()
	for i := 0; i < len(rsPutStream.writers); i++ {
		rsPutStream.writers[i].(*object_stream.TempPutStream).Commit(positive)
	}
}
