package diskimage

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type DiskImage struct {
	*SuperBlock
	src *os.File
}

func Open(file string) (*DiskImage, error) {
	src, err := os.OpenFile(file, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	super := &SuperBlock{}
	binary.Read(io.NewSectionReader(src, 512, 512), binary.LittleEndian, super)

	return &DiskImage{super, src}, nil
}
func (di *DiskImage) Close() {
	di.src.Close()
}

func (di *DiskImage) GetInode(n int64) (*Inode, error) {
	pos, err := di.calcInodePosition(n)
	if err != nil {
		return nil, err
	}

	inode := &Inode{}
	err = binary.Read(io.NewSectionReader(di.src, pos, 64), binary.LittleEndian, inode)
	if err != nil {
		return nil, err
	}

	return inode, nil
}
func (di *DiskImage) SetInode(n int64, inode *Inode) error {
	pos, err := di.calcInodePosition(n)
	if err != nil {
		return err
	}

	_, err = di.src.Seek(pos, os.SEEK_SET)
	if err != nil {
		return err
	}

	return binary.Write(di.src, binary.LittleEndian, inode)
}
func (di *DiskImage) AllocInode() (int64, error) {
	for i := int64(1); i < int64(di.NInodes); i++ {
		inode, err := di.GetInode(i)
		if err != nil {
			return -1, err
		}

		if inode.Type == T_UNUSED {
			return i, nil
		}
	}
	return -1, fmt.Errorf("No inodes left")
}

func (di *DiskImage) GetBitmap(n int64) (bool, error) {
	pos, mask, err := di.calcBitmapPosition(n)
	if err != nil {
		return false, err
	}

	buf := make([]byte, 1)
	_, err = di.src.ReadAt(buf, pos)
	if err != nil {
		return false, err
	}

	return buf[0]&mask > 0, nil
}
func (di *DiskImage) SetBitmap(n int64, b bool) error {
	pos, mask, err := di.calcBitmapPosition(n)
	if err != nil {
		return err
	}

	buf := make([]byte, 1)
	_, err = di.src.ReadAt(buf, pos)
	if err != nil {
		return err
	}

	if b {
		buf[0] |= mask
	} else {
		buf[0] &^= mask
	}

	_, err = di.src.WriteAt(buf, pos)
	return err
}

func (di *DiskImage) GetData(n int64) ([]byte, error) {
	pos, err := di.calcDataPosition(n)
	if err != nil {
		return nil, err
	}

	data := make([]byte, 512)
	di.src.ReadAt(data, pos)
	return data, nil
}
func (di *DiskImage) SetData(n int64, data []byte) error {
	pos, err := di.calcDataPosition(n)
	if err != nil {
		return err
	}

	_, err = di.src.WriteAt(data, pos)
	return err
}
func (di *DiskImage) AllocData() (int64, error) {
	for i := di.calcDataStart(); i < int64(di.Size); i++ {
		used, err := di.GetBitmap(i)
		if err != nil {
			return -1, err
		}

		if !used {
			return i, nil
		}
	}
	return -1, fmt.Errorf("No data blocks left")
}
