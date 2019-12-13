package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"sync"
)

type mode int

const (
	modePosition  mode = 0
	modeImmediate mode = 1
	modeRelative  mode = 2
)

type computer struct {
	pc      int
	relBase int
	memory  []int

	input  *intBuffer
	output *intBuffer

	running bool
}

func (c *computer) runProgram() {
	for c.running {
		code, modes := parseOpcodeModes(c.next())
		op, ok := opcodes[code]
		if !ok {
			panic(fmt.Sprintf("unknown opcode %d", code))
		}
		op(c, modes)
	}
}

func parseOpcodeModes(value int) (code int, modes []mode) {
	code = value % 100
	value /= 100
	for value > 0 {
		modes = append(modes, mode(value%10))
		value /= 10
	}
	return code, modes
}

func fillModes(modes []mode, size int) []mode {
	if len(modes) < size {
		filled := make([]mode, size)
		copy(filled, modes)
		modes = filled
	}
	return modes
}

func newComputer(prg []int, input, output *intBuffer) *computer {
	return &computer{
		memory:  append([]int{}, prg...),
		input:   input,
		output:  output,
		running: true,
	}
}

func (c *computer) next() int {
	n := c.memory[c.pc]
	c.pc++
	return n
}

func (c *computer) expandMemoryForIndex(idx int) {
	if idx >= len(c.memory) {
		mem := make([]int, idx+1)
		copy(mem, c.memory)
		c.memory = mem
	}
}

func (c *computer) read(m mode) int {
	idx := c.next()
	switch m {
	case modeImmediate:
		return idx
	case modePosition:
	case modeRelative:
		idx += c.relBase
	default:
		panic(fmt.Sprintf("unknown mode %d", m))
	}
	c.expandMemoryForIndex(idx)
	return c.memory[idx]
}

func (c *computer) write(n int, m mode) {
	idx := c.next()
	switch m {
	case modeImmediate:
		panic("immediate mode used for write")
	case modePosition:
	case modeRelative:
		idx += c.relBase
	}
	c.expandMemoryForIndex(idx)
	c.memory[idx] = n
}

func add(c *computer, modes []mode) {
	modes = fillModes(modes, 3)
	a := c.read(modes[0])
	b := c.read(modes[1])
	c.write(a+b, modes[2])
}

func mul(c *computer, modes []mode) {
	modes = fillModes(modes, 3)
	a := c.read(modes[0])
	b := c.read(modes[1])
	c.write(a*b, modes[2])
}

func input(c *computer, modes []mode) {
	modes = fillModes(modes, 1)
	n, _ := c.input.ReadInt()
	c.write(n, modes[0])
}

func output(c *computer, modes []mode) {
	modes = fillModes(modes, 1)
	c.output.WriteInt(c.read(modes[0]))
}

func jit(c *computer, modes []mode) {
	modes = fillModes(modes, 2)
	n := c.read(modes[0])
	v := c.read(modes[1])
	if n != 0 {
		c.pc = v
	}
}

func jif(c *computer, modes []mode) {
	modes = fillModes(modes, 2)
	n := c.read(modes[0])
	v := c.read(modes[1])
	if n == 0 {
		c.pc = v
	}
}

func lt(c *computer, modes []mode) {
	modes = fillModes(modes, 3)
	a := c.read(modes[0])
	b := c.read(modes[1])
	if a < b {
		c.write(1, modes[2])
	} else {
		c.write(0, modes[2])
	}
}

func eq(c *computer, modes []mode) {
	modes = fillModes(modes, 3)
	a := c.read(modes[0])
	b := c.read(modes[1])
	if a == b {
		c.write(1, modes[2])
	} else {
		c.write(0, modes[2])
	}
}

func rel(c *computer, modes []mode) {
	modes = fillModes(modes, 1)
	c.relBase += c.read(modes[0])
}

func halt(c *computer, _ []mode) {
	c.running = false
}

type opcode func(c *computer, modes []mode)

var opcodes map[int]opcode = map[int]opcode{
	1:  add,
	2:  mul,
	3:  input,
	4:  output,
	5:  jit,
	6:  jif,
	7:  lt,
	8:  eq,
	9:  rel,
	99: halt,
}

type intBuffer struct {
	wait   *sync.Cond
	ints   []int
	closed bool
}

func newIntBuffer() *intBuffer {
	var mu sync.Mutex
	return &intBuffer{
		wait: sync.NewCond(&mu),
	}
}

func (rw *intBuffer) ReadInt() (int, bool) {
	rw.wait.L.Lock()
	for len(rw.ints) == 0 && !rw.closed {
		rw.wait.Wait()
	}
	if rw.closed && len(rw.ints) == 0 {
		return 0, false
	}
	n := rw.ints[0]
	rw.ints = rw.ints[1:]
	rw.wait.L.Unlock()
	return n, true
}

func (rw *intBuffer) WriteInt(n int) {
	rw.wait.L.Lock()
	rw.ints = append(rw.ints, n)
	rw.wait.Broadcast()
	rw.wait.L.Unlock()
}

func (rw *intBuffer) Close() {
	rw.wait.L.Lock()
	rw.closed = true
	rw.wait.Broadcast()
	rw.wait.L.Unlock()
}

type direction int

const (
	up    direction = 0
	down  direction = 1
	left  direction = 2
	right direction = 3
)

type point struct {
	x int
	y int
}

type robot struct {
	c      *computer
	input  *intBuffer
	output *intBuffer

	wg sync.WaitGroup

	x    int
	y    int
	dir  direction
	grid map[point]int
}

func newRobot(prg []int) *robot {
	r := &robot{
		input:  newIntBuffer(),
		output: newIntBuffer(),
		dir:    up,
		grid:   make(map[point]int),
	}
	r.c = newComputer(prg, r.input, r.output)
	return r
}

func (r *robot) run() {
	r.input.WriteInt(1)
	r.wg.Add(1)
	go r.handleIO()
	r.c.runProgram()
	r.output.Close()
	r.wg.Wait()
	r.drawGrid()
}

func (r *robot) handleIO() {
	defer r.wg.Done()
	for {
		color, ok := r.output.ReadInt()
		if !ok {
			break
		}
		rot, ok := r.output.ReadInt()
		if !ok {
			break
		}

		if color != 0 && color != 1 {
			panic(fmt.Sprintf("unexpected color %d", color))
		}
		if rot != 0 && rot != 1 {
			panic(fmt.Sprintf("unexpected rotation %d", rot))
		}

		r.grid[point{r.x, r.y}] = color

		switch r.dir {
		case up:
			if rot == 0 {
				r.dir = left
				r.x--
			} else {
				r.dir = right
				r.x++
			}
		case down:
			if rot == 0 {
				r.dir = right
				r.x++
			} else {
				r.dir = left
				r.x--
			}
		case left:
			if rot == 0 {
				r.dir = down
				r.y++
			} else {
				r.dir = up
				r.y--
			}
		case right:
			if rot == 0 {
				r.dir = up
				r.y--
			} else {
				r.dir = down
				r.y++
			}
		default:
			panic(fmt.Sprintf("unexpected direction %d", r.dir))
		}
		r.input.WriteInt(r.grid[point{r.x, r.y}])
	}
}

func (r *robot) drawGrid() {
	fmt.Print("\033[2J")
	fmt.Print("\033[1;1H")

	var (
		minX int
		maxX int
		minY int
		maxY int

		minXSet bool
		maxXSet bool
		minYSet bool
		maxYSet bool
	)

	for p := range r.grid {
		if !minXSet || p.x < minX {
			minX = p.x
			minXSet = true
		}
		if !maxXSet || p.x > maxX {
			maxX = p.x
			maxXSet = true
		}
		if !minYSet || p.y < minY {
			minY = p.y
			minYSet = true
		}
		if !maxYSet || p.y > maxY {
			maxY = p.y
			maxYSet = true
		}
	}
	fmt.Println(len(r.grid))
	grid := make([][]rune, maxY-minY+1)
	for i := 0; i < len(grid); i++ {
		grid[i] = make([]rune, maxX-minX+1)
	}

	for p, color := range r.grid {
		x := p.x - minX
		y := p.y - minY
		switch color {
		case 0:
			grid[y][x] = ' '
		case 1:
			grid[y][x] = '█'
		}
	}

	for _, row := range grid {
		fmt.Println(string(row))
	}
}

func loadProgram(filename string) ([]int, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var prg []int
	for _, s := range bytes.Split(b, []byte(",")) {
		n, err := strconv.Atoi(strings.TrimSpace(string(s)))
		if err != nil {
			return nil, err
		}
		prg = append(prg, n)
	}
	return prg, nil
}

func run() error {
	prg, err := loadProgram("input.txt")
	if err != nil {
		return err
	}
	r := newRobot(prg)
	r.run()
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
