package main

import (
	"flag"
	"log"

	"bitbucket.org/sekai/xv6fs/diskimage"
	"bitbucket.org/sekai/xv6fs/filesystem"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 2 {
		log.Fatalln("Usage: xv6mount <IMAGE_FILE> <MOUNT_POINT>")
	}

	image, err := diskimage.Open(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}
	defer image.Close()

	root, err := filesystem.RootDir(image)
	if err != nil {
		log.Fatalln(err)
	}

	Mount(flag.Arg(1), root)
}
