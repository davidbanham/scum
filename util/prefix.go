package util

func Prefix(strs []string, prefix string) []string {
	ret := []string{}
	for _, str := range strs {
		ret = append(ret, prefix+str)
	}
	return ret
}
