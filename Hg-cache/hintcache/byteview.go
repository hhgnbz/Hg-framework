package hintcache

type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice returns copy of the data as a byte slice
// ByteView中数据为只读，为防止外部程序修改，用本方法返回一个克隆值
func (v ByteView) ByteSlice() []byte {
	return cloneSlice(v.b)
}

// String returns copy of the data as a string
func (v ByteView) String() string {
	return string(v.b)
}

// Deep copy
func cloneSlice(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
