package utils

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

func AnyToBig(v any) (*big.Int, error) {
	switch x := v.(type) {
	case *big.Int:
		return new(big.Int).Set(x), nil
	case string:
		if isHexPrefix(x) {
			b, err := FromHex(x)
			if err != nil {
				return nil, err
			}
			return new(big.Int).SetBytes(b), nil
		}
		bi, ok := new(big.Int).SetString(x, 10)
		if !ok {
			return nil, fmt.Errorf("bad int string: %q", x)
		}
		return bi, nil
	case int:
		return big.NewInt(int64(x)), nil
	case int64:
		return big.NewInt(x), nil
	case uint64:
		return new(big.Int).SetUint64(x), nil
	case float64: // JSON number
		i := int64(x)
		if float64(i) != x {
			return nil, fmt.Errorf("non-integer float: %v", x)
		}
		return big.NewInt(i), nil
	default:
		return nil, fmt.Errorf("unsupported int type %T", v)
	}
}

func StrToU64(s string) (uint64, error) {
	if isHexPrefix(s) {
		n := new(big.Int)
		if _, ok := n.SetString(s[2:], 16); !ok {
			return 0, fmt.Errorf("invalid hex: %s", s)
		}
		return n.Uint64(), nil
	}
	n := new(big.Int)
	if _, ok := n.SetString(s, 10); !ok {
		return 0, fmt.Errorf("invalid decimal: %s", s)
	}
	return n.Uint64(), nil
}

func StrToBig(s string) (*big.Int, error) {
	if isHexPrefix(s) {
		n := new(big.Int)
		if _, ok := n.SetString(s[2:], 16); !ok {
			return nil, fmt.Errorf("invalid hex: %s", s)
		}
		return n, nil
	}
	n := new(big.Int)
	if _, ok := n.SetString(s, 10); !ok {
		return nil, fmt.Errorf("invalid decimal: %s", s)
	}
	return n, nil
}

func U64ToBig(n uint64) *big.Int {
	return new(big.Int).SetUint64(n)
}

func IntLikeToWord(typ string, v any) ([]byte, error) {
	signed := strings.HasPrefix(typ, "int")
	bi, err := AnyToBig(v)
	if err != nil {
		return nil, err
	}
	if !signed && bi.Sign() < 0 {
		return nil, errors.New("unsigned negative")
	}
	if bi.Sign() < 0 {
		mod := new(big.Int).Lsh(big.NewInt(1), 256)
		bi.Add(bi, mod)
	}
	return LeftPad32(bi.Bytes()), nil
}

func AddressToWord(v any) ([]byte, error) {
	var b []byte
	switch x := v.(type) {
	case string:
		h, err := FromHex(x)
		if err != nil {
			return nil, err
		}
		b = h
	case []byte:
		b = x
	case [20]byte:
		tmp := x
		b = tmp[:]
	default:
		return nil, fmt.Errorf("address expects string/[]byte/[20]byte, got %T", v)
	}
	if len(b) != 20 {
		return nil, fmt.Errorf("address length must be 20, got %d", len(b))
	}
	return LeftPad32(b), nil
}

func BoolToWord(v any) ([]byte, error) {
	out := make([]byte, 32)
	switch x := v.(type) {
	case bool:
		if x {
			out[31] = 1
		}
	default:
		return nil, fmt.Errorf("bool expects bool, got %T", v)
	}
	return out, nil
}

func StrToWord(v any) ([]byte, error) {
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("string expects string, got %T", v)
	}
	return Keccak([]byte(s)), nil
}

func BytesDynToWord(v any) ([]byte, error) {
	switch x := v.(type) {
	case []byte:
		return Keccak(x), nil
	case string:
		b, err := FromHex(x)
		if err != nil {
			return nil, err
		}
		return Keccak(b), nil
	default:
		return nil, fmt.Errorf("bytes expects []byte or 0x-string, got %T", v)
	}
}

func BytesNToWord(n int, v any) ([]byte, error) {
	var b []byte
	switch x := v.(type) {
	case []byte:
		b = x
	case string:
		h, err := FromHex(x)
		if err != nil {
			return nil, err
		}
		b = h
	default:
		return nil, fmt.Errorf("bytes%d expects []byte or 0x-string, got %T", n, v)
	}
	if len(b) != n {
		return nil, fmt.Errorf("bytes%d length %d", n, len(b))
	}
	return RightPad32(b), nil
}

func EncodePrimitiveWord(typ string, v any) ([]byte, error) {
	switch typ {
	case "address":
		return AddressToWord(v)
	case "bool":
		return BoolToWord(v)
	case "string":
		return StrToWord(v)
	case "bytes":
		return BytesDynToWord(v)
	}
	if strings.HasPrefix(typ, "bytes") {
		// bytesN
		nStr := strings.TrimPrefix(typ, "bytes")
		n := 0
		if nStr == "" {
			return nil, fmt.Errorf("invalid bytes type: %q", typ)
		}
		var err error
		n, err = strconv.Atoi(nStr)
		if err != nil || n < 1 || n > 32 {
			return nil, fmt.Errorf("invalid bytesN: %q", typ)
		}
		return BytesNToWord(n, v)
	}
	if typ == "int" || typ == "uint" || strings.HasPrefix(typ, "int") || strings.HasPrefix(typ, "uint") {
		return IntLikeToWord(typ, v)
	}
	return nil, fmt.Errorf("unknown primitive type: %q", typ)
}
