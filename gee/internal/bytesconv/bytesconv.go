package bytesconv

import "unsafe"

// 不会发生内存拷贝的字符串转字节数组
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}
