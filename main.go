package main

import (
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

func main() {
	termbox.Init()
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(brickSlice))
	color := termbox.Attribute(rand.Int()%8) + 1
	defer termbox.Close()
	for {
		initGame(index, color)
	}
}
