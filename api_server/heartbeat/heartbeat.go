package heartbeat

import (
	"github.com/pojiang20/distribute-object-storage/src/err_utils"
	"github.com/pojiang20/distribute-object-storage/src/rabbitmq"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	dataServers = make(map[string]time.Time)
	mutex       sync.Mutex
)

func ListenHeartbeat() {
	mq := rabbitmq.New(os.Getenv("RABBITMQ_SERVER"))
	defer mq.Close()

	//绑定apiServers的exchange
	mq.BindExchange("apiServers")
	c := mq.Consume()

	//定期清除过期的服务
	go removeExpiredDataServer()
	for msg := range c {
		//接收来自数据服务的心跳，并将活跃的服务存储到dataServers中
		dataServer, err := strconv.Unquote(string(msg.Body))
		err_utils.Panic_NonNilErr(err)
		mutex.Lock()
		dataServers[dataServer] = time.Now()
		mutex.Unlock()
	}
}

func removeExpiredDataServer() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		<-ticker.C
		mutex.Lock()
		for server, time1 := range dataServers {
			if time1.Add(10 * time.Second).Before(time.Now()) {
				delete(dataServers, server)
			}
		}
		mutex.Unlock()
	}
}

func GetDataServers() []string {
	mutex.Lock()
	defer mutex.Unlock()
	servers := make([]string, 0, len(dataServers))
	for s := range dataServers {
		servers = append(servers, s)
	}
	return servers
}

func ChooseRandomDataServer() string {
	servers := GetDataServers()
	n := len(servers)
	if n == 0 {
		return ""
	}
	return servers[rand.Intn(n)]
}
