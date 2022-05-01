package main

import (
	"flag"
	"fmt"
	"github.com/KurenaiRyu/MkvAutoSubset/mkvlib"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
)

const appName = "MKV Tool"
const appVer = "v3.6.6"
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

var latestTag = ""

func main() {
	setWindowTitle(tTitle)
	go getLatestTag()
	s := ""
	data := ""
	dist := ""
	cache_p := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		cache_p = os.Getenv("USERPROFILE")
	}
	cache_p = path.Join(cache_p, ".mkvtool/fonts.cache")
	f := ""
	c := false
	d := false
	m := false
	n := false
	q := false
	v := false
	clean := false
	ans := false
	a2p := false
	apc := false
	l := false
	cc := false
	cfc := false
	mks := false
	ck := false
	cks := false
	sl, st := "", ""
	af, ao := "", ""
	flog := ""
	co := ""
	asses := new(arrayArg)
	pf := ""
	pr := ""
	flag.StringVar(&s, "s", "", "Source folder.")
	flag.StringVar(&f, "f", "", "MKV file. (join single mode)")
	flag.BoolVar(&c, "c", false, "Create mode.")
	flag.BoolVar(&d, "d", false, "Dump mode.")
	flag.BoolVar(&m, "m", false, "Make mode.")
	flag.BoolVar(&q, "q", false, "Query mode.")
	flag.BoolVar(&a2p, "a2p", false, "Enable ass2pgs. (need ass2bdnxml)")
	flag.BoolVar(&apc, "apc", false, "Ass and pgs coexist.")
	flag.BoolVar(&mks, "mks", false, "Enable mks mode.")
	flag.BoolVar(&l, "l", false, "Show fonts list.")
	flag.BoolVar(&cc, "cc", false, "Create fonts cache.")
	flag.Var(asses, "a", "ASS files. (multiple & join ass mode)")
	flag.BoolVar(&n, "n", false, "Not do ass font subset & not change font name.")
	flag.BoolVar(&clean, "clean", false, "Clean original file subtitles and fonts. (create mode only)")
	flag.BoolVar(&ck, "ck", false, "Enable check mode.")
	flag.BoolVar(&cks, "cks", false, "Enable strict mode for check.")
	flag.StringVar(&sl, "sl", "chi", "Subtitle language. (create & make mode only)")
	flag.StringVar(&st, "st", "", "Subtitle title. (create & make mode only)")
	flag.StringVar(&af, "af", "", "ASS fonts folder. (ass mode only)")
	flag.StringVar(&ao, "ao", "", "ASS output folder. (ass mode only)")
	flag.StringVar(&co, "co", "fonts", "Copy fonts from cache dist folder.")
	flag.StringVar(&cache_p, "cp", cache_p, "Fonts cache path. (cache mode only)")
	flag.BoolVar(&cfc, "cfc", false, "Copy fonts from cache.")
	flag.BoolVar(&ans, "ans", false, `ASS output not to the new "subsetted" folder. (ass mode only)`)
	flag.StringVar(&data, "data", "data", "Subtitles & Fonts folder (dump & make mode only)")
	flag.StringVar(&dist, "dist", "dist", "Results output folder (make mode only)")
	flag.StringVar(&flog, "log", "", "Log file path.")
	flag.StringVar(&pf, "pf", "23.976", "PGS frame rate:23.976, 24, 25, 30, 29.97, 50, 59.94, 60 or custom fps like 15/1. (ass2pgs only)")
	flag.StringVar(&pr, "pr", "1920*1080", "PGS resolution:720p, 1080p, 2k, or with custom resolution like 720*480. (ass2pgs only)")
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

	defer func() {
		if latestTag != "" && latestTag != appVer {
			log.Printf("New version available:%s", latestTag)
		}
		os.Exit(ec)
	}()

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
	processer.A2P(a2p, apc, pr, pf)
	processer.MKS(mks)
	processer.NRename(n)
	processer.Check(ck, cks)

	if cc && s != "" {
		list := processer.CreateFontsCache(s, cache_p, nil)
		el := len(list)
		if el > 0 {
			ec++
			log.Printf("Error list:(%d)\n%s", el, strings.Join(list, "\n"))
		}
		return
	}

	if cache_p != "" {
		processer.Cache(cache_p)
	}

	if l && s != "" {
		list := processer.GetFontsList(s, nil)
		if len(list) > 0 {
			fmt.Println(strings.Join(list, "\n"))
		}
		if cfc {
			if !processer.CopyFontsFromCache(s, co, nil) {
				ec++
				return
			}
		}
		return
	}

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
}

func getLatestTag() {
	if resp, err := http.DefaultClient.Get("https://api.github.com/repos/MkvAutoSubset/MkvAutoSubset/releases/latest"); err == nil {
		if data, err := ioutil.ReadAll(resp.Body); err == nil {
			reg, _ := regexp.Compile(`"tag_name":"([^"]+)"`)
			arr := reg.FindStringSubmatch(string(data))
			if len(arr) > 1 {
				latestTag = arr[1]
			}
		}
	}
}
