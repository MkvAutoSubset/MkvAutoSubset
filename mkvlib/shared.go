package mkvlib

import (
	"fmt"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const libName = "mkvlib"
const libVer = "v2.4.0"

const LibFName = libName + " " + libVer

const (
	logInfo byte = iota
	logWarning
	logSWarning
	logError
	logProgress
)

type logCallback func(byte, string)

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
	_, _ass2bdnxml := exec.LookPath(ass2bdnxml)
	_, _ffmpeg := exec.LookPath(ffmpeg)
	if _ttx != nil || _pyftsubset != nil {
		printLog(lcb, logError, `Missing dependency: fonttools (need "%s" & "%s").`, ttx, pyftsubset)
		ec++
	}
	if _mkvextract != nil || _mkvmerge != nil {
		printLog(lcb, logError, `Missing dependency: mkvtoolnix (need "%s" & "%s").`, mkvextract, mkvmerge)
		ec++
	}

	if _ass2bdnxml != nil {
		printLog(lcb, logWarning, `Missing dependency: ass2bdnxml.`)
		//ec++
	}

	if _ffmpeg != nil {
		printLog(lcb, logWarning, `Missing dependency: ffmpeg.`)
		//ec++
	}

	r := ec == 0
	if r {
		self.checked = true
		self.instance = new(mkvProcessor)
		self.instance.ass2bdnxml = _ass2bdnxml == nil
		self.instance.ffmpeg = _ffmpeg == nil
	}

	return r
}

func (self *processorGetter) GetProcessorInstance() *mkvProcessor {
	if self.checked {
		return self.instance
	}
	return nil
}

func printLog(lcb logCallback, l byte, f string, v ...interface{}) {
	str := fmt.Sprintf(f, v...)
	if lcb != nil {
		lcb(l, str)
	} else {
		color.New(color.FgWhite).Print(time.Now().Format("2006/01/02 15:04:05 "))
		switch l {
		case logInfo:
			color.Blue(str)
			break
		case logWarning:
			color.Yellow(str)
			break
		case logSWarning:
			color.HiYellow(str)
			break
		case logError:
			color.Red(str)
			break
		case logProgress:
			color.Green(str)
			break
		}
	}
}

func Version() string {
	return libVer
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
