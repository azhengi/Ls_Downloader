package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"godlvideo/tool"
	"io/ioutil"
	"net/http"
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

var url string = "https://1252524126.vod2.myqcloud.com/9764a7a5vodtransgzp1252524126/2d5f54705285890805166562636/drm/v.f230.m3u8"

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

	m, err := tool.Parser(lines)
	if err != nil {
		fmt.Printf("parse lines failed: %v\n", err)
		return
	}

	fmt.Printf("%+v\n", m.Segment[0])
}

func main() {
	run()
}
