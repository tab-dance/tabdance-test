package main

import (
	"github.com/tab-dance/tabdance-test/internal/server"
)

func main() {
	s := server.New()
	if err := s.Run(); err != nil {
		panic(err)
	}
}
