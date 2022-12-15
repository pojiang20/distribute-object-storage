# distribute-object-storage
Simple distributed object storage implemented by go

### 1 单机版对象存储
mkdir /tmp/objects

`LISTEN_ADDRESS=:12345 STORAGE_ROOT=/tmp go run server.go`

存对象`curl -v 127.0.0.1:12345/objects/test -XPUT  -d "this is a test object"`
取对象`curl -v 127.0.0.1:12345/objects/test`