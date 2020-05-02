package gocache

// ByteView 用于表示缓存的数据.
// 作为LRU中的value来使用
type ByteView struct {
	b []byte
}

func (v ByteView) Len() uint64 {
	return uint64(len(v.b))
}

func (v ByteView) String() string  {
	return string(v.b)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}


func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
