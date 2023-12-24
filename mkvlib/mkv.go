package mkvlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	mkvmerge   = `mkvmerge`
	mkvextract = `mkvextract`
	ass2bdnxml = `ass2bdnxml`
	ffmpeg     = `ffmpeg`
)

type mkvInfo struct {
	Attachments []struct {
		ID          int    `json:"id"`
		FileName    string `json:"file_name"`
		Size        int    `json:"size"`
		ContentType string `json:"content_type"`
	} `json:"attachments"`
	Tracks []struct {
		ID         int    `json:"id"`
		Type       string `json:"type"`
		Codec      string `json:"codec"`
		Properties struct {
			Language     string `json:"language"`
			TrackName    string `json:"track_name"`
			DefaultTrack bool   `json:"default_track"`
		} `json:"properties"`
	}
}

type mkvProcessor struct {
	a2p        bool
	apc        bool
	mks        bool
	pr         string
	pf         string
	caches     []string
	ass2bdnxml bool
	ffmpeg     bool
	nrename    bool
	check      bool
	strict     bool
	noverwrite bool
}

func (self *mkvProcessor) GetMKVInfo(file string) *mkvInfo {
	buf := bytes.NewBufferString("")
	if p, err := newProcess(nil, buf, nil, "", mkvmerge, "-J", file); err == nil {
		if s, err := p.Wait(); err == nil && s.ExitCode() == 0 {
			obj := new(mkvInfo)
			_ = json.Unmarshal(buf.Bytes(), obj)
			return obj
		}
	}
	return nil
}

func (self *mkvProcessor) DumpMKV(file, output string, subset bool, lcb logCallback) bool {
	ec := 0
	obj := self.GetMKVInfo(file)
	if obj == nil {
		printLog(lcb, logError, `Failed to get the file info: "%s".`, file)
		return false
	}
	attachments := make([]string, 0)
	tracks := make([]string, 0)
	for _, _item := range obj.Attachments {
		attachments = append(attachments, fmt.Sprintf(`%d:%s`, _item.ID, path.Join(output, "fonts", _item.FileName)))
	}
	for _, _item := range obj.Tracks {
		if _item.Type == "subtitles" {
			s := ""
			if _item.Properties.DefaultTrack {
				s = "#"
			}
			s += fmt.Sprintf(`%d_%s_%s`, _item.ID, _item.Properties.Language, _item.Properties.TrackName)
			if _item.Codec == "SubStationAlpha" {
				s += ".ass"
			} else {
				s += ".sub"
			}
			tracks = append(tracks, fmt.Sprintf(`%d:%s`, _item.ID, path.Join(output, s)))
		}
	}
	la := len(attachments)
	lt := len(tracks)
	if la > 0 || lt > 0 {
		args := make([]string, 0)
		args = append(args, file)
		if la > 0 {
			args = append(args, "attachments")
			args = append(args, attachments...)
		}
		if lt > 0 {
			args = append(args, "tracks")
			args = append(args, tracks...)
		}
		if p, err := newProcess(nil, nil, nil, "", mkvextract, args...); err == nil {
			s, err := p.Wait()
			ok := err == nil && s.ExitCode() == 0
			if ok {
				if subset {
					asses := make([]string, 0)
					for _, _item := range tracks {
						f := _item[strings.Index(_item, ":")+1:]
						if strings.HasSuffix(f, ".ass") {
							asses = append(asses, f)
						}
						if len(asses) > 0 {
							if !self.ASSFontSubset(asses, "", "", false, lcb) {
								ec++
							}
						}
					}
				}
			} else {
				ec++
			}
		} else {
			ec++
		}
	} else {
		printLog(lcb, logInfo, `This file is not has the subtitles & attachments: "%s"`, file)
	}
	return ec == 0
}

func (self *mkvProcessor) CheckSubset(file string, lcb logCallback) (bool, bool) {
	obj := self.GetMKVInfo(file)
	if obj == nil {
		printLog(lcb, logError, `Failed to get the file info: "%s".`, file)
		return false, true
	}
	ass := false
	ok := false
	reg, _ := regexp.Compile(`\.[A-Z0-9]{8}\.\S+$`)
	for _, track := range obj.Tracks {
		ass = track.Type == "subtitles" && track.Codec == "SubStationAlpha"
		if ass {
			break
		}
	}
	for _, attachment := range obj.Attachments {
		ok = !ass || (strings.HasPrefix(attachment.ContentType, "font/") && reg.MatchString(attachment.FileName))
		if ok {
			break
		}
	}
	return !ass || (ass && ok), false
}

func (self *mkvProcessor) CreateMKV(file string, tracks, attachments []string, output, slang, stitle string, clean bool) bool {
	args := make([]string, 0)
	if clean && !self.mks {
		args = append(args, "--no-subtitles", "--no-attachments")
	}
	d, _, _, ne := splitPath(output)
	if !self.mks {
		args = append(args, file)
		output = path.Join(d, ne+".mkv")
	} else {
		output = path.Join(d, ne+".mks")
	}
	args = append(args, "--output", output)
	for _, _item := range attachments {
		args = append(args, "--attach-file", _item)
	}
	for _, _item := range tracks {
		_, _, _, f := splitPath(_item)
		_arr := strings.Split(f, "_")
		_sl := slang
		_st := stitle
		if len(_arr) > 1 {
			_sl = _arr[1]
		}
		if len(_arr) > 2 {
			_st = strings.Join(_arr[2:], "_")
		}
		if _sl != "" {
			args = append(args, "--language", "0:"+_sl)
		}
		if _st != "" {
			args = append(args, "--track-name", "0:"+_st)
		}
		if !strings.HasPrefix(f, "#") {
			args = append(args, "--default-track-flag", "0:no")
		}
		args = append(args, _item)
	}
	if p, err := newProcess(nil, nil, nil, "", mkvmerge, args...); err == nil {
		s, err := p.Wait()
		ok := err == nil && s.ExitCode() != 2
		if !ok {
			_ = os.Remove(output)
		}
		return ok
	}
	return false
}

func (self *mkvProcessor) DumpMKVs(dir, output string, subset bool, lcb logCallback) bool {
	ok := true
	files := findMKVs(dir)
	l := len(files)
	_ok := 0
	for _, item := range files {
		p := strings.TrimPrefix(item, dir)
		d, _, _, f := splitPath(p)
		p = path.Join(output, d, f)
		if !self.DumpMKV(item, p, subset, lcb) {
			ok = false
			printLog(lcb, logError, `Failed to dump the file: "%s".`, item)
		} else {
			_ok++
			printLog(lcb, logProgress, "Dump (%d/%d) done.", _ok, l)
		}
	}
	return ok
}

func (self *mkvProcessor) QueryFolder(dir string, lcb logCallback) []string {
	lines := make([]string, 0)
	files := findMKVs(dir)
	l := len(files)
	for i, file := range files {
		a, b := self.CheckSubset(file, lcb)
		if b {
			printLog(lcb, logError, `Failed to check subset for file: "%s".`, file)
		} else if !a {
			lines = append(lines, file)
		}
		printLog(lcb, logProgress, "Query (%d/%d) done.", i+1, l)
	}
	return lines
}

func (self *mkvProcessor) CreateMKVs(vDir, sDir, fDir, tDir, oDir, slang, stitle string, clean bool, lcb logCallback) bool {
	ok := true
	if tDir == "" {
		tDir = os.TempDir()
	}
	tDir = path.Join(tDir, randomStr(8))
	files, _ := findPath(vDir, `\.\S+$`)
	l := len(files)
	_ok := 0
	for _, item := range files {
		ec := 0
		_, _, _, _f := splitPath(item)
		tmp, _ := findPath(sDir, `\.\S+$`)
		asses := make([]string, 0)
		subs := make([]string, 0)
		p := path.Join(tDir, _f)
		fn := path.Join(oDir, _f)
		s1 := path.Join(p, "asses")
		s2 := path.Join(p, "subs")
		if self.mks {
			fn += ".mks"
		} else {
			fn += ".mkv"
		}
		if _a, _ := isExists(fn); _a && self.noverwrite {
			printLog(lcb, logInfo, `Existing file: "%s",skip.`, fn)
			_ok++
			printLog(lcb, logProgress, "Create (%d/%d) done.", _ok, l)
			continue
		}
		for i, sub := range tmp {
			_, n, e, _ := splitPath(sub)
			reg, _ := regexp.Compile(fmt.Sprintf(`^#?(%s)(_[^_]*)*\.\S+$`, regexp.QuoteMeta(_f)))
			if !reg.MatchString(n) {
				continue
			}
			f := strings.Replace(n, _f, "", 1)
			g := ""
			if strings.HasPrefix(f, "#") {
				f = strings.TrimPrefix(f, "#")
				g = "#"
			}
			_s := fmt.Sprintf("%s%d%s", g, i, f)
			if e == ".ass" {
				_s = path.Join(s1, _s)
				asses = append(asses, _s)
			} else {
				_s = path.Join(s2, _s)
				subs = append(subs, _s)
			}
			_ = copyFileOrDir(sub, _s)
		}
		attachments := make([]string, 0)
		tracks := make([]string, 0)
		if len(asses) > 0 {
			if !self.ASSFontSubset(asses, fDir, "", false, lcb) {
				ec++
			} else {
				_tracks, _ := findPath(s1, `\.pgs$`)
				__p := path.Join(s1, "subsetted")
				attachments = findFonts(__p)
				tracks, _ = findPath(__p, `\.ass$`)
				tracks = append(tracks, _tracks...)
			}
		}
		tracks = append(tracks, subs...)
		if ec == 0 && !self.CreateMKV(item, tracks, attachments, fn, slang, stitle, clean) {
			ec++
		}
		if ec > 0 {
			ok = false
			printLog(lcb, logError, `Failed to create the file: "%s".`, item)
		} else {
			_ok++
			printLog(lcb, logProgress, "Create (%d/%d) done.", _ok, l)
		}
	}
	_ = os.RemoveAll(tDir)
	return ok
}

func (self *mkvProcessor) MakeMKVs(dir, data, output, slang, stitle string, subset bool, lcb logCallback) bool {
	dir, _ = filepath.Abs(dir)
	data, _ = filepath.Abs(data)
	output, _ = filepath.Abs(output)
	ok := true
	_files, _ := findPath(dir, `\.\S+$`)
	files := make([]string, 0)
	for _, item := range _files {
		if strings.HasPrefix(item, data) || strings.HasPrefix(item, output) {
			continue
		}
		files = append(files, item)
	}
	l := len(files)
	_ok := 0
	for _, item := range files {
		p := strings.TrimPrefix(item, dir)
		d, _, _, f := splitPath(p)
		fn := path.Join(output, d, f)
		if self.mks {
			fn += ".mks"
		} else {
			fn += ".mkv"
		}
		if _a, _ := isExists(fn); _a && self.noverwrite {
			printLog(lcb, logInfo, `Existing file: "%s",skip.`, fn)
			_ok++
			printLog(lcb, logProgress, "Make (%d/%d) done.", _ok, l)
			continue
		}
		p = path.Join(data, d, f)
		_p := path.Join(p, "subsetted")
		asses, _ := findPath(_p, `\.ass$`)
		attachments := findFonts(_p)
		if len(asses) == 0 && subset {
			asses, _ = findPath(p, `\.ass$`)
			if len(asses) > 0 {
				if !self.ASSFontSubset(asses, "", "", false, lcb) {
					ok = false
					printLog(lcb, logError, `Failed to make the file: "%s".`, item)
					continue
				}
				asses, _ = findPath(_p, `\.ass$`)
				attachments = findFonts(_p)
			}
		}
		subs, _ := findPath(p, `\.(sub)|(pgs)`)
		tracks := append(subs, asses...)
		if !self.CreateMKV(item, tracks, attachments, fn, slang, stitle, true) {
			ok = false
			printLog(lcb, logError, `Failed to make the file: "%s".`, item)
		} else {
			_ok++
			printLog(lcb, logProgress, "Make (%d/%d) done.", _ok, l)
		}
	}
	return ok
}

func (self *mkvProcessor) ASSFontSubset(files []string, fonts, output string, dirSafe bool, lcb logCallback) bool {
	if len(files) == 0 {
		return false
	}
	obj := new(assProcessor)
	obj.files = files
	obj._fonts = fonts
	obj.output = output
	obj.lcb = lcb
	obj.rename = !self.nrename
	obj.check = self.check
	obj.strict = self.strict
	d, _, _, _ := splitPath(obj.files[0])
	if obj._fonts == "" {
		obj._fonts = path.Join(d, "fonts")
	}
	if obj.output == "" {
		obj.output = d
		dirSafe = true
	}
	if dirSafe {
		obj.output = path.Join(obj.output, "subsetted")
	}
	obj.fonts = findFonts(obj._fonts)
	obj.loadCache(self.caches)
	r := obj.parse() && len(obj.matchFonts()) == 0 && obj.createFontsSubset() && obj.changeFontsName() && obj.replaceFontNameInAss()
	if !r {
		_ = os.RemoveAll(obj.output)
	}
	if r && self.a2p {
		r = self.ass2Pgs(obj._files, self.pr, self.pf, obj.output, d, lcb)
		if r && !self.apc {
			_ = os.RemoveAll(obj.output)
		}
	}
	return r
}

func (self *mkvProcessor) A2P(a2p, apc bool, pr, pf string) {
	self.a2p = self.ass2bdnxml && a2p
	self.apc = apc
	self.pr = pr
	self.pf = pf
}

func (self *mkvProcessor) ass2Pgs(input []string, resolution, frameRate, fontsDir, output string, lcb logCallback) bool {
	return self.a2p && ass2Pgs(input, resolution, frameRate, fontsDir, output, lcb)
}

func (self *mkvProcessor) GetFontsList(files []string, fonts string, lcb logCallback) [][]string {
	if len(files) > 0 {
		obj := new(assProcessor)
		obj.files = files
		obj.lcb = lcb
		d, _, _, _ := splitPath(obj.files[0])
		obj._fonts = fonts
		if obj._fonts == "" {
			obj._fonts = path.Join(d, "fonts")
		}
		obj.check = self.check
		obj.strict = self.strict
		obj.loadCache(self.caches)
		return obj.getFontsList()
	}
	return nil
}

func (self *mkvProcessor) CreateFontsCache(dir, output string, lcb logCallback) []string {
	obj := new(assProcessor)
	obj._fonts = dir
	obj.lcb = lcb
	return obj.createFontsCache(output)
}

func (self *mkvProcessor) CopyFontsFromCache(asses []string, dist string, lcb logCallback) bool {
	obj := new(assProcessor)
	obj.lcb = lcb
	obj.files = asses
	obj.output = dist
	obj.check = self.check
	obj.strict = self.strict
	obj.loadCache(self.caches)
	return obj.copyFontsFromCache()
}

func (self *mkvProcessor) GetFontInfo(p string) *fontCache {
	obj := new(assProcessor)
	return obj.createFontCache(p)
}

func (self *mkvProcessor) Cache(ccs []string) {
	self.caches = ccs
}

func (self *mkvProcessor) MKS(mks bool) {
	self.mks = mks
}

func (self *mkvProcessor) Check(check, strict bool) {
	self.check = check
	self.strict = strict
}

func (self *mkvProcessor) NRename(nrename bool) {
	self.nrename = nrename
}

func (self *mkvProcessor) NOverwrite(n bool) {
	self.noverwrite = n
}

func (self *mkvProcessor) CreateBlankOrBurnVideo(t int64, s, enc, ass, fontdir, output string) bool {
	if !self.ffmpeg {
		return false
	}
	args := make([]string, 0)
	args = append(args, "-y", "-hide_banner", "-loglevel", "quiet")
	if enc == "" {
		enc = "libx264"
	}
	if s == "" {
		args = append(args, "-f", "lavfi")
		args = append(args, "-i", fmt.Sprintf("color=c=0x000000:s=%s:r=%s", self.pr, self.pf))
	} else {
		args = append(args, "-i", s)
	}
	if ass != "" && fontdir != "" {
		t = new(assProcessor).getLength(ass).Milliseconds()
		fontdir = strings.ReplaceAll(fontdir, `\`, `/`)
		fontdir = strings.ReplaceAll(fontdir, `:`, `\\:`)
		fontdir = strings.ReplaceAll(fontdir, `[`, `\[`)
		fontdir = strings.ReplaceAll(fontdir, `]`, `\]`)
		ass = strings.ReplaceAll(ass, `\`, `/`)
		ass = strings.ReplaceAll(ass, `:`, `\\:`)
		ass = strings.ReplaceAll(ass, `[`, `\[`)
		ass = strings.ReplaceAll(ass, `]`, `\]`)
		args = append(args, "-vf", fmt.Sprintf(`subtitles=%s:fontsdir=%s`, ass, fontdir))
	}
	if s == "" {
		if t > 0 {
			args = append(args, "-t", fmt.Sprintf("%dms", t))
		} else {
			return false
		}
	}
	args = append(args, "-pix_fmt", "nv12", "-crf", "18")
	args = append(args, "-vcodec", enc)
	args = append(args, output)
	if p, err := newProcess(nil, nil, nil, "", ffmpeg, args...); err == nil {
		s, err := p.Wait()
		ok := err == nil && s.ExitCode() == 0
		if !ok {
			_ = os.Remove(output)
		}
		return ok
	}
	return false
}

func (self *mkvProcessor) CreateTestVideo(asses []string, s, fontdir, enc string, burn bool, lcb logCallback) bool {
	if s == "-" {
		s = ""
	}
	l := len(asses)
	if l == 0 {
		return false
	}
	if burn {
		ec := 0
		_ok := 0
		for _, v := range asses {
			d, _, _, ne := splitPath(v)
			_output := path.Join(d, fmt.Sprintf("%s-test.mp4", ne))
			ok := self.CreateBlankOrBurnVideo(0, s, enc, v, fontdir, _output)
			if !ok {
				ec++
				printLog(lcb, logError, `Failed to create the test video file: "%s"`, _output)
				_ = os.Remove(_output)
			} else {
				_ok++
				printLog(lcb, logProgress, "CT (%d/%d) done.", _ok, l)
			}
		}
		return ec == 0
	}
	_obj := new(assProcessor)
	var t time.Duration
	for _, v := range asses {
		_t := _obj.getLength(v)
		if _t > t {
			t = _t
		}
	}
	ok := true
	_fonts := findFonts(fontdir)
	if len(_fonts) > 0 {
		d, _, _, _ := splitPath(asses[0])
		n := randomStr(8)
		_t := s == ""
		if _t {
			s = path.Join(d, fmt.Sprintf("%s.mp4", n))
			if !self.CreateBlankOrBurnVideo(t.Milliseconds(), "", enc, "", "", s) {
				ok = false
				printLog(lcb, logError, `Failed to create the temp video file: "%s".`, s)
			}
		}
		if ok {
			output := path.Join(d, fmt.Sprintf("%s.mkv", n))
			if !self.CreateMKV(s, asses, _fonts, output, "", "", true) {
				ok = false
				printLog(lcb, logError, `Failed to create the test video file: "%s".`, output)
			} else {
				printLog(lcb, logProgress, "CT done.")
			}
		}
		if _t {
			_ = os.Remove(s)
		}
	} else {
		ok = false
	}
	return ok
}
