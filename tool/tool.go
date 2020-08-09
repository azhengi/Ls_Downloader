package tool

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

var lineParameterPattern = regexp.MustCompile(`([a-zA-Z-]+)=("[^"]+"|[^",]+)`)

func Get(u string, params map[string]string, headers map[string]string) (io.ReadCloser, error) {
	c := http.Client{Timeout: time.Second * time.Duration(60)}
	request, err := http.NewRequest(http.MethodGet, u, nil)
	values := request.URL.Query()
	for k, v := range params {
		values.Add(k, v)
	}
	request.URL.RawQuery = values.Encode()

	for k, v := range headers {
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

func ParseLineParams(line string) map[string]string {
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
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = Pkcs5UnPadding(origData)
	return origData, nil
}

func Pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}
