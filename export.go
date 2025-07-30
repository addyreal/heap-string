package heapstr

import (
	"os"
	"errors"
	"runtime"
	"golang.org/x/sys/unix"
)

func FromFile(p string) (error, *Buffer) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err, fd := openFile(p)
	if err != nil {
		return err, nil
	}
	defer unix.Close(fd)

	var stat unix.Stat_t
	err = unix.Fstat(fd, &stat)
	if err != nil || (stat.Mode & unix.S_IFMT) != unix.S_IFREG {
		return err, nil
	}

	uid := uint32(os.Getuid())
	if stat.Uid != uid && stat.Uid != 0 {
		return errors.New("Ownership check failed"), nil
	}

	if stat.Mode&0o777 != 0o600 {
		return errors.New("Permission check failed"), nil
	}

	err, s, as := getSizes(stat)
	if s <= 0 {
		return errors.New("Size check failed"), nil
	}

	b, err := unix.Mmap(
		-1, 0, as,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_ANON|unix.MAP_PRIVATE,
	)
	if err != nil {
		return err, nil
	}

	if len(b) == 0 || b == nil {
		return errors.New("Mmap failed"), nil
	}

	err = unix.Mlock(b)
	if err != nil {
		unix.Munmap(b)
		return err, nil
	}

	_ = unix.Madvise(b, unix.MADV_DONTDUMP)
	_ = unix.Madvise(b, unix.MADV_WIPEONFORK)

	t := 0
	for t < s {
		n, err := unix.Read(fd, b[t:s])
		if err != nil {
			wipe(b)
			unix.Munlock(b)
			unix.Munmap(b)
			return err, nil
		}

		if n == 0 {
			break
		}

		t += n
	}

	if t != s {
		wipe(b)
		unix.Munlock(b)
		unix.Munmap(b)
		return errors.New("Read failed"), nil 
	}

	_ = unix.Msync(b, unix.MS_SYNC)
	return nil, &Buffer{b: b[:s]}
}

func (x *Buffer) Get() []byte {
	return x.b
}

func (x *Buffer) Wipe() {
	wipe(x.b)
}

func (x *Buffer) Free() error {
	if x.b == nil {
		return nil
	}

	wipe(x.b)
	_ = unix.Msync(x.b, unix.MS_SYNC)

	err := unix.Munlock(x.b)
	if err != nil {
		return err
	}

	err = unix.Munmap(x.b)
	if err != nil {
		return err
	}

	x.b = nil
	return nil
}
