//go:build !windows || (windows && !amd64)

package mkvlib

func ass2Pgs(input []string, resolution, frameRate int, fontsDir string, output string, lcb logCallback) bool {
	printLog(lcb, "Only work in win64.")
	return false
}
