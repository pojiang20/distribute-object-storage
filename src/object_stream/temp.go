package object_stream

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type TempPutStream struct {
	Server string
	UUID   string
}

func NewTempPutStream(server, objectName string, size int64) (*TempPutStream, error) {
	//创建临时对象
	req, err := http.NewRequest(http.MethodPost, "http://"+server+"/temp/"+objectName, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("size", fmt.Sprintf("%d", size))
	httpCli := http.Client{}
	//执行请求后，会得到临时对象的UUID
	resp, err := httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	uuidBytes, err := io.ReadAll(resp.Body)
	//处理读取的uuid末尾的换行
	uuidBytes = bytes.ReplaceAll(uuidBytes, []byte("\n"), []byte(""))
	if err != nil {
		return nil, err
	}

	return &TempPutStream{
		Server: server,
		UUID:   string(uuidBytes),
	}, nil
}

// 实现write接口
func (t *TempPutStream) Write(p []byte) (n int, err error) {
	req, err := http.NewRequest(http.MethodPatch, "http://"+t.Server+"/temp/"+t.UUID, strings.NewReader(string(p)))
	if err != nil {
		return 0, err
	}
	httpCli := http.Client{}
	resp, err := httpCli.Do(req)
	if err != nil {
		return 0, nil
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("dataServer Error: STATUSCODE[%d]\n", resp.StatusCode)
	}
	return len(p), nil
}

// true将临时对象转正
func (t *TempPutStream) Commit(toFormal bool) {
	method := http.MethodDelete
	if toFormal {
		method = http.MethodPut
	}
	req, _ := http.NewRequest(method, "http://"+t.Server+"/temp/"+t.UUID, nil)
	cli := http.Client{}
	cli.Do(req)
}
