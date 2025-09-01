package utils

import "encoding/hex"

func isHexPrefix(s string) bool {
	return len(s) >= 2 && (s[0:2] == "0x" || s[0:2] == "0X")
}

func FromHex(s string) ([]byte, error) {
	if isHexPrefix(s) {
		s = s[2:]
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return hex.DecodeString(s)
}

func LeftPad32(b []byte) []byte {
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

func RightPad32(b []byte) []byte {
	out := make([]byte, 32)
	copy(out, b)
	return out
}
