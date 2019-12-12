package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"unicode"
)

func splitLayers(img []int, width, height int) [][]int {
	if len(img)%(width*height) != 0 {
		panic("invalid image: width * height doesn't evenly divide img")
	}
	layerSize := width * height
	numLayers := len(img) / layerSize
	var layers [][]int
	for i := 0; i < numLayers; i++ {
		layers = append(layers, img[i*layerSize:i*layerSize+layerSize])
	}
	return layers
}

func loadImage(filename string) ([]int, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var img []int
	for _, r := range b {
		if !unicode.IsDigit(rune(r)) {
			continue
		}
		n, err := strconv.Atoi(string(r))
		if err != nil {
			return nil, err
		}
		img = append(img, n)
	}
	return img, nil
}

func solvePartOne(layers [][]int) {
	min := len(layers[0]) + 1
	var minLayer []int
	for _, layer := range layers {
		nz := 0
		for _, n := range layer {
			if n == 0 {
				nz++
			}
		}
		if nz < min {
			min = nz
			minLayer = layer
		}
	}
	if minLayer == nil {
		panic("no layers contain zeroes")
	}
	numOnes := 0
	numTwos := 0
	for _, n := range minLayer {
		if n == 1 {
			numOnes++
		}
		if n == 2 {
			numTwos++
		}
	}
	fmt.Println(numOnes * numTwos)
}

func printLayer(layer []int, width, height int) {
	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			fmt.Print(layer[r*width+c])
		}
		fmt.Println()
	}
}

func run() error {
	img, err := loadImage("input.txt")
	if err != nil {
		return err
	}
	width := 25
	height := 6
	layers := splitLayers(img, width, height)
	solvePartOne(layers)

	layerSize := len(layers[0])
	final := make([]int, layerSize)
outer:
	for i := 0; i < layerSize; i++ {
		for _, layer := range layers {
			if layer[i] != 2 {
				final[i] = layer[i]
				continue outer
			}
		}
	}
	printLayer(final, width, height)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
