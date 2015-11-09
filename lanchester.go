package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

var f *os.File
var w *csv.Writer

//enums
type ActivationOrder int
type BatchMode int
type Outcome int

const (
	incomplete = iota
	redVictory
	blueVictory
	tie
)
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
	maxShots int
	shotProb float64
	health   int
}

type force struct {
	forces           []unit
	forceSize        int
	retreatThreshold float64
	shotProb         float64
	maxShots         int
	health           int
}

type modelSettings struct {
	Filename             string            `json:"filename"`
	WriteDynamics        bool              `json:"writeDynamics"`
	BatchMode            BatchMode         `json:"batchMode"`
	Niter                int               `json:"niter"`
	Verbose              bool              `json:"verbose"`
	ActivationOrder      []ActivationOrder `json:"activationOrder"`
	RedSize              [3]int            `json:"RedSize"`
	RedHealth            [3]int            `json:"RedHealth"`
	RedShotProb          [3]float64        `json:"RedShotProb"`
	RedMaxShots          [3]int            `json:"RedMaxShots"`
	RedRetreatThreshold  [3]float64        `json:"RedRetreatThreshold"`
	BlueSize             [3]int            `json:"BlueSize"`
	BlueHealth           [3]int            `json:"BlueHealth"`
	BlueShotProb         [3]float64        `json:"BlueShotProb"`
	BlueMaxShots         [3]int            `json:"BlueMaxShots"`
	BlueRetreatThreshold [3]float64        `json:"BlueRetreatThreshold"`
}

type parameters struct {
	Verbose              bool
	ActivationOrder      ActivationOrder
	RedSize              int
	RedHealth            int
	RedShotProb          float64
	RedMaxShots          int
	RedRetreatThreshold  float64
	BlueSize             int
	BlueHealth           int
	BlueShotProb         float64
	BlueMaxShots         int
	BlueRetreatThreshold float64
}

type casualties []int

var turns = 0
var par parameters
var set modelSettings
var runNum = 1
var writeToFile = false

//Implement Stringer
func (f force) String() string {
	return fmt.Sprintf("%v units, each with maximum health %v, a %v kill probability, and a retreat threshold of %v",
		len(f.forces), f.health, f.shotProb, f.retreatThreshold)
}

func (a ActivationOrder) String() string {
	if a == 0 {
		return fmt.Sprintf("random-synchronous")
	} else if a == 1 {
		return fmt.Sprintf("uniform-synchronous")
	} else if a == 2 {
		return fmt.Sprintf("random-asynchronous")
	} else if a == 3 {
		return fmt.Sprintf("uniform-asynchronous")
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

func (o Outcome) String() string {
	if o == 0 {
		return fmt.Sprintf("incomplete")
	} else if o == 1 {
		return fmt.Sprintf("red-victory")
	} else if o == 2 {
		return fmt.Sprintf("blue-victory")
	} else if o == 3 {
		return fmt.Sprintf("stalemate")
	} else {
		return fmt.Sprintf("error")
	}
}

//Initialize and return a force: a collection of units
func createForce(size, health, maxShots int, shotProb, retreatThreshold float64) force {
	f := force{forces: make([]unit, 0),
		forceSize:        size,
		retreatThreshold: retreatThreshold,
		shotProb:         shotProb,
		health:           health,
		maxShots:         maxShots}
	for i := 0; i < size; i++ {
		f.forces = append(f.forces, unit{maxShots, shotProb, health})
	}
	return f
}

func doCombatRandomSync(red, blue *force, par parameters) bool {
	for {
		// increment turn
		turns++

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
		if status := adjudicate(red, blue, red.forceSize, blue.forceSize, par); status != incomplete {
			if writeToFile {
				writeLine(*red, *blue, status)
			}

			return true
		}
		if writeToFile && set.WriteDynamics {
			writeLine(*red, *blue, incomplete)
		}
	}
	return false
}

func doCombatUniform(red, blue *force, par parameters) bool {
	for {
		// increment turn
		turns++

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
		if status := adjudicate(red, blue, red.forceSize, blue.forceSize, par); status != incomplete {
			if writeToFile {
				writeLine(*red, *blue, status)
			}

			return true
		}
		if writeToFile && set.WriteDynamics {
			writeLine(*red, *blue, incomplete)
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

			if status := adjudicate(red, blue, red.forceSize, blue.forceSize, par); status != incomplete {
				if writeToFile {
					writeLine(*red, *blue, status)
				}

				return true
			}
		}
		if writeToFile && set.WriteDynamics {
			writeLine(*red, *blue, incomplete)
		}
	}
	return false
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

			if status := adjudicate(red, blue, red.forceSize, blue.forceSize, par); status != incomplete {
				if writeToFile {
					writeLine(*red, *blue, status)
				}

				return true
			}
		}
		if writeToFile && set.WriteDynamics {
			writeLine(*red, *blue, incomplete)
		}
	}
	return false
}

// Determine if one force should retreat.
func adjudicate(red, blue *force, RedSize, BlueSize int, par parameters) Outcome {
	_ = "breakpoint"
	if float64(len(red.forces)) <= float64(RedSize)*red.retreatThreshold && float64(len(blue.forces)) <= float64(BlueSize)*blue.retreatThreshold {
		return tie
	} else if float64(len(red.forces)) <= float64(RedSize)*red.retreatThreshold {
		return blueVictory
	} else if float64(len(blue.forces)) <= float64(BlueSize)*blue.retreatThreshold {
		return redVictory
	}
	return incomplete
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
	x := a.maxShots
	for i := range target.forces { //TODO: non-random; doesn't matter unless I add heterogeneity
		if rand.Float64() < a.shotProb && x > 0 {
			target.forces[i].health--
		}
		x--
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

// Write one line to the csv
func writeLine(r, b force, status Outcome) {
	s := make([]string, 16)
	s[0] = fmt.Sprintf("%v", runNum)
	s[1] = fmt.Sprintf("%v", par.ActivationOrder)
	s[2] = fmt.Sprintf("%v", par.RedSize)
	s[3] = fmt.Sprintf("%v", par.RedHealth)
	s[4] = fmt.Sprintf("%.3v", par.RedShotProb)
	s[5] = fmt.Sprintf("%v", par.RedMaxShots)
	s[6] = fmt.Sprintf("%.3v", par.RedRetreatThreshold)
	s[7] = fmt.Sprintf("%v", len(r.forces))
	s[8] = fmt.Sprintf("%v", par.BlueSize)
	s[9] = fmt.Sprintf("%v", par.BlueHealth)
	s[10] = fmt.Sprintf("%.3v", par.BlueShotProb)
	s[11] = fmt.Sprintf("%v", par.BlueMaxShots)
	s[12] = fmt.Sprintf("%.3v", par.BlueRetreatThreshold)
	s[13] = fmt.Sprintf("%v", len(b.forces))
	s[14] = fmt.Sprintf("%v", status)
	s[15] = fmt.Sprintf("%v", turns)
	_ = "breakpoint"
	err := w.Write(s)
	if err != nil {
		panic(err)
	}
	w.Flush()
}

// Write csv headers
func writeHeader() {
	headers := []string{"run", "activation-order", "red-size", "red-health", "red-shot-prob", "red-max-shots", "red-retreat-threshold", "red-forces", "blue-size", "blue-health", "blue-shot-prob", "blue-max-shots", "blue-retreat-threshold", "blue-forces", "victor", "turns"}
	err := w.Write(headers)
	if err != nil {
		panic(err)
	}
}

// Wrapper function for handling different activation orders
func runModel(par parameters, runNum int) {
	// initialize forces
	red := createForce(par.RedSize, par.RedHealth, par.RedMaxShots, par.RedShotProb, par.RedRetreatThreshold)
	blue := createForce(par.BlueSize, par.BlueHealth, par.BlueMaxShots, par.BlueShotProb, par.BlueRetreatThreshold)

	//reset turns
	turns = 0

	//TODO: multiple verbosity levels
	if set.Verbose {
		fmt.Println()
		fmt.Printf("Starting run number %v \n", runNum)
		fmt.Println("Initial model state:")
		fmt.Printf("The red force has %v.\n", red)
		fmt.Printf("The blue force has %v.\n", blue)
		fmt.Printf("Running model with %v activation:\n", par.ActivationOrder)
	}
	if par.ActivationOrder == randomSynchronous {
		doCombatRandomSync(&red, &blue, par)
	} else if par.ActivationOrder == uniformSynchronous {
		doCombatUniform(&red, &blue, par)
	} else if par.ActivationOrder == randomAsynchronous {
		doCombatRandomAsync(&red, &blue, par)
	} else if par.ActivationOrder == uniformAsynchronous {
		doCombatUniformAsync(&red, &blue, par)
	}
	if set.Verbose {
		fmt.Printf("\nModel finished after %v turns.\n\n", turns)
		fmt.Println("Final model state:")
		fmt.Printf("The red force has %v.\n", red)
		fmt.Printf("The blue force %v.\n", blue)
	}
}

func main() {

	var file []byte
	if len(os.Args) <= 1 {
		// if no argument is specified, see if you can load the default file
		if _, err := os.Stat("parameters.json"); !os.IsNotExist(err) {
			fmt.Println("Using default parameter settings...")
			file, err = ioutil.ReadFile("parameters.json")
			if err != nil {
				fmt.Println("Error opening parameter file")
				os.Exit(2)
			}
		} else {
			fmt.Println("Please provide a JSON file with the appropriate model parameters")
			os.Exit(1)
		}
	} else {
		var err error
		file, err = ioutil.ReadFile(os.Args[1])
		if err != nil {
			fmt.Println("Error opening parameter file")
			os.Exit(2)
		}
	}
	// parse the JSON into model settings
	err := json.Unmarshal(file, &set)
	if err != nil {
		fmt.Println("Error parsing JSON")
		os.Exit(3)
	}

	// if there is a specified filename, writing to file is enabled, so create the file
	// this will clobber the file
	// TODO: prevent doing something stupid, like overwriting the source file
	if set.Filename != "" {
		writeToFile = true
		var err error // paranoid about shadowing f
		f, err = os.Create(set.Filename)
		if err != nil {
			panic(err)
		}

		defer f.Close()
		w = csv.NewWriter(f)
		defer w.Flush()

		writeHeader()

	}
	rand.Seed(time.Now().UnixNano())

	switch set.BatchMode {
	case singleRun:
		runOnce()

	case parameterSweep:
		// if running a parameter sweep, check with the user to make sure
		// they know how many runs they're doing
		sweepSize, ps := caluculateSweep()
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Run parameter sweep with %v runs? (Y/n):  ", sweepSize)
		text, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		text = strings.TrimRight(text, "\n")

		if text == "N" || text == "n" {
			fmt.Println("Cancelling")
		} else {
			fmt.Println("Continuing with parameter sweep")
			executeSweep(ps)
		}
	case monteCarlo:
		monteCarloRun()
	}
}
