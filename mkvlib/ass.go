package mkvlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/KurenaiRyu/MkvAutoSubset/mkvlib/parser"
	"github.com/KurenaiRyu/MkvAutoSubset/mkvlib/parser/sfnt"
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
	runes   []rune
	index   int
	oldName string
	newName string
	sFont   string
}

type fontCache struct {
	File  string     `json:"file"`
	Fonts [][]string `json:"fonts"`
	Types [][]string `json:"types"`
}

type cacheInfo struct {
	File  string
	Names [][]map[string]bool
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
	cache     []cacheInfo
	fg        map[string]string
	seps      []string
	rename    bool
	check     bool
	strict    bool
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
		__reg, _ := regexp.Compile(`nd[xyz]?\d+`)
		___reg, _ := regexp.Compile(`\d`)
		m := make(map[string]string)
		for k, v := range self.subtitles {
			subtitle, err := parser.ReadFromSSAWithOptions(strings.NewReader(v), opt)
			if err != nil {
				ec++
				printLog(self.lcb, `Failed to parse the ass file: "%s" [%s]`, k, err)
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
				if ___reg.MatchString(v) {
					v += "0123456789"
				}
				self.m[k] = new(fontInfo)
				self.m[k].runes = []rune(v)
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

func (self *assProcessor) dumpFont(file, out string) bool {
	ok := false
	_, n, _, _ := splitPath(file)
	reg, _ := regexp.Compile(`[\x00-\x1f]`)
	args := make([]string, 0)
	args = append(args, "-q")
	args = append(args, "-f")
	args = append(args, "-o", out)
	args = append(args, file)
	if p, err := newProcess(nil, nil, nil, "", ttx, args...); err == nil {
		s, err := p.Wait()
		ok = err == nil && s.ExitCode() == 0
		if ok {
			f, err := ioutil.ReadFile(out)
			if err == nil {
				str := string(f)
				str = reg.ReplaceAllString(str, "")
				ok = ioutil.WriteFile(out, []byte(str), os.ModePerm) == nil
			}
		}
	}
	if !ok {
		printLog(self.lcb, `Failed to dump font: "%s".`, n)
	}
	return ok
}

func (self *assProcessor) getFontName(p string) [][]map[string]bool {
	w := func(_font *sfnt.Font) []map[string]bool {
		names := make(map[string]bool)
		types := make(map[string]bool)
		id1, _ := _font.Name(nil, sfnt.NameIDFamily)
		id2, _ := _font.Name(nil, sfnt.NameIDSubfamily)
		id4, _ := _font.Name(nil, sfnt.NameIDFull)
		id6, _ := _font.Name(nil, sfnt.NameIDPostScript)
		if id1 != nil {
			for _, v := range id1 {
				if v != "" {
					names[v] = true
				}
			}
		}
		if id4 != nil {
			for _, v := range id4 {
				if v != "" {
					names[v] = true
				}
			}
		}
		if id6 != nil {
			for _, v := range id6 {
				if v != "" {
					names[v] = true
				}
			}
		}
		if id2 != nil {
			for _, v := range id2 {
				if v != "" {
					types[v] = true
				}
			}
		}
		return []map[string]bool{names, types}
	}
	list := make([][]map[string]bool, 0)
	f, err := openFile(p, true, false)
	defer func() { _ = f.Close() }()
	if err == nil {
		data, err := ioutil.ReadAll(f)
		if err == nil {
			fonts := make([]*sfnt.Font, 0)
			if strings.HasSuffix(strings.ToLower(p), ".ttc") {
				c, err := sfnt.ParseCollection(data)
				if err == nil {
					l := c.NumFonts()
					for i := 0; i < l; i++ {
						_f, err := c.Font(i)
						if err == nil {
							fonts = append(fonts, _f)
						}
					}
				}
			} else {
				_f, err := sfnt.Parse(data)
				if err == nil {
					fonts = append(fonts, _f)
				}
			}
			for _, _font := range fonts {
				list = append(list, w(_font))
			}
		}
	}
	return list
}

func (self *assProcessor) getFontsName(files []string) map[string][][]map[string]bool {
	l := len(files)
	wg := new(sync.WaitGroup)
	wg.Add(l)
	m := new(sync.Mutex)
	_m := make(map[string][][]map[string]bool)
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

func (self *assProcessor) checkFontMissing(f *fontInfo, i int, c bool) bool {
	_str := ""
	_runes := make([]rune, 0)
	_f, err := os.Open(f.file)
	if err == nil {
		defer func() { _ = _f.Name() }()
		data, err := ioutil.ReadAll(_f)
		if err == nil {
			var _font *sfnt.Font
			if strings.HasSuffix(strings.ToLower(f.file), ".ttc") {
				c, err := sfnt.ParseCollection(data)
				if err == nil {
					_font, _ = c.Font(f.index)
				}
			} else {
				_font, _ = sfnt.Parse(data)
			}
			if _font != nil {
				for _, r := range f.runes {
					n, _ := _font.GlyphIndex(nil, r)
					if n == 0 {
						if r == '\u0020' || r == '\u00a0' {
							_runes = append(_runes, r)
						} else {
							_str += string(r)
						}
					}
				}
			}
		}
	}
	h := "N"
	if c {
		h = "C"
	}
	if _str != "" {
		_str = stringDeduplication(_str)
		printLog(self.lcb, `{%s%02d}Font [%s] missing nomal char(s): "%s"`, h, i, f.oldName, _str)
	}
	if len(_runes) > 0 {
		_str = ""
		for _, _rune := range _runes {
			_str += fmt.Sprintf(`,0x%x`, _rune)
		}
		_str = _str[1:]
		printLog(self.lcb, `{%s%02d}Font [%s] missing special char(s): "%s"`, h, i, f.oldName, _str)
	}
	return _str == "" && len(_runes) == 0
}

func (self *assProcessor) matchFonts() bool {
	self.fg = make(map[string]string)
	fonts := findFonts(self._fonts)
	m := self.getFontsName(fonts)
	_count := make(map[string]int)
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
			for __k, v := range m {
				for ___k, _v := range v {
					if self.matchFontName(_v, _k) {
						self.m[k].file = __k
						self.m[k].index = ___k
						if self.check {
							_count[_k[0]]++
							if !self.checkFontMissing(self.m[k], _count[_k[0]], false) && self.strict {
								self.m[k].file = ""
								self.m[k].index = 0
								continue
							}
						}
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

func (self *assProcessor) fontNameToMap(m []map[string]bool) map[string]map[string]bool {
	_m := make(map[string]map[string]bool)
	for name, _ := range m[0] {
		for family, _ := range m[1] {
			if _, ok := _m[name]; !ok {
				_m[name] = make(map[string]bool)
			}
			_m[name][family] = true
		}
	}
	return _m
}

func (self *assProcessor) matchFontName(m []map[string]bool, _k []string) bool {
	names := make(map[string]string)

	names[_k[0]] = _k[1]
	for _, v := range self.seps {
		l := strings.LastIndex(_k[0], v)
		tk := ""
		if l > -1 && len(_k[0]) > 1 {
			tk = _k[0][l+1:]
		}
		if tk != "" {
			names[_k[0][:l]] = tk
			if !m[1][_k[1]] {
				names[_k[0]] = tk
			}
		}
	}
	for name, _ := range m[0] {
		if len(m[1]) == 0 {
			return _k[0] == name
		}
		for family, _ := range m[1] {
			if name != "" && family != "" && names[name] == family {
				return true
			}
		}
	}
	return false
}

func (self *assProcessor) reMap() {
	m := make(map[string]*fontInfo)
	for _, v := range self.m {
		if _, ok := m[v.newName]; !ok {
			m[v.newName] = v
		} else {
			m[v.newName].runes = append(m[v.newName].runes, v.runes...)
		}
		m[v.newName].runes = append(m[v.newName].runes, '\u00a0', '\u0020')
	}
	self.m = m
}

func (self *assProcessor) createFontSubset(font *fontInfo) bool {
	ok := false
	fn := fmt.Sprintf(`%s.txt`, randomStr(8))
	_, fn, _, _ = splitPath(fn)
	fn = path.Join(os.TempDir(), fn)
	_, n, e, _ := splitPath(font.file)
	if strings.ToLower(e) == ".ttc" {
		e = ".ttf"
	}
	e = strings.ToLower(e)
	if os.MkdirAll(self.output, os.ModePerm) != nil {
		printLog(self.lcb, "Failed to create the output folder.")
		return false
	}
	str := string(font.runes)
	str = stringDeduplication(str)
	if os.WriteFile(fn, []byte(str), os.ModePerm) == nil {
		defer func() {
			_ = os.Remove(fn)
		}()
		n := font.newName
		if !self.rename {
			n = font.oldName
		}
		_fn := fmt.Sprintf("%s%s", n, e)
		_fn = path.Join(self.output, _fn)
		args := make([]string, 0)
		args = append(args, "--text-file="+fn)
		args = append(args, "--output-file="+_fn)
		args = append(args, "--name-languages="+"*")
		args = append(args, "--font-number="+strconv.Itoa(font.index))
		args = append(args, font.file)
		if p, err := newProcess(nil, nil, nil, "", pyftsubset, args...); err == nil {
			s, err := p.Wait()
			ok = err == nil && s.ExitCode() == 0
		}
		if !ok {
			printLog(self.lcb, `Failed to subset font: "%s"[%d].`, font.oldName, font.index)
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
	ec := 0
	fn := fmt.Sprintf("%s.ttx", randomStr(8))
	fn = path.Join(os.TempDir(), fn)
	if self.dumpFont(font.sFont, fn) {
		f, err := openFile(fn, true, false)
		if err == nil {
			defer func() {
				_ = f.Close()
				_ = os.Remove(fn)
			}()
			n := font.newName
			if !self.rename {
				n = font.oldName
			}
			if xml, err := xmlquery.Parse(f); err == nil {
				for _, v := range xml.SelectElements(`ttFont/name/namerecord`) {
					id := v.SelectAttr("nameID")
					switch id {
					case "0":
						v.FirstChild.Data = "Processed by " + LibFName + " at " + time.Now().Format("2006-01-02 15:04:05")
						break
					case "1", "4", "6":
						v.FirstChild.Data = n
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
	return ok == l
}

func (self *assProcessor) replaceFontNameInAss() bool {
	ec := 0
	m := make(map[string]map[string]bool)
	if self.rename {
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
	}
	for f, s := range self.subtitles {
		if self.rename {
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
			}
		}
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
			ec++
			printLog(self.lcb, `Failed to write the new ass file: "%s".`, fn)
		}
	}
	return ec == 0
}

func (self *assProcessor) createFontCache(p string) *fontCache {
	_m := self.getFontName(p)
	_fonts := make([][]string, len(_m))
	_types := make([][]string, len(_m))
	for k, v := range _m {
		_list := make([]string, 0)
		for _k, _ := range v[0] {
			_list = append(_list, _k)
		}
		_fonts[k] = _list
		_list = make([]string, 0)
		for _k, _ := range v[1] {
			_list = append(_list, _k)
		}
		_types[k] = _list
	}
	if len(_fonts) > 0 && len(_types) > 0 {
		return &fontCache{p, _fonts, _types}
	}
	return nil
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
	wg := new(sync.WaitGroup)
	el := make([]string, 0)
	w := func(s, e int) {
		for i := s; i < e; i++ {
			go func(x int) {
				_item := fonts[x]
				m.Lock()
				c := self.createFontCache(_item)
				if c != nil {
					ok++
					cache = append(cache, *c)
					printLog(self.lcb, "Cache font (%d/%d) done.", ok, l)
				} else {
					el = append(el, _item)
				}
				m.Unlock()
				wg.Done()
			}(i)
		}
	}
	c := 5
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

func (self *assProcessor) loadCache(ccs []string) {
	if len(ccs) > 0 {
		for _, p := range ccs {
			if data, err := ioutil.ReadFile(p); err == nil {
				cache := make([]fontCache, 0)
				if json.Unmarshal(data, &cache) == nil {
					for _, v := range cache {
						list := make([][]map[string]bool, 0)
						l := len(v.Fonts)
						for i := 0; i < l; i++ {
							m := make([]map[string]bool, 2)
							for _, n := range v.Fonts[i] {
								if m[0] == nil {
									m[0] = make(map[string]bool)
								}
								m[0][n] = true
							}
							for _, f := range v.Types[i] {
								if m[1] == nil {
									m[1] = make(map[string]bool)
								}
								m[1][f] = true
							}
							list = append(list, m)
						}
						self.cache = append(self.cache, cacheInfo{v.File, list})
					}
				}
			}
		}
	}
}

func (self *assProcessor) matchCache(k string) (string, int) {
	ok := ""
	i := -1
	_count := 0
	_k := strings.Split(k, "^")
	for _, v := range self.cache {
		for q, list := range v.Names {
			if self.matchFontName(list, _k) {
				ok = v.File
				i = q
				if self.check {
					names := self.getFontName(v.File)
					if len(names) > 0 {
						_count++
						if !self.checkFontMissing(self.m[k], _count, true) && self.strict {
							ok = ""
							i = 0
							continue
						}
					} else {
						continue
					}
				}
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
	return ok, i
}
