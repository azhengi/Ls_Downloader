package def

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
	URI string
	Key *Key
}
