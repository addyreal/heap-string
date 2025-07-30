package heapstr

import (
	"golang.org/x/sys/unix"
)

func openFile(p string) (error, int) {
	fd, err := unix.Open(p, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return err, -1
	}

	return nil, fd
}

func getSizes(stat unix.Stat_t) (error, int, int) {
	size := int(stat.Size)

	ps := unix.Getpagesize()
	aligned := (size + ps - 1) & ^(ps - 1)

	return nil, size, aligned
}

func null(b []byte, n int) {
	for i := 0; i < n; i++ {
		b[i] = 0
	}
}
