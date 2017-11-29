package diskimage

const (
	NDIRECT = 12
)
const (
	T_DIR = 1 + iota
	T_FILE
	T_DEV
)

type Inode struct {
	Type  int16
	Major int16
	Minor int16
	Nlink int16
	Size  uint32
	Addrs [NDIRECT + 1]uint32
}
