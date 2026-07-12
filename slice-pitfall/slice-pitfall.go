package main

import "fmt"

func main() {
	runWithSideEffect()
	runWithoutSideEffect()
}

func modifyWithSideEffect(s []int) {
	s = append(s, 999) // 假設 cap 夠
	s[0] = 100         // 改寫 index 0 的值
}

func modifyWithoutSideEffect(s []int) []int {
	duplicate := make([]int, len(s))
	copy(duplicate, s)
	duplicate = append(duplicate, 999)
	duplicate[0] = 100
	return duplicate
}

func runWithSideEffect() {
	original := make([]int, 3, 5)
	modifyWithSideEffect(original)
	fmt.Println(len(original)) // 3,還是 3!
	fmt.Println(original)
}

func runWithoutSideEffect() {
	original := make([]int, 3, 5)
	original = modifyWithoutSideEffect(original)
	fmt.Println(len(original))
	fmt.Println(original)
}
