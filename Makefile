.PHONY: check
check:
	go fmt ./...
	go vet ./...

.PHONY: clean
clean:
	rm xv6mount/xv6mount

.PHONY: mount
mount: xv6mount
	umount ~/Desktop/xv6fs
	./xv6mount/xv6mount fs.img ~/Desktop/xv6fs

xv6mount: **/*.go
	cd xv6mount && go build
