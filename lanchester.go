package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

type unit struct {
	shotProb float64
	health   int
}
type force struct {
	forces           []unit
	retreatThreshold float64
	shotProb         float64
	health           int
}

var turns = 0

func (f force) String() string {
	return fmt.Sprintf("force has %v units, each with maximum health %v, a %v kill probability, and a retreat threshold of %v",
		len(f.forces), f.health, f.shotProb, f.retreatThreshold)
}

func createForce(size, health int, shotProb, retreatThreshold float64) force {

	f := force{forces: make([]unit, 0), retreatThreshold: retreatThreshold, shotProb: shotProb, health: health}
	for i := 0; i < size; i++ {
		f.forces = append(f.forces, unit{shotProb, health})

	}
	return f
}

func doCombat(red, blue *force) bool {
	redSize := len(red.forces)
	blueSize := len(blue.forces) //TODO move these into the structs
	for {
		// random (synchronous?) activation first, it's simplest
		//_ = "breakpoint"
		pool := len(red.forces) + len(blue.forces)
		for i := 0; i < pool; i++ {
			active := rand.Intn(pool)
			if active >= len(red.forces) {
				shoot(blue.forces[active-len(red.forces)], red)
			} else {
				shoot(red.forces[active], blue)
			}
		}

		//remove killed units
		removeKilled(red, blue)

		// increment turn
		turns++
		//adjudicate results
		if adjudicate(red, blue, redSize, blueSize) {
			return true
		}
	}
	return false
}

func adjudicate(red, blue *force, redSize, blueSize int) bool {
	_ = "breakpoint"
	if float64(len(red.forces)) < float64(redSize)*red.retreatThreshold || float64(len(blue.forces)) < float64(blueSize)*blue.retreatThreshold {
		return true
	}
	return false
}

func removeKilled(red, blue *force) {
	//_ = "breakpoint"
	for i := 0; i < len(red.forces); i++ {
		_ = "breakpoint"
		if red.forces[i].health <= 0 {
			if i < len(red.forces)-1 {
				red.forces = append(red.forces[:i], red.forces[i+1:]...)
			} else {
				red.forces = red.forces[:i]
			}
			i--
		}
	}
	for i := 0; i < len(blue.forces); i++ {
		if blue.forces[i].health <= 0 {
			if i < len(blue.forces)-1 {
				blue.forces = append(blue.forces[:i], blue.forces[i+1:]...)
			} else {
				blue.forces = blue.forces[:i]
			}
			i--
		}
	}
}

func shoot(a unit, target *force) {
	for i := range target.forces {
		if rand.Float64() < a.shotProb {
			target.forces[i].health--
		}
	}
}

func main() {
	// variable declarations before flag.Parse()
	var redSize int
	var redHealth int
	var redShotProb float64
	var redRetreatThreshold float64
	var blueSize int
	var blueHealth int
	var blueShotProb float64
	var blueRetreatThreshold float64

	// parse flags
	flag.IntVar(&redSize, "rs", 10, "number of red agents")
	flag.IntVar(&blueSize, "bs", 10, "number of blue agents")
	flag.IntVar(&redHealth, "rh", 1, "health of red agents")
	flag.IntVar(&blueHealth, "bh", 1, "health of blue agents")
	flag.Float64Var(&redShotProb, "rp", 0.05, "red shot probability")
	flag.Float64Var(&blueShotProb, "bp", 0.05, "blue shot probability")
	flag.Float64Var(&redRetreatThreshold, "rr", 0.4, "red retreat threshold")
	flag.Float64Var(&blueRetreatThreshold, "br", 0.4, "blue retreat threshold")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())
	red := createForce(redSize, redHealth, redShotProb, redRetreatThreshold)
	blue := createForce(blueSize, blueHealth, blueShotProb, blueRetreatThreshold)

	fmt.Println("Initial model state:")
	fmt.Printf("The red %v.\n", red)
	fmt.Printf("The blue %v.\n", blue)
	doCombat(&red, &blue)
	fmt.Printf("\nModel finished after %v turns.\n\n", turns)
	fmt.Println("Final model state:")
	fmt.Printf("The red %v.\n", red)
	fmt.Printf("The blue %v.\n", blue)
}
