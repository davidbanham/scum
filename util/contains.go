package util

func Contains(str []string, target string) bool {
	for _, s := range str {
		if s == target {
			return true
		}
	}
	return false
}
