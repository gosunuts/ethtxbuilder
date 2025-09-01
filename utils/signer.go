package utils

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

type SignFunc func(msgHash []byte) ([]byte, error)

const (
	LegacyTxType     byte = 0x00
	AccessListTxType byte = 0x01 // EIP-2930
	DynamicFeeTxType byte = 0x02 // EIP-1559
	BlobTxType       byte = 0x03 // EIP-4844
	SetCodeTxType    byte = 0x04 // EIP-7702
)

const (
	SignatureLength  = 65 // 64 bytes (r||s) + 1 byte recovery id (y-parity 0/1)
	RecoveryIDIndex  = 64
	HashLength       = 32
	VHomesteadOffset = 27 // homestead style V = 27/28
	VParityMax       = 1  // typed tx V is 0 or 1
)

var (
	// curve params
	secpN     = secp256k1.S256().Params().N
	secpHalfN = new(big.Int).Rsh(secpN, 1) // N/2

	// common big ints
	Big0 = big.NewInt(0)
	Big1 = big.NewInt(1)

	// errors
	ErrInvalidSig    = errors.New("invalid transaction v, r, s values")
	ErrInvalidPubKey = errors.New("invalid public key")
	ErrBadSignature  = errors.New("invalid signature (size or format)")
	ErrBadHash       = errors.New("invalid hash length (want 32 bytes)")
)

// ValidateSignatureValues verifies whether the signature values are valid with
// the given chain rules. The v value is assumed to be either 0 or 1.
func ValidateSignatureValues(v byte, r, s *big.Int, homestead bool) bool {
	// r,s >= 1
	if r.Cmp(Big1) < 0 || s.Cmp(Big1) < 0 {
		return false
	}
	// s <= N (strictly less than N)
	if r.Cmp(secpN) >= 0 || s.Cmp(secpN) >= 0 {
		return false
	}
	// v in {0,1}
	if v > VParityMax {
		return false
	}
	// low-S for Homestead+
	if homestead && s.Cmp(secpHalfN) > 0 {
		return false
	}
	return true
}

// IsLowS reports whether s is in the lower half of the curve order.
func IsLowS(s *big.Int) bool {
	return s.Cmp(secpHalfN) <= 0
}
func RecoverFrom(sighash []byte, R, S, Vb *big.Int, homestead bool) (string, error) {
	if len(sighash) != HashLength {
		return "", ErrBadHash
	}
	if Vb == nil || Vb.BitLen() > 8 {
		return "", ErrInvalidSig
	}
	vRaw := Vb.Uint64()
	if vRaw < VHomesteadOffset || vRaw > VHomesteadOffset+1 {
		return "", fmt.Errorf("%w: V must be 27 or 28 (got %d)", ErrInvalidSig, vRaw)
	}
	vParity := byte(vRaw - VHomesteadOffset)

	if !ValidateSignatureValues(vParity, R, S, homestead) {
		return "", ErrInvalidSig
	}
	// Build compact signature format expected by RecoverCompact:
	// btcsig = [recovery(27/28)] || r(32) || s(32)
	btcsig := make([]byte, SignatureLength)
	btcsig[0] = byte(vRaw) // 27 or 28
	rb, sb := R.Bytes(), S.Bytes()
	copy(btcsig[1+32-len(rb):1+32], rb)
	copy(btcsig[1+64-len(sb):1+64], sb)

	pub, _, err := ecdsa.RecoverCompact(btcsig, sighash)
	if err != nil {
		return "", err
	}
	uncompressed := pub.SerializeUncompressed()
	if len(uncompressed) == 0 || uncompressed[0] != 4 {
		return "", ErrInvalidPubKey
	}
	addr := PubkeyToAddr(uncompressed)
	return RawAddrToStr(addr), nil
}

// Ecrecover returns the uncompressed public key that created the given signature.
func Ecrecover(hash, sig []byte) ([]byte, error) {
	if len(hash) != HashLength {
		return nil, ErrBadHash
	}
	pub, err := sigToPub(hash, sig)
	if err != nil {
		return nil, err
	}
	return pub.SerializeUncompressed(), nil
}

func sigToPub(hash, sig []byte) (*secp256k1.PublicKey, error) {
	if len(hash) != HashLength {
		return nil, ErrBadHash
	}
	if len(sig) != SignatureLength {
		return nil, ErrBadSignature
	}
	vParity := sig[RecoveryIDIndex]
	if vParity > VParityMax {
		return nil, ErrBadSignature
	}
	// Convert to Compact signature: [v(27/28)] || r || s
	btcsig := make([]byte, SignatureLength)
	btcsig[0] = vParity + VHomesteadOffset
	copy(btcsig[1:], sig[:RecoveryIDIndex]) // r||s

	pub, _, err := ecdsa.RecoverCompact(btcsig, hash)
	return pub, err
}

// ParityToV27 converts 0/1 to 27/28.
func ParityToV27(vParity byte) (*big.Int, error) {
	if vParity > VParityMax {
		return nil, ErrInvalidSig
	}
	return big.NewInt(int64(VHomesteadOffset + vParity)), nil
}

// V27ToParity converts 27/28 to 0/1.
func V27ToParity(V *big.Int) (byte, error) {
	if V == nil || V.BitLen() > 8 {
		return 0, ErrInvalidSig
	}
	val := V.Uint64()
	if val < VHomesteadOffset || val > VHomesteadOffset+1 {
		return 0, ErrInvalidSig
	}
	return byte(val - VHomesteadOffset), nil
}
