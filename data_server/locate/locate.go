package locate

import (
	"github.com/pojiang20/distribute-object-storage/src/err_utils"
	"github.com/pojiang20/distribute-object-storage/src/rabbitmq"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
)

var (
	objects = make(map[string]struct{})
	lock    sync.Mutex
)

func ObjectExists(hash string) bool {
	lock.Lock()
	_, ok := objects[hash]
	lock.Unlock()
	return ok
}

func Add(hash string) {
	lock.Lock()
	objects[hash] = struct{}{}
	lock.Unlock()
}

func Del(hash string) {
	lock.Lock()
	delete(objects, hash)
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
			mq.Send(msg.ReplyTo, os.Getenv("LISTEN_ADDRESS"))
		}
	}
}

// 扫描节点上已有的对象文件，载入内存中
func CollectObjects() {
	files, _ := filepath.Glob(path.Join(os.Getenv("STORAGE_ROOT"), "objects/*"))
	for i := range files {
		hash := filepath.Base(files[i])
		objects[hash] = struct{}{}
	}
}
