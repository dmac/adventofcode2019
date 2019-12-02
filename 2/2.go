package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type computer struct {
	pc      int
	memory  []int
	running bool
}

func newComputer() *computer {
	return &computer{running: true}
}

func (c *computer) next() int {
	n := c.memory[c.pc]
	c.pc++
	return n
}

func add(c *computer) {
	a := c.memory[c.next()]
	b := c.memory[c.next()]
	r := c.next()
	c.memory[r] = a + b
}

func mul(c *computer) {
	a := c.memory[c.next()]
	b := c.memory[c.next()]
	r := c.next()
	c.memory[r] = a * b
}

func halt(c *computer) {
	c.running = false
}

type opcode func(*computer)

var opcodes map[int]opcode = map[int]opcode{
	1:  add,
	2:  mul,
	99: halt,
}

func tryWithInputs(prg []int, n, v int) int {
	c := newComputer()
	c.memory = append([]int{}, prg...)
	c.memory[1] = n
	c.memory[2] = v
	for c.running {
		op, ok := opcodes[c.next()]
		if !ok {
			panic(fmt.Sprintf("unknown opcode %d", c.memory[c.pc]))
		}
		op(c)
	}
	return c.memory[0]
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
	target := 19690720
	for n := 0; n < 100; n++ {
		for v := 0; v < 100; v++ {
			if tryWithInputs(prg, n, v) == target {
				fmt.Println(100*n + v)
				return nil
			}
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
