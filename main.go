package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1200
	screenHeight = 800

	centerX = screenWidth / 2
	centerY = screenHeight / 2
)

type Game struct {
	star Star
}

func (g *Game) Update() error {
	const (
		dt      = 1.0 / 60.0
		gravity = 100
	)

	dx := centerX - g.star.X
	dy := centerY - g.star.Y

	dist := math.Sqrt(dx*dx + dy*dy)

	dirX := dx / dist
	dirY := dy / dist

	g.star.AX = dirX * gravity
	g.star.AY = dirY * gravity

	g.star.VX += g.star.AX * dt
	g.star.VY += g.star.AY * dt

	g.star.X += g.star.VX * dt
	g.star.Y += g.star.VY * dt

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	vector.DrawFilledCircle(
		screen,
		float32(centerX),
		float32(centerY),
		6,
		color.White,
		false,
	)
	vector.DrawFilledCircle(
		screen,
		float32(g.star.X),
		float32(g.star.Y),
		3,
		color.White,
		false,
	)

	fmt.Printf("\rFPS: %.2f | TPS: %.2f", ebiten.ActualFPS(), ebiten.ActualTPS())
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := &Game{
		star: Star{
			X:    800,
			Y:    400,
			VX:   100,
			VY:   0,
			Mass: 1,
		},
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Galaxy Go")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
