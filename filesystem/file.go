package filesystem

import (
	"bytes"
	"encoding/binary"
	"fmt"

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
func (f File) InodeNum() int64 {
	return f.inodeNum
}
func (f *File) Size() uint64 {
	return uint64(f.inode.Size)
}

func (f *File) Buffer() (*bytes.Buffer, error) {
	addrs := make([]uint32, 0)
	for _, addr := range f.inode.Addrs[:diskimage.NDIRECT] {
		if addr == 0 {
			break
		}
		addrs = append(addrs, addr)
	}

	if f.inode.Addrs[diskimage.NDIRECT] != 0 {
		data, err := f.image.GetData(int64(f.inode.Addrs[diskimage.NDIRECT]))
		if err != nil {
			return nil, err
		}

		for x := 0; x < len(data); x += 4 {
			addr := binary.LittleEndian.Uint32(data[x : x+4])
			if addr == 0 {
				break
			}
			addrs = append(addrs, addr)
		}
	}

	fmt.Println("##### READ! ", addrs)

	buffer := bytes.NewBuffer(make([]byte, 0, 512*len(addrs)))
	for _, addr := range addrs {
		data, err := f.image.GetData(int64(addr))
		if err != nil {
			return nil, err
		}

		_, err = buffer.Write(data)
		if err != nil {
			return nil, err
		}
	}

	buffer.Truncate(int(f.inode.Size))
	return buffer, nil
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

	f.inode.Size = uint32(size)
	err := f.image.SetInode(f.inodeNum, f.inode)
	if err != nil {
		return err
	}

	return nil
}
func (f *File) writeBlock(data []byte) (int64, error) {
	i, err := f.image.AllocData()
	if err != nil {
		return -1, err
	}

	err = f.image.SetData(i, data)
	if err != nil {
		return -1, err
	}

	err = f.image.SetBitmap(i, true)
	if err != nil {
		return -1, err
	}

	return i, nil
}
func (f *File) Write(buffer *bytes.Buffer) error {
	if buffer.Len() > 71680 {
		buffer.Truncate(71680)
	}
	buffer = bytes.NewBuffer(bytes.Trim(buffer.Bytes(), "\x00"))

	f.inode.Size = uint32(buffer.Len())

	blocks := make([]uint32, 0, 32)

	for buffer.Len() > 0 {
		data := make([]byte, 512)

		_, err := buffer.Read(data)
		if err != nil {
			return err
		}

		i, err := f.writeBlock(data)
		if err != nil {
			return err
		}

		blocks = append(blocks, uint32(i))
	}

	fmt.Println("##### WRITE! ", blocks)

	copy(f.inode.Addrs[:], blocks[:diskimage.NDIRECT])

	if len(blocks) > diskimage.NDIRECT {
		data := make([]byte, 0, 512)

		for _, addr := range blocks[diskimage.NDIRECT:] {
			buf := make([]byte, 4)
			binary.LittleEndian.PutUint32(buf, addr)
			data = append(data, buf...)
		}

		i, err := f.writeBlock(data)
		if err != nil {
			return err
		}

		f.inode.Addrs[diskimage.NDIRECT] = uint32(i)
	} else {
		f.inode.Addrs[diskimage.NDIRECT] = 0
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
