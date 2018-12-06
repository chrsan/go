package crdt

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeZero(t *testing.T) {
	assert.Equal(t, []uint8{0}, EncodeUint32(0))
	assert.Equal(t, []uint8{0}, EncodeBigInt(big.NewInt(0)))
}

func TestEncodeSingleByteValues(t *testing.T) {
	assert.Equal(t, []uint8{43}, EncodeUint32(43))
	assert.Equal(t, []uint8{43}, EncodeBigInt(big.NewInt(43)))
}

func TestEncodeMultibyteValues(t *testing.T) {
	assert.Equal(t, []uint8{130, 249, 67}, EncodeUint32(48323))
	assert.Equal(t, []uint8{130, 249, 67}, EncodeBigInt(big.NewInt(48323)))
}

func TestDecodeZero(t *testing.T) {
	bytes := []uint8{0}
	n1, r1, err1 := DecodeUint32(bytes)
	n2, r2, err2 := DecodeBigInt(bytes)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, n1, uint32(0))
	assert.Equal(t, n2, big.NewInt(0))
	assert.Equal(t, 0, len(r1))
	assert.Equal(t, 0, len(r2))
}

func TestDecodeSingleByteValue(t *testing.T) {
	bytes := []uint8{124}
	n1, r1, err1 := DecodeUint32(bytes)
	n2, r2, err2 := DecodeBigInt(bytes)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, n1, uint32(124))
	assert.Equal(t, n2, big.NewInt(124))
	assert.Equal(t, 0, len(r1))
	assert.Equal(t, 0, len(r2))
}

func TestDecodeMultibyteValue(t *testing.T) {
	bytes := []uint8{130, 249, 67}
	n1, r1, err1 := DecodeUint32(bytes)
	n2, r2, err2 := DecodeBigInt(bytes)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, n1, uint32(48323))
	assert.Equal(t, n2, big.NewInt(48323))
	assert.Equal(t, 0, len(r1))
	assert.Equal(t, 0, len(r2))
}

func TestDecodeMultipleValues(t *testing.T) {
	bytes := []uint8{130, 249, 67, 124, 0}
	n1, r1, err1 := DecodeUint32(bytes)
	n2, r2, err2 := DecodeUint32(r1)
	n3, r3, err3 := DecodeUint32(r2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Nil(t, err3)
	assert.Equal(t, uint32(48323), n1)
	assert.Equal(t, uint32(124), n2)
	assert.Equal(t, uint32(0), n3)
	assert.Equal(t, bytes[3:5], r1)
	assert.Equal(t, bytes[4:5], r2)
	assert.True(t, len(r3) == 0)
	n4, r4, err4 := DecodeBigInt(bytes)
	n5, r5, err5 := DecodeBigInt(r1)
	n6, r6, err6 := DecodeBigInt(r2)
	assert.Nil(t, err4)
	assert.Nil(t, err5)
	assert.Nil(t, err6)
	assert.Equal(t, big.NewInt(48323), n4)
	assert.Equal(t, big.NewInt(124), n5)
	assert.Equal(t, &big0, n6)
	assert.Equal(t, bytes[3:5], r4)
	assert.Equal(t, bytes[4:5], r5)
	assert.True(t, len(r6) == 0)
}

func TestDecodeInvalidValue(t *testing.T) {
	bytes := []uint8{130, 249, 129}
	_, _, err1 := DecodeUint32(bytes)
	_, _, err2 := DecodeBigInt(bytes)
	assert.Error(t, err1)
	assert.Error(t, err2)
}

func TestEncodeDecode(t *testing.T) {
	vlq := EncodeBigInt(big.NewInt(10382))
	vlq = append(vlq, EncodeBigInt(big.NewInt(4834))...)
	vlq = append(vlq, EncodeUint32(81023)...)
	n1, r1, err1 := DecodeBigInt(vlq)
	n2, r2, err2 := DecodeBigInt(r1)
	n3, r3, err3 := DecodeUint32(r2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Nil(t, err3)
	assert.Equal(t, big.NewInt(10382), n1)
	assert.Equal(t, big.NewInt(4834), n2)
	assert.Equal(t, uint32(81023), n3)
	assert.Equal(t, vlq[2:7], r1)
	assert.Equal(t, vlq[4:7], r2)
	assert.True(t, len(r3) == 0)
}
