package transaction

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/gosunuts/ethtxbuilder/utils"
)

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"` // e.g. "address", "uint256", "MyStruct", "MyStruct[]"
}
type Types map[string][]Field

type Domain struct {
	Name              string
	Version           string
	ChainID           *big.Int
	VerifyingContract string
	Salt              string // 0x-hex
}

type TypedData struct {
	Types       Types
	PrimaryType string
	Domain      Domain
	Message     map[string]any
}

/* ---------------- Domain helpers ---------------- */

func (d *Domain) Map() map[string]any {
	m := map[string]any{}
	if d.Name != "" {
		m["name"] = d.Name
	}
	if d.Version != "" {
		m["version"] = d.Version
	}
	if d.ChainID != nil {
		m["chainId"] = d.ChainID
	}
	if d.VerifyingContract != "" {
		m["verifyingContract"] = d.VerifyingContract
	}
	if d.Salt != "" {
		m["salt"] = d.Salt
	}
	return m
}

/* ---------------- Type graph & encoding ---------------- */

func isRefType(t string) bool {
	return t != "" && unicode.IsUpper([]rune(t)[0])
}
func baseType(t string) string {
	if strings.HasSuffix(t, "[]") {
		return t[:len(t)-2]
	}
	return t
}
func (td *TypedData) deps(primary string, seen map[string]bool, order *[]string) {
	primary = baseType(primary)
	if seen[primary] || td.Types[primary] == nil {
		return
	}
	seen[primary] = true
	*order = append(*order, primary)
	for _, f := range td.Types[primary] {
		if isRefType(baseType(f.Type)) {
			td.deps(f.Type, seen, order)
		}
	}
}

func (td *TypedData) EncodeType(primary string) []byte {
	var order []string
	td.deps(primary, map[string]bool{}, &order)
	if len(order) > 1 {
		tail := append([]string{}, order[1:]...)
		sort.Strings(tail)
		order = append([]string{order[0]}, tail...)
	}
	var buf bytes.Buffer
	for _, typ := range order {
		buf.WriteString(typ)
		buf.WriteByte('(')
		fs := td.Types[typ]
		for i, f := range fs {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(f.Type)
			buf.WriteByte(' ')
			buf.WriteString(f.Name)
		}
		buf.WriteByte(')')
	}
	return buf.Bytes()
}

func (td *TypedData) TypeHash(primary string) []byte {
	return utils.Keccak(td.EncodeType(primary))
}

/* ---------------- Data encoding (EIP-712) ---------------- */

func (td *TypedData) EncodeData(primary string, data map[string]any) ([]byte, error) {
	fields := td.Types[primary]
	if fields == nil {
		return nil, fmt.Errorf("unknown type %q", primary)
	}
	var buf bytes.Buffer
	buf.Write(td.TypeHash(primary))

	for _, f := range fields {
		ft, fv := f.Type, data[f.Name]

		// arrays
		if strings.HasSuffix(ft, "[]") {
			bt := baseType(ft)
			digest, err := td.encodeArray(bt, fv)
			if err != nil {
				return nil, fmt.Errorf("%s[]: %w", bt, err)
			}
			buf.Write(digest) // already 32-byte hash
			continue
		}

		// nested struct
		if isRefType(ft) {
			sub, ok := fv.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%s expects object", f.Name)
			}
			enc, err := td.EncodeData(ft, sub)
			if err != nil {
				return nil, err
			}
			buf.Write(utils.Keccak(enc))
			continue
		}

		// primitive
		word, err := encodePrimitive(ft, fv)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f.Name, err)
		}
		buf.Write(word)
	}
	return buf.Bytes(), nil
}

func (td *TypedData) HashStruct(primary string, data map[string]any) ([]byte, error) {
	enc, err := td.EncodeData(primary, data)
	if err != nil {
		return nil, err
	}
	return utils.Keccak(enc), nil
}

func (td *TypedData) ensureDomainType() {
	if td.Types["EIP712Domain"] != nil {
		return
	}
	var fs []Field
	if td.Domain.Name != "" {
		fs = append(fs, Field{"name", "string"})
	}
	if td.Domain.Version != "" {
		fs = append(fs, Field{"version", "string"})
	}
	if td.Domain.ChainID != nil {
		fs = append(fs, Field{"chainId", "uint256"})
	}
	if td.Domain.VerifyingContract != "" {
		fs = append(fs, Field{"verifyingContract", "address"})
	}
	if td.Domain.Salt != "" {
		fs = append(fs, Field{"salt", "bytes32"})
	}
	if len(fs) > 0 {
		td.Types["EIP712Domain"] = fs
	}
}

func (td *TypedData) DomainSeparator() ([]byte, error) {
	td.ensureDomainType()
	return (&TypedData{
		Types:       td.Types,
		PrimaryType: "EIP712Domain",
		Domain:      td.Domain,
		Message:     td.Domain.Map(),
	}).HashStruct("EIP712Domain", td.Domain.Map())
}

func (td *TypedData) HashTypedData() ([]byte, error) {
	ds, err := td.DomainSeparator()
	if err != nil {
		return nil, err
	}
	msg, err := td.HashStruct(td.PrimaryType, td.Message)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(0x19)
	buf.WriteByte(0x01)
	buf.Write(ds)
	buf.Write(msg)
	return utils.Keccak(buf.Bytes()), nil
}

/* ---------------- Array & primitive helpers ---------------- */

func (td *TypedData) encodeArray(elemType string, v any) ([]byte, error) {
	s, ok := v.([]any)
	if !ok {
		switch vv := v.(type) {
		case []map[string]any:
			tmp := make([]any, len(vv))
			for i := range vv {
				tmp[i] = vv[i]
			}
			s = tmp
		case [][]byte, []string, []bool, []*big.Int, []uint64, []int64, []int:
			s2 := make([]any, 0)
			switch arr := vv.(type) {
			case [][]byte:
				for _, it := range arr {
					s2 = append(s2, it)
				}
			case []string:
				for _, it := range arr {
					s2 = append(s2, it)
				}
			case []bool:
				for _, it := range arr {
					s2 = append(s2, it)
				}
			case []*big.Int:
				for _, it := range arr {
					s2 = append(s2, it)
				}
			case []uint64:
				for _, it := range arr {
					s2 = append(s2, it)
				}
			case []int64:
				for _, it := range arr {
					s2 = append(s2, it)
				}
			case []int:
				for _, it := range arr {
					s2 = append(s2, it)
				}
			}
			s, ok = s2, true
		default:
			return nil, fmt.Errorf("array expects slice")
		}
	}
	var cat bytes.Buffer
	for _, it := range s {
		if isRefType(elemType) {
			obj, ok := it.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("array elem expects object")
			}
			enc, err := td.EncodeData(elemType, obj)
			if err != nil {
				return nil, err
			}
			cat.Write(utils.Keccak(enc))
		} else {
			w, err := encodePrimitive(elemType, it)
			if err != nil {
				return nil, err
			}
			cat.Write(w)
		}
	}
	return utils.Keccak(cat.Bytes()), nil
}

/* --------- Primitive encoding to 32-byte word --------- */

func encodePrimitive(t string, v any) ([]byte, error) {
	switch t {
	case "address":
		// 20B -> 32B left-pad
		addr := utils.StrToRawAddr(v.(string)) // 20 bytes
		out := make([]byte, 32)
		copy(out[12:], addr)
		return out, nil
	case "bool":
		return utils.BoolToWord(v)
	case "string":
		return utils.StrToWord(v)
	case "bytes":
		return utils.BytesDynToWord(v)
	}
	if strings.HasPrefix(t, "bytes") {
		n, err := strconv.Atoi(t[len("bytes"):])
		if err != nil || n < 1 || n > 32 {
			return nil, fmt.Errorf("invalid %s", t)
		}
		return utils.BytesNToWord(n, v)
	}
	if t == "int" || t == "uint" || strings.HasPrefix(t, "int") || strings.HasPrefix(t, "uint") {
		return utils.IntLikeToWord(t, v)
	}
	return nil, fmt.Errorf("unknown type %q", t)
}
