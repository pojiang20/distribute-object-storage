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
可以在网页中添加exchanges

添加用户并配置权限
`sudo ./rabbitmqctl add_user test test`
`sudo ./rabbitmqctl set_permissions -p / test ".*" ".*" ".*"`

配置环境变量`export RABBITMQ_SERVER=amqp://test:test@localhost:5672`

mac中也可以使用homebrew直接启动`brew services start rabbitmq`

#### 测试
开启多个ip地址 ``for i in `seq 1 6`; do sudo ifconfig lo0 alias 10.29.1.$i/16 255.255.255.0; done``
可以通过 `ping 10.29.1.2` 来检测
`sudo ifconfig lo0 -alias 10.29.1.1` 可以删除别名

创建文件 for i in `seq 1 6`; do mkdir -p /tmp/stg/$i/objects; done

export RABBITMQ_SERVER=amqp://test:test@localhost:5672
for i in `seq 1 6`; do LISTEN_ADDRESS=10.29.1.$i:12345 STORAGE_ROOT=/tmp/stg/$i go run ./data_server/data_server.go &; done

for i in `seq 1 2`; do sudo ifconfig lo0 alias 10.29.2.$i/16 255.255.255.0; done
for i in `seq 1 2`; do LISTEN_ADDRESS=10.29.2.$i:12345 go run ./api_server/api_server.go &; done

curl -v http://10.29.2.2:12345/objects/test2 -XPUT -d "this is object test2"

#### 项目结构
![img.png](img.png)
项目结构如图，使用rabbitMQ进行一对多的广播是关键。apiServer需要读取对象时，向所有dataServer广播，只有存在对象的dataServer节点发送响应。dataServer通过广播向所有apiServer发送心跳，告知它们自己服务可用。

#### 数据流
项目中的数据流处理很巧妙，好处就是，从apiServer获取的put请求一直到apiServer发送请求给dataServer这个过程，都没有存储过数据，而是一个个读写接口传递，最后传递给httpRequest中的io.reader。这样即使传输内容非常大。
apiServer端也不会收到影响。
这里是关系是：request.Body实现了io.reader（A），且io.Pipe返回了writer（B）和reader（C）。io.copy将A的内容写入B，pipe会将B的内容写入C，C作为io.reader直接放到HttpRequest中发送。

### 第三章
本章添加了元数据来对对象进行版本管理，apiServer部分提供的接口进行相关操作也都需要向元数据数据库中读、写，dataServer不受影响但是是以元数据中获得对象的散列值作为对象名来读取对象。
#### es安装
`brew tap elastic/tap`、`brew install elastic/tap/elasticsearch-full`即可安装最新的elasticsearch
`brew services start elasticsearch-full`执行
```shell
# 建立索引
curl -XPUT localhost:9200/metadata -H 'Content-Type:application/json' -d '{
  "mappings": {
    "properties": {
      "name": {"type": "keyword"},
      "version": {"type": "integer"},
      "size": {"type": "integer"},
      "hash": {"type": "keyword"}
    }
  }
}'
```
输入`localhost:9200/metadata`可以看到elasticsearch相关数据
`export ES_SERVER=localhost:9200`配置环境变量
#### 测试
创建temp文件，``for i in `seq 1 6`; do mkdir -p /tmp/stg/$i/temp; done``
```shell script
curl -v 10.29.2.2:12345/objects/test -XPUT -d "This is object content version-1" -H "Digest: SHA-256=9AimTha2kCISf8bVfi1jPXo2BzY="
```
可以看到如下结果
```text
☁  stg  curl -v http://10.29.2.2:12345/objects/test -XPUT -d "This is object content version-1" -H "Digest: SHA-256=DrkE6D0mP4aGOU37mPqUrpauEk5LRso0LBt6jnwISv0="
*   Trying 10.29.2.2:12345...
* Connected to 10.29.2.2 (10.29.2.2) port 12345 (#0)
> PUT /objects/test HTTP/1.1
> Host: 10.29.2.2:12345
> User-Agent: curl/7.77.0
> Accept: */*
> Digest: SHA-256=DrkE6D0mP4aGOU37mPqUrpauEk5LRso0LBt6jnwISv0=
> Content-Length: 32
> Content-Type: application/x-www-form-urlencoded
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Mon, 26 Dec 2022 16:05:42 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.2 left intact
```
定位测试
```shell script
☁  stg  curl -v 10.29.2.2:12345/locate/DrkE6D0mP4aGOU37mPqUrpauEk5LRso0LBt6jnwISv0= 
*   Trying 10.29.2.2:12345...
* Connected to 10.29.2.2 (10.29.2.2) port 12345 (#0)
> GET /locate/DrkE6D0mP4aGOU37mPqUrpauEk5LRso0LBt6jnwISv0= HTTP/1.1
> Host: 10.29.2.2:12345
> User-Agent: curl/7.77.0
> Accept: */*
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Mon, 26 Dec 2022 16:07:45 GMT
< Content-Length: 17
< Content-Type: text/plain; charset=utf-8
< 
* Connection #0 to host 10.29.2.2 left intact
"10.29.1.2:12345"%    
```
获取对象
```shell
☁  stg  curl -v 10.29.2.2:12345/objects/test
*   Trying 10.29.2.2:12345...
* Connected to 10.29.2.2 (10.29.2.2) port 12345 (#0)
> GET /objects/test HTTP/1.1
> Host: 10.29.2.2:12345
> User-Agent: curl/7.77.0
> Accept: */*
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Tue, 27 Dec 2022 02:54:32 GMT
< Content-Length: 32
< Content-Type: text/plain; charset=utf-8
< 
* Connection #0 to host 10.29.2.2 left intact
This is object content version-1%
```

```shell script
# 查看对象名为test的所有版本的元数据信息
curl -v 10.29.2.2:12345/objects/test
# 查看对象名为test的指定版本的元数据信息，注意携带的参数的问号要转义
curl -v 10.29.2.2:12345/objects/test\?version
```
curl -v 10.29.2.2:12345/objects/DrkE6D0mP4aGOU37mPqUrpauEk5LRso0LBt6jnwISv0=
```shell script
curl -v 10.29.2.2:12345/objects/test -XDELETE
```

### 第四章

#### 测试

### 第五章

#### 测试效果

发送请求和响应，响应200表示存储成功
```text
curl -v 10.29.2.2:12345/objects/test5 -XPUT -d "this object will be seperated to 4+2" -H "Digest: SHA-256=2MRHhgqkvILs04RzLIEpZBzpgos/9QlNKRXEh7cMngY="
*   Trying 10.29.2.2:12345...
* Connected to 10.29.2.2 (10.29.2.2) port 12345 (#0)
> PUT /objects/test5 HTTP/1.1
> Host: 10.29.2.2:12345
> User-Agent: curl/7.77.0
> Accept: */*
> Digest: SHA-256=2MRHhgqkvILs04RzLIEpZBzpgos/9QlNKRXEh7cMngY=
> Content-Length: 36
> Content-Type: application/x-www-form-urlencoded
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Sat, 07 Jan 2023 04:56:45 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.2 left intact
```
查看存储情况
```text
.
├── 1
│   ├── objects
│   │   └── 2MRHhgqkvILs04RzLIEpZBzpgos%2F9QlNKRXEh7cMngY=.4.3bG4EYy18vq+9JU%2F0vzbeRxDjFglFg2OBa26TlEt8gA=
│   └── temp
├── 2
│   ├── objects
│   │   └── 2MRHhgqkvILs04RzLIEpZBzpgos%2F9QlNKRXEh7cMngY=.2.syDfa6hGt2PmnadawXz8SSr2zaalOCW6zCT3Ky7BcW4=
│   └── temp
├── 3
│   ├── objects
│   │   └── 2MRHhgqkvILs04RzLIEpZBzpgos%2F9QlNKRXEh7cMngY=.1.S8JnDgiwWnfNud06pGE+%2FEHgmBzNpFAEg6s%2FgR0J4Cw=
│   └── temp
├── 4
│   ├── objects
│   │   └── 2MRHhgqkvILs04RzLIEpZBzpgos%2F9QlNKRXEh7cMngY=.0.%2FQbxVU4EOPqz%2F2DYStzA19gOT5KBh2eEeuktmV3NFgY=
│   └── temp
├── 5
│   ├── objects
│   │   └── 2MRHhgqkvILs04RzLIEpZBzpgos%2F9QlNKRXEh7cMngY=.3.0aX0oLlWMm02ane45F2KxC79KTwjBUmuh1GwXcfRqH4=
│   └── temp
└── 6
    ├── objects
    │   └── 2MRHhgqkvILs04RzLIEpZBzpgos%2F9QlNKRXEh7cMngY=.5.5GrT1Vp7C1sI4V5brEWSJ+heWjWxDpJvVE75nVY8bdk=
    └── temp

```
查看对象
```text
curl 10.29.2.1:12345/objects/test5
this object will be seperated to 4+2%                                                   
```

删除和修改
```text
rm /tmp/stg/1/objects/2MRHhgqkvILs04RzLIEpZBzpgos%2F9QlNKRXEh7cMngY=.4.3bG4EYy18vq+9JU%2F0vzbeRxDjFglFg2OBa26TlEt8gA=
echo errMsg > /tmp/stg/4/objects/2MRHhgqkvILs04RzLIEpZBzpgos%2F9QlNKRXEh7cMngY=.0.%2FQbxVU4EOPqz%2F2DYStzA19gOT5KBh2eEeuktmV3NFgY=

☁  stg  tree ./1        
./1
├── objects
└── temp
```
再次查看对象，可以获取信息，并且删除和修改的文件得到了修复。
```text
☁  stg  tree ./1
./1
├── objects
│   └── 2MRHhgqkvILs04RzLIEpZBzpgos%2F9QlNKRXEh7cMngY=.4.S8JnDgiwWnfNud06pGE+%2FEHgmBzNpFAEg6s%2FgR0J4Cw=
└── temp
```
删除2个以上的文件
```text
rm [1-3]/objects/*
```
再次请求则返回404
```text
☁  distribute-object-storage [dev] ⚡  curl -v 10.29.2.1:12345/objects/test5
*   Trying 10.29.2.1:12345...
* Connected to 10.29.2.1 (10.29.2.1) port 12345 (#0)
> GET /objects/test5 HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.77.0
> Accept: */*
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 404 Not Found
< Date: Sat, 07 Jan 2023 05:10:01 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.1 left intact
```

### 第六章
