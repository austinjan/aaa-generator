package utils

import (
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func CreateDirIfNotExists(path string) error {
	if !FileExists(path) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := CreateDirIfNotExists(filepath.Dir(dst)); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = srcFile.WriteTo(dstFile)
	return err
}
