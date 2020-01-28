package main

import (
	"math/rand"
	"time"
)

func generateUUID() string {
	rand.Seed(time.Now().UnixNano())
	s := ""

	for i := 0; i < 5; i++ {
		t := ""
		for j := 0; j < 5; j++ {
			t += string(rand.Int()%26 + 65)
		}
		s += t + "-"
	}

	return s[0:29]
}
