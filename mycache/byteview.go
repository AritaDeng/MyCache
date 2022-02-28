package mycache

type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}
func cloneBytes(b []byte) []byte {
	clo := make([]byte, len(b))
	copy(b, clo)
	return clo
}
func (v ByteView) String() string {
	return string(v.b)
}
