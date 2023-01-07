# 传输数据
#"this object will be seperated to 4+2"

# 存储请求
curl -v 10.29.2.2:12345/objects/test5 -XPUT -d "this object will be seperated to 4+2" -H "Digest: SHA-256=2MRHhgqkvILs04RzLIEpZBzpgos/9QlNKRXEh7cMngY="

# 读取请求
curl 10.29.2.1:12345/objects/test5

# 删除和修改
rm /tmp/stg/1/objects/2MRHhgqkvILs04RzLIEpZBzpgos%2F9QlNKRXEh7cMngY=.4.3bG4EYy18vq+9JU%2F0vzbeRxDjFglFg2OBa26TlEt8gA=
echo errMsg > /tmp/stg/4/objects/2MRHhgqkvILs04RzLIEpZBzpgos%2F9QlNKRXEh7cMngY=.0.%2FQbxVU4EOPqz%2F2DYStzA19gOT5KBh2eEeuktmV3NFgY=