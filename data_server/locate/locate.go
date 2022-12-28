package locate

import (
	"github.com/pojiang20/distribute-object-storage/src/err_utils"
	"github.com/pojiang20/distribute-object-storage/src/rabbitmq"
	"github.com/pojiang20/distribute-object-storage/src/types"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	objectMap = make(map[string]int)
	lock      sync.Mutex
)

const (
	//分片不在该服务节点上
	NOT_FOUND = -1
	//"分片所属的对象哈希"+"."+"当前分片的编号"+"."+"当前分片数据的哈希"
	SHARD_NAME_COMPONENT_NUM = 3
)

func init() {
	CollectObjects()
}

func ObjectExists(hash string) bool {
	return objectMap[hash] != NOT_FOUND
}

func Add(hash string, id int) {
	lock.Lock()
	objectMap[hash] = id
	lock.Unlock()
}

func Del(hash string) {
	lock.Lock()
	delete(objectMap, hash)
	lock.Unlock()
}

func ListenLocate() {
	mq := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	defer mq.Close()
	//绑定名为dataServers的交换器
	mq.BindExchange("dataServers")
	//获取相应的channel
	c := mq.Consume()

	for msg := range c {
		hash, err := strconv.Unquote(string(msg.Body))
		err_utils.Panic_NonNilErr(err)
		if ObjectExists(hash) {
			mq.Send(msg.ReplyTo, types.LocateMessage{
				Addr: os.Getenv("LISTEN_ADDRESS"),
				ID:   objectMap[hash],
			})
		}
	}
}

// 扫描节点上已有的对象文件，载入内存中
func CollectObjects() {
	files, _ := filepath.Glob(path.Join(os.Getenv("STORAGE_ROOT"), "objects", "*"))
	for i := range files {
		shardNameComponents := strings.Split(filepath.Base(files[i]), ".")
		if len(shardNameComponents) != SHARD_NAME_COMPONENT_NUM {
			log.Fatalf("Error: shard %v name is invalid, it should be 3 compoments [objectHash.ID.shardHash]\n", shardNameComponents)
		}
		objectHash := shardNameComponents[0]
		shardID, err := strconv.Atoi(shardNameComponents[1])
		err_utils.Panic_NonNilErr(err)
		objectMap[objectHash] = shardID
	}
}
