package heartbeat

import (
	"log"
	"testing"
	"time"
)

// 心跳用ticker更加合适，下面是对比示例
func Test_sleep(t *testing.T) {
	log.Println("-----time.Sleep(2s)-----")
	for i := 0; i < 5; i++ {
		log.Println(time.Now())
		do()
		time.Sleep(2 * time.Second)
	}

	log.Println("-----ticker(2s)-----")
	ticker := time.NewTicker(2 * time.Second)
	for i := 0; i < 5; i++ {
		log.Println(time.Now())
		do()
		<-ticker.C
	}
}
func do() {
	time.Sleep(time.Second)
}
