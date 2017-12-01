package main

import (
	"bytes"
	"log"

	"bitbucket.org/sekai/xv6fs/filesystem"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

type xv6file struct {
	nodefs.File
	entity *filesystem.File
	buffer *bytes.Buffer
}

func NewFile(file *filesystem.File) (*xv6file, error) {
	buffer, err := file.Buffer()
	if err != nil {
		return nil, err
	}
	return &xv6file{nodefs.NewDataFile(buffer.Bytes()), file, buffer}, nil
}

func (f *xv6file) Truncate(size uint64) fuse.Status {
	log.Println(">> Write", f.entity.Name())

	f.buffer.Truncate(int(size))
	err := f.entity.Truncate(size)
	if err != nil {
		return fuse.EIO
	}

	return fuse.OK
}
func (f *xv6file) Write(data []byte, off int64) (uint32, fuse.Status) {
	log.Println(">> Write", f.entity.Name())

	f.buffer.Truncate(int(off))
	written, err := f.buffer.Write(data)
	if err != nil {
		return 0, fuse.EIO
	}

	err = f.entity.Write(f.buffer)
	if err != nil {
		return 0, fuse.EIO
	}

	return uint32(written), fuse.OK
}
