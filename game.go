package main

import (
	"time"

	"github.com/nsf/termbox-go"
)

const (
	layout_width  = 40
	layout_height = 25
)

type layoutShift struct {
	left int
	top  int
}

var (
	layout = layoutShift{20, 1}
	height = layout_height

	brickSlice = [][][]interface{}{
		{
			{1, 1, 0, 0},
			{1, 0, 0, 0},
			{1, 0, 0, 0},
		},
		{
			{1, 1, 0, 0},
			{0, 1, 0, 0},
			{0, 1, 0, 0},
		},
		{
			{0, 1, 1, 0},
			{1, 1, 0, 0},
			{0, 0, 0, 0},
		},
		{
			{1, 1, 0, 0},
			{0, 1, 1, 0},
			{0, 0, 0, 0},
		},
		{
			{1, 1, 1, 1},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
		{
			{1, 1, 0, 0},
			{1, 1, 0, 0},
			{0, 0, 0, 0},
		},
		{
			{0, 1, 0, 0},
			{1, 1, 1, 0},
			{0, 0, 0, 0},
		},
	}

	backColor = termbox.ColorBlue
)

// left1 top0 y23 x 14/24
func drawBackGround() {
	// term_size, _ := termbox.Size()

	// term_width := term_size.x
	// term_height := term_size.y
	for i := 0; i < layout_height; i++ {
		for j := 0; j < layout_width; j++ {
			termbox.SetCell(j+layout.left, i+layout.top, ' ', backColor, termbox.ColorBlack)
		}
	}

}

func createBrick(metrix [][]interface{}, color termbox.Attribute, shift_x int, shift_y int) {
	for y, s := range metrix {
		for x, v := range s {
			if v == 1 {
				termbox.SetCell(layout.left+layout_width/2+2*x+shift_x, layout.top+y+shift_y, ' ', backColor, color)
				termbox.SetCell(layout.left+layout_width/2+2*x-1+shift_x, layout.top+y+shift_y, ' ', backColor, color)
			}
		}
	}
}

func move(s [][]interface{}, color termbox.Attribute) {
	ticker := time.NewTicker(1000 * time.Millisecond)

	func(t *time.Ticker) {
		<-t.C
		createBrick(s, color, 0, layout_height-height)
	}(ticker)
	height = height - 1
}

func initGame(index int, color termbox.Attribute) {
	termbox.Clear(backColor, backColor)
	drawBackGround()
	move(brickSlice[index], color)
	termbox.Flush()
}
