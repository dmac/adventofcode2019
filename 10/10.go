package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"sort"
	"strconv"
)

type asteroid struct {
	x int
	y int
}

type fraction struct {
	n   int
	d   int
	nan bool
}

func newFraction(n, d int) fraction {
	f := fraction{n: n, d: d}
	if d == 0 {
		f.nan = true
	}
	return f
}

type intLine struct {
	slope      fraction
	yIntercept fraction
}

func (l intLine) String() string {
	var m, b string

	switch {
	case l.slope.n == 0:
		m = "0"
	case l.slope.nan:
		m = "?"
	case l.slope.d == 1:
		m = strconv.Itoa(l.slope.n)
	default:
		m = fmt.Sprintf("%d/%d", l.slope.n, l.slope.d)
	}

	switch {
	case l.yIntercept.n == 0:
		b = "0"
	case l.yIntercept.nan:
		b = "?"
	case l.yIntercept.d == 1:
		b = strconv.Itoa(l.yIntercept.n)
	default:
		b = fmt.Sprintf("%d/%d", l.yIntercept.n, l.yIntercept.d)
	}

	return fmt.Sprintf("y=%sx+%s", m, b)
}

func loadAsteroids(filename string) ([]asteroid, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var asts []asteroid
	x := 0
	y := 0
	for _, c := range b {
		switch c {
		case '\n':
			x = 0
			y++
		case '.':
			x++
		case '#':
			ast := asteroid{x: x, y: y}
			asts = append(asts, ast)
			x++
		default:
			return nil, fmt.Errorf("unknown character %q", c)
		}
	}
	return asts, nil
}

func factors(n int) []int {
	m := make(map[int]struct{})
	dir := 1
	if n < 0 {
		dir = -1
	}
	for i := 1 * dir; ; i += dir {
		if n%i == 0 {
			m[i] = struct{}{}
			m[n/i] = struct{}{}
		}
		if i == n {
			break
		}
	}
	var facts []int
	for k := range m {
		facts = append(facts, k)
	}
	sort.Ints(facts)
	return facts
}

func gcd(a, b int) int {
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	af := factors(a)
	bf := factors(b)
	m := make(map[int]int)
	for _, f := range af {
		m[f]++
	}
	for _, f := range bf {
		m[f]++
	}
	var common []int
	for f, count := range m {
		if count == 2 {
			common = append(common, f)
		}
	}
	if len(common) == 0 {
		return 0
	}
	sort.Ints(common)
	return common[len(common)-1]
}

func simplifyFraction(frac fraction) fraction {
	n := frac.n
	d := frac.d
	if n < 0 && d < 0 {
		n *= -1
		d *= -1
	}
	if n >= 0 && d < 0 {
		n *= -1
		d *= -1
	}
	div := gcd(n, d)
	n /= div
	d /= div
	return newFraction(n, d)
}

func findLine(a, b asteroid) intLine {
	run := a.x - b.x
	rise := a.y - b.y
	slope := simplifyFraction(newFraction(rise, run))
	if slope.nan {
		return intLine{
			slope:      slope,
			yIntercept: fraction{nan: true},
		}
	}

	x := a.x
	y := a.y
	prevX := x
	prevY := y
	for x > 0 {
		prevX = x
		prevY = y
		x -= slope.d
		y -= slope.n
	}
	var yIntercept fraction
	if x == 0 {
		yIntercept = newFraction(y, 1)
	} else {
		yIntercept = newFraction(prevY-y, prevX-x)
	}
	return intLine{
		slope:      slope,
		yIntercept: yIntercept,
	}
}

func partitionAndSortColinearAsteroids(ast asteroid, colinear []asteroid) (lefts, rights []asteroid) {
	line := findLine(ast, colinear[0])
	all := make(map[asteroid]struct{})
	for _, co := range colinear {
		all[co] = struct{}{}
	}
	left := ast
	right := ast
	for len(all) > 0 {
		left.x -= line.slope.d
		left.y -= line.slope.n
		if _, ok := all[left]; ok {
			lefts = append(lefts, left)
			delete(all, left)
		}
		right.x += line.slope.d
		right.y += line.slope.n
		if _, ok := all[right]; ok {
			rights = append(rights, right)
			delete(all, right)
		}
	}
	return lefts, rights
}

func findColinearAsteroids(ast asteroid, others []asteroid) map[intLine][]asteroid {
	colinear := make(map[intLine][]asteroid)
	for _, other := range others {
		line := findLine(ast, other)
		colinear[line] = append(colinear[line], other)
	}
	return colinear
}

func countDetections(ast asteroid, others []asteroid) int {
	colinear := findColinearAsteroids(ast, others)
	detections := 0
	for _, asts := range colinear {
		lefts, rights := partitionAndSortColinearAsteroids(ast, asts)
		if len(lefts) > 0 {
			detections++
		}
		if len(rights) > 0 {
			detections++
		}
	}
	return detections
}

func findBestLocation(asts []asteroid) (ast asteroid, detections int) {
	others := make([]asteroid, len(asts)-1)
	max := 0
	var best asteroid
	for i := 0; i < len(asts); i++ {
		ast := asts[i]
		copy(others[:i], asts[:i])
		copy(others[i:], asts[i+1:])
		detections := countDetections(ast, others)
		if detections > max {
			max = detections
			best = ast
		}
	}
	return best, max
}

func rad2deg(rad float64) float64 {
	return rad * 180 / math.Pi
}

func computeAngle(a, b asteroid) float64 {
	y := float64(b.y - a.y)
	x := float64(b.x - a.x)
	angle := rad2deg(math.Atan2(y, x)) + 90
	for angle > 360 {
		angle -= 360
	}
	for angle < 0 {
		angle += 360
	}
	return angle
}

func nthVaporizedAsteroid(ast asteroid, others []asteroid, n int) asteroid {
	colinear := findColinearAsteroids(ast, others)
	var vectors [][]asteroid
	for _, asts := range colinear {
		lefts, rights := partitionAndSortColinearAsteroids(ast, asts)
		if len(lefts) > 0 {
			vectors = append(vectors, lefts)
		}
		if len(rights) > 0 {
			vectors = append(vectors, rights)
		}
	}
	sort.Slice(vectors, func(i, j int) bool {
		a := vectors[i][0]
		b := vectors[j][0]
		angleA := computeAngle(ast, a)
		angleB := computeAngle(ast, b)
		return angleA < angleB
	})
	i := 0
	nth := 0
	for {
		if nth == n-1 {
			return vectors[i][0]
		}
		vectors[i] = vectors[i][1:]
		nth++
		i = (i + 1) % len(vectors)
		for len(vectors[i]) == 0 {
			i = (i + 1) % len(vectors)
		}
	}
}

func run() error {
	asts, err := loadAsteroids("input.txt")
	if err != nil {
		return err
	}
	best, detections := findBestLocation(asts)
	fmt.Println(best, detections)

	others := make([]asteroid, len(asts)-1)
	for i, ast := range asts {
		if ast == best {
			copy(others[:i], asts[:i])
			copy(others[i:], asts[i+1:])
			break
		}
	}
	nth := nthVaporizedAsteroid(best, others, 200)
	fmt.Println(nth.x*100 + nth.y)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
