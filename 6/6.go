package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func loadOrbits(filename string) (kOrbitsV map[string]string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	orbits := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		bodies := strings.Split(line, ")")
		if len(bodies) != 2 {
			return nil, fmt.Errorf("got %d bodies on line %q", len(bodies), line)
		}
		orbits[bodies[1]] = bodies[0]
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return orbits, nil
}

func countOrbits(body string, orbits map[string]string) int {
	count := 0
	for body != "COM" {
		count++
		body = orbits[body]
	}
	return count
}

func allOrbits(body string, orbits map[string]string) map[string]int {
	all := make(map[string]int)
	transfers := 0
	for body != "COM" {
		next := orbits[body]
		all[next] = transfers
		body = next
		transfers++
	}
	return all
}

func run() error {
	orbits, err := loadOrbits("input.txt")
	if err != nil {
		return err
	}
	santaOrbits := allOrbits("SAN", orbits)
	youTransfers := 0
	body := orbits["YOU"]
	for body != "COM" {
		if santaTransfers, ok := santaOrbits[body]; ok {
			fmt.Println(youTransfers + santaTransfers)
			return nil
		}
		body = orbits[body]
		youTransfers++
	}
	return fmt.Errorf("no common orbit found")
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
