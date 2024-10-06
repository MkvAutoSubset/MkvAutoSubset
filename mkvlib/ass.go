package mkvlib

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MkvAutoSubset/MkvAutoSubset/mkvlib/c"
	"github.com/MkvAutoSubset/MkvAutoSubset/mkvlib/parser"
	"github.com/MkvAutoSubset/MkvAutoSubset/mkvlib/parser/sfnt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type fontInfo struct {
	file        string
	runes       []rune
	family      string
	index       int
	matchedName string
	oldNames    []string
	newName     string
	fontNames   []string
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
	newNames  map[string]string
	seps      []string
	rename    bool
	check     bool
	strict    bool
}

func (self *assProcessor) getLength(p string) time.Duration {
	f, err := openFile(p, true, false)
	if err != nil {
		return 0
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return 0
	}
	str := string(data)
	opt := parser.SSAOptions{}
	subtitle, err := parser.ReadFromSSAWithOptions(strings.NewReader(str), opt)
	if err == nil {
		var s, e time.Duration
		for _, v := range subtitle.Items {
			if v.StartAt < s {
				s = v.StartAt
			}
			if v.EndAt > e {
				e = v.EndAt
			}
		}
		return e - s
	}
	return 0
}

func restoreSubsetted(str string) string {
	reg := regexp.MustCompile(`; Font [Ss]ubset: (\w+) - ([\S ]+)`)
	x := 1
	y := 2
	s := "; Font subset restore: ${1} - ${2}"
	if strings.Contains(str, "[Assfonts Rename Info]") {
		str = strings.Replace(str, "[Assfonts Rename Info]", "[Assfonts Rename Restore Info]", 1)
		reg = regexp.MustCompile(`([\S ]+) ---- (\w+)`)
		x = 2
		y = 1
		s = `${2} - ${1}`
	}
	for _, v := range reg.FindAllStringSubmatch(str, -1) {
		if len(v) > 2 {
			_reg, _ := regexp.Compile(fmt.Sprintf(`(Style:[^,\r\n]+,|\\fn)(@?)%s([,\\\}])`, v[x]))
			if _reg.MatchString(str) {
				r := fmt.Sprintf("${1}${2}%s${3}", v[y])
				str = _reg.ReplaceAllString(str, r)
				str = reg.ReplaceAllString(str, s)
				printLog(nil, logWarning, `Font subset restore: "%s" -> "%s".`, v[x], v[y])
			}
		}
	}
	return str
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
			data, err := io.ReadAll(f)
			if err == nil {
				str := toUTF8(data)
				str = restoreSubsetted(str)
				self.subtitles[file] = str
			} else {
				ec++
			}
		}
		if ec > 0 {
			printLog(self.lcb, logError, `Failed to read the ass file: "%s"`, file)
		}
		_ = f.Close()
	}
	if ec == 0 {
		opt := parser.SSAOptions{}
		reg, _ := regexp.Compile(`\\[^\r\n\\\}]+`)
		_reg, _ := regexp.Compile(`\\fn@?(.*)`)
		__reg, _ := regexp.Compile(`\\([bir])(.*)`)
		___reg, _ := regexp.Compile(`nd[xyz]?\d+`)
		____reg, _ := regexp.Compile(`\d`)
		m := make(map[string]string)
		for k, v := range self.subtitles {
			subtitle, err := parser.ReadFromSSAWithOptions(strings.NewReader(v), opt)
			if err != nil {
				ec++
				printLog(self.lcb, logError, `Failed to parse the ass file: "%s" [%s]`, k, err)
				continue
			}
			for _, item := range subtitle.Items {
				name := ""
				_b := *item.Style.InlineStyle.SSABold
				_i := *item.Style.InlineStyle.SSAItalic
				for _, _item := range item.Lines {
					for _, __item := range _item.Items {
						if __item.InlineStyle != nil {
							items := reg.FindAllString(__item.InlineStyle.SSAEffect, -1)
							for _, ___item := range items {
								arr := _reg.FindStringSubmatch(___item)
								if len(arr) > 1 {
									name = arr[1]
									continue
								}
								_arr := __reg.FindAllStringSubmatch(___item, -1)
								for _, v := range _arr {
									if len(v) > 2 {
										switch v[1] {
										case "b":
											i, err := strconv.Atoi(v[2])
											if err == nil {
												if i == 0 || (i > 1 && i < 612) {
													_b = false
												} else if i == 1 || i > 611 {
													_b = true
												}
											}
											break
										case "i":
											_i = v[2] == "1"
											break
										case "r":
											if ___reg.MatchString(v[2]) {
												break
											}
											if v[2] == "*Default" {
												v[2] = "Default"
											}
											if v[2] == "" {
												name = ""
												_b = *item.Style.InlineStyle.SSABold
												_i = *item.Style.InlineStyle.SSAItalic
											} else if s, ok := subtitle.Styles[v[2]]; ok {
												name = s.InlineStyle.SSAFontName
												_b = *s.InlineStyle.SSABold
												_i = *s.InlineStyle.SSAItalic
											} else {
												printLog(self.lcb, logError, `Not found style in the ass file:"%s" [%s].`, k, v[2])
												ec++
											}
											break
										}
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
						s := m[_name] + strings.ReplaceAll(strings.ReplaceAll(__item.Text, "\\n", ""), "\\N", "")
						if len(s) > 1000 {
							_m := make(map[rune]bool)
							chars := make([]rune, 0)
							for _, _v := range s {
								if _, ok := _m[_v]; !ok {
									_m[_v] = true
									chars = append(chars, _v)
								}
							}
							s = string(chars)
						}
						m[_name] = s
					}
				}
			}
		}
		self.m = make(map[string]*fontInfo)
		for k, v := range m {
			if v != "" {
				if ____reg.MatchString(v) {
					v += "0123456789"
				}
				v += "a\u0020\u00a0"
				self.m[k] = new(fontInfo)
				self.m[k].runes = []rune(v)
			}
		}
		if len(self.m) == 0 {
			printLog(self.lcb, logWarning, `Not found item in the ass file(s): "%d"`, len(self.files))
		}
	}
	return ec == 0
}

func (self *assProcessor) getFontsList() [][]string {
	if !self.parse() {
		return nil
	}
	list := make([]string, 0)
	for k, _ := range self.m {
		list = append(list, k)
	}
	list2 := make([]string, 0)
	if self.check {
		list2 = self.matchFonts()
	}
	sort.Strings(list)
	sort.Strings(list2)
	return [][]string{list, list2}
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
		data, err := io.ReadAll(f)
		if err == nil {
			fonts := make([]*sfnt.Font, 0)
			if strings.HasSuffix(strings.ToLower(p), ".ttc") {
				c, err := sfnt.ParseCollection(data)
				if err == nil {
					l := c.NumFonts()
					for i := 0; i < l; i++ {
						_f, err := c.Font(i)
						if err == nil {
							if g, _ := _f.GlyphIndex(nil, '\u0020'); g == 0 {
								printLog(self.lcb, logWarning, `Font: "%s"[%d] is not defined '\u0020',skip.`, p, i)
							} else {
								fonts = append(fonts, _f)
							}
						}
					}
				}
			} else {
				_f, err := sfnt.Parse(data)
				if err == nil {
					if g, _ := _f.GlyphIndex(nil, '\u0020'); g == 0 {
						printLog(self.lcb, logWarning, `Font: "%s" is not defined '\u0020',skip.`, p)
					} else {
						fonts = append(fonts, _f)
					}
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
	_f, err := os.Open(f.file)
	if err == nil {
		defer func() { _ = _f.Close() }()
		data, err := io.ReadAll(_f)
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
					if r == '\u00a0' || r == '\u0009' {
						continue
					}
					n, _ := _font.GlyphIndex(nil, r)
					if n == 0 {
						_str += string(r)
					}
				}
			} else {
				return false
			}
		}
	} else {
		return false
	}
	h := "N"
	if c {
		h = "C"
	}
	if _str != "" {
		_str = stringDeduplication(_str)
		printLog(self.lcb, logWarning, `{%s%02d}Font [%s] -> "%s"[%d] missing char(s): "%s"`, h, i, f.matchedName, f.file, f.index, _str)
	}
	return _str == ""
}

func (self *assProcessor) matchFonts() []string {
	self.newNames = make(map[string]string)
	fonts := findFonts(self._fonts)
	m := self.getFontsName(fonts)
	_count := make(map[string]int)
	w := func(fb int) {
		for k, _ := range self.m {
			_k := strings.Split(k, "^")
			if self.m[k].file != "" || (fb == 1 && _k[1] == "Regular") {
				continue
			}
			if fb > 0 && _k[1] != "Regular" {
				if fb == 1 {
					printLog(self.lcb, logWarning, `Font fallback:[%s^%s] -> [%s^Regular]`, _k[0], _k[1], _k[0])
				}
				_k[1] = "Regular"
			}
			for __k, v := range m {
				for ___k, _v := range v {
					if n, names := self.matchFontName(_v, _k, fb == 2); n != "" {
						self.m[k].file = __k
						self.m[k].index = ___k
						self.m[k].matchedName = n
						self.m[k].fontNames = names
						if self.check {
							_count[n]++
							if !self.checkFontMissing(self.m[k], _count[_k[0]], false) && self.strict {
								self.m[k].file = ""
								self.m[k].index = 0
								continue
							}
						}
						self.m[k].family = _k[1]
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
			if f, i, n, names := self.matchCache(fmt.Sprintf("%s^%s", _k[0], _k[1]), k, fb == 2); f != "" {
				self.m[k].file, self.m[k].index, self.m[k].matchedName = f, i, n
				self.m[k].family = _k[1]
				self.m[k].fontNames = names
			}
		}
	}
	w(0)
	w(1)
	w(2)
	el := make([]string, 0)
	for k, _ := range self.m {
		if self.m[k].file == "" {
			el = append(el, k)
			printLog(self.lcb, logError, `Missing the font: "%s".`, k)
		}
	}
	return el
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

func (self *assProcessor) matchFontName(m []map[string]bool, _k []string, b bool) (string, []string) {
	fontNames := make([]string, 0)
	for name, _ := range m[0] {
		fontNames = append(fontNames, name)
	}
	names := make(map[string]string)
	_name := strings.TrimSpace(_k[0])
	_family := _k[1]
	names[_name] = _family
	names[strings.ToLower(_name)] = _family
	if !m[1][_family] {
		for _, v := range self.seps {
			l := strings.LastIndex(_name, v)
			if l > -1 && l < len(_name)-1 {
				tk := _name[l+1:]
				names[_name] = tk
				names[strings.ToLower(_name)] = tk
			}
		}
	}
	for name, _ := range m[0] {
		for family, _ := range m[1] {
			if name != "" && family != "" {
				if names[name] == family {
					return name, fontNames
				}
				if b && names[strings.ToLower(name)] == family {
					printLog(self.lcb, logSWarning, `Font bottom fallback:[%s^%s] -> [%s^%s]`, _name, _family, name, family)
					return name, fontNames
				}
			}
		}
		if b && (_name == name || strings.ToLower(_name) == strings.ToLower(name)) {
			if len(m[1]) > 0 {
				families := make([]string, 0)
				for family, _ := range m[1] {
					if family == "" {
						continue
					}
					if family == "Regular" {
						return "", nil
					}
					families = append(families, family)
				}
				if len(families) > 1 {
					printLog(self.lcb, logSWarning, `Font bottom fallback:[%s^%s] -> [%s^(%s)]`, _name, _family, name, strings.Join(families, ","))
				} else {
					printLog(self.lcb, logSWarning, `Font bottom fallback:[%s^%s] -> [%s^%s]`, _name, _family, name, families[0])
				}
			} else {
				printLog(self.lcb, logSWarning, `Font bottom fallback:[%s^%s] -> [%s]`, _name, _family, name)
			}
			return name, fontNames
		}
	}
	return "", nil
}

func (self *assProcessor) reMap() {
	m := make(map[string]*fontInfo)
	for k, v := range self.m {
		if v.file == "" {
			continue
		}
		_k := strings.Split(k, "^")[0]
		_f := fmt.Sprintf("%s.%d", v.file, v.index)
		if _, ok := m[_f]; !ok {
			m[_f] = v
			m[_f].oldNames = []string{_k}
		} else {
			m[_f].runes = append(m[_f].runes, v.runes...)
			m[_f].oldNames = append(m[_f].oldNames, _k)
		}
	}
	fontNameMap := make(map[string][]string)
	for file, font := range m {
		font.oldNames = stringArrayDeduplication(font.oldNames)
		printLog(self.lcb, logInfo, `Font selected:[%s]^%s -> "%s"[%d]`, strings.Join(font.oldNames, ","), font.family, font.file, font.index)
		for _, name := range font.fontNames {
			fontNameMap[name] = append(fontNameMap[name], file)
		}
	}
	uniqueMap := make(map[string]int)
	for name, files := range fontNameMap {
		for _, file := range files {
			if uniqueMap[file] < len(files) {
				m[file].oldNames = append(m[file].oldNames, m[file].matchedName)
				m[file].matchedName = name
				newName := self.newNames[name]
				if newName == "" {
					newName = randomStr(8)
					self.newNames[name] = newName
				}
				m[file].newName = newName
				uniqueMap[file] = len(files)
			}
		}
	}
	self.m = m
}

func (self *assProcessor) createFontSubset(font *fontInfo) bool {
	_, n, e, _ := splitPath(font.file)
	e = strings.ToLower(e)
	if e == ".ttc" {
		e = ".ttf"
	}
	if os.MkdirAll(self.output, os.ModePerm) != nil {
		printLog(self.lcb, logError, "Failed to create the output folder.")
		return false
	}
	str := string(font.runes)
	str = stringDeduplication(str)
	n = font.newName
	if !self.rename {
		n = font.matchedName
	}
	dest := "Processed by " + LibFName + " at " + time.Now().Format("2006-01-02 15:04:05")
	fn := fmt.Sprintf("%s.%s%s", n, randomStr(8), e)
	fn = path.Join(self.output, fn)
	if !c.Subset(font.file, font.index, fn, n, dest, str) {
		printLog(self.lcb, logError, `Failed to subset font: "%s"(%s)[%d].`, font.matchedName, font.file, font.index)
		return false
	}
	return true
}

func (self *assProcessor) createFontsSubset() bool {
	self.reMap()
	err := os.RemoveAll(self.output)
	if !(err == nil || errors.Is(err, os.ErrNotExist)) {
		printLog(self.lcb, logError, "Failed to clean the output folder.")
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

func (self *assProcessor) replaceFontNameInAss() bool {
	ec := 0
	m := make(map[string]map[string]bool)
	for _, v := range self.m {
		if self.rename || len(v.oldNames) > 1 {
			for f, s := range self.subtitles {
				if m[f] == nil {
					m[f] = make(map[string]bool)
				}
				if !self.rename {
					v.newName = v.matchedName
				}
				for _, _v := range v.oldNames {
					n := regexp.QuoteMeta(_v)
					reg, _ := regexp.Compile(fmt.Sprintf(`(Style:[^,\r\n]+,|\\fn)(@?)%s([,\\\}])`, n))
					if reg.MatchString(s) {
						r := fmt.Sprintf("${1}${2}%s${3}", v.newName)
						s = reg.ReplaceAllString(s, r)
						m[f][v.matchedName] = true
						self.subtitles[f] = s
					}
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
					if v.matchedName == k {
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
			printLog(self.lcb, logError, `Failed to write the new ass file: "%s".`, fn)
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
	if len(_fonts) > 0 {
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
					printLog(self.lcb, logProgress, "Cache font (%d/%d) done.", ok, l)
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
		if os.WriteFile(output, data, os.ModePerm) != nil {
			printLog(self.lcb, logError, `Failed to write cache file: "%s"`, output)
		}
	}
	return el
}

func (self *assProcessor) copyFontsFromCache() bool {
	ec := 0
	if self.parse() {
		i := 0
		self.matchFonts()
		self.reMap()
		l := len(self.m)
		for k, v := range self.m {
			if v.file != "" {
				_, fn, _, _ := splitPath(v.file)
				fn = path.Join(self.output, fn)
				if copyFile(v.file, fn) == nil {
					i++
					printLog(self.lcb, logProgress, "Copy (%d/%d) done.", i, l)
				}
			} else {
				ec++
				printLog(self.lcb, logError, `Missing the font: "%s".`, k)
			}
		}
	}
	return ec == 0
}

func (self *assProcessor) loadCache(ccs []string) {
	if len(ccs) > 0 {
		for _, p := range ccs {
			if data, err := os.ReadFile(p); err == nil {
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

func (self *assProcessor) matchCache(k, o string, b bool) (string, int, string, []string) {
	ok := ""
	i := -1
	_count := 0
	_k := strings.Split(k, "^")
	otfFile := ""
	otfName := ""
	otfNames := make([]string, 0)
	n := ""
	names := make([]string, 0)
	for _, v := range self.cache {
		for q, list := range v.Names {
			if n, names = self.matchFontName(list, _k, b); n != "" {
				ok = v.File
				i = q
				if self.check {
					names := self.getFontName(v.File)
					if len(names) > 0 {
						_count++
						f := new(fontInfo)
						f.matchedName = n
						f.file = ok
						f.index = i
						f.runes = self.m[o].runes
						if !self.checkFontMissing(f, _count, true) && self.strict {
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
		_, _, e, _ := splitPath(ok)
		e = strings.ToLower(e)
		if e == ".otf" && otfFile == "" {
			otfFile = ok
			otfName = n
			otfNames = names
			ok = ""
		}
		if ok != "" {
			break
		}
	}
	if ok == "" && otfFile != "" {
		ok = otfFile
		n = otfName
		names = otfNames
		i = 0
	}
	if _, err := os.Stat(ok); err != nil {
		ok = ""
	}
	return ok, i, n, names
}
