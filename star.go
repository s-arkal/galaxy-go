package main

type StarPopulation int

const (
	DiskStar StarPopulation = iota
	BulgeStar
)

type Star struct {
	X, Y       float64
	VX, VY     float64
	AX, AY     float64
	Mass       float64
	Luminosity float64
	Population StarPopulation
}
