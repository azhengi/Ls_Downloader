package def

type KeyURL string

const (
	EXTM3U    = "#EXTM3U"
	EXT_X_KEY = "#EXT-X-KEY"
)

var FakePermitHeader = map[string]string{
	"Accept":          "*/*",
	"Accept-Encoding": "gzip, deflate, br",
	"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
	"Cache-Control":   "no-cache",
	"Connection":      "keep-alive",
	"Host":            "encrypt-k-vod.xiaoe-tech.com",
	"Origin":          "https://www.lianshiclass.com",
	"Pragma":          "no-cache",
	"Referer":         "https://www.lianshiclass.com/detail/v_5f1a91bce4b0a1003caeff45/3?from=p_5e0d5bb80e33e_14PVeh6x&type=5",
	"Sec-Fetch-Dest":  "empty",
	"Sec-Fetch-Mode":  "cors",
	"Sec-Fetch-Site":  "cross-site",
	"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.113 Safari/537.36",
}

var NormalClientHeader = map[string]string{
	"Referer":    "https://www.lianshiclass.com/detail/v_5f1a91bce4b0a1003caeff45/3?from=p_5e0d5bb80e33e_14PVeh6x&type=5",
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.113 Safari/537.36",
}
