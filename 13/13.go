package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
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
	n := c.input.ReadInt()
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
	wait *sync.Cond
	ints []int
}

func newIntBuffer() *intBuffer {
	var mu sync.Mutex
	return &intBuffer{
		wait: sync.NewCond(&mu),
	}
}

func (rw *intBuffer) ReadInt() int {
	rw.wait.L.Lock()
	for len(rw.ints) == 0 {
		rw.wait.Wait()
	}
	n := rw.ints[0]
	rw.ints = rw.ints[1:]
	rw.wait.L.Unlock()
	return n
}

func (rw *intBuffer) WriteInt(n int) {
	rw.wait.L.Lock()
	rw.ints = append(rw.ints, n)
	rw.wait.Broadcast()
	rw.wait.L.Unlock()
}

type arcade struct {
	c      *computer
	input  *intBuffer
	output *intBuffer

	mu     sync.Mutex
	screen [][]rune
	score  int
}

func newArcade(prg []int) *arcade {
	prg[0] = 2
	input := newIntBuffer()
	output := newIntBuffer()
	c := newComputer(prg, input, output)
	a := &arcade{
		c:      c,
		input:  input,
		output: output,
	}
	return a
}

func (a *arcade) run() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		a.c.runProgram()
		wg.Done()
	}()
	go a.handleInput()
	go a.handleOutput()
	go a.loopDrawScreen()
	wg.Wait()
	time.Sleep(5 * time.Second)
}

var tileRunes = map[int]rune{
	0: ' ',
	1: '|',
	2: 'B',
	3: '-',
	4: 'o',
}

func (a *arcade) handleInput() {
	for {
		var ballX int
		var paddleX int
		a.mu.Lock()
		for _, row := range a.screen {
			for x, r := range row {
				switch r {
				case 'o':
					ballX = x
				case '-':
					paddleX = x
				}
			}
		}
		a.mu.Unlock()

		switch {
		case paddleX < ballX:
			a.input.WriteInt(1)
		case paddleX > ballX:
			a.input.WriteInt(-1)
		default:
			a.input.WriteInt(0)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (a *arcade) handleOutput() {
	for {
		x := a.output.ReadInt()
		y := a.output.ReadInt()
		tile := a.output.ReadInt()

		a.mu.Lock()
		if y >= len(a.screen) {
			screen := make([][]rune, y+1)
			copy(screen, a.screen)
			a.screen = screen
		}
		for i := 0; i < len(a.screen); i++ {
			row := a.screen[i]
			if x >= len(row) {
				newRow := make([]rune, x+1)
				copy(newRow, row)
				a.screen[i] = newRow
			}
		}

		if x == -1 && y == 0 {
			a.score = tile
		} else {
			r, ok := tileRunes[tile]
			if !ok {
				panic(fmt.Sprintf("unknown tile %d", tile))
			}
			a.screen[y][x] = r
		}
		a.mu.Unlock()
	}
}

func (a *arcade) loopDrawScreen() {
	for range time.Tick(100 * time.Millisecond) {
		a.drawScreen()
		fmt.Println(a.score)
	}
}

func (a *arcade) drawScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[1;1H")
	a.mu.Lock()
	for _, row := range a.screen {
		fmt.Println(string(row))
	}
	a.mu.Unlock()
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
	a := newArcade(prg)
	a.run()
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
