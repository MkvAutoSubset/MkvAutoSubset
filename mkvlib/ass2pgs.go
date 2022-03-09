//go:build !windows || (windows && 386)

package mkvlib

func ass2Pgs(input []string, resolution, frameRate int, fontsDir string, output string, lcb logCallback) bool {
	return false
}
