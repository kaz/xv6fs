package main

import (
	"log"
	"strings"

	"bitbucket.org/sekai/xv6fs/filesystem"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type xv6fs struct {
	pathfs.FileSystem
	root *filesystem.Directory
}

func Mount(mountPoint string, root *filesystem.Directory) {
	pnfs := pathfs.NewPathNodeFs(&xv6fs{pathfs.NewDefaultFileSystem(), root}, nil)
	conn := nodefs.NewFileSystemConnector(pnfs.Root(), nil)
	server, err := fuse.NewServer(conn.RawFS(), mountPoint, &fuse.MountOptions{
		Name:                 "xv6fs",
		FsName:               "xv6fs",
		DisableXAttrs:        true,
		IgnoreSecurityLabels: true,
	})
	if err != nil {
		panic(err)
	}
	server.Serve()
}

func (x *xv6fs) fetchEntry(name string) filesystem.Entry {
	var currentEntry filesystem.Entry = *x.root

	if name == "" {
		return currentEntry
	}

	for _, fragment := range strings.Split(name, "/") {
		switch current := currentEntry.(type) {
		case filesystem.Directory:
			ents, err := current.Entries()
			if err != nil {
				return nil
			}

			currentEntry = nil
			for _, ent := range ents {
				if ent.Name() == fragment {
					currentEntry = ent
					break
				}
			}
		default:
			return nil
		}
	}

	return currentEntry
}

func (x *xv6fs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	log.Println(">> GetAttr", name)

	switch entry := x.fetchEntry(name).(type) {

	case filesystem.File:
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0644,
			Size: entry.Size(),
		}, fuse.OK

	case filesystem.Directory:
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK

	default:
		return nil, fuse.ENOENT

	}
}
func (x *xv6fs) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	log.Println(">> OpenDir", name)

	switch entry := x.fetchEntry(name).(type) {

	case filesystem.File:
		return nil, fuse.ENOTDIR

	case filesystem.Directory:
		entries, err := entry.Entries()
		if err != nil {
			return nil, fuse.EIO
		}

		fuseEntries := make([]fuse.DirEntry, 0, len(entries))
		for _, ent := range entries {
			mode := fuse.S_IFREG
			if ent.IsDir() {
				mode = fuse.S_IFDIR
			}

			fuseEntries = append(fuseEntries, fuse.DirEntry{
				Name: ent.Name(),
				Mode: uint32(mode),
			})
		}

		return fuseEntries, fuse.OK

	default:
		return nil, fuse.ENOENT

	}
}
func (x *xv6fs) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	log.Println(">> Open", name)

	switch entry := x.fetchEntry(name).(type) {

	case filesystem.File:
		file, err := NewFile(&entry)
		if err != nil {
			return nil, fuse.EIO
		}

		return file, fuse.OK

	default:
		return nil, fuse.ENOENT

	}
}
func (x *xv6fs) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	fragments := strings.Split(name, "/")

	switch entry := x.fetchEntry(strings.Join(fragments[:len(fragments)-1], "/")).(type) {

	case filesystem.Directory:
		rawFile, err := entry.AddFile(fragments[len(fragments)-1])
		if err != nil {
			return nil, fuse.EIO
		}

		file, err := NewFile(rawFile)
		if err != nil {
			return nil, fuse.EIO
		}

		return file, fuse.OK

	default:
		return nil, fuse.ENOTDIR

	}
}
func (x *xv6fs) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	log.Println(">> Mkdir", name)

	fragments := strings.Split(name, "/")

	switch entry := x.fetchEntry(strings.Join(fragments[:len(fragments)-1], "/")).(type) {

	case filesystem.Directory:
		_, err := entry.AddDirectory(fragments[len(fragments)-1])
		if err != nil {
			return fuse.EIO
		}

		return fuse.OK

	default:
		return fuse.ENOTDIR

	}
}

func (x *xv6fs) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	log.Println(">> Unlink", name)

	switch entry := x.fetchEntry(name).(type) {

	case filesystem.File:
		err := entry.Delete()
		if err != nil {
			return fuse.EIO
		}

	default:
		return fuse.EIO

	}

	fragments := strings.Split(name, "/")

	switch entry := x.fetchEntry(strings.Join(fragments[:len(fragments)-1], "/")).(type) {

	case filesystem.Directory:
		err := entry.RemoveEntry(fragments[len(fragments)-1], 0)
		if err != nil {
			return fuse.EIO
		}

	default:
		return fuse.ENOTDIR

	}

	return fuse.OK
}
func (x *xv6fs) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	log.Println(">> Rmdir", name)

	switch entry := x.fetchEntry(name).(type) {

	case filesystem.Directory:
		err := entry.Delete()
		if err != nil {
			return fuse.EIO
		}

	default:
		return fuse.ENOTDIR

	}

	fragments := strings.Split(name, "/")

	switch entry := x.fetchEntry(strings.Join(fragments[:len(fragments)-1], "/")).(type) {

	case filesystem.Directory:
		err := entry.RemoveEntry(fragments[len(fragments)-1], 1)
		if err != nil {
			return fuse.EIO
		}

	default:
		return fuse.ENOTDIR

	}

	return fuse.OK
}
func (x *xv6fs) Rename(oldName string, newName string, context *fuse.Context) (code fuse.Status) {
	log.Println(">> Rename", oldName, newName)

	fragments := strings.Split(oldName, "/")

	switch entry := x.fetchEntry(strings.Join(fragments[:len(fragments)-1], "/")).(type) {

	case filesystem.Directory:
		err := entry.RenameEntry(fragments[len(fragments)-1], newName)
		if err != nil {
			return fuse.EIO
		}

		return fuse.OK

	default:
		return fuse.ENOTDIR

	}
}
func (x *xv6fs) Link(oldName string, newName string, context *fuse.Context) (code fuse.Status) {
	log.Println(">> Link", oldName, newName)

	fragments := strings.Split(newName, "/")

	switch entry := x.fetchEntry(strings.Join(fragments[:len(fragments)-1], "/")).(type) {

	case filesystem.Directory:
		err := entry.LinkEntry(fragments[len(fragments)-1], x.fetchEntry(oldName).InodeNum())
		if err != nil {
			return fuse.EIO
		}

		return fuse.OK

	default:
		return fuse.ENOTDIR

	}
}
