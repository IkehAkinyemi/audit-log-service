package utils

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const alphabets = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomInt generates random integer between min and max.
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generates a random string of length n.
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabets)

	for i := 0; i < n; i++ {
		c := alphabets[rand.Intn(k)]
		sb.WriteByte(c)
	}

	sb.WriteString("-" + strconv.FormatInt(RandomInt(10, 1000), 10))

	return sb.String()
}

// RandomOwner generates a random owner name.
func RandomServiceID() string {
	return RandomString(6)
}
