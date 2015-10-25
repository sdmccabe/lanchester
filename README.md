Lanchester Combat Model
Stefan McCabe
CSS 739

This is a basic agent-based model to simulate Lanchesterian combat, inspired by Ilachniski (2003). A set of red and blue forces are specified with parameters *size*, *kill probability*, *health*, and *retreat threshold*. Combat works as follows: Agents are selected (using some specified activation order) at random and shoot at the opposing force. Like in Ilachinski (2003), one agent shoots at all enemy agents simultaneously.  With some probability, a shot reduces the enemy's health by one. At zero health, the enemy is killed and removed from the model.  If either force sustains casualties proportionate to their retreat theshold, the model ends. 

USAGE: The model takes a JSON file describing the model parameters as an argument at runtime. This seemed like a more modular approach than hard-coding the parameters into the model.  At some point I intend to write a simple script to more easily generate this file. A default parameter file, parameters.json, is included in the repository. 


TODO: 
- Finish model parameter input for easy batch runs
- Allow a parameter to limit the number of shots fired by an agent per turn (as in EINSTein)
- Verify that my decision to "kill" agents by removing them from the array rather than changing some state variable isn't biasing activation.
- Formalize model outputs.
