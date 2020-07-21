package tool

import (
	"encoding/hex"
	"fmt"
	"godlvideo/def"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

var lineParameterPattern = regexp.MustCompile(`([a-zA-Z-]+)=("[^"]+"|[^",]+)`)

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

func Parser(lines []string) (*def.M3u8, error) {
	var (
		m       *def.M3u8 = &def.M3u8{}
		key     *def.Key
		seg     *def.Segment
		lineLen = len(lines)
	)

	for i := 0; i < lineLen; i++ {
		line := strings.TrimSpace(lines[i])
		if i == 0 {
			if line != def.EXTM3U {
				return nil, fmt.Errorf("invalid m3u8, missing #EXTM3U in line 1")
			}
			continue
		}

		switch {
		case line == "":

		case strings.HasPrefix(line, def.EXT_X_KEY):
			params := parseLineParams(line)
			if len(params) == 0 {
				return nil, fmt.Errorf("invalid #EXT-X-KEY: %s, line: %d", line, i+1)
			}
			key = new(def.Key)
			method := def.CryptMethod(params["METHOD"])
			uri, uriOK := params["URI"]
			iv, ivOK := params["IV"]
			if method == def.CryptNONE {
				if uriOK || ivOK {
					return nil, fmt.Errorf("invalid #EXT-X-KEY: %s, line: %d", line, i+1)
				}
			}
			if method != "" && method != def.CryptAES && method != def.CryptNONE {
				return nil, fmt.Errorf("invalid #EXT-X-KEY method: %s, line: %d", method, i+1)
			}

			key.Method = method
			key.URI = uri
			key.IV = iv
		// 解析 ts
		case !strings.HasPrefix(line, "#"):
			seg = new(def.Segment)
			seg.URI = line
			seg.Key = key
			m.Segment = append(m.Segment, seg)
		default:
		}
	}

	GetTsKeyText(m)

	return m, nil
}

func parseLineParams(line string) map[string]string {
	r := lineParameterPattern.FindAllStringSubmatch(line, -1)
	// [[METHOD=AES-128. METHOD, AES-128], [URI=http://www.g.com, URI, http://www.g.com]]
	params := make(map[string]string)
	for _, cr := range r {
		params[cr[1]] = strings.Trim(cr[2], "\"")
	}

	return params
}

func GetTsKeyText(m *def.M3u8) (*def.M3u8, error) {

	limitChan := make(chan int, 200)

	wg := &sync.WaitGroup{}
	wg.Add(len(m.Segment))

	for _, v := range m.Segment {
		go func(seg *def.Segment) {
			uri := seg.Key.URI
			defer func() {
				<-limitChan
				wg.Done()
			}()

			body, err := Get(uri)
			if err != nil {
				fmt.Printf("Get %s key file err:%v\n", uri, err)
				return
			}
			defer body.Close()

			b, err := ioutil.ReadAll(body)
			if err != nil {
				fmt.Printf("Read %s body err:%v\n", uri, err)
				return
			}

			actualText := hex.EncodeToString(b)
			seg.DecodeKey = actualText
		}(v)

		limitChan <- 20
	}

	return m, nil
}
