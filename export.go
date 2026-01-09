package heapstr

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"runtime"
)

type Buffer struct {
	b []byte
	n int
}

func FromFile(p string) (*Buffer, error) {
	// Lock thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Open file
	fd, err := unix.Open(p, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd)

	// Stat file
	var stat unix.Stat_t
	err = unix.Fstat(fd, &stat)
	if err != nil || (stat.Mode&unix.S_IFMT) != unix.S_IFREG {
		return nil, err
	}

	// Check uid
	if stat.Uid != uint32(os.Getuid()) && stat.Uid != 0 {
		return nil, errors.New("Ownership check failed")
	}

	// Check mode
	if stat.Mode&0o777 != 0o400 {
		return nil, errors.New("Permission check failed")
	}

	// Get sizes
	s := int(stat.Size)
	as := (s + unix.Getpagesize() - 1) & ^(unix.Getpagesize() - 1)
	if s <= 0 {
		return nil, errors.New("Size check failed")
	}

	// Mmap
	b, err := unix.Mmap(
		-1, 0, as,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_ANON|unix.MAP_PRIVATE,
	)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 || b == nil {
		return nil, errors.New("Mmap failed")
	}

	// Mlock
	err = unix.Mlock(b)
	if err != nil {
		unix.Munmap(b)
		return nil, err
	}

	// Madvise
	_ = unix.Madvise(b, unix.MADV_DONTDUMP)
	_ = unix.Madvise(b, unix.MADV_WIPEONFORK)

	// Read
	t := 0
	for t < s {
		n, err := unix.Read(fd, b[t:s])
		if err != nil {
			for i := range s {
				b[i] = 0
			}
			unix.Munlock(b)
			unix.Munmap(b)
			return nil, err
		}

		if n == 0 {
			break
		}

		t += n
	}
	if t != s {
		for i := range s {
			b[i] = 0
		}
		unix.Munlock(b)
		unix.Munmap(b)
		return nil, errors.New("Read failed")
	}

	// Msync
	_ = unix.Msync(b, unix.MS_SYNC)

	return &Buffer{b: b, n: s}, nil
}

func (x *Buffer) GetRaw() []byte {
	return x.b[:x.n]
}

func (x *Buffer) Get() []byte {
	a := x.b[:x.n]
	s := len(a) - 1
	if a[s] == '\n' {
		return a[:s]
	}

	return a
}

func (x *Buffer) Wipe() {
	for i := range x.n {
		x.b[i] = 0
	}
	runtime.KeepAlive(x.b)
}

func (x *Buffer) Free() error {
	if x.b == nil {
		return nil
	}

	// Wipe
	x.Wipe()
	_ = unix.Msync(x.b, unix.MS_SYNC)

	// Munlock
	err := unix.Munlock(x.b)
	if err != nil {
		return err
	}

	// Munmap
	err = unix.Munmap(x.b)
	if err != nil {
		return err
	}

	x.b = nil
	return nil
}
