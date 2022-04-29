package mkvlib

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/KurenaiRyu/MkvAutoSubset/mkvlib/parser"
	"github.com/antchfx/xmlquery"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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
	File  string     `json:"file"`
	Fonts [][]string `json:"fonts"`
	Types [][]string `json:"types"`
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
	fg        map[string]string
	seps      []string
}

func (self *assProcessor) parse() bool {
	ec := 0
	self.seps = []string{"-", " "}
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
		opt := parser.SSAOptions{}
		reg, _ := regexp.Compile(`\\fn@?([^\r\n\\\}]*)`)
		_reg, _ := regexp.Compile(`\\([bir])([^\r\n\\\}]*)`)
		__reg, _ := regexp.Compile(`nd\d+`)
		m := make(map[string]string)
		for k, v := range self.subtitles {
			subtitle, err := parser.ReadFromSSAWithOptions(strings.NewReader(v), opt)
			if err != nil {
				ec++
				printLog(self.lcb, `Failed to read the ass file: "%s"`, k)
				continue
			}
			for _, item := range subtitle.Items {
				name := ""
				_b := *item.Style.InlineStyle.SSABold
				_i := *item.Style.InlineStyle.SSAItalic
				for _, _item := range item.Lines {
					for _, __item := range _item.Items {
						if __item.InlineStyle != nil {
							arr := reg.FindStringSubmatch(__item.InlineStyle.SSAEffect)
							if len(arr) > 1 {
								name = arr[1]
							}
							_arr := _reg.FindAllStringSubmatch(__item.InlineStyle.SSAEffect, -1)
							for _, v := range _arr {
								if len(v) > 2 {
									switch v[1] {
									case "b":
										_b = v[2] == "1"
										break
									case "i":
										_i = v[2] == "1"
										break
									case "r":
										if __reg.MatchString(v[2]) {
											break
										}
										v[2] = strings.TrimPrefix(v[2], "*")
										if v[2] == "" {
											name = ""
											_b = *item.Style.InlineStyle.SSABold
											_i = *item.Style.InlineStyle.SSAItalic
										} else if s, ok := subtitle.Styles[v[2]]; ok {
											name = s.InlineStyle.SSAFontName
											_b = *s.InlineStyle.SSABold
											_i = *s.InlineStyle.SSAItalic
										} else {
											printLog(self.lcb, `Not Found style in the ass file:"%s" [%s].`, k, v[2])
											ec++
										}
										break
									}
								}
							}
						}
						if name == "" || name == "0" {
							name = item.Style.InlineStyle.SSAFontName
						}
						if strings.HasPrefix(name, "@") && len(name) > 1 {
							name = name[1:]
						}
						arr := make([]string, 0)
						if _b {
							arr = append(arr, "Bold")
						}
						if _i {
							arr = append(arr, "Italic")
						}
						if !_b && !_i {
							arr = append(arr, "Regular")
						}
						_name := fmt.Sprintf("%s^%s", name, strings.Join(arr, " "))
						m[_name] += __item.Text
					}
				}
			}
		}
		self.m = make(map[string]*fontInfo)
		for k, v := range m {
			if v != "" {
				self.m[k] = new(fontInfo)
				self.m[k].str = v
				self.m[k].oldName = strings.Split(k, "^")[0]
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
	reg, _ := regexp.Compile(`[\x00-\x1f]|(&#([0-9]|[12][0-9]|3[01]);)`)
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
			if ok {
				f, err := ioutil.ReadFile(fn)
				if err == nil {
					str := string(f)
					str = reg.ReplaceAllString(str, "")
					ok = ioutil.WriteFile(fn, []byte(str), os.ModePerm) == nil
				}
			}
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

func (self *assProcessor) getFontName(p string) []map[string]bool {
	f, err := openFile(p, true, false)
	if err == nil {
		defer func() { _ = f.Close() }()
		names := make(map[string]bool)
		types := make(map[string]bool)
		xml, err := xmlquery.Parse(f)
		if err == nil {
			for _, v := range xml.SelectElements(`ttFont/name/namerecord`) {
				id := v.SelectAttr("nameID")
				name := strings.TrimSpace(v.FirstChild.Data)
				switch id {
				case "1", "3", "4", "6":
					names[name] = true
					break
				case "2", "17":
					types[name] = true
					break
				}
			}
		}
		return []map[string]bool{names, types}
	}
	return nil
}

func (self *assProcessor) getFontsName(files []string) map[string][]map[string]bool {
	l := len(files)
	wg := new(sync.WaitGroup)
	wg.Add(l)
	m := new(sync.Mutex)
	_m := make(map[string][]map[string]bool)
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
	self.fg = make(map[string]string)
	reg, _ := regexp.Compile(`_(\d+)\.ttx$`)
	m := make(map[string]map[string][]map[string]bool)
	for font, ttxs := range self._m {
		_m := self.getFontsName(ttxs)
		if len(_m) > 0 {
			m[font] = _m
		}
	}
	w := func(fb bool) {
		for k, _ := range self.m {
			_k := strings.Split(k, "^")
			if self.m[k].file != "" || (fb && _k[1] == "Regular") {
				continue
			}
			if fb {
				printLog(self.lcb, `Font fallback:[%s^%s] -> [%s^Regular]`, _k[0], _k[1], _k[0])
				_k[1] = "Regular"
			}
			seps0 := make([]string, 0)
			seps1 := make([]string, 0)
			for _, v := range self.seps {
				l := strings.LastIndex(_k[0], v)
				tk := ""
				if l > -1 && len(_k[0]) > 1 {
					tk = _k[0][l+1:]
				}
				if tk != "" {
					seps0 = append(seps0, _k[0][:l])
					seps1 = append(seps1, tk)
				}
			}
			_tk := func(q1 bool, qk map[string]bool) bool {
				arr := seps0
				if q1 {
					arr = seps1
				}
				for _, v := range arr {
					if qk[v] {
						return true
					}

				}
				return false
			}
			for __k, v := range m {
				for ___k, _v := range v {
					if (_v[0][_k[0]] || _tk(false, _v[0])) && (_v[1][_k[1]] || _tk(true, _v[1])) {
						self.m[k].file = __k
						self.m[k].index = reg.FindStringSubmatch(___k)[1]
						n := self.fg[_k[0]]
						if n == "" {
							n = randomStr(8)
							self.fg[_k[0]] = n
						}
						self.m[k].newName = n
						break
					}
				}
				if self.m[k].file != "" {
					break
				}
			}
			if self.m[k].file != "" {
				continue
			}
			if f, i := self.matchCache(fmt.Sprintf("%s^%s", _k[0], _k[1])); f != "" {
				self.m[k].file, self.m[k].index = f, i
				n := self.fg[_k[0]]
				if n == "" {
					n = randomStr(8)
					self.fg[_k[0]] = n
				}
				self.m[k].newName = n
			}
		}
	}
	w(false)
	w(true)
	ok := true
	for k, _ := range self.m {
		if self.m[k].file == "" {
			ok = false
			printLog(self.lcb, `Missing the font: "%s".`, k)
		}
	}
	return ok
}

func (self *assProcessor) reMap() {
	m := make(map[string]*fontInfo)
	for _, v := range self.m {
		if _, ok := m[v.newName]; !ok {
			m[v.newName] = v
		} else {
			m[v.newName].str += v.str
		}
	}
	self.m = m
}

func (self *assProcessor) createFontSubset(font *fontInfo) bool {
	ok := false
	fn := fmt.Sprintf(`%s.txt`, font.newName)
	_, fn, _, _ = splitPath(fn)
	fn = path.Join(self.tDir, fn)
	_, n, e, _ := splitPath(font.file)
	if strings.ToLower(e) == ".ttc" {
		e = ".ttf"
	}
	e = strings.ToLower(e)
	if os.MkdirAll(self.output, os.ModePerm) != nil {
		printLog(self.lcb, "Failed to create the output folder.")
		return false
	}
	if os.WriteFile(fn, []byte(font.str), os.ModePerm) == nil {
		_fn := fmt.Sprintf("%s%s", font.newName, e)
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
	self.reMap()
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
								if strings.Contains(buf.String(), "ERROR: Unhandled exception has occurred") {
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
						_ = os.Remove(font.sFont)
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
			for _, v := range self.m {
				if strings.Split(v.oldName, "^")[0] == k {
					comments = append(comments, fmt.Sprintf("; Font subset: %s - %s", v.newName, k))
					break
				}
			}
		}
		if len(comments) > 2 {
			comments = append(comments, "")
			comments = append(comments, "; Processed by "+LibFName+" at "+time.Now().Format("2006-01-02 15:04:05"))
			comments = append(comments, "; -----  Font subset end  -----")
			comments = append(comments, "")
			r := "[Script Info]\n"
			_r := "\n"
			if strings.Contains(s, "[Script Info]\r\n") {
				r = "[Script Info]\r\n"
				_r = "\r\n"
			}
			s = strings.Replace(s, r, strings.Join(comments, _r), 1)
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

func (self *assProcessor) createFontsCache(output string) []string {
	cache := make([]fontCache, 0)
	if !filepath.IsAbs(self._fonts) {
		self._fonts, _ = filepath.Abs(self._fonts)
	}
	fonts := findFonts(self._fonts)
	ok := 0
	l := len(fonts)
	m := new(sync.Mutex)
	if self.tDir == "" {
		self.tDir = path.Join(os.TempDir(), randomStr(8))
		if os.MkdirAll(self.tDir, os.ModePerm) != nil {
			printLog(self.lcb, `Failed to create temp dir: "%s"`, self.tDir)
			return []string{""}
		}
	}
	reg, _ := regexp.Compile(`_(\d+)\.ttx$`)
	wg := new(sync.WaitGroup)
	el := make([]string, 0)
	w := func(s, e int) {
		for i := s; i < e; i++ {
			go func(x int) {
				_item := fonts[x]
				list := self.dumpFont(_item, self.tDir, false)
				if len(list) > 0 {
					m.Lock()
					_m := self.getFontsName(list)
					_fonts := make([][]string, len(_m))
					_types := make([][]string, len(_m))
					for k, v := range _m {
						index := reg.FindStringSubmatch(k)[1]
						q, _ := strconv.Atoi(index)
						_list := make([]string, 0)
						for _k, _ := range v[0] {
							_list = append(_list, _k)
						}
						_fonts[q] = _list
						_list = make([]string, 0)
						for _k, _ := range v[1] {
							_list = append(_list, _k)
						}
						_types[q] = _list
					}
					if len(_fonts) > 0 && len(_types) > 0 {
						ok++
						cache = append(cache, fontCache{_item, _fonts, _types})
						printLog(self.lcb, "Cache font (%d/%d) done.", ok, l)
					} else {
						el = append(el, _item)
					}
					m.Unlock()
				} else {
					el = append(el, _item)
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
	defer func() { _ = os.RemoveAll(self.tDir) }()
	if len(cache) > 0 {
		data, _ := json.Marshal(cache)
		d, _, _, _ := splitPath(output)
		_ = os.MkdirAll(d, os.ModePerm)
		if ioutil.WriteFile(output, data, os.ModePerm) != nil {
			printLog(self.lcb, `Failed to write cache file: "%s"`, output)
		}
	}
	return el
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
	_k := strings.Split(k, "^")
	seps0 := make([]string, 0)
	seps1 := make([]string, 0)
	for _, v := range self.seps {
		l := strings.LastIndex(_k[0], v)
		tk := ""
		if l > -1 && len(_k[0]) > 1 {
			tk = _k[0][l+1:]
		}
		if tk != "" {
			seps0 = append(seps0, _k[0][:l])
			seps1 = append(seps1, tk)
		}
	}
	_tk := func(q1 bool, qk string) bool {
		arr := seps0
		if q1 {
			arr = seps1
		}
		for _, v := range arr {
			if v == qk {
				return true
			}
		}
		return false
	}
	for _, v := range self.cache {
		for q, _v := range v.Fonts {
			for _, __v := range _v {
				if __v == _k[0] || _tk(false, __v) {
					for _, ___v := range v.Types[q] {
						if ___v == _k[1] || _tk(true, ___v) {
							ok = v.File
							i = q
							break
						}
					}
				}
				if ok != "" {
					break
				}
			}
			if ok != "" {
				break
			}
		}
	}
	if _, err := os.Stat(ok); err != nil {
		ok = ""
	}
	return ok, strconv.Itoa(i)
}
