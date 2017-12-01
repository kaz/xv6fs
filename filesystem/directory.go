package filesystem

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"bitbucket.org/sekai/xv6fs/diskimage"
)

type Directory struct {
	*File
}

func RootDir(image *diskimage.DiskImage) (*Directory, error) {
	inode, err := image.GetInode(1)
	if err != nil {
		return nil, err
	}
	if inode.Type != diskimage.T_DIR {
		return nil, fmt.Errorf("Expected Type=%d, but actual TYPE=%d", diskimage.T_DIR, inode.Type)
	}
	return &Directory{&File{image, inode, 1, ""}}, nil
}

func (d Directory) IsDir() bool {
	return true
}

func (d *Directory) RemoveEntry(name string) error {
	newBuffer := bytes.NewBuffer([]byte{})
	buffer, err := d.Buffer()
	if err != nil {
		return err
	}

	for buffer.Len() > 0 {
		ent := diskimage.DirEnt{}
		binary.Read(buffer, binary.LittleEndian, &ent)

		if strings.Trim(string(ent.Name[:]), "\x00") != name {
			binary.Write(newBuffer, binary.LittleEndian, &ent)
		}
	}

	d.Truncate(0)
	if err != nil {
		return err
	}

	return d.Write(newBuffer)
}
func (d *Directory) AddFile(name string) (*File, error) {
	inodeNum, err := d.image.AllocInode()
	if err != nil {
		return nil, err
	}

	ent := diskimage.DirEnt{INum: uint16(inodeNum)}
	copy(ent.Name[:], []byte(name))

	buffer, err := d.Buffer()
	if err != nil {
		return nil, err
	}

	binary.Write(buffer, binary.LittleEndian, &ent)
	if err != nil {
		return nil, err
	}

	d.Truncate(0)
	if err != nil {
		return nil, err
	}

	d.Write(buffer)
	if err != nil {
		return nil, err
	}

	return &File{
		image: d.image,
		inode: &diskimage.Inode{
			Type:  diskimage.T_FILE,
			Nlink: 1,
		},
		inodeNum: inodeNum,
		name:     name,
	}, nil
}

func (d *Directory) Entries() ([]Entry, error) {
	entries := []Entry{}

	buffer, err := d.Buffer()
	if err != nil {
		return nil, err
	}

	for buffer.Len() > 0 {
		ent := diskimage.DirEnt{}
		binary.Read(buffer, binary.LittleEndian, &ent)

		if ent.INum != 0 {
			inode, err := d.image.GetInode(int64(ent.INum))
			if err != nil {
				return nil, err
			}

			f := File{
				image:    d.image,
				inode:    inode,
				inodeNum: int64(ent.INum),
				name:     strings.Trim(string(ent.Name[:]), "\x00"),
			}
			if inode.Type == diskimage.T_DIR {
				entries = append(entries, Directory{&f})
			} else {
				entries = append(entries, f)
			}
		}
	}

	return entries, nil
}
