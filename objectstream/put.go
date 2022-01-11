package objectstream

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
	reader, writer := io.Pipe()
	c := make(chan error)
	go func() {
		req, _ := http.NewRequest("PUT", "http://"+server+"/objects/"+object, reader)
		client := http.Client{}
		r, err := client.Do(req)
		if err == nil && r.StatusCode != http.StatusOK {
			err = fmt.Errorf("dataServer return http code %d", r.StatusCode)
		}
		c <- err
	}()
	return &PutStream{writer, c}
}

func (w *PutStream) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w *PutStream) Close() error {
	w.writer.Close()
	return <-w.c
}
