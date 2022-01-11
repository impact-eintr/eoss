package objectstream

import (
	"fmt"
	"io"
	"net/http"
)

type GetStream struct {
	reader io.Reader
}

func NewGetStream(server, object string) (*GetStream, error) {
	if server == "" || object == "" {
		return nil, fmt.Errorf("invalid server %s object %s", server, object)
	}
	return newGetStream("http://" + server + "/objects/" + object)
}

func newGetStream(url string) (*GetStream, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dataServer return http code %d", r.StatusCode)
	}
	return &GetStream{r.Body}, nil
}

func (r *GetStream) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}
