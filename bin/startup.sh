# brew启动es和rabbitmq
brew services start elasticsearch-full
# 创建metadata索引和映射
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
echo ""
echo "INFO: finish start elasticsearch && create index 'metadata' and mappings"

brew services start rabbitmq

# 设置文件
for i in `seq 1 6`; do mkdir -p /tmp/stg/$i/objects; done
for i in `seq 1 6`; do mkdir -p /tmp/stg/$i/temp; done

# 设置环境变量
export RABBITMQ_SERVER=amqp://test:test@localhost:5672
export ES_SERVER=localhost:9200

# 设置ip地址
for i in `seq 1 6`; do sudo ifconfig lo0 alias 10.29.1.$i/16 255.255.255.0; done
for i in `seq 1 2`; do sudo ifconfig lo0 alias 10.29.2.$i/16 255.255.255.0; done

# 运行
for i in `seq 1 2`; do LISTEN_ADDRESS=10.29.2.$i:12345 go run ../api_server/api_server.go &; done
for i in `seq 1 6`; do LISTEN_ADDRESS=10.29.1.$i:12345 STORAGE_ROOT=/tmp/stg/$i go run ../data_server/data_server.go &; done
