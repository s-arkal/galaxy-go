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

	armCount       = 4
	armStartRadius = 40.0
	pitchDegrees   = 18.0

	patternSpeed   = 0.05
	spiralStrength = 5.0

	haloSpeed      = 14.0
	haloCoreRadius = 120.0
)

func updateStar(star *Star, simTime float64) {
	dx := centerX - star.X
	dy := centerY - star.Y

	distSquared := dx*dx + dy*dy
	softened := math.Pow(distSquared+epsilon*epsilon, 1.5)

	star.AX = G * centralMass * dx / softened
	star.AY = G * centralMass * dy / softened

	haloAX, haloAY := haloAcceleration(star)
	star.AX += haloAX
	star.AY += haloAY

	if star.Population == DiskStar {
		spiralAX, spiralAY := spiralAcceleration(star, simTime)
		star.AX += spiralAX
		star.AY += spiralAY
	}

	star.VX += star.AX * dt
	star.VY += star.AY * dt

	star.X += star.VX * dt
	star.Y += star.VY * dt

}

func circularSpeed(radius float64) float64 {
	r2 := radius * radius
	softenedRadius := math.Pow(r2+epsilon*epsilon, 1.5)
	centralVC2 := G * centralMass * r2 / softenedRadius

	haloVC2 := haloSpeed * haloSpeed * r2 / (r2 + haloCoreRadius*haloCoreRadius)

	totalVC2 := centralVC2 + haloVC2
	orbitalSpeed := math.Sqrt(totalVC2)
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

func spiralAcceleration(star *Star, simTime float64) (float64, float64) {
	dx := star.X - centerX
	dy := star.Y - centerY

	radius := math.Sqrt(dx*dx + dy*dy)

	if radius < armStartRadius {
		return 0, 0
	}

	theta := math.Atan2(dy, dx)

	pitch := pitchDegrees * math.Pi / 180.0
	b := math.Tan(pitch)

	m := float64(armCount)

	phase := m * (theta + math.Log(radius/armStartRadius)/b - patternSpeed*simTime)

	s := math.Sin(phase)

	radialAcceleration := -spiralStrength * m * s / (b * radius)

	tangentialAcceleration := -spiralStrength * m * s / radius

	cosTheta := math.Cos(theta)
	sinTheta := math.Sin(theta)

	ax := radialAcceleration*cosTheta - tangentialAcceleration*sinTheta
	ay := radialAcceleration*sinTheta + tangentialAcceleration*cosTheta

	return ax, ay
}

func haloAcceleration(star *Star) (float64, float64) {
	dx := centerX - star.X
	dy := centerY - star.Y
	r2 := dx*dx + dy*dy

	if r2 == 0 {
		return 0, 0
	}

	factor := haloSpeed * haloSpeed / (r2 + haloCoreRadius*haloCoreRadius)

	return factor * dx, factor * dy

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

	vx += radialVelocity * math.Cos(angle)
	vy += radialVelocity * math.Sin(angle)

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

		armFraction = 0.92

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

	var angle float64

	if radius < armStartRadius {
		angle = rng.Float64() * 2 * math.Pi
	} else if rng.Float64() < armFraction {
		arm := rng.Intn(armCount)

		armOffset := 2 * math.Pi * float64(arm) / float64(armCount)

		pitch := pitchDegrees * math.Pi / 180.0
		b := math.Tan(pitch)

		spiralAngle := -math.Log(radius/armStartRadius) / b

		scatter := 0.04 + 0.06*(radius/maxRadius)

		angle = armOffset + spiralAngle + scatter*rng.NormFloat64()
	} else {
		angle = rng.Float64() * 2 * math.Pi
	}

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
	stars   []Star
	simTime float64
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
				updateStar(&g.stars[i], g.simTime)
			}
		}(start, end)
	}

	wg.Wait()

	g.simTime += dt

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
		diskStarCount  = 5000
		bulgeStarCount = 500
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
