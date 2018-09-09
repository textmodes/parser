package drawing

import (
	"image"
	"math"
)

type Polygon []Point

func PolygomFromImagePoints(points ...image.Point) Polygon {
	polygon := make(Polygon, len(points))
	for i, point := range points {
		polygon[i] = PointFromImagePoint(point)
	}
	return polygon
}

func (polygon Polygon) ImagePoints() []image.Point {
	points := make([]image.Point, len(polygon))
	for i, point := range polygon {
		points[i] = point.ImagePoint()
	}
	return points
}

// Add a point to the polygon.
func (polygon *Polygon) Add(point Point) {
	*polygon = append(*polygon, point)
}

// Bounds are the polygon bounding box.
func (polygon Polygon) Bounds() Rectangle {
	b := Rectangle{
		Min: Point{
			X: math.Inf(+1),
			Y: math.Inf(+1),
		},
		Max: Point{
			X: math.Inf(-1),
			Y: math.Inf(-1),
		},
	}
	for _, point := range polygon {
		b.Min.X = math.Min(b.Min.X, point.X)
		b.Max.X = math.Max(b.Max.X, point.X)
		b.Min.Y = math.Min(b.Min.Y, point.Y)
		b.Max.Y = math.Max(b.Max.Y, point.Y)
	}
	return b
}

// Checks if a point is inside a contour using the "point in polygon" raycast method.
// This works for all polygons, whether they are clockwise or counter clockwise,
// convex or concave.
func (polygon Polygon) Contains(point Point) bool {
	var intersect int
	for i, curr := range polygon {
		j := i + 1
		if j == len(polygon) {
			j = 0
		}
		next := polygon[j]

		bot, top := curr, next
		if bot.Y > top.Y {
			bot, top = top, bot
		}
		if point.Y < bot.Y || point.Y >= top.Y {
			continue
		}

		if point.X >= math.Max(curr.X, next.X) || next.Y == curr.Y {
			continue
		}

		// Find where the line intersects
		xint := (point.Y-curr.Y)*(next.X-curr.X)/(next.Y-curr.Y) + curr.X
		if curr.X != next.X && point.X > xint {
			continue
		}

		intersect++
	}
	return intersect%2 != 0
}

// Len is the number of points in the polygon.
func (polygon Polygon) Len() int {
	return len(polygon)
}
