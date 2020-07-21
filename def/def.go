package def

type KeyURL string

type M3u8 struct {
	Segment []*Segment
}

type CryptMethod string

const (
	CryptAES  CryptMethod = "AES-128"
	CryptNONE CryptMethod = "NONE"
)

type Key struct {
	Method CryptMethod
	URI    string
	IV     string
}

type Segment struct {
	URI       string
	DecodeKey string
	Key       *Key
}

const (
	EXTM3U    = "#EXTM3U"
	EXT_X_KEY = "#EXT-X-KEY"
)
