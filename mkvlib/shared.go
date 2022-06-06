package mkvlib

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const libName = "mkvlib"
const libVer = "v2.1.6"

const LibFName = libName + " " + libVer

const (
	LogInfo byte = iota
	LogWarning
	LogSWarning
	LogError
	LogProgress
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
		PrintLog(lcb, LogError, `Missing dependency: fonttools (need "%s" & "%s").`, ttx, pyftsubset)
		ec++
	}
	if _mkvextract != nil || _mkvmerge != nil {
		PrintLog(lcb, LogError, `Missing dependency: mkvtoolnix (need "%s" & "%s").`, mkvextract, mkvmerge)
		ec++
	}

	if _ass2bdnxml != nil {
		PrintLog(lcb, LogWarning, `Missing dependency: ass2bdnxml.`)
		//ec++
	}

	if _ffmpeg != nil {
		PrintLog(lcb, LogWarning, `Missing dependency: ffmpeg.`)
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

func PrintLog(lcb logCallback, l byte, f string, v ...interface{}) {
	str := fmt.Sprintf(f, v...)
	if lcb != nil {
		lcb(l, str)
	} else {
		switch l {
		case LogInfo:
			str = color.BlueString(str)
			break
		case LogWarning:
			str = color.YellowString(str)
			break
		case LogSWarning:
			str = color.HiYellowString(str)
			break
		case LogError:
			str = color.RedString(str)
			break
		case LogProgress:
			str = color.GreenString(str)
			break
		}
		log.Print(str)
	}
}

func Version() string {
	return libVer
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
