package main

import (
	"flag"
	"fmt"
	"github.com/MkvAutoSubset/MkvAutoSubset/mkvlib"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const appName = "MKV Tool"
const appVer = "v4.3.7"
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
	cache_p = path.Join(cache_p, ".mkvtool", "caches")
	ccs, _ := findPath(cache_p, `\.cache$`)
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
	nck := false
	ncks := false
	b := false
	no := false
	t := ""
	e := ""
	sl, st := "", ""
	af, ao := "", ""
	flog := ""
	co := ""
	i := ""
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
	flag.BoolVar(&clean, "clean", false, "Clean original file subtitles and fonts for create mode, or clean old caches for create cache mode.")
	flag.BoolVar(&nck, "nck", false, "Disable check mode.")
	flag.BoolVar(&ncks, "ncks", false, "Disable strict mode for check.")
	flag.StringVar(&sl, "sl", "chi", "Subtitle language. (create & make mode only)")
	flag.StringVar(&st, "st", "", "Subtitle title. (create & make mode only)")
	flag.StringVar(&af, "af", "", "ASS fonts folder. (ass mode only)")
	flag.StringVar(&ao, "ao", "", "ASS output folder. (ass mode only)")
	flag.StringVar(&co, "co", "fonts", "Copy fonts from cache dist folder.")
	flag.StringVar(&cache_p, "cp", cache_p, "Fonts caches dir path. (cache mode only)")
	flag.StringVar(&i, "i", "", "Show font info.")
	flag.BoolVar(&cfc, "cfc", false, "Copy fonts from cache.")
	flag.BoolVar(&ans, "ans", false, `ASS output not to the new "subsetted" folder. (ass mode only)`)
	flag.StringVar(&data, "data", "data", "Subtitles & Fonts folder (dump & make mode only)")
	flag.StringVar(&dist, "dist", "dist", "Results output folder (make mode only)")
	flag.StringVar(&flog, "log", "", "Log file path.")
	flag.StringVar(&pf, "pf", "23.976", "PGS or blank video frame rate:23.976, 24, 25, 30, 29.97, 50, 59.94, 60 or custom fps like 15/1.")
	flag.StringVar(&pr, "pr", "1920*1080", "PGS or blank video resolution:720p, 1080p, 2k, or with custom resolution like 720*480.")
	flag.StringVar(&t, "t", "", `Create test video source path(enter "-" for blank video).`)
	flag.BoolVar(&b, "b", false, `Create test video with burn subtitle.`)
	flag.StringVar(&e, "e", "libx264", `Create test video use encoder.`)
	flag.BoolVar(&no, "no", false, `Disable overwrite mode.`)
	flag.BoolVar(&v, "v", false, "Show app info.")
	flag.Parse()

	if flog != "" {
		lf, err := os.Create(flog)
		if err != nil {
			color.Red(`Failed to create log file: "%s"`, flog)
		}
		mw := io.MultiWriter(colorable.NewColorableStdout(), lf)
		color.Output = mw
		color.NoColor = true
	}

	ec := 0
	defer func() {
		if latestTag != "" && latestTag != appVer {
			color.Green("New version available:%s", latestTag)
		}
		os.Exit(ec)
	}()

	if v {
		color.Green("%s (powered by %s)", appFN, mkvlib.LibFName)
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
	processer.Check(!nck, !ncks)
	processer.NOverwrite(no)

	if i != "" {
		info := processer.GetFontInfo(i)
		if info != nil {
			color.Blue("File: \t%s\n", info.File)
			l := len(info.Fonts)
			for _i := 0; _i < l; _i++ {
				color.Magenta("\nIndex:\t%d\n", _i)
				color.Green("\tNames:\t%s\n", strings.Join(info.Fonts[_i], "\n\t\t"))
				color.HiGreen("\tTypes:\t%s\n", strings.Join(info.Types[_i], "\n\t\t"))
			}
		} else {
			color.Red("Failed to get font info: [%s]", i)
			ec++
		}
		return
	}

	if cc && s != "" {
		if clean {
			_ = os.RemoveAll(cache_p)
		}
		p := path.Join(cache_p, path2MD5(s)+".cache")
		list := processer.CreateFontsCache(s, p, nil)
		el := len(list)
		if el > 0 {
			ec++
			color.Yellow("Error list:(%d)\n%s", el, strings.Join(list, "\n"))
		}
		return
	}

	if cache_p != "" {
		processer.Cache(ccs)
	}

	if l && (s != "" || f != "") {
		files := []string{f}
		if s != "" {
			files, _ = findPath(s, `\.ass$`)
		}
		list := processer.GetFontsList(files, af, nil)
		if len(list[0]) > 0 {
			color.Yellow("Need list: \t%s\n", strings.Join(list[0], "\n\t\t"))
			if len(list[1]) > 0 {
				color.HiYellow("\nMissing list: \t%s\n", strings.Join(list[1], "\n\t\t"))
			} else if !nck {
				color.Green("\n*** All included fonts are found. ***")
			}
		} else {
			color.Yellow("!!! No fonts found. !!!")
		}
		if cfc {
			if !processer.CopyFontsFromCache(files, co, nil) {
				ec++
				return
			}
		}
		return
	}

	if len(*asses) > 0 {
		if !processer.ASSFontSubset(*asses, af, ao, !ans, nil) {
			ec++
		} else if t != "" {
			d, _, _, _ := splitPath((*asses)[0])
			if ao == "" {
				ao = path.Join(d, "subsetted")
			}
			_asses, _ := findPath(ao, `\.ass$`)
			if len(_asses) > 0 {
				processer.CreateTestVideo(_asses, t, ao, e, b, nil)
			}
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
				color.Blue("Need font subset: %v", !r)
			}
			return

		}
	}
	if s != "" {
		if q {
			lines := processer.QueryFolder(s, nil)
			if len(lines) > 0 {
				color.Blue("Has item(s).")
				data := []byte(strings.Join(lines, "\n"))
				if os.WriteFile("list.txt", data, os.ModePerm) != nil {
					color.Red("Failed to write the result file")
					ec++
				}
			} else {
				color.Blue("No item.")
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
			if !processer.MakeMKVs(s, data, dist, sl, st, !n, nil) {
				ec++
			}
			return
		}
		s, _ = filepath.Abs(s)
		dist, _ = filepath.Abs(dist)
		files, _ := findPath(s, `\.mkv$`)
		for _, item := range files {
			if strings.HasPrefix(item, dist) {
				continue
			}
			p := strings.TrimPrefix(item, s)
			_d, _, _, _f := splitPath(p)
			p = path.Join(data, _d, _f)
			if !processer.DumpMKV(item, p, true, nil) {
				ec++
				break
			}
		}
		if ec == 0 {
			if !processer.MakeMKVs(s, data, dist, sl, st, false, nil) {
				ec++
			}
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
