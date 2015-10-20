package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

type ActivationOrder int

const (
	randomSynchronous = iota
	uniformSynchronous
	randomAsynchronous
	uniformAsynchronous
)

type unit struct {
	shotProb float64
	health   int
}

type force struct {
	forces           []unit
	forceSize        int
	retreatThreshold float64
	shotProb         float64
	health           int
}

type parameters struct {
	ActivationOrder      ActivationOrder `json:"activationOrder"`
	RedSize              int             `json:"RedSize"`
	RedHealth            int             `json:"RedHealth"`
	RedShotProb          float64         `json:"RedShotProb"`
	RedRetreatThreshold  float64         `json:"RedRetreatThreshold"`
	BlueSize             int             `json:"BlueSize"`
	BlueHealth           int             `json:"BlueHealth"`
	BlueShotProb         float64         `json:"BlueShotProb"`
	BlueRetreatThreshold float64         `json:"BlueRetreatThreshold"`
}

var turns = 0

//Implement Stringer
func (f force) String() string {
	return fmt.Sprintf("%v units, each with maximum health %v, a %v kill probability, and a retreat threshold of %v",
		len(f.forces), f.health, f.shotProb, f.retreatThreshold)
}

func (a ActivationOrder) String() string {
	if a == 0 {
		return fmt.Sprintf("random synchronous")
	} else if a == 1 {
		return fmt.Sprintf("uniform synchronous")
	} else if a == 2 {
		return fmt.Sprintf("random asynchronous")
	} else if a == 3 {
		return fmt.Sprintf("uniform asynchronous")
	}
	return "undefined"
}

//Initialize and return a force: a collection of units
func createForce(size, health int, shotProb, retreatThreshold float64) force {

	f := force{forces: make([]unit, 0),
		forceSize:        size,
		retreatThreshold: retreatThreshold,
		shotProb:         shotProb,
		health:           health}
	for i := 0; i < size; i++ {
		f.forces = append(f.forces, unit{shotProb, health})

	}
	return f
}

//Adjucate combat. This should probably be restructured into a scheduling function.
func doCombat(red, blue *force) bool {
	for {
		// increment turn
		turns++

		// random (synchronous?) activation first, it's simplest
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
		_ = removeKilled(red, blue)

		//adjudicate results
		if adjudicate(red, blue, red.forceSize, blue.forceSize) {
			return true
		}
	}
	return false
}

func doCombatUniform(red, blue *force) bool {
	for {
		// increment turn
		turns++

		// uniform (synchronous?) activation
		turnList := rand.Perm(len(red.forces) + len(blue.forces))
		for _, e := range turnList {
			if e >= len(red.forces) {
				shoot(blue.forces[e-len(red.forces)], red)
			} else {
				shoot(red.forces[e], blue)
			}

		}

		// remove killed units
		_ = removeKilled(red, blue)

		//adjudicate results
		if adjudicate(red, blue, red.forceSize, blue.forceSize) {
			return true
		}

	}
	return false
}
func doCombatRandomAsync(red, blue *force) bool {
	for {
		turns++
		for i := 0; i < len(red.forces)+len(blue.forces); i++ {
			x := rand.Intn(len(red.forces) + len(blue.forces))
			if x < len(red.forces) {
				shoot(red.forces[x], blue)
			} else {
				shoot(blue.forces[x-len(red.forces)], red)
			}
			_ := removeKilled(red, blue)
			if adjudicate(red, blue, red.forceSize, blue.forceSize) {
				return true
			}

		}

	}
}

func doCombatUniformAsync(red, blue *force) bool {
	for {
		turns++
		for i := 0; i < len(red.forces)+len(blue.forces); i++ {
			x := rand.Intn(len(red.forces) + len(blue.forces))
			if x < len(red.forces) {
				shoot(red.forces[x], blue)
			} else {
				shoot(blue.forces[x-len(red.forces)], red)
			}
			y := removeKilled(red, blue)
			fmt.Printf("Killed: %v\n", y)
			if adjudicate(red, blue, red.forceSize, blue.forceSize) {
				return true
			}

		}

	}
}

//Determine if one force should retreat. This should be refactored to determine a winner/loser.
func adjudicate(red, blue *force, RedSize, BlueSize int) bool {
	_ = "breakpoint"
	if float64(len(red.forces)) < float64(RedSize)*red.retreatThreshold || float64(len(blue.forces)) < float64(BlueSize)*blue.retreatThreshold {
		return true
	}
	return false
}

//Remove all forces with health = 0. Return array of killed units.
func removeKilled(red, blue *force) []int {
	killed := make([]int, 0)
	redSize := len(red.forces)
	for i := 0; i < len(red.forces); i++ {
		if red.forces[i].health <= 0 {
			killed = append(killed, i)
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
			killed = append(killed, i+redSize)
			if i < len(blue.forces)-1 {
				blue.forces = append(blue.forces[:i], blue.forces[i+1:]...)
			} else {
				blue.forces = blue.forces[:i]
			}
			i--
		}
	}
	return killed
}

//One agent shoots at all opposing agents.
func shoot(a unit, target *force) {
	for i := range target.forces {
		if rand.Float64() < a.shotProb {
			target.forces[i].health--
		}
	}
}

func main() {
	_ = "breakpoint"
	file, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("Error opening parameter file")
		os.Exit(1)
	}
	var par parameters
	err = json.Unmarshal(file, &par)
	if err != nil {
		fmt.Println("Error parsing JSON")
		os.Exit(2)
	}
	rand.Seed(time.Now().UnixNano())
	// initialize forces
	red := createForce(par.RedSize, par.RedHealth, par.RedShotProb, par.RedRetreatThreshold)
	blue := createForce(par.BlueSize, par.BlueHealth, par.BlueShotProb, par.BlueRetreatThreshold)

	fmt.Println("Initial model state:")
	fmt.Printf("The red force has %v.\n", red)
	fmt.Printf("The blue force has %v.\n", blue)
	fmt.Printf("Running model with %v activation:\n", par.ActivationOrder)
	if par.ActivationOrder == randomSynchronous {
		doCombat(&red, &blue)
	} else if par.ActivationOrder == uniformSynchronous {
		doCombatUniform(&red, &blue)
	} else if par.ActivationOrder == randomAsynchronous {
		doCombatRandomAsync(&red, &blue)
	} else if par.ActivationOrder == uniformAsynchronous {
		doCombatUniformAsync(&red, &blue)
	}
	fmt.Printf("\nModel finished after %v turns.\n\n", turns)
	fmt.Println("Final model state:")
	fmt.Printf("The red force has %v.\n", red)
	fmt.Printf("The blue force %v.\n", blue)
}
