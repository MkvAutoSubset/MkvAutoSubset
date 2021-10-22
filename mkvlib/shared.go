package mkvlib

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const libName = "mkvlib"
const libVer = "v1.1.3"

const LibFName = libName + " " + libVer

type logCallback func(string)

type processorGetter struct {
	checked  bool
	instance *mkvProcessor
}

var _instance = new(processorGetter)

func GetProcessorGetterInstance() *processorGetter {
	return _instance
}

func (self *processorGetter) InitProcessorInstance(lcb logCallback) bool {
	self.checked = false
	self.instance = nil

	ec := 0
	n := "PATH"
	s := ":"
	if runtime.GOOS == "windows" {
		n = "path"
		s = ";"
	}
	p := os.Getenv(n)
	if !strings.HasSuffix(p, s) {
		p += s
	}
	e, _ := os.Executable()
	e, _ = filepath.Split(e)
	p += e
	_ = os.Setenv(n, p)
	_, _ttx := exec.LookPath(ttx)
	_, _pyftsubset := exec.LookPath(pyftsubset)
	_, _mkvextract := exec.LookPath(mkvextract)
	_, _mkvmerge := exec.LookPath(mkvmerge)
	if _ttx != nil || _pyftsubset != nil {
		printLog(lcb, `Missing dependency: fonttools (need "%s" & "%s").`, ttx, pyftsubset)
		ec++
	}
	if _mkvextract != nil || _mkvmerge != nil {
		printLog(lcb, `Missing dependency: mkvtoolnix (need "%s" & "%s").`, mkvextract, mkvmerge)
		ec++
	}

	r := ec == 0
	if r {
		self.checked = true
		self.instance = new(mkvProcessor)
	}

	return r
}

func (self *processorGetter) GetProcessorInstance() *mkvProcessor {
	if self.checked {
		return self.instance
	}
	return nil
}

func printLog(lcb logCallback, f string, v ...interface{}) {
	if lcb != nil {
		lcb(fmt.Sprintf(f, v...))
	} else {
		log.Printf(f, v...)
	}
}
