package filesystem

import (
	"bytes"
	"encoding/binary"

	"bitbucket.org/sekai/xv6fs/diskimage"
)

type File struct {
	image    *diskimage.DiskImage
	inode    *diskimage.Inode
	inodeNum int64
	name     string
}

func (f File) IsDir() bool {
	return false
}
func (f File) Name() string {
	return f.name
}
func (f *File) Size() uint64 {
	return uint64(f.inode.Size)
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
func (f *File) Buffer() (*bytes.Buffer, error) {
	data, err := f.read(make([]byte, 0, f.inode.Size/512*512), f.inode.Addrs[:], len(f.inode.Addrs)-1)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(data[:f.inode.Size]), nil
}
func (f *File) Truncate(size uint64) error {
	remain := size / 512
	if size%512 > 0 {
		remain += 1
	}

	for i, addr := range f.inode.Addrs[remain:] {
		err := f.image.SetBitmap(int64(addr), false)
		if err != nil {
			return err
		}

		f.inode.Addrs[i] = 0
	}

	err := f.image.SetInode(f.inodeNum, f.inode)
	if err != nil {
		return err
	}

	return nil
}
func (f *File) Write(buffer *bytes.Buffer) error {
	f.inode.Size = uint32(buffer.Len())

	blocks := make([]uint32, 0, 32)

	for buffer.Len() > 0 {
		data := make([]byte, 512)

		_, err := buffer.Read(data)
		if err != nil {
			return err
		}

		i, err := f.image.AllocData()
		if err != nil {
			return err
		}

		err = f.image.SetData(i, data)
		if err != nil {
			return err
		}

		err = f.image.SetBitmap(i, true)
		if err != nil {
			return err
		}

		blocks = append(blocks, uint32(i))
	}

	copy(f.inode.Addrs[:], blocks[:diskimage.NDIRECT])

	if len(blocks) > diskimage.NDIRECT {
		data := make([]byte, 0, 512)

		for _, addr := range blocks[diskimage.NDIRECT:] {
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, addr)
			data = append(data, buf...)
		}

		i, err := f.image.AllocData()
		if err != nil {
			return err
		}

		err = f.image.SetData(i, data)
		if err != nil {
			return err
		}

		f.inode.Addrs[diskimage.NDIRECT] = uint32(i)
	}

	err := f.image.SetInode(f.inodeNum, f.inode)
	if err != nil {
		return err
	}

	return nil
}
func (f *File) Delete() error {
	f.inode.Nlink -= 1

	if f.inode.Nlink == 0 {
		err := f.Truncate(0)
		if err != nil {
			return err
		}

		f.inode = &diskimage.Inode{}
	}

	return f.image.SetInode(f.inodeNum, f.inode)
}
