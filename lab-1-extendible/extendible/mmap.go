package extendible

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func mmapFileRead(path string) (*os.File, []byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}
	if info.Size() <= 0 {
		_ = f.Close()
		return nil, nil, fmt.Errorf("cannot mmap empty file: %s", path)
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, int(info.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}

	return f, data, nil
}

func mmapFileWrite(path string, size int) (*os.File, []byte, error) {
	if size <= 0 {
		return nil, nil, fmt.Errorf("invalid mmap size: %d", size)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, nil, err
	}

	if err := f.Truncate(int64(size)); err != nil {
		_ = f.Close()
		return nil, nil, err
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}

	return f, data, nil
}

func msync(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	_, _, errno := syscall.Syscall(
		syscall.SYS_MSYNC,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		uintptr(syscall.MS_SYNC),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func closeMappedFile(f *os.File, data []byte, sync bool) error {
	var firstErr error

	if sync {
		if err := msync(data); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if len(data) > 0 {
		if err := syscall.Munmap(data); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if f != nil {
		if err := f.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
