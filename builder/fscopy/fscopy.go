package fscopy

import (
	"io"
	"os"
	"path/filepath"

	billy "gopkg.in/src-d/go-billy.v4"
)

// Copy copies src to dest, doesn't matter if src is a directory or a file
func Copy(src, dest string, srcFs, dstFs billy.Filesystem) error {
	info, err := srcFs.Stat(src)
	if err != nil {
		return err
	}
	return copyInternal(src, dest, info, srcFs, dstFs)
}

// "info" must be given here, NOT nil.
func copyInternal(src, dest string, info os.FileInfo, srcFs, dstFs billy.Filesystem) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	if info.IsDir() {
		return dcopy(src, dest, info, srcFs, dstFs)
	}
	return fcopy(src, dest, info, srcFs, dstFs)
}

func fcopy(src, dest string, info os.FileInfo, srcFs, dstFs billy.Filesystem) error {

	f, err := dstFs.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	s, err := srcFs.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

func dcopy(src, dest string, info os.FileInfo, srcFs, dstFs billy.Filesystem) error {

	if err := dstFs.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}

	infos, err := srcFs.ReadDir(src)
	if err != nil {
		return err
	}

	for _, info := range infos {
		if err := copyInternal(
			filepath.Join(src, info.Name()),
			filepath.Join(dest, info.Name()),
			info,
			srcFs,
			dstFs,
		); err != nil {
			return err
		}
	}

	return nil
}
