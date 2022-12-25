package locate

import (
	"encoding/json"
	"github.com/pojiang20/distribute-object-storage/src/err_utils"
	"github.com/pojiang20/distribute-object-storage/src/rabbitmq"
	"github.com/pojiang20/distribute-object-storage/src/utils"
	"log"
	"net/http"
	"os"
	"strconv"
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
	return Locate(name) != ""
}

func Locate(name string) string {
	mq := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	//向交换机发布消息
	mq.Publish("dataServers", name)
	c := mq.Consume()

	//TODO 这个是做什么的？publish()后，设置超时关闭连接，以判断资源是否存在
	go func() {
		time.Sleep(time.Second)
		mq.Close()
	}()
	msg := <-c
	res, err := strconv.Unquote(string(msg.Body))
	err_utils.Panic_NonNilErr(err)
	log.Printf("INFO: object at server '%s'\n", res)
	return res
}
