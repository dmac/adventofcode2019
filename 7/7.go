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
)

type computer struct {
	pc     int
	memory []int

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

func (c *computer) read(m mode) int {
	n := c.next()
	switch m {
	case modePosition:
		return c.memory[n]
	case modeImmediate:
		return n
	default:
		panic(fmt.Sprintf("unknown mode %d", m))
	}
}

func (c *computer) write(n int) {
	c.memory[c.next()] = n
}

func add(c *computer, modes []mode) {
	modes = fillModes(modes, 3)
	a := c.read(modes[0])
	b := c.read(modes[1])
	c.write(a + b)
}

func mul(c *computer, modes []mode) {
	modes = fillModes(modes, 3)
	a := c.read(modes[0])
	b := c.read(modes[1])
	c.write(a * b)
}

func input(c *computer, modes []mode) {
	modes = fillModes(modes, 1)
	n := c.input.ReadInt()
	c.write(n)
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
		c.write(1)
	} else {
		c.write(0)
	}
}

func eq(c *computer, modes []mode) {
	modes = fillModes(modes, 3)
	a := c.read(modes[0])
	b := c.read(modes[1])
	if a == b {
		c.write(1)
	} else {
		c.write(0)
	}
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
	99: halt,
}

type amplifier struct {
	c      *computer
	input  *intBuffer
	output *intBuffer
}

func newAmplifier(prg []int, phase int, input, output *intBuffer) *amplifier {
	c := newComputer(prg, input, output)
	input.WriteInt(phase)
	amp := &amplifier{
		c:      c,
		input:  input,
		output: output,
	}
	return amp
}

func (a *amplifier) addInput(in int) {
	a.input.WriteInt(in)
}

func (a *amplifier) runProgram() {
	a.c.runProgram()
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

func tryPhases(prg, phases []int) int {
	pipes := make([]*intBuffer, len(phases))
	for i := range phases {
		pipes[i] = newIntBuffer()
	}
	amps := make([]*amplifier, len(phases))
	for i, phase := range phases {
		in := pipes[i]
		out := pipes[(i+1)%len(phases)]
		amps[i] = newAmplifier(prg, phase, in, out)
	}
	amps[0].addInput(0)
	var wg sync.WaitGroup
	for i := range phases {
		i := i
		amp := amps[i]
		wg.Add(1)
		go func() {
			amp.runProgram()
			wg.Done()
		}()
	}
	wg.Wait()
	lastAmp := amps[len(amps)-1]
	return lastAmp.output.ReadInt()
}

func permutations(s []int) [][]int {
	if len(s) <= 1 {
		return [][]int{s}
	}
	var all [][]int
	for i, n := range s {
		rest := append(append([]int{}, s[:i]...), s[i+1:]...)
		tails := permutations(rest)
		for _, tail := range tails {
			all = append(all, append([]int{n}, tail...))
		}
	}
	return all
}

func run() error {
	prg, err := loadProgram("input.txt")
	if err != nil {
		return err
	}
	max := 0
	for _, perm := range permutations([]int{5, 6, 7, 8, 9}) {
		output := tryPhases(prg, perm)
		if output > max {
			max = output
		}
	}
	fmt.Println(max)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
