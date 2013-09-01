package util

func InArray(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

func a() {
	arr := [...]string{"life", "xx", "ou"}
	InArray(arr[0:], "life")
}