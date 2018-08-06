package randint

import (
	"math/rand"
	"time"
	"unsafe"
)

const a = "0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewRandomIntString returns random int string
func NewRandomIntString(len int) string {
	if len < 1 {
		return ""
	}

	res := make([]byte, 0, len)

	for i := 0; i < len; i++ {
		res = append(res, a[rand.Intn(10)])
	}

	return *(*string)(unsafe.Pointer(&res))
}
