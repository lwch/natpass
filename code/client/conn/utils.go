package conn

func dup(data []byte) []byte {
	ret := make([]byte, len(data))
	copy(ret, data)
	return ret
}
