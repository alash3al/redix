//go:build linux || darwin

package filesystem

import (
	"io/ioutil"
	"os"
	"syscall"
)

// ReadFileWithSharedLock reads the specified file using a shared-lock
func ReadFileWithSharedLock(filename string) ([]byte, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0775)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fd := int(f.Fd())

	if err := syscall.Flock(fd, syscall.LOCK_SH); err != nil {
		return nil, err
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// WriteFileWithExclusiveLock reads the specified file using a shared-lock
func WriteFileWithExclusiveLock(filename string, data []byte) (int, error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC|os.O_SYNC, 07775)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	fd := int(f.Fd())

	if err := syscall.Flock(fd, syscall.LOCK_EX); err != nil {
		return 0, err
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)

	written, err := f.Write(data)
	if err != nil {
		return 0, err
	}

	return written, nil
}
