package tool

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"godlvideo/def"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

var lineParameterPattern = regexp.MustCompile(`([a-zA-Z-]+)=("[^"]+"|[^",]+)`)

func Get(u string, params map[string]string) (io.ReadCloser, error) {
	c := http.Client{Timeout: time.Second * time.Duration(60)}
	request, err := http.NewRequest(http.MethodGet, u, nil)
	values := request.URL.Query()
	for k, v := range params {
		values.Add(k, v)
	}
	request.URL.RawQuery = values.Encode()

	for k, v := range def.FakeHeader {
		request.Header.Add(k, v)
	}
	resp, err := c.Do(request)
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
			m.Segments = append(m.Segments, seg)
		default:
		}
	}
	return m, nil
}

func parseLineParams(line string) map[string]string {
	r := lineParameterPattern.FindAllStringSubmatch(line, -1)
	params := make(map[string]string)
	for _, cr := range r {
		params[cr[1]] = strings.Trim(cr[2], "\"")
	}

	return params
}

func ResolveURL(u *url.URL, p string) string {
	if strings.HasPrefix(p, "https://") || strings.HasPrefix(p, "http://") {
		return p
	}
	var baseURL string
	if strings.Index(p, "/") == 0 {
		baseURL = u.Scheme + "://" + u.Host
	} else {
		tU := u.String()
		baseURL = tU[0:strings.LastIndex(tU, "/")]
	}
	return baseURL + path.Join("/", p)
}

func Aes128Decrypt(crypted, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	if len(iv) == 0 {
		iv = key
	}
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = pkcs5UnPadding(origData)
	return origData, nil
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}
