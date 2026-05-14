package world

import "math"

const ellipseSATSegments = 24

// ellipsePoly аппроксимирует эллипс выпуклым многоугольником (центр, полуоси по касательной и радиально, угол наклона).
func ellipsePoly(cx, cy, semiT, semiR, ang float64, n int) [][2]float64 {
	if n < 3 {
		n = 3
	}
	sinT, cosT := math.Sin(ang), math.Cos(ang)
	out := make([][2]float64, n)
	for i := 0; i < n; i++ {
		phi := 2 * math.Pi * float64(i) / float64(n)
		cp, sp := math.Cos(phi), math.Sin(phi)
		wx := cx + semiT*cp*cosT + semiR*sp*sinT
		wy := cy - semiT*cp*sinT + semiR*sp*cosT
		out[i] = [2]float64{wx, wy}
	}
	return out
}

func convexPolygonsIntersect(a, b [][2]float64) bool {
	if len(a) < 3 || len(b) < 3 {
		return false
	}
	if hasSeparatingAxis(a, b) {
		return false
	}
	if hasSeparatingAxis(b, a) {
		return false
	}
	return true
}

func hasSeparatingAxis(ref, other [][2]float64) bool {
	n := len(ref)
	for i := 0; i < n; i++ {
		p0, p1 := ref[i], ref[(i+1)%n]
		ex := p1[0] - p0[0]
		ey := p1[1] - p0[1]
		ax, ay := -ey, ex
		al := math.Hypot(ax, ay)
		if al < 1e-12 {
			continue
		}
		ax /= al
		ay /= al
		if axisSeparates(ax, ay, ref, other) {
			return true
		}
	}
	return false
}

func axisSeparates(ax, ay float64, p, q [][2]float64) bool {
	minP, maxP := projectInterval(p, ax, ay)
	minQ, maxQ := projectInterval(q, ax, ay)
	return maxP < minQ || maxQ < minP
}

func projectInterval(poly [][2]float64, ax, ay float64) (min, max float64) {
	for i, pt := range poly {
		d := pt[0]*ax + pt[1]*ay
		if i == 0 || d < min {
			min = d
		}
		if i == 0 || d > max {
			max = d
		}
	}
	return min, max
}

// orientedEllipsesOverlap проверяет пересечение двух эллипсов в плоскости XY (аппроксимация SAT по многоугольникам).
func orientedEllipsesOverlap(ax, ay, aSemiT, aSemiR, aAng, bx, by, bSemiT, bSemiR, bAng float64) bool {
	pa := ellipsePoly(ax, ay, aSemiT, aSemiR, aAng, ellipseSATSegments)
	pb := ellipsePoly(bx, by, bSemiT, bSemiR, bAng, ellipseSATSegments)
	return convexPolygonsIntersect(pa, pb)
}
