package main

import (
	"fmt"
	"math/rand"
)

//parameter set holds the potential values for a parameter sweep
type parameterSet struct {
	Verbose              bool
	ActivationOrder      []ActivationOrder
	RedSize              []int
	RedHealth            []int
	RedShotProb          []float64
	RedMaxShots          []int
	RedRetreatThreshold  []float64
	BlueSize             []int
	BlueHealth           []int
	BlueShotProb         []float64
	BlueMaxShots         []int
	BlueRetreatThreshold []float64
}

// runOnce uses the base values from parameters.json to feed one run
func runOnce() {
	par = parameters{
		Verbose:              set.Verbose,
		ActivationOrder:      set.ActivationOrder[0],
		RedSize:              set.RedSize[0],
		RedHealth:            set.RedHealth[0],
		RedShotProb:          set.RedShotProb[0],
		RedMaxShots:          set.RedMaxShots[0],
		RedRetreatThreshold:  set.RedRetreatThreshold[0],
		BlueSize:             set.BlueSize[0],
		BlueHealth:           set.BlueHealth[0],
		BlueShotProb:         set.BlueShotProb[0],
		BlueMaxShots:         set.BlueMaxShots[0],
		BlueRetreatThreshold: set.BlueRetreatThreshold[0],
	}
	runModel(par, runNum)

}

// calculateSweep generates the parameter set and returns
// its size so that the user can be warned

// TODO: There is a bug in this function causing it to return
// inaccurate parameter sets due to inaccuracies in floating-point
// arithmetic. This should not be too difficult to fix at some point.
func caluculateSweep() (int, parameterSet) {
	//TODO: Use reflect to make this less horrible.
	_ = "breakpoint"
	sweepSize := 1
	sweepSize *= len(set.ActivationOrder)
	parSet := parameterSet{
		ActivationOrder:      set.ActivationOrder,
		RedSize:              make([]int, 0),
		RedHealth:            make([]int, 0),
		RedShotProb:          make([]float64, 0),
		RedMaxShots:          make([]int, 0),
		RedRetreatThreshold:  make([]float64, 0),
		BlueSize:             make([]int, 0),
		BlueHealth:           make([]int, 0),
		BlueShotProb:         make([]float64, 0),
		BlueMaxShots:         make([]int, 0),
		BlueRetreatThreshold: make([]float64, 0),
	}

	if set.RedSize[2] != 0 {
		sweepSize *= (set.RedSize[1] - (set.RedSize[0] - set.RedSize[2])) / set.RedSize[2]
		for i := set.RedSize[0]; i <= set.RedSize[1]; i += set.RedSize[2] {
			parSet.RedSize = append(parSet.RedSize, i)
		}
	} else {
		parSet.RedSize = append(parSet.RedSize, set.RedSize[0])
	}
	if set.BlueSize[2] != 0 {
		sweepSize *= (set.BlueSize[1] - (set.BlueSize[0] - set.BlueSize[2])) / set.BlueSize[2]
		for i := set.BlueSize[0]; i <= set.BlueSize[1]; i += set.BlueSize[2] {
			parSet.BlueSize = append(parSet.BlueSize, i)
		}
	} else {
		parSet.BlueSize = append(parSet.BlueSize, set.BlueSize[0])
	}

	if set.RedHealth[2] != 0 {
		sweepSize *= (set.RedHealth[1] - (set.RedHealth[0] - set.RedHealth[2])) / set.RedHealth[2]
		for i := set.RedHealth[0]; i <= set.RedHealth[1]; i += set.RedHealth[2] {
			parSet.RedHealth = append(parSet.RedHealth, i)
		}
	} else {
		parSet.RedHealth = append(parSet.RedHealth, set.RedHealth[0])
	}
	if set.BlueHealth[2] != 0 {
		sweepSize *= (set.BlueHealth[1] - (set.BlueHealth[0] - set.BlueHealth[2])) / set.BlueHealth[2]
		for i := set.BlueHealth[0]; i <= set.BlueHealth[1]; i += set.BlueHealth[2] {
			parSet.BlueHealth = append(parSet.BlueHealth, i)
		}
	} else {
		parSet.BlueHealth = append(parSet.BlueHealth, set.BlueHealth[0])
	}

	if set.RedShotProb[2] != 0.0 {
		sweepSize *= int((set.RedShotProb[1] - (set.RedShotProb[0] - set.RedShotProb[2])) / set.RedShotProb[2])
		for i := set.RedShotProb[0]; i <= set.RedShotProb[1]; i += set.RedShotProb[2] {
			parSet.RedShotProb = append(parSet.RedShotProb, i)
		}
	} else {
		parSet.RedShotProb = append(parSet.RedShotProb, set.RedShotProb[0])
	}
	if set.BlueShotProb[2] != 0.0 {
		sweepSize *= int((set.BlueShotProb[1] - (set.BlueShotProb[0] - set.BlueShotProb[2])) / set.BlueShotProb[2])
		for i := set.BlueShotProb[0]; i <= set.BlueShotProb[1]; i += set.BlueShotProb[2] {
			parSet.BlueShotProb = append(parSet.BlueShotProb, i)
		}
	} else {
		parSet.BlueShotProb = append(parSet.BlueShotProb, set.BlueShotProb[0])
	}

	if set.RedMaxShots[2] != 0.0 {
		sweepSize *= int((set.RedMaxShots[1] - (set.RedMaxShots[0] - set.RedMaxShots[2])) / set.RedMaxShots[2])
		for i := set.RedMaxShots[0]; i <= set.RedMaxShots[1]; i += set.RedMaxShots[2] {
			parSet.RedMaxShots = append(parSet.RedMaxShots, i)
		}
	} else {
		parSet.RedMaxShots = append(parSet.RedMaxShots, set.RedMaxShots[0])
	}
	if set.BlueMaxShots[2] != 0.0 {
		sweepSize *= int((set.BlueMaxShots[1] - (set.BlueMaxShots[0] - set.BlueMaxShots[2])) / set.BlueMaxShots[2])
		for i := set.BlueMaxShots[0]; i <= set.BlueMaxShots[1]; i += set.BlueMaxShots[2] {
			parSet.BlueMaxShots = append(parSet.BlueMaxShots, i)
		}
	} else {
		parSet.BlueMaxShots = append(parSet.BlueMaxShots, set.BlueMaxShots[0])
	}

	if set.RedRetreatThreshold[2] != 0.0 {
		sweepSize *= int((set.RedRetreatThreshold[1] - (set.RedRetreatThreshold[0] - set.RedRetreatThreshold[2])) / set.RedRetreatThreshold[2])
		for i := set.RedRetreatThreshold[0]; i <= set.RedRetreatThreshold[1]; i += set.RedRetreatThreshold[2] {
			parSet.RedRetreatThreshold = append(parSet.RedRetreatThreshold, i)
		}
	} else {
		parSet.RedRetreatThreshold = append(parSet.RedRetreatThreshold, set.RedRetreatThreshold[0])
	}
	if set.BlueRetreatThreshold[2] != 0.0 {
		sweepSize *= int((set.BlueRetreatThreshold[1] - (set.BlueRetreatThreshold[0] - set.BlueRetreatThreshold[2])) / set.BlueRetreatThreshold[2])
		for i := set.BlueRetreatThreshold[0]; i <= set.BlueRetreatThreshold[1]; i += set.BlueRetreatThreshold[2] {
			parSet.BlueRetreatThreshold = append(parSet.BlueRetreatThreshold, i)
		}
	} else {
		parSet.BlueRetreatThreshold = append(parSet.BlueRetreatThreshold, set.BlueRetreatThreshold[0])
	}
	return sweepSize * set.Niter, parSet
}

// Execute the parameter sweep
func executeSweep(ps parameterSet) {
	ps.Verbose = set.Verbose
	par.Verbose = ps.Verbose
	// set the parameters to the initial values
	par = parameters{
		Verbose:              set.Verbose,
		ActivationOrder:      set.ActivationOrder[0],
		RedSize:              set.RedSize[0],
		RedHealth:            set.RedHealth[0],
		RedShotProb:          set.RedShotProb[0],
		RedMaxShots:          set.RedMaxShots[0],
		RedRetreatThreshold:  set.RedRetreatThreshold[0],
		BlueSize:             set.BlueSize[0],
		BlueHealth:           set.BlueHealth[0],
		BlueShotProb:         set.BlueShotProb[0],
		BlueMaxShots:         set.BlueMaxShots[0],
		BlueRetreatThreshold: set.BlueRetreatThreshold[0],
	}

	// TRIGGER WARNING
	for _, e := range ps.ActivationOrder {
		par.ActivationOrder = e
		for _, e := range ps.RedSize {
			par.RedSize = e
			for _, e := range ps.RedHealth {
				par.RedHealth = e
				for _, e := range ps.RedShotProb {
					par.RedShotProb = e
					for _, e := range ps.RedMaxShots {
						par.RedMaxShots = e
						for _, e := range ps.RedRetreatThreshold {
							par.RedRetreatThreshold = e
							for _, e := range ps.BlueSize {
								par.BlueSize = e
								for _, e := range ps.BlueHealth {
									par.BlueHealth = e
									for _, e := range ps.BlueShotProb {
										par.BlueShotProb = e
										for _, e := range ps.BlueMaxShots {
											par.BlueMaxShots = e
											for _, e := range ps.BlueRetreatThreshold {
												par.BlueRetreatThreshold = e
												for i := 0; i < set.Niter; i++ {
													runModel(par, runNum)
													runNum++
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func monteCarloRun() {
	for i := 0; i < set.Niter; i++ {

		par = parameters{
			ActivationOrder:      set.ActivationOrder[rand.Intn(len(set.ActivationOrder))],
			RedSize:              rand.Intn(set.RedSize[1]-set.RedSize[0]+1) + set.RedSize[0],
			RedHealth:            rand.Intn(set.RedHealth[1]-set.RedHealth[0]+1) + set.RedHealth[0],
			RedShotProb:          set.RedShotProb[0] + (set.RedShotProb[1]-set.RedShotProb[0])*rand.Float64(),
			RedMaxShots:          rand.Intn(set.RedMaxShots[1]-set.RedMaxShots[0]+1) + set.RedMaxShots[0],
			RedRetreatThreshold:  set.RedRetreatThreshold[0] + (set.RedRetreatThreshold[1]-set.RedRetreatThreshold[0])*rand.Float64(),
			BlueSize:             rand.Intn(set.BlueSize[1]-set.BlueSize[0]+1) + set.BlueSize[0],
			BlueHealth:           rand.Intn(set.BlueHealth[1]-set.BlueHealth[0]+1) + set.BlueHealth[0],
			BlueShotProb:         set.BlueShotProb[0] + (set.BlueShotProb[1]-set.BlueShotProb[0])*rand.Float64(),
			BlueMaxShots:         rand.Intn(set.BlueMaxShots[1]-set.BlueMaxShots[0]+1) + set.BlueMaxShots[0],
			BlueRetreatThreshold: set.BlueRetreatThreshold[0] + (set.BlueRetreatThreshold[1]-set.BlueRetreatThreshold[0])*rand.Float64(),
		}
		runModel(par, runNum)

	}
}

func latinHypercubeRun() {
	fmt.Println("Error: Not yet implemented")

	return
}
