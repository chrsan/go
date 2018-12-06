package crdt

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"math/rand"
)

const (
	baseLevel uint = 20
	maxLevel  uint = 64
	boundary  uint = 40
)

type UID struct {
	Position *big.Int
	SiteID   SiteID
	Counter  Counter
}

var (
	MaxUID UID
	MinUID UID
)

func init() {
	MinUID = UID{
		Position: big.NewInt(1 << baseLevel),
		SiteID:   SiteID(0),
		Counter:  Counter(0),
	}
	MaxUID = UID{
		Position: big.NewInt((1 << (baseLevel + 1)) - 1),
		SiteID:   SiteID(math.MaxUint32),
		Counter:  Counter(math.MaxUint32),
	}
}

func NewUID(uid1, uid2 UID, dot Dot) UID {
	position1, position2 := uid1.Position, uid2.Position
	position, sBits := big.NewInt(1), uint(1)
	for level := baseLevel; level <= maxLevel; level++ {
		sBits += level
		pos1, ok := getPos(position1, level, sBits)
		if !ok {
			pos1 = 0
		}
		pos2, ok := getPos(position2, level, sBits)
		if !ok {
			pos2 = uint64((1 << level) - 1)
		}
		var z big.Int
		if pos1+1 < pos2 {
			pos := generatePos(pos1, pos2, level)
			z.SetUint64(uint64(pos))
			position.Lsh(position, level)
			position.Add(position, &z)
			return UID{
				Position: position,
				SiteID:   dot.SiteID,
				Counter:  dot.Counter,
			}
		}
		z.SetUint64(uint64(pos1))
		position.Lsh(position, level)
		position.Add(position, &z)
	}
	panic(fmt.Sprintf("UID cannot have more than %d levels", maxLevel))
}

func (u UID) Dot() Dot {
	return Dot{
		SiteID:  u.SiteID,
		Counter: u.Counter,
	}
}

func (u UID) Cmp(uid UID) int {
	pos1, pos2 := u.Position, uid.Position
	bLen1, bLen2 := u.Position.BitLen(), uid.Position.BitLen()
	var z big.Int
	if bLen1 > bLen2 {
		pos1 = z.Rsh(pos1, uint(bLen1-bLen2))
	} else {
		pos2 = z.Rsh(pos2, uint(bLen2-bLen1))
	}
	switch pos1.Cmp(pos2) {
	case -1:
		return -1
	case 1:
		return 1
	}
	if u.SiteID < uid.SiteID {
		return -1
	}
	if u.SiteID > uid.SiteID {
		return 1
	}
	if u.Counter < uid.Counter {
		return -1
	}
	if u.Counter > uid.Counter {
		return 1
	}
	return 0
}

func (u UID) ToVLQ() []uint8 {
	vlq := EncodeBigInt(u.Position)
	vlq = append(vlq, EncodeUint32(uint32(u.SiteID))...)
	vlq = append(vlq, EncodeUint32(uint32(u.Counter))...)
	return vlq
}

func UIDFromVLQ(vlq []uint8) (UID, error) {
	pos, rest1, err := DecodeBigInt(vlq)
	if err != nil {
		return UID{}, err
	}
	s, rest2, err := DecodeUint32(rest1)
	if err != nil {
		return UID{}, err
	}
	c, _, err := DecodeUint32(rest2)
	if err != nil {
		return UID{}, err
	}
	return UID{
		Position: pos,
		SiteID:   SiteID(s),
		Counter:  Counter(c),
	}, nil
}

func getPos(position *big.Int, level, sBits uint) (uint64, bool) {
	bLen := uint(position.BitLen())
	if bLen >= sBits {
		var mask big.Int
		mask.SetUint64((1 << level) - 1)
		var pos big.Int
		pos.Rsh(position, uint(bLen-sBits))
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
	return uint(rand.Intn(int(hi-lo))) + lo
}
