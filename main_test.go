package main

import (
	"fmt"
	"testing"
)

func BenchmarkTest(b *testing.B) {
	first, second, fullName := "janko", "kondic", ""
	for i := 0; i < b.N; i++ {
		fullName = first + " " + second
	}

	fmt.Println(fullName)
}

func BenchmarkTest2(b *testing.B) {
	first, second, fullName := "janko", "kondic", ""
	for i := 0; i < b.N; i++ {
		fullName = fmt.Sprintf("%s %s", first, second)
	}

	fmt.Println(fullName)
}
