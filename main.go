package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"ls_Downloader/def"
	"ls_Downloader/m3u8"
	"ls_Downloader/tool"
)

var (
	url   string
	tsUrl string
	store string

	Params map[string]string = map[string]string{
		"t":     "5f2fb2a2",
		"exper": "0",
		"us":    "82mlhJlJWNUt",
		"sign":  "c6be3fe962b10f699eab9aee4f64c438",
		"whref": "www.lianshiclass.com",
	}
)

func init() {
	flag.StringVar(&url, "u", "", "目标 m3u8 的 url 地址")
	flag.StringVar(&tsUrl, "tu", "", "ts 请求路径 的 url 地址")
	flag.StringVar(&store, "s", "", "ts 文件的存储目录")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `dlvideo version: 0.0.1
Usage: dlvideo [-hvVtTq] [-u url] [-s store]

Options:
`)
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if len(url) == 0 || len(store) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	tsUrl = filepath.Dir(url)
	doRun()
}

func doRun() {
	body, err := tool.Get(url, Params, def.FakePermitHeader)
	if err != nil {
		fmt.Printf("Get m3u8 failed: %v\n", err)
		return
	}
	defer body.Close()

	enti := m3u8.New()
	err = enti.ExtractParser(body)
	if err != nil {
		fmt.Printf("parse lines failed: %v\n", err)
		return
	}

	_ = enti.GetTsCryptKey()

	if err := os.MkdirAll(store, 0777); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	// 防止协程启动过多，限制频率
	limitChan := make(chan byte, 20)
	for idx, val := range enti.Segments {
		wg.Add(1)
		go func(index int, seg *m3u8.Segment) {
			defer func() {
				wg.Done()
				<-limitChan
			}()
			// 以需要命名文件
			fullURL := tsUrl + seg.URI
			body, err := tool.Get(fullURL, nil, def.FakePermitHeader)
			if err != nil {
				fmt.Printf("Download failed [%s] %s\n", err.Error(), fullURL)
				return
			}
			defer body.Close()
			// 创建存在 TS 数据的文件
			tsFile := filepath.Join(store, strconv.Itoa(index)+".ts")
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
			if seg.Key != nil {
				// key := seg.DecodeKey
				bytes, err = tool.Aes128Decrypt(bytes,
					[]byte(seg.DecodeKey),
					[]byte(seg.Key.IV),
				)
				if err != nil {
					fmt.Printf("decryt TS failed: %s\n", err.Error())
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
		}(idx, val)
		limitChan <- 1
	}
	wg.Wait()
}
