package locate

import (
	"encoding/json"
	"github.com/pojiang20/distribute-object-storage/src/rabbitmq"
	"github.com/pojiang20/distribute-object-storage/src/rs"
	"github.com/pojiang20/distribute-object-storage/src/types"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"net/http"
	"os"
	"time"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	if m != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	info := Locate(utils.GetObjectName(r.URL.EscapedPath()))
	if len(info) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	b, _ := json.Marshal(info)
	w.Write(b)
}

func Exist(name string) bool {
	//RS码规则：当获取的分片大于等于数据分片的值时，则可以进行数据恢复。这种情况认为可以定位到该数据
	return len(Locate(name)) >= rs.DATA_SHARDS
}

func Locate(name string) (locateInfo map[int]string) {
	mq := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	//向交换机发布消息
	mq.Publish("dataServers", name)
	c := mq.Consume()

	//TODO 这个是做什么的？publish()后，设置超时关闭连接，以判断资源是否存在
	go func() {
		time.Sleep(time.Second)
		mq.Close()
	}()
	locateInfo = make(map[int]string)
	for i := 0; i < rs.ALL_SHARDS; i++ {
		msg := <-c
		if len(msg.Body) == 0 {
			return
		}
		var info types.LocateMessage
		json.Unmarshal(msg.Body, &info)
		locateInfo[info.ID] = info.Addr
	}
	return
}
