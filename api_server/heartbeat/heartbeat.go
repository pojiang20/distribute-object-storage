package heartbeat

import (
	"github.com/pojiang20/distribute-object-storage/src/err_utils"
	"github.com/pojiang20/distribute-object-storage/src/rabbitmq"
	"github.com/pojiang20/distribute-object-storage/src/rs"
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

// ChooseServers: 选取dataServersNum个数据服务节点用于存放分片数据，该函数有两种使用方式：
// 1. 第一次存储对象分片数据，则dataServersNum等于ALL_SHARDS，unbrokenShardServerMap为nil
// 2. 从可用的数据服务节点中排除对象分片数据正常的节点，获取可用于存储修复的分片数据的数据服务节点
func ChooseServers(dataServerNum int, unbrokenShardServerMap map[int]string) (dataServers []string) {
	// 所需的用于存储的分片的节点数与已存放正常分片数据的节点数之和应等于一个对象的分片数之和，否则应直接中断程序执行
	if dataServerNum+len(unbrokenShardServerMap) != rs.ALL_SHARDS {
		panic("apiServer Error: the sum of brokenShards number and unbrokenShards number is not equal to ALL_SHARDS\n")
	}
	// 用于存放分片数据的候选服务节点，该切片长度不小于dataServersNum时，才可以进行随机选择
	candidates := make([]string, 0)
	reverseUnbrokenShardServerMap := make(map[string]int)
	for id, addr := range unbrokenShardServerMap {
		reverseUnbrokenShardServerMap[addr] = id
	}
	// 获取可以作为分片存储节点的服务节点，实际上就是获取存储了需要修复分片数据的服务节点
	servers := GetDataServers()
	for i := range servers {
		if _, in := reverseUnbrokenShardServerMap[servers[i]]; !in {
			candidates = append(candidates, servers[i])
		}
	}
	//若是候选服务节点数小于所需的数据服务节点数在直接返回空，这说明没有足够的服务节点满足对象分片的存储需求
	if len(candidates) < dataServerNum {
		return
	}
	//打乱并随机选择所需的数目
	randomIds := rand.Perm(len(candidates))
	for i := 0; i < dataServerNum; i++ {
		dataServers = append(dataServers, candidates[randomIds[i]])
	}
	return
}
