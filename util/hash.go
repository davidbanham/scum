package util

import (
	"crypto/md5"
	"encoding/hex"
)

func Hash(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}
