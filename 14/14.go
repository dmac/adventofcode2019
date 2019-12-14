package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type chemical struct {
	name   string
	amount int
}

func (c *chemical) String() string {
	return fmt.Sprintf("%d %s", c.amount, c.name)
}

type reaction struct {
	inputs []*chemical
	output *chemical
}

func (r *reaction) String() string {
	var sb strings.Builder
	for i, chem := range r.inputs {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(chem.String())
	}
	sb.WriteString(" => ")
	sb.WriteString(r.output.String())
	return sb.String()
}

func parseChemical(s string) (*chemical, error) {
	halves := strings.Split(s, " ")
	if len(halves) != 2 {
		panic(fmt.Sprintf("error parsing chemical %q", s))
	}
	amount, err := strconv.Atoi(halves[0])
	if err != nil {
		return nil, err
	}
	chem := &chemical{
		name:   halves[1],
		amount: amount,
	}
	return chem, nil
}

func parseReaction(s string) (*reaction, error) {
	halves := strings.Split(s, "=>")
	if len(halves) != 2 {
		panic("reaction parsing expected exactly one =>")
	}

	lhs := strings.Split(halves[0], ",")
	var inputs []*chemical
	for _, term := range lhs {
		chem, err := parseChemical(strings.TrimSpace(term))
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, chem)
	}

	rhs := halves[1]
	output, err := parseChemical(strings.TrimSpace(rhs))
	if err != nil {
		return nil, err
	}

	rx := &reaction{
		inputs: inputs,
		output: output,
	}
	return rx, nil
}

func loadReactions(filename string) ([]*reaction, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var reactions []*reaction
	for scanner.Scan() {
		rx, err := parseReaction(scanner.Text())
		if err != nil {
			return nil, err
		}
		reactions = append(reactions, rx)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return reactions, nil
}

func oreRequiredForFuel(reactions []*reaction, fuel int) int {
	goals := []*chemical{{name: "FUEL", amount: fuel}}
	storage := make(map[string]*chemical)
	i := 0
	for {
		// Cycle through all goals.
		if i >= len(goals) {
			i = 0
		}
		goal := goals[i]

		// If we're only left with ORE, we're done. Otherwise, skip it.
		if goal.name == "ORE" {
			if len(goals) == 1 {
				break
			}
			i++
			continue
		}

		// Find the reaction that creates the current goal.
		var rx *reaction
		for _, rx0 := range reactions {
			if rx0.output.name == goal.name {
				rx = rx0
				break
			}
		}

		// If we have the goal chemical in storage, reduce the goal by
		// that amount and remove the corresponding amount from storage.
		stored, ok := storage[goal.name]
		if ok {
			if goal.amount < stored.amount {
				stored.amount -= goal.amount
				goals = append(goals[:i], goals[i+1:]...)
				continue
			}
			goal.amount -= stored.amount
			delete(storage, stored.name)
		}

		// Determine how many times the reaction must run to satisfy the
		// current goal.
		mult := 1
		for mult*rx.output.amount < goal.amount {
			mult++
		}

		// Add any extra output to storage.
		extra := mult*rx.output.amount - goal.amount
		if extra > 0 {
			stored, ok := storage[goal.name]
			if !ok {
				stored = &chemical{name: goal.name}
				storage[goal.name] = stored
			}
			stored.amount += extra
		}

		// Replace the current goal with the scaled reaction inputs.
		goals = append(goals[:i], goals[i+1:]...)
		for _, chem := range rx.inputs {
			newGoal := &chemical{
				name:   chem.name,
				amount: chem.amount * mult,
			}
			found := false
			for _, g := range goals {
				if g.name == newGoal.name {
					g.amount += newGoal.amount
					found = true
				}
			}
			if !found {
				goals = append(goals, newGoal)
			}
		}
	}
	return goals[0].amount
}

func run() error {
	reactions, err := loadReactions("input.txt")
	if err != nil {
		return err
	}

	lo := 4_000_000
	hi := 5_000_000
	for {
		mid := lo + ((hi - lo) / 2)
		if mid == lo {
			fmt.Println(mid)
			break
		}
		ore := oreRequiredForFuel(reactions, mid)
		fmt.Printf("%d ORE => %d FUEL\n", ore, mid)
		if ore > 1_000_000_000_000 {
			hi = mid
			continue
		}
		if ore < 1_000_000_000_000 {
			lo = mid
			continue
		}
		fmt.Println(mid)
		break
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
