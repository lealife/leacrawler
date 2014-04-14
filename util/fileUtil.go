package util

import "os"

func IsExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
        return false;
    }
    return true
}
