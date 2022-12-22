# distribute-object-storage
Simple distributed object storage implemented by go

### 第一章 单机版对象存储
mkdir /tmp/objects

`LISTEN_ADDRESS=:12345 STORAGE_ROOT=/tmp go run server.go`

存对象`curl -v 127.0.0.1:12345/objects/test -XPUT  -d "this is a test object"`
```text
*   Trying 127.0.0.1:12345...
* Connected to 127.0.0.1 (127.0.0.1) port 12345 (#0)
> PUT /objects/test HTTP/1.1
> Host: 127.0.0.1:12345
> User-Agent: curl/7.77.0
> Accept: */*
> Content-Length: 21
> Content-Type: application/x-www-form-urlencoded
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Thu, 15 Dec 2022 10:06:43 GMT
< Content-Length: 0
< 
* Connection #0 to host 127.0.0.1 left intact

```
取对象`curl -v 127.0.0.1:12345/objects/test`
```text
*   Trying 127.0.0.1:12345...
* Connected to 127.0.0.1 (127.0.0.1) port 12345 (#0)
> GET /objects/test HTTP/1.1
> Host: 127.0.0.1:12345
> User-Agent: curl/7.77.0
> Accept: */*
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Thu, 15 Dec 2022 10:07:34 GMT
< Content-Length: 21
< Content-Type: text/plain; charset=utf-8
< 
* Connection #0 to host 127.0.0.1 left intact
```

### 第二章

接口服务和数据服务的objects/heartbeat/locate三个包有很大区别。
数据服务中，objects包负责对象在本地的存取；heartbeat发送心跳消息；locate包用于接收定位消息、定位对象以及发送反馈消息
接口服务中，objects包负责将对象请求转发给数据服务；heartbeat包用于接收数据服务节点的心跳；locate包用于发送定位消息并处理反馈消息

#### rabbitmq的安装和使用
brew install rabbitmq
打开`cd /usr/local/Cellar/rabbitmq/版本号/sbin/`，运行`rabbitmq-server`启动服务
开启插件 `./rabbitmq-plugins enable rabbitmq_management`(关闭插件 `sudo ./rabbitmq-plugins disable rabbitmq_management`)，输入` http://localhost:15672/#/ `登录即可。
可以在网页中添加exchange

添加用户并配置权限
sudo ./rabbitmqctl add_user test test
sudo ./rabbitmqctl set_permissions -p / test ".*" ".*" ".*"


#### 测试
for i in `seq 1 6`; do sudo ifconfig en0 alias 10.29.1.$i/16 255.255.255.0; done

可以通过 `ping 10.29.1.2/16` 来检测
sudo ifconfig en0 -alias 10.29.1.1 可以删除别名

cd /tmp/stg
创建文件 for i in `seq 1 6`; do mkdir -p /tmp/stg/$i/objects; done

export RABBITMQ_SERVER=amqp://test:test@localhost:5672
for i in `seq 1 6`; do LISTEN_ADDRESS=10.29.1.$i:12345 STORAGE_ROOT=/tmp/stg/$i go run ./data_server/data_server.go &; done

for i in `seq 1 2`; do sudo ifconfig en0 alias 10.29.2.$i/16 255.255.255.0; done
for i in `seq 1 2`; do LISTEN_ADDRESS=10.29.2.$i:12345 go run ./api_server/api_server.go &; done

curl -v http://10.29.2.2:12345/objects/test2 -XPUT -d "this is object test2"