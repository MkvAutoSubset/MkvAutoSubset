package main

import (
	"flag"
	"fmt"
	"github.com/KurenaiRyu/MkvAutoSubset/mkvlib"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
)

const appName = "MKV Tool"
const appVer = "v3.2.5"
const tTitle = appName + " " + appVer

var appFN = fmt.Sprintf("%s %s %s/%s", appName, appVer, runtime.GOOS, runtime.GOARCH)

type arrayArg []string

func (self *arrayArg) String() string {
	return fmt.Sprintf("%v", []string(*self))
}

func (self *arrayArg) Set(value string) error {
	*self = append(*self, value)
	return nil
}

func main() {
	setWindowTitle(tTitle)
	s := ""
	data := ""
	dist := ""
	f := ""
	c := false
	d := false
	m := false
	n := false
	q := false
	v := false
	clean := false
	ans := false
	sl, st := "", ""
	af, ao := "", ""
	flog := ""
	asses := new(arrayArg)
	flag.StringVar(&s, "s", "", "Source folder.")
	flag.StringVar(&f, "f", "", "MKV file. (join single mode)")
	flag.BoolVar(&c, "c", false, "Create mode.")
	flag.BoolVar(&d, "d", false, "Dump mode.")
	flag.BoolVar(&m, "m", false, "Make mode.")
	flag.BoolVar(&q, "q", false, "Query mode.")
	flag.Var(asses, "a", "ASS files. (multiple & join ass mode)")
	flag.BoolVar(&n, "n", false, "Not do ass font subset. (dump mode only)")
	flag.BoolVar(&clean, "clean", false, "Clean original file subtitles and fonts. (create mode only)")
	flag.StringVar(&sl, "sl", "chi", "Subtitle language. (create & make mode only)")
	flag.StringVar(&st, "st", "", "Subtitle title. (create & make mode only)")
	flag.StringVar(&af, "af", "", "ASS fonts folder. (ass mode only)")
	flag.StringVar(&ao, "ao", "", "ASS output folder. (ass mode only)")
	flag.BoolVar(&ans, "ans", false, `ASS output not to the new "subsetted" folder. (ass mode only)`)
	flag.StringVar(&data, "data", "data", "Subtitles & Fonts folder (dump & make mode only)")
	flag.StringVar(&dist, "dist", "dist", "Results output folder (make mode only)")
	flag.StringVar(&flog, "log", "", "Log file path")

	flag.BoolVar(&v, "v", false, "Show app info.")
	flag.Parse()

	if flog != "" {
		lf, err := os.Create(flog)
		if err != nil {
			log.Printf(`Failed to create log file: "%s"`, flog)
		}
		mw := io.MultiWriter(os.Stdout, lf)
		log.SetOutput(mw)
	}

	ec := 0
	if v {
		log.Printf("%s (powered by %s)", appFN, mkvlib.LibFName)
		return
	}
	getter := mkvlib.GetProcessorGetterInstance()
	if !getter.InitProcessorInstance(nil) {
		ec++
		return
	}

	processer := getter.GetProcessorInstance()

	if len(*asses) > 0 {
		if !processer.ASSFontSubset(*asses, af, ao, !ans, nil) {
			ec++
		}
		return
	}
	if f != "" {
		if d {
			if !processer.DumpMKV(f, data, !n, nil) {
				ec++
			}
			return
		}
		if q {
			r, err := processer.CheckSubset(f, nil)
			if err {
				ec++
			} else {
				log.Printf("Need font subset: %v", !r)
			}
			return

		}
	}
	if s != "" {
		if q {
			lines := processer.QueryFolder(s, nil)
			if len(lines) > 0 {
				log.Printf("Has item(s).")
				data := []byte(strings.Join(lines, "\n"))
				if os.WriteFile("list.txt", data, os.ModePerm) != nil {
					log.Printf("Faild to write the result file")
					ec++
				}
			} else {
				log.Printf("No item.")
			}
			return
		}
		if c {
			_v := path.Join(s, "v")
			_s := path.Join(s, "s")
			_f := path.Join(s, "f")
			_o := path.Join(s, "o")
			if !processer.CreateMKVs(_v, _s, _f, "", _o, sl, st, clean, nil) {
				ec++
			}
			return
		}
		if d {
			if !processer.DumpMKVs(s, data, !n, nil) {
				ec++
			}
			return
		}
		if m {
			if !processer.MakeMKVs(s, data, dist, sl, st, nil) {
				ec++
			}
			return
		}
		if !processer.DumpMKVs(s, data, true, nil) {
			ec++
		} else if !processer.MakeMKVs(s, data, dist, sl, st, nil) {
			ec++
		}
		return
	} else {
		ec++
		flag.PrintDefaults()
	}
	defer os.Exit(ec)
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
