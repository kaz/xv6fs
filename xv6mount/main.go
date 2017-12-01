package main

import (
	"flag"
	"fmt"

	"bitbucket.org/sekai/xv6fs/diskimage"
	"bitbucket.org/sekai/xv6fs/filesystem"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 2 {
		fmt.Println("Usage: xv6mount <IMAGE_FILE> <MOUNT_POINT>")
		return
	}

	image, err := diskimage.Open(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	defer image.Close()

	root, err := filesystem.RootDir(image)
	if err != nil {
		panic(err)
	}

	Mount(flag.Arg(1), root)
}
