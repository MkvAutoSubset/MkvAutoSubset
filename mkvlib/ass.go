package mkvlib

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/asticode/go-astisub"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ttx        = "ttx"
	pyftsubset = "pyftsubset"
)

type fontInfo struct {
	file    string
	str     string
	index   string
	oldName string
	newName string
	sFont   string
}

type fontCache struct {
	Font   string      `json:"font"`
	Caches []cacheInfo `json:"caches"`
}

type cacheInfo struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
}

type assProcessor struct {
	files     []string
	_files    []string
	_fonts    string
	output    string
	m         map[string]*fontInfo
	fonts     []string
	subtitles map[string]string
	lcb       logCallback
	cache     []fontCache
	tDir      string
	_m        map[string][]string
}

func (self *assProcessor) parse() bool {
	ec := 0
	self.subtitles = make(map[string]string)
	for _, file := range self.files {
		f, err := openFile(file, true, false)
		if err != nil {
			ec++
		} else {
			data, _ := io.ReadAll(f)
			str := string(data)
			if err == nil {
				self.subtitles[file] = str
			} else {
				ec++
			}
		}
		if ec > 0 {
			printLog(self.lcb, `Failed to read the ass file: "%s"`, file)
		}
	}
	if ec == 0 {
		opt := astisub.SSAOptions{}
		reg, _ := regexp.Compile(`\{?\\fn@?([^\r\n\\\}]+)[\\\}]`)
		m := make(map[string]map[rune]bool)
		for k, v := range self.subtitles {
			subtitle, err := astisub.ReadFromSSAWithOptions(strings.NewReader(v), opt)
			if err != nil {
				ec++
				printLog(self.lcb, `Failed to read the ass file: "%s"`, k)
				continue
			}
			for _, item := range subtitle.Items {
				for _, _item := range item.Lines {
					for _, __item := range _item.Items {
						name := item.Style.InlineStyle.SSAFontName
						if __item.InlineStyle != nil {
							arr := reg.FindStringSubmatch(__item.InlineStyle.SSAEffect)
							if len(arr) > 1 {
								name = arr[1]
							}
						}
						if strings.HasPrefix(name, "@") && len(name) > 1 {
							name = name[1:]
						}
						if m[name] == nil {
							m[name] = make(map[rune]bool)
						}
						str := __item.Text
						for _, char := range str {
							m[name][char] = true
						}
					}
				}
			}
		}
		self.m = make(map[string]*fontInfo)
		reg, _ = regexp.Compile("[A-Za-z0-9]]")
		for k, v := range m {
			str := ""
			for _k, _ := range v {
				str += string(_k)
			}
			str = strings.TrimSpace(str)
			if str != "" {
				str = reg.ReplaceAllString(str, "")
				str += "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
				reg, _ = regexp.Compile("[１２３４５６７８９０]")
				if reg.MatchString(str) {
					str = reg.ReplaceAllString(str, "")
					str += "１２３４５６７８９０"
				}
				self.m[k] = new(fontInfo)
				self.m[k].str = str
				self.m[k].oldName = k
			}
		}
	}
	if len(self.m) == 0 {
		printLog(self.lcb, `Not Found item in the ass file(s): "%d"`, len(self.files))
	}
	return ec == 0
}

func (self *assProcessor) getFontsList() []string {
	list := make([]string, 0)
	for k, _ := range self.m {
		list = append(list, k)
	}
	return list
}

func (self *assProcessor) getTTCCount(file string) int {
	f, err := openFile(file, true, false)
	if err == nil {
		defer func() { _ = f.Close() }()
		data := make([]byte, 4)
		if n, err := f.ReadAt(data, 8); err == nil && n == 4 {
			return int(binary.BigEndian.Uint32(data))
		}
	}
	return 0
}

func (self *assProcessor) dumpFont(file, out string, full bool) []string {
	ok := false
	count := 1
	_, n, _, _ := splitPath(file)
	list := make([]string, 0)
	if strings.HasSuffix(strings.ToLower(n), ".ttc") {
		count = self.getTTCCount(file)
		if count < 1 {
			printLog(self.lcb, `Failed to get the ttc font count: "%s".`, n)
			return list
		}
	}
	for i := 0; i < count; i++ {
		fn := fmt.Sprintf("%s_%d.ttx", file, i)
		if out != "" {
			_, fn, _, _ = splitPath(fn)
			fn = path.Join(out, fn)
		}
		args := make([]string, 0)
		args = append(args, "-q")
		args = append(args, "-f")
		args = append(args, "-y", strconv.Itoa(i))
		args = append(args, "-o", fn)
		if !full {
			args = append(args, "-t", "name")
		}
		args = append(args, file)
		if p, err := newProcess(nil, nil, nil, "", ttx, args...); err == nil {
			s, err := p.Wait()
			ok = err == nil && s.ExitCode() == 0
		}
		if !ok {
			printLog(self.lcb, `Failed to dump font(%t): "%s"[%d].`, full, n, i)
		} else {
			list = append(list, fn)
		}
	}
	return list
}

func (self *assProcessor) dumpFonts(files []string, full bool) bool {
	if self.tDir == "" {
		self.tDir = path.Join(os.TempDir(), randomStr(8))
		if os.MkdirAll(self.tDir, os.ModePerm) != nil {
			return false
		}
	}
	ok := 0
	l := len(files)
	wg := new(sync.WaitGroup)
	wg.Add(l)
	m := new(sync.Mutex)
	if self._m == nil {
		self._m = make(map[string][]string)
	}
	for _, item := range files {
		go func(_item string) {
			_ok := self.dumpFont(_item, self.tDir, full)
			if len(_ok) > 0 {
				m.Lock()
				ok++
				self._m[_item] = _ok
				m.Unlock()
			}
			wg.Done()
		}(item)
	}
	wg.Wait()
	return ok == l
}

func (self *assProcessor) getFontName(p string) []string {
	f, err := openFile(p, true, false)
	if err == nil {
		defer func() { _ = f.Close() }()
		names := make([]string, 0)
		if xml, err := xmlquery.Parse(f); err == nil {
			for _, v := range xml.SelectElements(`ttFont/name/namerecord[@platformID=3]`) {
				id := v.SelectAttr("nameID")
				name := strings.TrimSpace(v.FirstChild.Data)
				switch id {
				case "1":
					names = append(names, name)
					break
				case "4":
					names = append(names, name)
					break
				}
			}
		}
		return names
	}
	return nil
}

func (self *assProcessor) getFontsName(files []string) map[string][]string {
	l := len(files)
	wg := new(sync.WaitGroup)
	wg.Add(l)
	m := new(sync.Mutex)
	_m := make(map[string][]string)
	for _, item := range files {
		go func(_item string) {
			names := self.getFontName(_item)
			if len(names) > 0 {
				m.Lock()
				_m[_item] = names
				m.Unlock()
			}
			wg.Done()
		}(item)
	}
	wg.Wait()
	return _m
}

func (self *assProcessor) matchFonts() bool {
	if !self.dumpFonts(self.fonts, false) {
		return false
	}
	reg, _ := regexp.Compile(`_(\d+)\.ttx$`)
	for font, ttxs := range self._m {
		m := self.getFontsName(ttxs)
		if len(m) > 0 {
			for k, _ := range self.m {
				for _k, v := range m {
					for _, _v := range v {
						if _v == k {
							self.m[k].file = font
							self.m[k].index = reg.FindStringSubmatch(_k)[1]
							self.m[k].newName = randomStr(8)
							break
						}
					}
				}
			}
		}
	}
	ok := true
	for k, v := range self.m {
		if v.file == "" {
			if f, i := self.matchCache(k); f != "" {
				self.m[k].file, self.m[k].index = f, i
				self.m[k].newName = randomStr(8)
			} else {
				ok = false
				printLog(self.lcb, `Missing the font: "%s".`, v.oldName)
			}
		}
	}
	return ok
}

func (self *assProcessor) createFontSubset(font *fontInfo) bool {
	ok := false
	fn := fmt.Sprintf(`%s.txt`, font.file)
	_, fn, _, _ = splitPath(fn)
	fn = path.Join(self.tDir, fn)
	_, n, e, ne := splitPath(font.file)
	if strings.ToLower(e) == ".ttc" {
		e = ".ttf"
	}
	e = strings.ToLower(e)
	if os.MkdirAll(self.output, os.ModePerm) != nil {
		printLog(self.lcb, "Failed to create the output folder.")
		return false
	}
	if os.WriteFile(fn, []byte(font.str), os.ModePerm) == nil {
		_fn := fmt.Sprintf("%s.%s%s", ne, font.newName, e)
		_fn = path.Join(self.output, _fn)
		args := make([]string, 0)
		args = append(args, "--text-file="+fn)
		args = append(args, "--output-file="+_fn)
		args = append(args, "--name-languages="+"*")
		args = append(args, "--font-number="+font.index)
		args = append(args, font.file)
		if p, err := newProcess(nil, nil, nil, "", pyftsubset, args...); err == nil {
			s, err := p.Wait()
			ok = err == nil && s.ExitCode() == 0
		}
		if !ok {
			printLog(self.lcb, `Failed to subset font: "%s"[%s].`, n, font.index)
		} else {
			font.sFont = _fn
		}

	} else {
		printLog(self.lcb, `Failed to write the font text: "%s".`, n)
	}
	return ok
}

func (self *assProcessor) createFontsSubset() bool {
	err := os.RemoveAll(self.output)
	if !(err == nil || err == os.ErrNotExist) {
		printLog(self.lcb, "Failed to clean the output folder.")
		return false
	}
	ok := 0
	l := len(self.m)
	wg := new(sync.WaitGroup)
	wg.Add(l)
	m := new(sync.Mutex)
	for _, item := range self.m {
		go func(_item *fontInfo) {
			_ok := self.createFontSubset(_item)
			if _ok {
				m.Lock()
				ok++
				m.Unlock()
			}
			wg.Done()
		}(item)
	}
	wg.Wait()
	return ok == l
}

func (self *assProcessor) changeFontName(font *fontInfo) bool {
	if self.tDir == "" {
		self.tDir = path.Join(os.TempDir(), randomStr(8))
		if os.MkdirAll(self.tDir, os.ModePerm) != nil {
			return false
		}
	}
	ec := 0
	if len(self.dumpFont(font.sFont, self.tDir, true)) > 0 {
		fn := fmt.Sprintf("%s_0.ttx", font.sFont)
		_, fn, _, _ = splitPath(fn)
		fn = path.Join(self.tDir, fn)
		f, err := openFile(fn, true, false)
		if err == nil {
			defer func() {
				_ = f.Close()
				_ = os.Remove(fn)
			}()
			if xml, err := xmlquery.Parse(f); err == nil {
				for _, v := range xml.SelectElements(`ttFont/name/namerecord`) {
					id := v.SelectAttr("nameID")
					switch id {
					case "0":
						v.FirstChild.Data = "Processed by " + LibFName + " at " + time.Now().Format("2006-01-02 15:04:05")
						break
					case "1", "3", "4", "6":
						v.FirstChild.Data = font.newName
						break
					}
				}
				str := `<?xml version="1.0" encoding="UTF-8"?>`
				str += xml.SelectElement("ttFont").OutputXML(true)
				if os.WriteFile(fn, []byte(str), os.ModePerm) == nil {
					args := make([]string, 0)
					args = append(args, "-q")
					args = append(args, "-f")
					args = append(args, "-o", font.sFont)
					args = append(args, fn)
					ok := false
					buf := bytes.NewBufferString("")
					if p, err := newProcess(nil, nil, buf, "", ttx, args...); err == nil {
						r := true
						go func() {
							for r {
								time.Sleep(500 * time.Millisecond)
								if strings.Contains(buf.String(), "(Hit any key to exit)") {
									_ = p.Kill()
									break
								}
							}
						}()
						s, err := p.Wait()
						r = false
						ok = err == nil && s.ExitCode() == 0
					}
					if !ok {
						ec++
						_, n, _, _ := splitPath(font.sFont)
						printLog(self.lcb, `Failed to compile the font: "%s".`, n)
					}
				}
			} else {
				printLog(self.lcb, `Faild to change the font name: "%s".`, font.oldName)
			}
		}
	}
	return ec == 0
}

func (self *assProcessor) changeFontsName() bool {
	ok := 0
	l := len(self.m)
	wg := new(sync.WaitGroup)
	wg.Add(l)
	m := new(sync.Mutex)
	for _, item := range self.m {
		go func(_item *fontInfo) {
			_ok := self.changeFontName(_item)
			if _ok {
				m.Lock()
				ok++
				m.Unlock()
			}
			wg.Done()
		}(item)
	}
	wg.Wait()
	_ = os.RemoveAll(self.tDir)
	return ok == l
}

func (self *assProcessor) replaceFontNameInAss() bool {
	ec := 0
	m := make(map[string]map[string]bool)
	for _, v := range self.m {
		for f, s := range self.subtitles {
			if m[f] == nil {
				m[f] = make(map[string]bool)
			}
			n := regexp.QuoteMeta(v.oldName)
			reg, _ := regexp.Compile(fmt.Sprintf(`(Style:[^,\r\n]+,|\\fn)(@?)%s([,\\\}])`, n))
			if reg.MatchString(s) {
				r := fmt.Sprintf("${1}${2}%s${3}", v.newName)
				s = reg.ReplaceAllString(s, r)
				m[f][v.oldName] = true
				self.subtitles[f] = s
			}
		}
	}
	for f, s := range self.subtitles {
		comments := make([]string, 0)
		comments = append(comments, "[Script Info]")
		comments = append(comments, "; ----- Font subset begin -----")
		for k, _ := range m[f] {
			comments = append(comments, fmt.Sprintf("; Font subset: %s - %s", self.m[k].newName, k))
		}
		if len(comments) > 2 {
			comments = append(comments, "")
			comments = append(comments, "; Processed by "+LibFName+" at "+time.Now().Format("2006-01-02 15:04:05"))
			comments = append(comments, "; -----  Font subset end  -----")
			comments = append(comments, "")
			s = strings.Replace(s, "[Script Info]\r\n", strings.Join(comments, "\r\n"), 1)
			_, n, _, _ := splitPath(f)
			fn := path.Join(self.output, n)
			ok := false
			if os.WriteFile(fn, []byte(s), os.ModePerm) == nil {
				ok = true
				self._files = append(self._files, fn)
			} else {
				ec++
			}
			if !ok {
				printLog(self.lcb, `Failed to write the new ass file: "%s".`, fn)
			}
		}
	}
	return ec == 0
}

func (self *assProcessor) createFontsCache(output string) bool {
	cache := make([]fontCache, 0)
	fonts := findFonts(self._fonts)
	ok := 0
	l := len(fonts)
	m := new(sync.Mutex)
	if self.tDir == "" {
		self.tDir = path.Join(os.TempDir(), randomStr(8))
		if os.MkdirAll(self.tDir, os.ModePerm) != nil {
			return false
		}
	}
	reg, _ := regexp.Compile(`_(\d+)\.ttx$`)
	wg := new(sync.WaitGroup)
	w := func(s, e int) {
		for i := s; i < e; i++ {
			go func(x int) {
				_item := fonts[x]
				list := self.dumpFont(_item, self.tDir, false)
				if len(list) > 0 {
					m.Lock()
					ok++
					_m := self.getFontsName(list)
					__m := make(map[string]bool)
					caches := make([]cacheInfo, 0)
					for k, v := range _m {
						index := reg.FindStringSubmatch(k)[1]
						for _, _v := range v {
							__m[_v] = true
						}
						for _k, _ := range __m {
							q, _ := strconv.Atoi(index)
							caches = append(caches, cacheInfo{_k, q})
						}
					}
					cache = append(cache, fontCache{_item, caches})
					printLog(self.lcb, "Cache font (%d/%d) done.", ok, l)
					m.Unlock()
				}
				wg.Done()
			}(i)
		}
	}

	c := 100
	x := l / c
	y := l % c
	for i := 0; i < x; i++ {
		wg.Add(c)
		w(i*c, (i+1)*c)
		wg.Wait()
	}
	if y > 0 {
		wg.Add(y)
		w(x*c, l)
		wg.Wait()
	}
	data, _ := json.Marshal(cache)
	defer func() { _ = os.RemoveAll(self.tDir) }()
	d, _, _, _ := splitPath(output)
	_ = os.MkdirAll(d, os.ModePerm)
	return ioutil.WriteFile(output, data, os.ModePerm) == nil
}

func (self *assProcessor) copyFontsFromCache() bool {
	ec := 0
	if self.parse() {
		l := len(self.m)
		i := 0
		for k, _ := range self.m {
			ok, _ := self.matchCache(k)
			if ok != "" {
				_, fn, _, _ := splitPath(ok)
				fn = path.Join(self.output, fn)
				if copyFile(ok, fn) == nil {
					i++
					printLog(self.lcb, "Copy (%d/%d) done.", i, l)
				}
			} else {
				ec++
				printLog(self.lcb, `Missing the font: "%s".`, k)
			}
		}
	}
	return ec == 0
}

func (self *assProcessor) loadCache(p string) {
	if data, err := ioutil.ReadFile(p); err == nil {
		self.cache = make([]fontCache, 0)
		_ = json.Unmarshal(data, &self.cache)
	}
}

func (self *assProcessor) matchCache(k string) (string, string) {
	ok := ""
	i := -1
	for _, v := range self.cache {
		for _, _v := range v.Caches {
			if k == _v.Name {
				ok = v.Font
				i = _v.Index
			}
			if ok != "" {
				break
			}
		}
		if ok != "" {
			break
		}
	}
	if _, err := os.Stat(ok); err != nil {
		ok = ""
	}
	return ok, strconv.Itoa(i)
}
