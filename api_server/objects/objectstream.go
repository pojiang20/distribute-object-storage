package objects

import (
	"fmt"
	"io"
	"net/http"
)

type PutStream struct {
	writer *io.PipeWriter
	c      chan error
}

func NewPutStream(server, object string) *PutStream {
	reader, w := io.Pipe()
	errCh := make(chan error)
	go func() {
		reqeust, _ := http.NewRequest("PUT", fmt.Sprintf("http://%s/objects/%s", server, object), reader)
		client := http.Client{}
		res, err := client.Do(reqeust)
		if err != nil && res.StatusCode != http.StatusOK {
			err = fmt.Errorf("dataServer retturn http code %d", res.StatusCode)
		}
		errCh <- err
	}()
	return &PutStream{w, errCh}
}

func (w *PutStream) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w *PutStream) Close() error {
	//close之后，管道另一端会读到io.EOF
	w.writer.Close()
	return <-w.c
}

type GetStream struct {
	reader io.Reader
}

func newGetStream(url string) (*GetStream, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dataServer return http code %d", resp.StatusCode)
	}
	return &GetStream{resp.Body}, nil
}

// 对外隐藏url细节，只需要提供服务节点地址和对象名就好
func NewGetStream(server, object string) (*GetStream, error) {
	if server == "" || object == "" {
		return nil, fmt.Errorf("invalid server %s object %s", server, object)
	}
	return newGetStream(fmt.Sprintf("http://%s/objects/%s", server, object))
}

func (g *GetStream) Read(p []byte) (n int, err error) {
	return g.reader.Read(p)
}
