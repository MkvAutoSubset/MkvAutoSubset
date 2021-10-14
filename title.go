//go:build !windows

package main

import "fmt"

func setWindowTitle(title string) {
	fmt.Printf("\033]0;%s\007", title)
}
