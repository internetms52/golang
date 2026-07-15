package main

import "fmt"

func main() {
	// empty slice literal
	var x = []int{}
	printSlice("empty= ", x)
	// with default values
	var y = []int{2, 4, 6, 8}
	printSlice("default value=", y)
	var z = make([]int, 5)
	printSlice("make=", z)
}

func printSlice(s string, numbers []int) {
	fmt.Print(s)
	for value := range numbers {
		fmt.Printf("%d,", value)
	}
	fmt.Print("\n")
}
