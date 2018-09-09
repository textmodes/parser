package drawing

import "math"

type Matrix struct {
	XX, YX, XY, YY, X0, Y0 float64
}

func Identity() Matrix {
	return Matrix{
		1, 0,
		0, 1,
		0, 0,
	}
}

func Scale(x, y float64) Matrix {
	return Matrix{
		x, 0,
		0, y,
		0, 0,
	}
}

func Rotate(angle float64) Matrix {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return Matrix{
		c, s,
		-s, c,
		0, 0,
	}
}

func Shear(x, y float64) Matrix {
	return Matrix{
		1, y,
		x, 1,
		0, 0,
	}
}

func (matrix Matrix) Transform(x, y float64) (ox float64, oy float64) {
	ox = matrix.XX*x + matrix.XY*y + matrix.X0
	oy = matrix.YX*x + matrix.YY*y + matrix.Y0
	return
}
