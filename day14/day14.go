package day14

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

type Reagant = struct {
	chemical string
	quantity uint
}

func parseReagant(s string) Reagant {
	fields := strings.Fields(s)
	if len(fields) != 2 {
		log.Fatalf("Passed string with more than 2 fields: %s", s)
	}
	quantity, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return Reagant{fields[1], uint(quantity)}
}

func parseLine(s string) ([]Reagant, Reagant) {
	res := strings.Split(s, "=>")
	lhsStrings := strings.Split(res[0], ",")
	lhs := make([]Reagant, len(lhsStrings))
	for n, r := range lhsStrings {
		lhs[n] = parseReagant(r)
	}
	rhs := parseReagant(res[1])
	return lhs, rhs
}

func getNonOre(rlist map[string]uint, order map[string]int) string {
	if len(rlist) == 0 {
		return ""
	}

	// prioitize the order
	bestSoFar := math.MaxInt
	bestChoice := ""
	for r := range rlist {
		if r != "ORE" && order[r] < bestSoFar {
			bestSoFar = order[r]
			bestChoice = r
		}
	}
	return bestChoice
}

func getMinimumOreForFuel(
	fuelAmount uint,
	order map[string]int,
	reagants map[string]Reagant,
	productions map[string][]Reagant,
) uint {
	needed := make(map[string]uint)
	surplus := make(map[string]uint)
	needed["FUEL"] = fuelAmount

	for chemical := "FUEL"; chemical != ""; chemical = getNonOre(needed, order) {
		if surplus[chemical] >= needed[chemical] {
			// eat it from the surplus
			surplus[chemical] -= needed[chemical]
			delete(needed, chemical)
			continue
		}

		// amount applies to the production rule, r has a different required
		// amount
		required := productions[chemical]
		amount := reagants[chemical].quantity
		times := needed[chemical] / amount
		generatesSurplus := needed[chemical]%amount != 0
		if generatesSurplus {
			times += 1
		}
		surplus[chemical] += (amount*times - needed[chemical])

		delete(needed, chemical)
		for _, ingredient := range required {
			// these are new requirements on top of the existing requirements.
			needed[ingredient.chemical] += ingredient.quantity * times
		}
		// is that it?  I assume we need to figure out some excess thing.
	}
	// I suppose we could try to use the surplus to undo some required ore.
	// doesn't seem to work for example3
	// we can do a combinatorial explosion.  that might not work though.
	return needed["ORE"]
}

func Run(f *os.File, partTwo bool) {
	scanner := bufio.NewScanner(f)
	reagants := make(map[string]Reagant)
	productions := make(map[string][]Reagant)
	for scanner.Scan() {
		line := scanner.Text()
		dependents, production := parseLine(line)
		reagants[production.chemical] = production
		productions[production.chemical] = dependents
	}

	graph := make(map[string][]Reagant)
	for k, v := range productions {
		graph[k] = v
	}
	inDegree := make(map[string]int)
	for _, v := range graph {
		for _, e := range v {
			inDegree[e.chemical]++
		}
	}

	queue := make([]string, 0)
	queue = append(queue, "FUEL")
	count := 0
	order := make(map[string]int, 0)
	for len(queue) > 0 {
		item := queue[0]
		order[item] = count
		count++
		queue = queue[1:]
		for _, e := range graph[item] {
			inDegree[e.chemical]--
			if inDegree[e.chemical] == 0 {
				queue = append(queue, e.chemical)
			}
		}
		graph[item] = []Reagant{}
	}

	if !partTwo {
		fmt.Println(getMinimumOreForFuel(1, order, reagants, productions))
		return
	}

	maxOre := uint(1_000_000_000_000)
	minOreForOne := getMinimumOreForFuel(1, order, reagants, productions)
	fuelGuessLow := maxOre / minOreForOne
	fuelGuessHigh := maxOre / minOreForOne * 2

	if getMinimumOreForFuel(fuelGuessLow, order, reagants, productions) > maxOre {
		panic("Minimum guess was too large")
	}

	if getMinimumOreForFuel(fuelGuessHigh, order, reagants, productions) < maxOre {
		panic("Maximum guess was too high")
	}

	for {
		if fuelGuessHigh == fuelGuessLow || fuelGuessHigh == fuelGuessLow+1 {
			// this is our answer (or I guess -1?)  we can run again to confirm.
			fmt.Println(fuelGuessLow)
			break
		}
		fuelGuess := (fuelGuessHigh + fuelGuessLow) / 2
		ore := getMinimumOreForFuel(fuelGuess, order, reagants, productions)
		if ore < maxOre {
			fuelGuessLow = fuelGuess
		} else if ore > maxOre {
			fuelGuessHigh = fuelGuess
		} else if ore == maxOre {
			// I guess this is our answer too
			fmt.Println(fuelGuess)
			break
		}
	}
}
