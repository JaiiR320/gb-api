package api

import (
	"crypto/rand"
	"encoding/hex"
)

func UUID() string {
	src := make([]byte, 8)
	n, _ := rand.Read(src) // ignore error as per docs

	dst := make([]byte, hex.EncodedLen(n))
	hex.Encode(dst, src)

	str := string(dst)

	return str[0:4] + "-" + str[4:8] + "-" + str[8:12] + "-" + str[12:]
}
