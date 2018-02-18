package getter

import (
	"sync/atomic"

	"errors"

	"os"

	"gopkg.in/src-d/go-billy.v4"
)

var OutOfSpace = errors.New("out of space")

func NewWriteLimitedFilesystem(fs billy.Filesystem, max uint64) billy.Filesystem {
	return &WriteLimitedFilesystem{
		Filesystem: fs,
		written:    new(uint64),
		max:        max,
	}
}

type WriteLimitedFilesystem struct {
	billy.Filesystem
	written *uint64 // use atomic.AddUint64 to increment
	max     uint64
}

func (w *WriteLimitedFilesystem) Chroot(path string) (billy.Filesystem, error) {
	fs, err := w.Filesystem.Chroot(path)
	if err != nil {
		return nil, err
	}
	return &WriteLimitedFilesystem{fs, w.written, w.max}, nil
}

func (w *WriteLimitedFilesystem) TempFile(dir, prefix string) (billy.File, error) {
	f, err := w.Filesystem.TempFile(dir, prefix)
	if err != nil {
		return nil, err
	}
	return &WriteLimitedFile{f, w}, nil
}

func (w *WriteLimitedFilesystem) Create(filename string) (billy.File, error) {
	f, err := w.Filesystem.Create(filename)
	if err != nil {
		return nil, err
	}
	return &WriteLimitedFile{f, w}, nil
}

func (w *WriteLimitedFilesystem) Open(filename string) (billy.File, error) {
	f, err := w.Filesystem.Open(filename)
	if err != nil {
		return nil, err
	}
	return &WriteLimitedFile{f, w}, nil
}

func (w *WriteLimitedFilesystem) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	f, err := w.Filesystem.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}
	return &WriteLimitedFile{f, w}, nil
}

type WriteLimitedFile struct {
	billy.File
	fs *WriteLimitedFilesystem
}

func (w *WriteLimitedFile) Write(p []byte) (n int, err error) {
	if atomic.AddUint64(w.fs.written, uint64(len(p))) > w.fs.max {
		return 0, OutOfSpace
	}
	return w.File.Write(p)
}
