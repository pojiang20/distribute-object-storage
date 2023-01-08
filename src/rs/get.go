package rs

import (
	"fmt"
	"github.com/pojiang20/distribute-object-storage/src/object_stream"
	"io"
	"log"
)

type RSGetStream struct {
	*rsDecoder
}

func NewRSGetStream(locateInfo map[int]string, dataServers []string, hash string, size int64) (*RSGetStream, error) {
	if len(locateInfo)+len(dataServers) != ALL_SHARDS {
		return nil, fmt.Errorf("Error: dataServer number %d + %d is not equal to %d\n", len(dataServers), len(locateInfo), ALL_SHARDS)
	}
	//创建对象分片的读对象
	readers := make([]io.Reader, ALL_SHARDS)
	for i := 0; i < ALL_SHARDS; i++ {
		server := locateInfo[i]
		//如果数据服务节点为空串，说明该编号对应的对象分片需要修复，因此给该编号分配一个可用的随机服务节点，存储修复的分片
		if server == "" {
			locateInfo[i] = dataServers[0]
			dataServers = dataServers[1:]
			//否则，创建分片编号对应的读对象
		} else {
			reader, err := object_stream.NewGetStream(server, fmt.Sprintf("%s.%d", hash, i))
			if err == nil {
				readers[i] = reader
			}
		}
	}
	//为对应编号缺失的分片创建写对象，以便修复后写入数据服务节点
	writers := make([]io.Writer, ALL_SHARDS)
	perShard := (size + DATA_SHARDS - 1) / DATA_SHARDS
	var err error
	for i := 0; i < ALL_SHARDS; i++ {
		if readers[i] == nil {
			writers[i], err = object_stream.NewTempPutStream(locateInfo[i], fmt.Sprintf("%s.%d", hash, i), perShard)
			if err != nil {
				return nil, err
			}
		}
	}
	dec := NewDecoder(readers, writers, size)
	return &RSGetStream{dec}, nil
}

// 将修复的数据写入数据节点中
func (s *RSGetStream) Close() {
	for i := 0; i < len(s.writers); i++ {
		if s.writers[i] != nil {
			s.writers[i].(*object_stream.TempPutStream).Commit(true)
		}
	}
}

// whence:何处
// 从whence（何处）开始要跳过offset字节
func (s *RSGetStream) Seek(offset int64, whence int) (int64, error) {
	//只支持从当前位置起跳
	if whence != io.SeekCurrent {
		log.Fatalln("Fatal: only support SeekCurrent")
	}
	//负数不能跳
	if offset < 0 {
		log.Fatalln("Fatal: offset should >=0")
	}
	//每次读取BLOCK_SIZE大小的内容并丢弃
	for offset != 0 {
		length := int64(BLOCK_SIZE)
		if length > offset {
			length = offset
		}
		buff := make([]byte, length)
		io.ReadFull(s, buff)
		offset -= length
	}
	return offset, nil
}
