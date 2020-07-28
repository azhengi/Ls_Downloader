package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"godlvideo/def"
	"godlvideo/tool"
)

const (
	StoreFolder = "F:/assets/down"
	TsBaseURL   = "https://encrypt-k-vod.xiaoe-tech.com/9764a7a5vodtransgzp1252524126/d62a10165285890805593265719/drm/"
	CryptKey    = "DF94B77EF2D324477C7A2DDAEFD29236"
)

var (
	URL string = "https://encrypt-k-vod.xiaoe-tech.com/9764a7a5vodtransgzp1252524126/d62a10165285890805593265719/drm/v.f230.m3u8"

	Params map[string]string = map[string]string{
		"t":     "5f21985d",
		"exper": "0",
		"us":    "wQl73trtFLZL",
		"sign":  "74c3dee01031091d2d984e0d90288b2a",
		"whref": "www.lianshiclass.com",
	}
)

func main() {
	doRun()
}

func doRun() {
	body, err := tool.Get(URL, Params)
	if err != nil {
		fmt.Printf("Get m3u8 failed: %v\n", err)
		return
	}
	defer body.Close()

	// file, err := os.Open("./temp/v.f230.m3u8")
	// if err != nil {
	// 	fmt.Printf("Open m3u8 file failed: %v\n", err)
	// 	return
	// }

	lines := make([]string, 0, 3000)
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	m, err := tool.Parser(lines)
	if err != nil {
		fmt.Printf("parse lines failed: %v\n", err)
		return
	}

	GetTsCryptKey(m.Segments)

	if err := os.MkdirAll(StoreFolder, 0777); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	// 防止协程启动过多，限制频率
	limitChan := make(chan byte, 20)
	for idx, seg := range m.Segments {
		wg.Add(1)
		go func(i int, s *def.Segment) {
			defer func() {
				wg.Done()
				<-limitChan
			}()
			// 以需要命名文件
			fullURL := TsBaseURL + seg.URI
			body, err := tool.Get(fullURL, nil)
			if err != nil {
				fmt.Printf("Download failed [%s] %s\n", err.Error(), fullURL)
				return
			}
			defer body.Close()
			// 创建存在 TS 数据的文件
			tsFile := filepath.Join(StoreFolder, strconv.Itoa(i)+".ts")
			tsFileTmpPath := tsFile + "_tmp"
			tsFileTmp, err := os.Create(tsFileTmpPath)
			if err != nil {
				fmt.Printf("Create TS file failed: %s\n", err.Error())
				return
			}
			//noinspection GoUnhandledErrorResult
			defer tsFileTmp.Close()
			bytes, err := ioutil.ReadAll(body)
			if err != nil {
				fmt.Printf("Read TS file failed: %s\n", err.Error())
				return
			}
			// 解密 TS 数据
			if s.Key != nil {
				// key := seg.DecodeKey
				bytes, err = tool.Aes128Decrypt(bytes, []byte(CryptKey), []byte(s.Key.IV))
				if err != nil {
					fmt.Printf("decryt TS failed: %s\n", err.Error())
				}
			}

			syncByte := uint8(71) //0x47
			bLen := len(bytes)
			for j := 0; j < bLen; j++ {
				if bytes[j] == syncByte {
					bytes = bytes[j:]
					break
				}
			}

			if _, err := tsFileTmp.Write(bytes); err != nil {
				fmt.Printf("Save TS file failed:%s\n", err.Error())
				return
			}
			_ = tsFileTmp.Close()
			// 重命名为正式文件
			if err = os.Rename(tsFileTmpPath, tsFile); err != nil {
				fmt.Printf("Rename TS file failed: %s\n", err.Error())
				return
			}
			fmt.Printf("下载成功：%s\n", fullURL)
		}(idx, seg)
		limitChan <- 1
	}
	wg.Wait()
}

func GetTsCryptKey(segments []*def.Segment) ([]*def.Segment, error) {

	limitChan := make(chan int, 200)

	wg := &sync.WaitGroup{}
	wg.Add(len(segments))

	for _, v := range segments {
		go func(seg *def.Segment) {
			uri := seg.Key.URI
			defer func() {
				<-limitChan
				wg.Done()
			}()

			body, err := tool.Get(uri, nil)
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

	return segments, nil
}
