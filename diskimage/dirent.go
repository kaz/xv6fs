package diskimage

const (
	DIRSIZ = 14
)

type DirEnt struct {
	INum uint16
	Name [DIRSIZ]byte
}
