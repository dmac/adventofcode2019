package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

type mode int

const (
	modePosition  mode = 0
	modeImmediate mode = 1
)

type computer struct {
	pc     int
	memory []int

	input  io.Reader
	output io.Writer

	running bool
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

func newComputer(prg []int, input io.Reader, output io.Writer) *computer {
	return &computer{
		memory:  prg,
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

func in(c *computer, modes []mode) {
	modes = fillModes(modes, 1)
	b, err := ioutil.ReadAll(c.input)
	if err != nil {
		panic(fmt.Sprintf("error reading input: %s", err))
	}
	n, err := strconv.Atoi(string(b))
	if err != nil {
		panic(fmt.Sprintf("error parsing input: %s", err))
	}
	c.write(n)
}

func out(c *computer, modes []mode) {
	modes = fillModes(modes, 1)
	s := strconv.Itoa(c.read(modes[0]))
	c.output.Write([]byte(s))
	c.output.Write([]byte("\n"))
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
	3:  in,
	4:  out,
	5:  jit,
	6:  jif,
	7:  lt,
	8:  eq,
	99: halt,
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
	c := newComputer(prg, bytes.NewBufferString("5"), os.Stdout)
	for c.running {
		code, modes := parseOpcodeModes(c.next())
		op, ok := opcodes[code]
		if !ok {
			panic(fmt.Sprintf("unknown opcode %d", code))
		}
		op(c, modes)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
