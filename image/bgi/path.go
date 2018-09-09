package bgi

import "image"

type Path []image.Point

func NewPath(points ...image.Point) Path {
	return Path(points)
}

func (p *Path) Add(point image.Point) {
	*p = append(*p, point)
}

func (p Path) Bounds() image.Rectangle {
	var b image.Rectangle
	for _, point := range p {
		b.Min.X = min(b.Min.X, point.X)
		b.Min.Y = min(b.Min.Y, point.Y)
		b.Max.X = max(b.Max.X, point.X)
		b.Max.Y = max(b.Max.Y, point.Y)
	}
	return b
}

func (p Path) Contains(point image.Point) bool {
	b := p.Bounds()
	if !point.In(b) {
		return false
	}

	var (
		vertices  = len(p)
		intersect bool
		i, j      int
	)
	for i = 1; i < vertices; i, j = i+1, j+1 {
		if (p[i].Y > point.Y) != (p[j].Y > point.Y) && (point.X < (p[j].X-p[i].X)*(point.Y-p[i].Y)/(p[j].Y-p[i].Y)+p[i].X) {
			intersect = !intersect
		}
	}

	return intersect
}

func (p Path) Last() image.Point {
	var l = len(p)
	if l == 0 {
		return image.ZP
	}
	return p[l-1]
}
