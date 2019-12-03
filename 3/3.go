package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type point [2]int

type wire []point

func distance(p, q point) int {
	if p[0] > q[0] {
		p, q = q, p
	}
	if p[1] > q[1] {
		p, q = q, p
	}
	if p[0] == q[0] {
		return q[1] - p[1]
	} else {
		return q[0] - p[0]
	}
}

func intersection(p0, p1, q0, q1 point) (point, bool) {
	if p0[0] == p1[0] && q0[0] == q1[0] {
		return point{}, false
	}
	if p0[1] == p1[1] && q0[1] == q1[1] {
		return point{}, false
	}
	if p0[0] != p1[0] {
		// p0[1] == p1[1]
		// q0[0] == q1[0]
		if p0[0] > p1[0] {
			p0, p1 = p1, p0
		}
		if q0[1] > q1[1] {
			q0, q1 = q1, q0
		}
		x := p0[0] <= q0[0] && q0[0] <= p1[0]
		y := q0[1] <= p0[1] && p0[1] <= q1[1]
		if x && y {
			ix := point{q0[0], p0[1]}
			return ix, ix[0] != 0 || ix[1] != 0
		}
		return point{}, false
	} else {
		// q0[1] == q1[1]
		// p0[0] == p1[0]
		if p0[1] > p1[1] {
			p0, p1 = p1, p0
		}
		if q0[0] > q1[0] {
			q0, q1 = q1, q0
		}
		x := q0[0] <= p0[0] && p0[0] <= q1[0]
		y := p0[1] <= q0[1] && q0[1] <= p1[1]
		if x && y {
			ix := point{p0[0], q0[1]}
			return ix, ix[0] != 0 || ix[1] != 0
		}
		return point{}, false
	}
}

func parseWire(dirs []string) (wire, error) {
	w := wire{{0, 0}}
	p := w[0]
	for _, dir := range dirs {
		n, err := strconv.Atoi(dir[1:])
		if err != nil {
			return nil, err
		}
		switch dir[0] {
		case 'U':
			p[1] += n
		case 'D':
			p[1] -= n
		case 'L':
			p[0] -= n
		case 'R':
			p[0] += n
		default:
			return nil, fmt.Errorf("unexpected direction %s", dir)
		}
		w = append(w, p)
	}
	return w, nil
}

func loadWires(filename string) ([]wire, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var wires []wire
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		dirs := strings.Split(scanner.Text(), ",")
		w, err := parseWire(dirs)
		if err != nil {
			return nil, err
		}
		wires = append(wires, w)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return wires, nil
}

func run() error {
	wires, err := loadWires("input.txt")
	if err != nil {
		return err
	}
	if len(wires) != 2 {
		return fmt.Errorf("expected 2 wires, got %d", len(wires))
	}
	min := 0
	pSteps := 0
	for i := 0; i < len(wires[0])-1; i++ {
		p0 := wires[0][i]
		p1 := wires[0][i+1]
		qSteps := 0
		for j := 0; j < len(wires[1])-1; j++ {
			q0 := wires[1][j]
			q1 := wires[1][j+1]
			ix, ok := intersection(p0, p1, q0, q1)
			if ok {
				dist := pSteps + distance(p0, ix) + qSteps + distance(q0, ix)
				if min == 0 || dist < min {
					min = dist
				}
			}
			qSteps += distance(q0, q1)
		}
		pSteps += distance(p0, p1)
	}
	fmt.Println(min)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
