package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1200
	screenHeight = 800
)

type Game struct {
	star Star
}

func (g *Game) Update() error {
	const dt = 1.0 / 60.0

	g.star.X += g.star.VX * dt
	g.star.Y += g.star.VY * dt

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
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
			X:    600,
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
