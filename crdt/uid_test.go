package crdt

import (
	"math"
	"math/big"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dot = Dot{SiteID: 3, Counter: 2}

func TestMin(t *testing.T) {
	assert.Equal(t, "100000000000000000000", MinUID.Position.Text(2))
	assert.Equal(t, SiteID(0), MinUID.Dot.SiteID)
	assert.Equal(t, Counter(0), MinUID.Dot.Counter)
}

func TestMax(t *testing.T) {
	assert.Equal(t, "111111111111111111111", MaxUID.Position.Text(2))
	assert.Equal(t, SiteID(math.MaxUint32), MaxUID.Dot.SiteID)
	assert.Equal(t, Counter(math.MaxUint32), MaxUID.Dot.Counter)
}

func TestCmp(t *testing.T) {
	uid0 := MinUID
	uid1 := newUID(bigInt("0b100000000000000101001"), 8, 382)
	uid2 := newUID(bigInt("0b100000000000101010010"), 1, 5)
	uid3 := newUID(bigInt("0b100000000000101010010"), 1, 5)
	uid4 := newUID(bigInt("0b100000000001011110010"), 4, 4)
	uid5 := newUID(bigInt("0b100000000001011111101"), 4, 4)
	uid6 := MaxUID
	uids := []UID{uid0, uid1, uid2, uid3, uid4, uid5, uid6}
	sort.SliceStable(uids, func(i, j int) bool { return uids[i].Cmp(&uids[j]) == -1 })
	assert.Equal(t, uid0, uids[0])
	assert.Equal(t, uid1, uids[1])
	assert.Equal(t, uid2, uids[2])
	assert.Equal(t, uid3, uids[3])
	assert.Equal(t, uid4, uids[4])
	assert.Equal(t, uid5, uids[5])
	assert.Equal(t, uid6, uids[6])
}

func TestBetweenTrivial(t *testing.T) {
	uid := UIDBetween(&MinUID, &MaxUID, &dot)
	assert.Equal(t, 1, uid.Position.Cmp(bigInt("0b100000000000000000000")))
	assert.Equal(t, -1, uid.Position.Cmp(bigInt("0b111111111111111111111")))
	assert.Equal(t, SiteID(3), uid.Dot.SiteID)
	assert.Equal(t, Counter(2), uid.Dot.Counter)
}

func TestBetweenBasic(t *testing.T) {
	uid1 := newUID(bigInt("0b101111111111111111110"), 1, 1)
	uid2 := newUID(bigInt("0b110000000000000000000"), 1, 1)
	uid := UIDBetween(&uid1, &uid2, &dot)
	assert.Equal(t, bigInt("0b101111111111111111111"), &uid.Position)
}

func TestBetweenMultiLevel(t *testing.T) {
	uid1 := newUID(bigInt("0b111111000000000000000"), 1, 1)
	uid2 := newUID(bigInt("0b111111000000000000001"), 1, 1)
	uid := UIDBetween(&uid1, &uid2, &dot)
	assert.Equal(t, 1, uid.Position.Cmp(bigInt("0b111111000000000000000000000000000000000000")))
	assert.Equal(t, -1, uid.Position.Cmp(bigInt("0b111111000000000000000000000000000000101001")))
}

func TestBetweenSqueeze(t *testing.T) {
	uid1 := newUID(bigInt("0b1111110000000000000000011010100101010101011010101010101010101010"), 1, 1)
	uid2 := newUID(bigInt("0b1111110000000000000000011010100101010101111010101010101010101010"), 1, 1)
	uid := UIDBetween(&uid1, &uid2, &dot)
	assert.Equal(t, bigInt("0b111111000000000000000001101010010101010110"), &uid.Position)
}

func TestBetweenEquals(t *testing.T) {
	uid1 := newUID(bigInt("0b100110011100000000010"), 1, 1)
	uid2 := newUID(bigInt("0b100110011100000000010"), 2, 1)
	uid := UIDBetween(&uid1, &uid2, &dot)
	assert.Equal(t, 1, uid.Position.Cmp(bigInt("0b100110011100000000010000000000000000000000")))
	assert.Equal(t, -1, uid.Position.Cmp(bigInt("0b100110011100000000010000000000000000101001")))
}

func TestBetweenFirstIsShorter(t *testing.T) {
	uid1 := newUID(bigInt("0b111111000000000000000"), 1, 1)
	uid2 := newUID(bigInt("0b111111000000000000000001101010010101010101"), 2, 1)
	uid := UIDBetween(&uid1, &uid2, &dot)
	assert.Equal(t, 1, uid.Position.Cmp(bigInt("0b111111000000000000000000000000000000000000")))
	assert.Equal(t, -1, uid.Position.Cmp(bigInt("0b111111000000000000000000000000000000101001")))
}

func TestBetweenFirstIsLonger(t *testing.T) {
	uid1 := newUID(bigInt("0b111111000000000000000001101010010101010110"), 1, 1)
	uid2 := newUID(bigInt("0b111111000000000000000"), 2, 1)
	uid := UIDBetween(&uid1, &uid2, &dot)
	assert.Equal(t, 1, uid.Position.Cmp(bigInt("0b111111000000000000000001101010010101010110")))
	assert.Equal(t, -1, uid.Position.Cmp(bigInt("0b111111000000000000000001101010010101111111")))
}

func TestBase64(t *testing.T) {
	uid := newUID(bigInt("0b10101010"), 4, 83)
	encoded := uid.String()
	decoded, err := UIDFromBase64String(encoded)
	assert.Nil(t, err)
	assert.Equal(t, "gSoEUw", encoded)
	assert.Equal(t, uid, decoded)
}

func TestInvalidBase64(t *testing.T) {
	_, err := UIDFromBase64String("bjad%%")
	assert.Error(t, err)
}

func newUID(pos *big.Int, siteID SiteID, counter Counter) UID {
	return UID{*pos, Dot{siteID, counter}}
}

func bigInt(s string) *big.Int {
	var x big.Int
	if _, ok := x.SetString(s, 0); !ok {
		panic(s)
	}
	return &x
}
