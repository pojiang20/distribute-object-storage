package rs

import (
	"github.com/klauspost/reedsolomon"
	"io"
)

type rsDecoder struct {
	//writers用于在读取数据的同时需要进行可能的数据修复
	readers   []io.Reader         //可正常读且数据完好，取对象分片的数据节点的文件读对象
	writers   []io.Writer         //不可正常读取或数据缺失的对象分片的数据节点的文件写对象
	rsEnc     reedsolomon.Encoder //对对象进行分片、编码、解码及数据的恢复
	size      int64               //对象数据的大小，数据分片中的实际数据量
	cache     []byte
	cacheSize int
	total     int64 //读取数据时计数
}

func NewDecoder(readers []io.Reader, writers []io.Writer, size int64) *rsDecoder {
	enc, _ := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	return &rsDecoder{
		readers:   readers,
		writers:   writers,
		rsEnc:     enc,
		size:      size,
		cache:     make([]byte, 0, BLOCK_PER_SHARD),
		cacheSize: 0,
		total:     0,
	}
}

func (rsdec *rsDecoder) Read(p []byte) (n int, err error) {
	if rsdec.cacheSize == 0 {
		err = rsdec.getData()
		if err != nil {
			return 0, err
		}
	}
	dataLen := len(p)
	if rsdec.cacheSize < dataLen {
		dataLen = rsdec.cacheSize
	}
	rsdec.cacheSize -= dataLen
	copy(p, rsdec.cache[:dataLen])
	rsdec.cache = rsdec.cache[dataLen:]
	return dataLen, nil
}

// 解码与修复
func (rsDec *rsDecoder) getData() error {
	if rsDec.total == rsDec.size {
		return io.EOF
	}

	shards := make([][]byte, ALL_SHARDS)
	repairIds := make([]int, 0)
	for i := 0; i < len(shards); i++ {
		if rsDec.readers[i] == nil {
			repairIds = append(repairIds, i)
		} else {
			shards[i] = make([]byte, BLOCK_PER_SHARD)
			readCount, err := io.ReadFull(rsDec.readers[i], shards[i])
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				shards[i] = nil
			} else if readCount != BLOCK_PER_SHARD {
				shards[i] = shards[i][:readCount]
			}
		}
	}
	//如果存在需要修复的分片，则进行分片修复
	if len(repairIds) > 0 {
		err := rsDec.rsEnc.Reconstruct(shards)
		if err != nil {
			return err
		}
		//将恢复的分片写入对应的服务节点
		for i := 0; i < len(repairIds); i++ {
			id := repairIds[i]
			rsDec.writers[id].Write(shards[i])
		}
	}
	//解码数据分片，还原数据
	for i := 0; i < DATA_SHARDS; i++ {
		shardSize := int64(len(shards[i]))
		//如果处理的是最后一块数据分片，存在填充数据，则只取实际数据
		if rsDec.total+shardSize > rsDec.size {
			shardSize -= rsDec.total + shardSize - rsDec.size
		}
		//将数据分片的数据存入缓存中，同时计算缓存数据总量
		rsDec.cache = append(rsDec.cache, shards[i][:shardSize]...)
		rsDec.cacheSize += int(shardSize)
		rsDec.total += shardSize
	}
	return nil
}
