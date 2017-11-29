package filesystem

import (
	"encoding/binary"

	"bitbucket.org/sekai/xv6fs/diskimage"
)

type File struct {
	image *diskimage.DiskImage
	inode *diskimage.Inode
	name  string
}

func (f File) IsDir() bool {
	return false
}
func (f File) Name() string {
	return f.name
}

func (f *File) read(buf []byte, addrs []uint32, indirect int) ([]byte, error) {
	for i, addr := range addrs {
		if addr == 0 {
			break
		}

		data, err := f.image.GetData(int64(addr))
		if err != nil {
			return nil, err
		}

		if i == indirect {
			indirectAddrs := make([]uint32, 0, len(data)/4)
			for x := 0; x < len(data); x += 4 {
				indirectAddrs = append(indirectAddrs, binary.LittleEndian.Uint32(data[x:x+4]))
			}
			return f.read(buf, indirectAddrs, -1)
		}

		buf = append(buf, data...)
	}
	return buf, nil
}
func (f *File) Read() ([]byte, error) {
	data, err := f.read(make([]byte, 0, f.inode.Size/512*512), f.inode.Addrs[:], len(f.inode.Addrs)-1)
	if err != nil {
		return nil, err
	}
	return data[:f.inode.Size], nil
}

func (f *File) Size() uint64 {
	return uint64(f.inode.Size)
}
