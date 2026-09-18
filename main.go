package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
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

func circularSpeed(radius float64) float64 {
	softenedRadius := math.Pow(radius*radius+epsilon*epsilon, 1.5)
	orbitalSpeed := math.Sqrt(G * centralMass * radius * radius / softenedRadius)
	return orbitalSpeed
}

func truncatedNormal(rng *rand.Rand, limit float64) float64 {
	for {
		z := rng.NormFloat64()
		if math.Abs(z) <= limit {
			return z
		}
	}
}

func luminosityFromMass(mass float64) float64 {
	return math.Pow(mass, 3.5)
}

func starRenderRadius(star *Star) float32 {
	brightness := math.Log1p(star.Luminosity)
	radius := 1.0 + 0.7*brightness

	if radius > 3.0 {
		radius = 3.0
	}
	return float32(radius)
}

func starColor(star *Star) color.Color {
	if star.Population == BulgeStar {
		return color.RGBA{
			R: 255,
			G: 210,
			B: 140,
			A: 255,
		}
	}

	switch {
	case star.Mass < 0.8:
		return color.RGBA{
			R: 255,
			G: 190,
			B: 140,
			A: 255,
		}
	case star.Mass < 1.2:
		return color.RGBA{
			R: 255,
			G: 244,
			B: 220,
			A: 255,
		}
	default:
		return color.RGBA{
			R: 190,
			G: 215,
			B: 255,
			A: 255,
		}
	}
}

func newBulgeStar(rng *rand.Rand) Star {
	const (
		bulgeSigma     = 45.0
		maxBulgeRadius = 130.0

		meanMass = 1.0
		massStd  = 0.2
	)

	var dx, dy, radius float64

	for {
		dx = bulgeSigma * rng.NormFloat64()
		dy = bulgeSigma * rng.NormFloat64()

		radius = math.Sqrt(dx*dx + dy*dy)

		if radius <= maxBulgeRadius {
			break
		}
	}

	x := centerX + dx
	y := centerY + dy

	angle := math.Atan2(dy, dx)
	vCircular := circularSpeed(radius)
	tangentialSpeed := vCircular * (1.0 + 0.15*truncatedNormal(rng, 2.5))

	direction := 1.0

	if rng.Float64() < 0.5 {
		direction = -1.0
	}

	tangentialSpeed *= direction

	vx := -math.Sin(angle) * tangentialSpeed
	vy := math.Cos(angle) * tangentialSpeed

	radialVelocity := 0.15 * vCircular * truncatedNormal(rng, 2.5)

	vx += radialVelocity * math.Sin(angle)
	vy += radialVelocity * math.Cos(angle)

	mass := meanMass + massStd*rng.NormFloat64()

	if mass < 0.1 {
		mass = 0.1
	}

	luminosity := luminosityFromMass(mass)

	return Star{
		X:          x,
		Y:          y,
		VX:         vx,
		VY:         vy,
		Mass:       mass,
		Luminosity: luminosity,
		Population: BulgeStar,
	}
}

func newDiskStar(rng *rand.Rand) Star {
	const (
		diskScale = 120.0
		maxRadius = 380.0

		meanMass = 1.0
		massStd  = 0.2

		velocityDispersion = 0.05
		radialDispersion   = 2.0
	)

	var radius float64

	for {
		r1 := -diskScale * math.Log(1-rng.Float64())
		r2 := -diskScale * math.Log(1-rng.Float64())

		radius = r1 + r2

		if radius <= maxRadius {
			break
		}
	}

	angle := rng.Float64() * 2 * math.Pi
	x := centerX + radius*math.Cos(angle)
	y := centerY + radius*math.Sin(angle)

	mass := meanMass + massStd*rng.NormFloat64()

	if mass < 0.1 {
		mass = 0.1
	}

	speedNoise := 1.0 + velocityDispersion*rng.NormFloat64()
	orbitalSpeed := circularSpeed(radius) * speedNoise

	vx := -math.Sin(angle) * orbitalSpeed
	vy := math.Cos(angle) * orbitalSpeed

	radialVelocity := radialDispersion * rng.NormFloat64()

	vx += radialVelocity * math.Cos(angle)
	vy += radialVelocity * math.Sin(angle)

	luminosity := luminosityFromMass(mass)

	return Star{
		X:          x,
		Y:          y,
		VX:         vx,
		VY:         vy,
		Mass:       mass,
		Luminosity: luminosity,
		Population: DiskStar,
	}
}

type Game struct {
	stars []Star
}

func (g *Game) Update() error {
	workerCount := runtime.GOMAXPROCS(0)

	if workerCount > len(g.stars) {
		workerCount = len(g.stars)
	}

	chunkSize := (len(g.stars) + workerCount - 1) / workerCount

	var wg sync.WaitGroup

	for worker := 0; worker < workerCount; worker++ {
		start := worker * chunkSize
		end := start + chunkSize
		if end > len(g.stars) {
			end = len(g.stars)
		}

		if start >= len(g.stars) {
			break
		}

		wg.Add(1)

		go func(start, end int) {
			defer wg.Done()

			for i := start; i < end; i++ {
				updateStar(&g.stars[i])
			}
		}(start, end)
	}

	wg.Wait()

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
	for i := range g.stars {
		star := &g.stars[i]

		vector.DrawFilledCircle(
			screen,
			float32(star.X),
			float32(star.Y),
			starRenderRadius(star),
			starColor(star),
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

	const (
		diskStarCount  = 1500
		bulgeStarCount = 300
	)

	stars := make([]Star, 0, diskStarCount+bulgeStarCount)

	for i := 0; i < diskStarCount; i++ {
		stars = append(stars, newDiskStar(rng))
	}

	for i := 0; i < bulgeStarCount; i++ {
		stars = append(stars, newBulgeStar(rng))
	}

	game := &Game{
		stars: stars,
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Galaxy Go")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
