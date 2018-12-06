package crdt

import (
	"errors"
	"math/big"
)

var (
	bigInt0              big.Int
	bigInt127            big.Int
	errNoTerminatingByte = errors.New("VLQ: no terminating byte")
)

func init() {
	bigInt127.SetUint64(0x7f)
}

func encodeUint32(n uint32) []byte {
	if n == 0 {
		return []byte{0}
	}
	bytes := make([]byte, 0, 4)
	for n > 0 {
		b := byte(n & 0x7f)
		n >>= 7
		if len(bytes) != 0 {
			b |= 0x80
		}
		bytes = append(bytes, b)
	}
	reverse(bytes)
	return bytes
}

func decodeUint32(bytes []byte) (uint32, []byte, error) {
	n := uint32(0)
	for i, b := range bytes {
		d := b & 0x7f
		n = (n << 7) + uint32(d)
		if b < 0x80 {
			return n, bytes[i+1:], nil
		}
	}
	return 0, nil, errNoTerminatingByte
}

func encodeBigInt(x *big.Int) []byte {
	if bigInt0.Cmp(x) == 0 {
		return []byte{0}
	}
	var y, z big.Int
	y.Set(x)
	bytes := make([]byte, 0, 4)
	for bigInt0.Cmp(&y) != 0 {
		z.And(&y, &bigInt127)
		y.Rsh(&y, 7)
		b := byte(z.Uint64())
		if len(bytes) != 0 {
			b |= 0x80
		}
		bytes = append(bytes, b)
	}
	reverse(bytes)
	return bytes
}

func decodeBigInt(bytes []byte) (big.Int, []byte, error) {
	var x, y big.Int
	for i, b := range bytes {
		y.SetUint64(uint64(b & 0x7f))
		x.Lsh(&x, 7)
		x.Add(&x, &y)
		if b < 0x80 {
			return x, bytes[i+1:], nil
		}
	}
	return x, nil, errNoTerminatingByte
}

func reverse(bytes []byte) {
	for i := len(bytes)/2 - 1; i >= 0; i-- {
		j := len(bytes) - 1 - i
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
}
