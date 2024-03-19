package utils

import "os"

func DirExists(path string) bool {
	_, err := os.ReadDir(path)

	return !os.IsNotExist(err)
}
