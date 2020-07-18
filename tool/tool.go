package tool

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func Get(u string) (io.ReadCloser, error) {
	// http.Client http.Reqeust
	c := http.Client{Timeout: time.Second * time.Duration(60)}
	resp, err := c.Get(u)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error: status code %d", resp.StatusCode)
	}

	return resp.Body, nil
}
