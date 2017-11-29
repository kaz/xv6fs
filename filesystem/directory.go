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
	return &Directory{&File{image, inode, ""}}, nil
}

func (d Directory) IsDir() bool {
	return true
}

func (d *Directory) Entries() ([]Entry, error) {
	entries := []Entry{}

	data, err := d.Read()
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(data)
	for r.Len() > 0 {
		ent := diskimage.DirEnt{}
		binary.Read(r, binary.LittleEndian, &ent)

		if ent.INum != 0 {
			inode, err := d.image.GetInode(int64(ent.INum))
			if err != nil {
				return nil, err
			}

			f := File{d.image, inode, strings.Trim(string(ent.Name[:]), "\x00")}
			if inode.Type == diskimage.T_DIR {
				entries = append(entries, Directory{&f})
			} else {
				entries = append(entries, f)
			}
		}
	}

	return entries, nil
}
