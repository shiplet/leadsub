package utils

import "fmt"

//Printout holds the values for the realtime printouts
type Printout struct {
	Workers  string
	Progress string
}

//Printing is a utility variable for use with HandlePrinting
var Printing Printout

//HandlePrinting prints the multiline readout
func HandlePrinting() {
	fmt.Print("\u001b[1000D")
	fmt.Print("\u001b[2A")
	fmt.Println(Printing.Workers)
	fmt.Println(Printing.Progress)
}

//Min handles math.min style comparison, but with ints instead of float64
func Min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

//Max handles math.max style comparison, but with ints instead of float64
func Max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
