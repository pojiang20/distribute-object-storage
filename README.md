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