package locate

import (
	"github.com/pojiang20/distribute-object-storage/src/err_utils"
	"github.com/pojiang20/distribute-object-storage/src/rabbitmq"
	"os"
	"strconv"
)

func ListenLocate() {
	mq := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	defer mq.Close()
	//绑定名为dataServers的交换器
	mq.BindExchange("dataServers")
	//获取相应的channel
	c := mq.Consume()

	for msg := range c {
		object, err := strconv.Unquote(string(msg.Body))
		err_utils.Panic_NonNilErr(err)
		//对象+对应的存储目录作为文件名
		//判断文件是否存在，存在则做出响应
		if pathExist(os.Getenv("STORAGE_ROOT") + "/objects/" + object) {
			mq.Send(msg.ReplyTo, os.Getenv("LISTEN_ADDRESS"))
		}
	}
}

func pathExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsExist(err)
}
