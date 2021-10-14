package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

const pName = "MKV Tool v3.0.2"

type arrayArg []string

func (self *arrayArg) String() string {
	return fmt.Sprintf("%v", []string(*self))
}

func (self *arrayArg) Set(value string) error {
	*self = append(*self, value)
	return nil
}

func main() {
	setWindowTitle(pName)
	s := ""
	c := false
	d := false
	m := false
	n := false
	q := false
	sl, st := "", ""
	af, ao := "", ""
	arr := new(arrayArg)
	flag.StringVar(&s, "s", "", "Source folder.")
	flag.BoolVar(&c, "c", false, "Create mode.")
	flag.BoolVar(&d, "d", false, "Dump mode.")
	flag.BoolVar(&m, "m", false, "Make mode.")
	flag.BoolVar(&q, "q", false, "Query mode.")
	flag.Var(arr, "a", "ASS files. (multiple & join ass mode)")
	flag.BoolVar(&n, "n", false, "Not do ass font subset. (dump mode only)")
	flag.StringVar(&sl, "sl", "chi", " Subtitle language. (create mode only)")
	flag.StringVar(&st, "st", "", " Subtitle title. (create mode only)")
	flag.StringVar(&af, "af", "", " ASS fonts folder. (ASS mode only)")
	flag.StringVar(&ao, "ao", "", " ASS output folder. (ASS mode only)")
	flag.Parse()

	ec := 0
	if len(*arr) > 0 {
		if !genASSes(*arr, af, ao) {
			ec++
		}
		return
	} else if s != "" {
		if q {
			if !queryFolder(s) {
				ec++
			}
			return
		}
		if c {
			if sl != "" {
				if !createMKVs(s, sl, st) {
					ec++
				}
				return
			}
		}
		if d {
			if !dumpMKVs(s, !n) {
				ec++
			}
			return
		}
		if m {
			if !makeMKVs(s) {
				ec++
			}
			return
		}
		if !dumpMKVs(s, true) {
			ec++
		} else if !makeMKVs(s) {
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
