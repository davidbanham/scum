package util

func Pluralise(num int, singular string, plural string) string {
	if num > 1 {
		return plural
	}
	return singular
}
