package main

import "fmt"

func main() {
	min := 372037
	max := 905157
	count := 0
	for i := min; i < max; i++ {
		digits := []int{
			i / 100000 % 10,
			i / 10000 % 10,
			i / 1000 % 10,
			i / 100 % 10,
			i / 10 % 10,
			i / 1 % 10,
		}
		pair := false
		increasing := true
		for j := 0; j < len(digits)-1; j++ {
			if digits[j] == digits[j+1] &&
				(j == 0 || digits[j-1] != digits[j]) &&
				(j+2 == len(digits) || digits[j+2] != digits[j]) {
				pair = true
			}
			if digits[j] > digits[j+1] {
				increasing = false
			}
		}
		if pair && increasing {
			count++
		}
	}
	fmt.Println(count)
}
