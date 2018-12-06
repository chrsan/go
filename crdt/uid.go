package crdt

import (
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"math/rand"
)

type UID struct {
	Position big.Int
	Dot      Dot
}

var (
	MinUID = UID{
		Position: *big.NewInt(1 << baseLevel),
		Dot:      Dot{0, 0},
	}
	MaxUID = UID{
		Position: *big.NewInt((1 << (baseLevel + 1)) - 1),
		Dot:      Dot{math.MaxUint32, math.MaxUint32},
	}
)

func UIDBetween(uid1, uid2 *UID, dot *Dot) UID {
	var position big.Int
	position.SetUint64(1)
	for level, sBits := baseLevel, uint(1); level <= maxLevel; level++ {
		position.Lsh(&position, level)
		sBits += level
		pos1, _ := pos(&uid1.Position, level, sBits)
		pos2, ok := pos(&uid2.Position, level, sBits)
		if !ok {
			pos2 = uint((1 << level) - 1)
		}
		var x big.Int
		if pos1+1 < pos2 {
			x.SetUint64(uint64(generatePos(pos1, pos2, level)))
			position.Add(&position, &x)
			return UID{
				Position: position,
				Dot:      *dot,
			}
		}
		x.SetUint64(uint64(pos1))
		position.Add(&position, &x)
	}
	panic(fmt.Sprintf("UID cannot have more than %d levels", maxLevel))
}

func (u UID) Cmp(uid *UID) int {
	pos1, pos2 := u.Position, uid.Position
	bLen1, bLen2 := uint(u.Position.BitLen()), uint(uid.Position.BitLen())
	var x big.Int
	if bLen1 > bLen2 {
		x.Rsh(&pos1, bLen1-bLen2)
		pos1 = x
	} else {
		x.Rsh(&pos2, bLen2-bLen1)
		pos2 = x
	}
	switch pos1.Cmp(&pos2) {
	case -1:
		return -1
	case 1:
		return 1
	}
	if u.Dot.SiteID < uid.Dot.SiteID {
		return -1
	}
	if u.Dot.SiteID > uid.Dot.SiteID {
		return 1
	}
	if u.Dot.Counter < uid.Dot.Counter {
		return -1
	}
	if u.Dot.Counter > uid.Dot.Counter {
		return 1
	}
	return 0
}

func (u UID) ToVLQ() []byte {
	vlq := encodeBigInt(&u.Position)
	vlq = append(vlq, encodeUint32(uint32(u.Dot.SiteID))...)
	vlq = append(vlq, encodeUint32(uint32(u.Dot.Counter))...)
	return vlq
}

func UIDFromVLQ(vlq []byte) (UID, error) {
	position, rest1, err := decodeBigInt(vlq)
	if err != nil {
		return UID{}, err
	}
	siteID, rest2, err := decodeUint32(rest1)
	if err != nil {
		return UID{}, err
	}
	counter, _, err := decodeUint32(rest2)
	if err != nil {
		return UID{}, err
	}
	return UID{position, Dot{SiteID(siteID), Counter(counter)}}, nil
}

func (u UID) String() string {
	vlq := u.ToVLQ()
	return base64.RawURLEncoding.EncodeToString(vlq)
}

func UIDFromBase64String(s string) (UID, error) {
	vlq, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return UID{}, err
	}
	return UIDFromVLQ(vlq)
}

const (
	baseLevel = uint(20)
	maxLevel  = bits.UintSize
	boundary  = uint(40)
	maxUint   = ^uint(0)
	maxInt    = maxUint >> 1
)

func pos(position *big.Int, level, significantBits uint) (uint, bool) {
	bLen := uint(position.BitLen())
	if bLen >= significantBits {
		var mask big.Int
		mask.SetUint64(uint64((1 << level) - 1))
		var pos big.Int
		pos.Rsh(position, bLen-significantBits)
		pos.And(&pos, &mask)
		if pos.BitLen() > bits.UintSize {
			panic(pos.Text(2))
		}
		return uint(pos.Uint64()), true
	}
	return 0, false
}

func generatePos(pos1, pos2, level uint) uint {
	lo, hi := pos1+1, pos1+boundary
	if pos2 < hi {
		hi = pos2
	}
	if lo > maxInt {
		panic(fmt.Sprintf("lo %d is too big", lo))
	}
	if hi > maxInt {
		panic(fmt.Sprintf("hi %d is too big", hi))
	}
	return uint(rand.Intn(int(hi-lo))) + lo
}
