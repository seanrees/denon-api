package util

// To bytes converts a string to a slice of bytes, terminated by '\r'.
func ToBytes(s string) []byte {
	buf := make([]byte, len(s)+1)
	copy(buf, s)
	buf[len(s)] = '\r'
	return buf
}
