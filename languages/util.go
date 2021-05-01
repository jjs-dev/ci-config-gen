package languages

import "os"

func checkPathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
