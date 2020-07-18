package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"godlvideo/tool"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"godlvideo/def"
)

func tempRun() {
	r, err := http.Get("https://app.xiaoe-tech.com/get_video_key.php?" +
		"edk=CiDBYPYN8qJyMuRHWIQTiyShgpQH49ezJhgbJETCbV6QFxCO08TAChiaoOvUBCokYjRhNjFiNTgtMmVhNy00OWYxLTgwZGMtZTE0NTIyODc5YWIy" +
		"&fileId=5285890805166562636&keySource=VodBuildInKMS")
	if err != nil {
		fmt.Printf("request failed: %v", err)
		return
	}

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("read body failed: %v", err)
		return
	}
	str := hex.EncodeToString(data)

	fmt.Printf("%v", str)
}

const (
	EXTM3U    = "#EXTM3U"
	EXT_X_KEY = "#EXT-X-KEY"
)

var lineParameterPattern = regexp.MustCompile(`([a-zA-Z-]+)=("[^"]+"|[^",]+)`)
var url string = "https://1252524126.vod2.myqcloud.com/9764a7a5vodtransgzp1252524126/2d5f54705285890805166562636/drm/v.f230.m3u8"

func parser(lines []string) (*def.M3u8, error) {
	var (
		m       *def.M3u8 = &def.M3u8{}
		key     *def.Key
		seg     *def.Segment
		lineLen = len(lines)
	)

	for i := 0; i < lineLen; i++ {
		line := strings.TrimSpace(lines[i])
		if i == 0 {
			if line != EXTM3U {
				return nil, fmt.Errorf("invalid m3u8, missing #EXTM3U in line 1")
			}
			continue
		}

		switch {
		case line == "":

		case strings.HasPrefix(line, EXT_X_KEY):
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

	return m, nil
}

func parseLineParams(line string) map[string]string {
	r := lineParameterPattern.FindAllStringSubmatch(line, -1)
	// [[METHOD, AES-128], [URI, http://www.g.com]]
	params := make(map[string]string)
	for _, cr := range r {
		params[cr[0]] = cr[1]
	}

	return params
}

func run() {
	body, err := tool.Get(url)
	if err != nil {
		fmt.Printf("Get m3u8 failed: %v\n", err)
		return
	}
	defer body.Close()

	lines := make([]string, 0, 3000)
	s := bufio.NewScanner(body)
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	m, err := parser(lines)
	if err != nil {
		fmt.Printf("parse lines faile: %v\n", err)
		return
	}
	fmt.Printf("%v", m.Segment[0].URI)
}

func main() {
	run()
}
