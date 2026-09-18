package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1200
	screenHeight = 800

	centerX = screenWidth / 2
	centerY = screenHeight / 2

	dt          = 1.0 / 60.0
	G           = 1000.0
	centralMass = 100.0
	epsilon     = 20.0
)

func updateStar(star *Star) {
	dx := centerX - star.X
	dy := centerY - star.Y

	distSquared := dx*dx + dy*dy
	softened := math.Pow(distSquared+epsilon*epsilon, 1.5)

	star.AX = G * centralMass * dx / softened
	star.AY = G * centralMass * dy / softened

	star.VX += star.AX * dt
	star.VY += star.AY * dt

	star.X += star.VX * dt
	star.Y += star.VY * dt

}

func newRandomStar(rng *rand.Rand) Star {
	const (
		meanRadius = 250.0
		radiusStd  = 80.0

		meanMass = 1.0
		massStd  = 0.2

		velocityDispersion = 0.05
		radialDispersion   = 2.0
	)

	radius := meanRadius + radiusStd*rng.NormFloat64()

	if radius < 30 {
		radius = 30
	}

	if radius > 500 {
		radius = 500
	}

	angle := rng.Float64() * 2 * math.Pi
	x := centerX + radius*math.Cos(angle)
	y := centerY + radius*math.Sin(angle)

	mass := 1.0 + 0.2*rng.NormFloat64()

	if mass < 0.1 {
		mass = 0.1
	}

	speedNoise := 1.0 + velocityDispersion*rng.NormFloat64()
	orbitalSpeed := math.Sqrt(G * centralMass / radius)
	orbitalSpeed *= speedNoise

	vx := -math.Sin(angle) * orbitalSpeed
	vy := math.Cos(angle) * orbitalSpeed

	radialVelocity := radialDispersion * rng.NormFloat64()

	vx += radialVelocity * math.Cos(angle)
	vy += radialVelocity * math.Sin(angle)

	return Star{
		X:    x,
		Y:    y,
		VX:   vx,
		VY:   vy,
		Mass: mass,
	}
}

type Game struct {
	star []Star
}

func (g *Game) Update() error {
	for i := range g.star {
		updateStar(&g.star[i])
	}

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
	for i := range g.star {
		star := &g.star[i]
		vector.DrawFilledCircle(
			screen,
			float32(star.X),
			float32(star.Y),
			3,
			color.White,
			false,
		)
	}

	fmt.Printf("\rFPS: %.2f | TPS: %.2f", ebiten.ActualFPS(), ebiten.ActualTPS())
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	const starCount = 100

	stars := make([]Star, starCount)

	for i := range stars {
		stars[i] = newRandomStar(rng)
	}

	game := &Game{
		star: stars,
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Galaxy Go")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
