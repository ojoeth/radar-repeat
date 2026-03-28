//go:build debug

package main

func debug(args ...any) {
	for _, arg := range args {
		print(arg)
		print(" ")
	}
	println()
}
