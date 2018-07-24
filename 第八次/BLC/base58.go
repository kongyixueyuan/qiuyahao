package BLC

import (
	"bytes"
	"math/big"
)

var QYH_b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Base58Encode encodes a byte array to Base58
func QYH_Base58Encode(input []byte) []byte {

	var result []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(QYH_b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, QYH_b58Alphabet[mod.Int64()])
	}

	// https://en.bitcoin.it/wiki/Base58Check_encoding#Version_bytes
	if input[0] == 0x00 {
		result = append(result, QYH_b58Alphabet[0])
	}

	ReverseBytes(result)

	return result
}

// Base58Decode decodes Base58-encoded data
func QYH_Base58Decode(input []byte) []byte {
	result := big.NewInt(0)

	for _, b := range input {
		charIndex := bytes.IndexByte(QYH_b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()

	if input[0] == QYH_b58Alphabet[0] {
		decoded = append([]byte{0x00}, decoded...)
	}

	return decoded
}
