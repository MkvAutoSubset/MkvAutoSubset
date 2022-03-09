//go:build windows && amd64

package mkvlib

import (
	"fmt"
	"path"
	"strconv"
	"syscall"
	"unsafe"
)

func ass2Pgs(input []string, resolution, frameRate int, fontsDir string, output string, lcb logCallback) bool {
	fonts := findFonts(fontsDir)
	r := addFontResource(fonts, lcb)
	if r {
		for _, item := range input {
			_, _, _, _f := splitPath(item)
			fn := path.Join(output, _f+".pgs")
			args := make([]string, 0)
			args = append(args, "-x0")
			args = append(args, "-v144")
			args = append(args, "-s", strconv.Itoa(resolution))
			args = append(args, "-r", strconv.Itoa(frameRate))
			args = append(args, "-i", item)
			args = append(args, fn)
			if p, err := newProcess(nil, nil, nil, "", spp2pgs, args...); err == nil {
				s, err := p.Wait()
				r = err == nil && s.ExitCode() == 1
				if !r {
					printLog(lcb, fmt.Sprintf(`Failed to Ass2Pgs:"%s"`, item))
				}
			}
		}
	}
	removeFontResource(fonts, lcb)
	return r
}

var gdi32 = syscall.NewLazyDLL("gdi32.dll")
var addFontResourceW = gdi32.NewProc("AddFontResourceW")
var removeFontResourceW = gdi32.NewProc("RemoveFontResourceW")

func addFontResource(fonts []string, lcb logCallback) bool {
	ec := 0
	for _, item := range fonts {
		p, _ := syscall.UTF16FromString(item)
		r, _, _ := addFontResourceW.Call(uintptr(unsafe.Pointer(&p[0])))
		if r == 0 {
			printLog(lcb, fmt.Sprintf(`Failed to load font:"%s"`, item))
			ec++
		}
	}
	return ec == 0
}

func removeFontResource(fonts []string, lcb logCallback) bool {
	ec := 0
	for _, item := range fonts {
		p, _ := syscall.UTF16FromString(item)
		r, _, _ := removeFontResourceW.Call(uintptr(unsafe.Pointer(&p[0])))
		if r == 0 {
			printLog(lcb, fmt.Sprintf(`Failed to unload font:"%s"`, item))
			ec++
		}
	}
	return ec == 0
}
