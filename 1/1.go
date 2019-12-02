package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
)

func fuelRequired(mass int) int {
	return mass/3 - 2
}

func totalFuelRequired(mass int) int {
	fuel := 0
	for {
		curr := fuelRequired(mass)
		if curr <= 0 {
			break
		}
		fuel += curr
		mass = curr
	}
	return fuel
}

func run() error {
	f, err := os.Open("input.txt")
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	fuel := 0
	for scanner.Scan() {
		line := scanner.Text()
		mass, err := strconv.Atoi(line)
		if err != nil {
			return err
		}
		fuel += totalFuelRequired(mass)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	fmt.Println(fuel)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
