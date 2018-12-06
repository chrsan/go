package crdt

import (
	"math"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dot Dot

func init() {
	dot = Dot{
		SiteID:  3,
		Counter: 2,
	}
}

func TestMin(t *testing.T) {
	assert.Equal(t, "100000000000000000000", MinUID.Position.Text(2))
	assert.Equal(t, SiteID(0), MinUID.SiteID)
	assert.Equal(t, Counter(0), MinUID.Counter)
}

func TestMax(t *testing.T) {
	assert.Equal(t, "111111111111111111111", MaxUID.Position.Text(2))
	assert.Equal(t, SiteID(math.MaxUint32), MaxUID.SiteID)
	assert.Equal(t, Counter(math.MaxUint32), MaxUID.Counter)
}

func TestCmp(t *testing.T) {
	uid0 := MinUID
	uid1 := UID{Position: bigInt("100000000000000101001"), SiteID: 8, Counter: 382}
	uid2 := UID{Position: bigInt("100000000000101010010"), SiteID: 1, Counter: 5}
	uid3 := UID{Position: bigInt("100000000000101010010"), SiteID: 1, Counter: 5}
	uid4 := UID{Position: bigInt("100000000001011110010"), SiteID: 4, Counter: 4}
	uid5 := UID{Position: bigInt("100000000001011111101"), SiteID: 4, Counter: 4}
	uid6 := MaxUID
	uids := []UID{uid0, uid1, uid2, uid3, uid4, uid5, uid6}
	sort.SliceStable(uids, func(i, j int) bool { return uids[i].Cmp(uids[j]) == -1 })
	assert.Equal(t, uid0, uids[0])
	assert.Equal(t, uid1, uids[1])
	assert.Equal(t, uid2, uids[2])
	assert.Equal(t, uid3, uids[3])
	assert.Equal(t, uid4, uids[4])
	assert.Equal(t, uid5, uids[5])
	assert.Equal(t, uid6, uids[6])
}

func TestBetweenTrivial(t *testing.T) {
	uid := NewUID(MinUID, MaxUID, dot)
	assert.Equal(t, -1, bigInt("100000000000000000000").Cmp(uid.Position))
	assert.Equal(t, 1, bigInt("111111111111111111111").Cmp(uid.Position))
	assert.Equal(t, SiteID(3), uid.SiteID)
	assert.Equal(t, Counter(2), uid.Counter)
}

func TestBetweenBasic(t *testing.T) {
	uid1 := UID{Position: bigInt("101111111111111111110"), SiteID: 1, Counter: 1}
	uid2 := UID{Position: bigInt("110000000000000000000"), SiteID: 1, Counter: 1}
	uid := NewUID(uid1, uid2, dot)
	assert.Equal(t, bigInt("101111111111111111111"), uid.Position)
}

func TestBetweenMultiLevel(t *testing.T) {
	uid1 := UID{Position: bigInt("111111000000000000000"), SiteID: 1, Counter: 1}
	uid2 := UID{Position: bigInt("111111000000000000001"), SiteID: 1, Counter: 1}
	uid := NewUID(uid1, uid2, dot)
	assert.Equal(t, 1, uid.Position.Cmp(bigInt("111111000000000000000000000000000000000000")))
	assert.Equal(t, -1, uid.Position.Cmp(bigInt("111111000000000000000000000000000000101001")))
}

func TestBetweenSqueeze(t *testing.T) {
	uid1 := UID{Position: bigInt("1111110000000000000000011010100101010101011010101010101010101010"), SiteID: 1, Counter: 1}
	uid2 := UID{Position: bigInt("1111110000000000000000011010100101010101111010101010101010101010"), SiteID: 1, Counter: 1}
	uid := NewUID(uid1, uid2, dot)
	assert.Equal(t, bigInt("111111000000000000000001101010010101010110"), uid.Position)
}

func TestBetweenEquals(t *testing.T) {
	uid1 := UID{Position: bigInt("100110011100000000010"), SiteID: 1, Counter: 1}
	uid2 := UID{Position: bigInt("100110011100000000010"), SiteID: 2, Counter: 1}
	uid := NewUID(uid1, uid2, dot)
	assert.Equal(t, 1, uid.Position.Cmp(bigInt("100110011100000000010000000000000000000000")))
	assert.Equal(t, -1, uid.Position.Cmp(bigInt("100110011100000000010000000000000000101001")))
}

func TestBetweenFirstIsShorter(t *testing.T) {
	uid1 := UID{Position: bigInt("111111000000000000000"), SiteID: 1, Counter: 1}
	uid2 := UID{Position: bigInt("111111000000000000000001101010010101010101"), SiteID: 2, Counter: 1}
	uid := NewUID(uid1, uid2, dot)
	assert.Equal(t, 1, uid.Position.Cmp(bigInt("111111000000000000000000000000000000000000")))
	assert.Equal(t, -1, uid.Position.Cmp(bigInt("111111000000000000000000000000000000101001")))
}

func TestBetweenFirstIsLonger(t *testing.T) {
	uid1 := UID{Position: bigInt("111111000000000000000001101010010101010110"), SiteID: 1, Counter: 1}
	uid2 := UID{Position: bigInt("111111000000000000000"), SiteID: 2, Counter: 1}
	uid := NewUID(uid1, uid2, dot)
	assert.Equal(t, 1, uid.Position.Cmp(bigInt("111111000000000000000001101010010101010110")))
	assert.Equal(t, -1, uid.Position.Cmp(bigInt("111111000000000000000001101010010101111111")))
}

/*
func bigInt(s string) *big.Int {
	var z big.Int
	if _, ok := z.SetString(s, 2); !ok {
		panic(s)
	}
	return &z
}
*/
