package util

func Truncate(num int, input string) string {
	if len(input) < num {
		return input
	}
	return input[:num]
}
