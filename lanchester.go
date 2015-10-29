package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

type ActivationOrder int
type BatchMode int

const (
	randomSynchronous = iota
	uniformSynchronous
	randomAsynchronous
	uniformAsynchronous
)

const (
	singleRun = iota
	parameterSweep
	monteCarlo
	latinHypercube
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

type modelSettings struct {
	BatchMode            BatchMode          `json:"batchMode"`
	Niter                int                `json:"niter"`
	Verbose              bool               `json:"verbose"`
	ActivationOrder      [3]ActivationOrder `json:"activationOrder"`
	RedSize              [3]int             `json:"RedSize"`
	RedHealth            [3]int             `json:"RedHealth"`
	RedShotProb          [3]float64         `json:"RedShotProb"`
	RedRetreatThreshold  [3]float64         `json:"RedRetreatThreshold"`
	BlueSize             [3]int             `json:"BlueSize"`
	BlueHealth           [3]int             `json:"BlueHealth"`
	BlueShotProb         [3]float64         `json:"BlueShotProb"`
	BlueRetreatThreshold [3]float64         `json:"BlueRetreatThreshold"`
}

type parameters struct {
	Verbose              bool
	ActivationOrder      ActivationOrder
	RedSize              int
	RedHealth            int
	RedShotProb          float64
	RedRetreatThreshold  float64
	BlueSize             int
	BlueHealth           int
	BlueShotProb         float64
	BlueRetreatThreshold float64
}

type casualties []int

var turns = 0
var par parameters
var set modelSettings
var runNum = 1

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
func (c casualties) String() string {
	var buffer bytes.Buffer
	for _, x := range c {
		buffer.WriteString(fmt.Sprintf("%v ", x))
	}
	return buffer.String()
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
func doCombatRandomSync(red, blue *force, par parameters) bool {
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
		redKilled, blueKilled := removeKilled(red, blue)
		if par.Verbose {
			printCasualties(redKilled, blueKilled)
		}
		//adjudicate results
		if adjudicate(red, blue, red.forceSize, blue.forceSize, par) {
			return true
		}
	}
	return false
}

func doCombatUniform(red, blue *force, par parameters) bool {
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
		//remove killed units
		redKilled, blueKilled := removeKilled(red, blue)
		if par.Verbose {
			printCasualties(redKilled, blueKilled)
		}

		//adjudicate results
		if adjudicate(red, blue, red.forceSize, blue.forceSize, par) {
			return true
		}

	}
	return false
}
func doCombatRandomAsync(red, blue *force, par parameters) bool {
	for {
		turns++
		for i := 0; i < len(red.forces)+len(blue.forces); i++ {
			x := rand.Intn(len(red.forces) + len(blue.forces))
			if x < len(red.forces) {
				shoot(red.forces[x], blue)
			} else {
				shoot(blue.forces[x-len(red.forces)], red)
			}
			//remove killed units
			redKilled, blueKilled := removeKilled(red, blue)
			if par.Verbose {
				printCasualties(redKilled, blueKilled)
			}

			if adjudicate(red, blue, red.forceSize, blue.forceSize, par) {
				return true
			}

		}

	}
}

func doCombatUniformAsync(red, blue *force, par parameters) bool {
	for {
		turns++
		for i := 0; i < len(red.forces)+len(blue.forces); i++ {
			x := rand.Intn(len(red.forces) + len(blue.forces))
			if x < len(red.forces) {
				shoot(red.forces[x], blue)
			} else {
				shoot(blue.forces[x-len(red.forces)], red)
			}
			//remove killed units
			redKilled, blueKilled := removeKilled(red, blue)
			if par.Verbose {
				printCasualties(redKilled, blueKilled)
			}

			if adjudicate(red, blue, red.forceSize, blue.forceSize, par) {
				return true
			}

		}

	}
}

//Determine if one force should retreat. This should be refactored to determine a winner/loser.
func adjudicate(red, blue *force, RedSize, BlueSize int, par parameters) bool {
	_ = "breakpoint"
	if float64(len(red.forces)) < float64(RedSize)*red.retreatThreshold || float64(len(blue.forces)) < float64(BlueSize)*blue.retreatThreshold {
		return true
	}
	return false
}

//Remove all forces with health = 0. Return array of killed units.
func removeKilled(red, blue *force) (casualties, casualties) {
	redKilled := make([]int, 0)
	blueKilled := make([]int, 0)
	for i := 0; i < len(red.forces); i++ {
		if red.forces[i].health <= 0 {
			redKilled = append(redKilled, i)
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
			blueKilled = append(blueKilled, i)
			if i < len(blue.forces)-1 {
				blue.forces = append(blue.forces[:i], blue.forces[i+1:]...)
			} else {
				blue.forces = blue.forces[:i]
			}
			i--
		}
	}
	return redKilled, blueKilled
}

//One agent shoots at all opposing agents.
func shoot(a unit, target *force) {
	for i := range target.forces {
		if rand.Float64() < a.shotProb {
			target.forces[i].health--
		}
	}
}

func printCasualties(r, b casualties) {
	printR := len(r) > 0
	printB := len(b) > 0
	if printR || printB {
		fmt.Println()
		if printR {
			fmt.Printf("Red forces: %vkilled\n", r.String())
		} else {
			fmt.Printf("Blue forces: %vkilled\n", b.String())
		}
	}

}

func runModel(par parameters, runNum int) {
	// initialize forces
	red := createForce(par.RedSize, par.RedHealth, par.RedShotProb, par.RedRetreatThreshold)
	blue := createForce(par.BlueSize, par.BlueHealth, par.BlueShotProb, par.BlueRetreatThreshold)

	fmt.Println()
	fmt.Printf("Starting run number %v \n", runNum)
	fmt.Println("Initial model state:")
	fmt.Printf("The red force has %v.\n", red)
	fmt.Printf("The blue force has %v.\n", blue)
	fmt.Printf("Running model with %v activation:\n", par.ActivationOrder)
	if par.ActivationOrder == randomSynchronous {
		doCombatRandomSync(&red, &blue, par)
	} else if par.ActivationOrder == uniformSynchronous {
		doCombatUniform(&red, &blue, par)
	} else if par.ActivationOrder == randomAsynchronous {
		doCombatRandomAsync(&red, &blue, par)
	} else if par.ActivationOrder == uniformAsynchronous {
		doCombatUniformAsync(&red, &blue, par)
	}
	fmt.Printf("\nModel finished after %v turns.\n\n", turns)
	fmt.Println("Final model state:")
	fmt.Printf("The red force has %v.\n", red)
	fmt.Printf("The blue force %v.\n", blue)

}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Please provide a JSON file with the appropriate model parameters")
		os.Exit(1)
	}
	file, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("Error opening parameter file")
		os.Exit(2)
	}
	err = json.Unmarshal(file, &set)
	if err != nil {
		fmt.Println("Error parsing JSON")
		os.Exit(3)
	}

	rand.Seed(time.Now().UnixNano())

	if set.BatchMode == 0 {
		par = parameters{
			Verbose:              set.Verbose,
			ActivationOrder:      set.ActivationOrder[0],
			RedSize:              set.RedSize[0],
			RedHealth:            set.RedHealth[0],
			RedShotProb:          set.RedShotProb[0],
			RedRetreatThreshold:  set.RedRetreatThreshold[0],
			BlueSize:             set.BlueSize[0],
			BlueHealth:           set.BlueHealth[0],
			BlueShotProb:         set.BlueShotProb[0],
			BlueRetreatThreshold: set.BlueRetreatThreshold[0],
		}
		runModel(par, runNum)
	}

}
