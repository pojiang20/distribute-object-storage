package rs

import (
	"github.com/klauspost/reedsolomon"
	"io"
)

// 先写入到cache中，当数据量为BLOCK_SIZE时再写入到指定位置
type rsEncoder struct {
	writers  []io.Writer //将对象分片数据写入到指定的储存位置
	rsEncode reedsolomon.Encoder
	cache    []byte //缓存待写入的数据，大小一般为BLOCK_SIZE，一次性可写入ALL_SHARDS个切片数据
}

func NewRSEncoder(writers []io.Writer) *rsEncoder {
	enc, _ := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	return &rsEncoder{
		writers:  writers,
		rsEncode: enc,
		cache:    make([]byte, 0),
	}
}

// 若p超过可支持的一次性写入的最大数据量时，需要分批次写入
// 如果p的量不足一次性写入的最大量，则会延迟到commit时才刷新写入对应的temp中
func (rsEnc *rsEncoder) Write(p []byte) (n int, err error) {
	dataLen := len(p)
	start := 0
	for dataLen != 0 {
		end := BLOCK_SIZE - len(rsEnc.cache)
		if end > dataLen {
			end = dataLen
		}
		rsEnc.cache = append(rsEnc.cache, p[start:end]...)
		//如果cache已经达到最大可写入数据量，则先flush然后再次读取
		if len(rsEnc.cache) == BLOCK_SIZE {
			rsEnc.Flush()
		}
		start += end
		dataLen -= end
	}
	return len(p), nil
}

// 写入缓存
func (rsEnc *rsEncoder) Flush() {
	if len(rsEnc.cache) == 0 {
		return
	}
	shards, _ := rsEnc.rsEncode.Split(rsEnc.cache)
	rsEnc.rsEncode.Encode(shards)
	for i := 0; i < len(shards); i++ {
		rsEnc.writers[i].Write(shards[i])
	}
	rsEnc.cache = []byte{}
}
