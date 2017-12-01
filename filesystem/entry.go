package filesystem

type Entry interface {
	IsDir() bool
	Name() string
	InodeNum() int64
}
