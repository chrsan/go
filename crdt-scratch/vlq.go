package crdt

import (
	"errors"
	"math/big"
)

var (
	big0   big.Int
	big127 = big.NewInt(0x7f)
)

func EncodeUint32(n uint32) []uint8 {
	if n == 0 {
		return []uint8{0}
	}
	s := make([]uint8, 0, 4)
	for n > 0 {
		b := uint8(n & 0x7f)
		n >>= 7
		if len(s) != 0 {
			b |= 0x80
		}
		s = append(s, b)
	}
	reverse(s)
	return s
}

func EncodeBigInt(x *big.Int) []uint8 {
	if isZero(x) {
		return []uint8{0}
	}
	s := make([]uint8, 0, 4)
	var y big.Int
	y.Set(x)
	for !isZero(&y) {
		var z big.Int
		z.And(&y, big127)
		b := uint8(z.Uint64())
		y.Rsh(&y, 7)
		if len(s) != 0 {
			b |= 0x80
		}
		s = append(s, b)
	}
	reverse(s)
	return s
}

func DecodeUint32(bytes []uint8) (uint32, []uint8, error) {
	n := uint32(0)
	for i, b := range bytes {
		d := b & 0x7f
		n = (n << 7) + uint32(d)
		if b < 0x80 {
			return n, bytes[i+1:], nil
		}
	}
	return 0, nil, errors.New("VLQ: no terminating byte")
}

func DecodeBigInt(bytes []uint8) (*big.Int, []uint8, error) {
	x := new(big.Int)
	for i, b := range bytes {
		d := b & 0x7f
		x.Lsh(x, 7)
		x.Add(x, big.NewInt(int64(d)))
		if b < 0x80 {
			return x, bytes[i+1:], nil
		}
	}
	return nil, nil, errors.New("VLQ: no terminating byte")
}

func isZero(n *big.Int) bool {
	return n.Cmp(&big0) == 0
}

func reverse(s []uint8) {
	for i := len(s)/2 - 1; i >= 0; i-- {
		j := len(s) - 1 - i
		s[i], s[j] = s[j], s[i]
	}
}
