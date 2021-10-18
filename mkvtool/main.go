package main

import (
	"flag"
	"fmt"
	"github.com/KurenaiRyu/MkvAutoSubset/mkvlib"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
)

const appName = "MKV Tool"
const appVer = "v3.1.4"
const tTitle = appName + " " + appVer

var processer = mkvlib.GetInstance()

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

	flag.BoolVar(&v, "v", false, "Show app info.")
	flag.Parse()

	ec := 0
	if v {
		fmt.Println(appFN + " (powered by " + mkvlib.LibFName + ")")
		return
	}

	if processer == nil {
		ec++
		return
	}

	if len(*asses) > 0 {
		if !processer.ASSFontSubset(*asses, af, ao, !ans) {
			ec++
		}
		return
	}
	if f != "" {
		if d {
			if !processer.DumpMKV(f, data, !n) {
				ec++
			}
			return
		}
		if q {
			r, err := processer.CheckSubset(f)
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
			lines := processer.QueryFolder(s)
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
			v := path.Join(s, "v")
			s := path.Join(s, "s")
			f := path.Join(s, "f")
			o := path.Join(s, "o")
			if !processer.CreateMKVs(v, s, f, "", o, sl, st, clean) {
				ec++
			}
			return
		}
		if d {
			if !processer.DumpMKVs(s, data, !n) {
				ec++
			}
			return
		}
		if m {
			if !processer.MakeMKVs(s, data, dist, sl, st) {
				ec++
			}
			return
		}
		if !processer.DumpMKVs(s, data, true) {
			ec++
		} else if !processer.MakeMKVs(s, data, dist, sl, st) {
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
