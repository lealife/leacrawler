package util

func InArray(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}