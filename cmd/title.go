//go:build linux

package main

func setWindowTitle(title string) {
	fmt.Printf("\033]0;%s\007", title)
}
