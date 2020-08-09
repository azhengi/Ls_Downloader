package m3u8

import (
	"bufio"
	"fmt"
	"godlvideo/def"
	"godlvideo/tool"
	"io"
	"io/ioutil"
	"strings"
	"sync"
)

type CryptMethod string

const (
	CryptAES  CryptMethod = "AES-128"
	CryptNONE CryptMethod = "NONE"
)

type Segment struct {
	URI       string
	DecodeKey []byte
	Key       *Key
}

type Key struct {
	Method CryptMethod
	URI    string
	IV     string
}

type Entity struct {
	Segments []*Segment
}

func New() *Entity {
	return new(Entity)
}

func (m *Entity) ExtractParser(body io.ReadCloser) error {
	lines := make([]string, 0, 3000)
	scanner := bufio.NewScanner(body)
	isFirstRow := true
	row := 1
	var key *Key
	for scanner.Scan() {
		row++
		text := scanner.Text()
		lines = append(lines, text)

		ln := strings.TrimSpace(text)
		if isFirstRow {
			isFirstRow = false
			if ln != def.EXTM3U {
				return fmt.Errorf("invalid m3u8, missing #EXTM3U in line 1")
			}
			continue
		}

		switch {
		case ln == "":
			// 提取 key

		case strings.HasPrefix(ln, def.EXT_X_KEY):
			params := tool.ParseLineParams(ln)
			if len(params) == 0 {
				return fmt.Errorf("invalid #EXT-X-KEY: %s, line: %d", ln, row)
			}

			key = new(Key)

			method := CryptMethod(params["METHOD"])
			uri, uriOK := params["URI"]
			iv, ivOK := params["IV"]
			if method == CryptNONE {
				if uriOK || ivOK {
					return fmt.Errorf("invalid #EXT-X-KEY: %s, line: %d", ln, row)
				}
			}
			if method != "" && method != CryptAES && method != CryptNONE {
				return fmt.Errorf("invalid #EXT-X-KEY method: %s, line: %d", method, row)
			}

			key.Method = method
			key.URI = uri
			key.IV = iv

		case !strings.HasPrefix(ln, "#"):
			seg := new(Segment)
			seg.URI = ln
			seg.Key = key
			m.Segments = append(m.Segments, seg)

		default:
		}
	}

	return nil
}

func FetchCryptKeyText(seg *Segment, uri string) error {
	body, err := tool.Get(uri, nil, def.NormalClientHeader)
	if err != nil {
		return fmt.Errorf("Get %s key file err:%v\n", uri, err)
	}
	defer body.Close()
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return fmt.Errorf("Read %s body err:%v\n", uri, err)
	}
	seg.DecodeKey = b
	return nil
}

func (m *Entity) GetTsCryptKey() error {

	limitChan := make(chan int, 200)

	wg := &sync.WaitGroup{}
	wg.Add(len(m.Segments))

	for _, v := range m.Segments {
		go func(seg *Segment) {
			uri := seg.Key.URI
			defer func() {
				<-limitChan
				wg.Done()
			}()

			body, err := tool.Get(uri, nil, def.NormalClientHeader)
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
			// actualText := hex.EncodeToString(b)
			seg.DecodeKey = b
		}(v)

		limitChan <- 1
	}

	return nil
}
