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
const libVer = "3.1.2"

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

var _getter = new(processorGetter)
var _processor = new(mkvProcessor)

func GetProcessorGetterInstance() *processorGetter {
	return _getter
}

func (self *processorGetter) InitProcessorInstance(lcb logCallback) *mkvProcessor {
	self.checked = false
	self.instance = nil

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
	_, _mkvextract := exec.LookPath(mkvextract)
	_, _mkvmerge := exec.LookPath(mkvmerge)
	_, _ffmpeg := exec.LookPath(ffmpeg)
	if _mkvextract != nil || _mkvmerge != nil {
		printLog(lcb, logWarning, `Missing dependency: mkvtoolnix (need "%s" & "%s").`, mkvextract, mkvmerge)
	}

	if _ffmpeg != nil {
		printLog(lcb, logWarning, `Missing dependency: ffmpeg.`)
	}

	self.instance = _processor
	self.instance.ffmpeg = _ffmpeg == nil
	self.instance.mkvextract = _mkvextract == nil
	self.instance.mkvmerge = _mkvmerge == nil

	return self.instance
}

func (self *processorGetter) GetProcessorDummyInstance() *mkvProcessor {
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
