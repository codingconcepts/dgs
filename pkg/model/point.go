package model

import "fmt"

type Point struct {
	Lat float64
	Lon float64
}

func (p Point) String() string {
	return fmt.Sprintf("Point(%f %f)", p.Lon, p.Lat)
}
