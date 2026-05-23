package main

import (
	"math/rand"
	"time"
)

func main() {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	var keep [][]byte
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		n := 4 + rnd.Intn(12)
		for i := 0; i < n; i++ {
			b := make([]byte, 64*1024+rnd.Intn(512*1024))
			b[0] = byte(rnd.Intn(256))
			keep = append(keep, b)
		}

		if len(keep) > 256 {
			keep = keep[len(keep)-128:]
		}
	}
}
