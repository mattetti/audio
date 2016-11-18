package id3v2

type TagHeader [10]byte

const (
	id3v2tag_magic = "ID3"
)

type Frame struct {
	Header [10]byte
	Data   []byte
}

type Tag struct {
	header         *TagHeader
	extendedHeader []byte
	frameSets      map[string][]*Frame
}
