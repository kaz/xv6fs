package diskimage

import (
	"fmt"
)

type SuperBlock struct {
	Size       uint32
	NBlocks    uint32
	NInodes    uint32
	NLog       uint32
	LogStart   uint32
	InodeStart uint32
	BmapStart  uint32
}

func (sb *SuperBlock) calcInodeBlocks() int64 {
	return int64(sb.NInodes)/8 + 1
}
func (sb *SuperBlock) calcInodePosition(n int64) (int64, error) {
	if n < 0 {
		return 0, fmt.Errorf("MUST BE: n >= 0")
	}
	if n >= int64(sb.NInodes) {
		return 0, fmt.Errorf("MUST BE: n < %d", sb.NInodes)
	}
	return 512*int64(sb.InodeStart) + 64*n, nil
}

func (sb *SuperBlock) calcBitmapBlocks() int64 {
	return int64(sb.Size)/(512*8) + 1
}
func (sb *SuperBlock) calcBitmapPosition(n int64) (int64, byte, error) {
	if n < 0 {
		return 0, 0, fmt.Errorf("MUST BE: n >= 0")
	}
	if n >= int64(sb.Size) {
		return 0, 0, fmt.Errorf("MUST BE: n < %d", sb.Size)
	}
	return 512*int64(sb.BmapStart) + n/8, 1 << byte(n%8), nil
}

func (sb *SuperBlock) calcDataStart() int64 {
	return 2 + int64(sb.NLog) + sb.calcInodeBlocks() + sb.calcBitmapBlocks()
}
func (sb *SuperBlock) calcDataPosition(n int64) (int64, error) {
	dataBlockStart := sb.calcDataStart()
	if n < dataBlockStart {
		return 0, fmt.Errorf("MUST BE: n >= %d", dataBlockStart)
	}
	if n >= int64(sb.Size) {
		return 0, fmt.Errorf("MUST BE: n < %d", sb.Size)
	}
	return 512 * n, nil
}
