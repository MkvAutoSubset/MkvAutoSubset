package mkvlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

const (
	mkvmerge   = `mkvmerge`
	mkvextract = `mkvextract`
	ass2bdnxml = `ass2bdnxml`
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
			Language  string `json:"language"`
			TrackName string `json:"track_name"`
		} `json:"properties"`
	}
}

type mkvProcessor struct {
	a2p        bool
	apc        bool
	mks        bool
	pr         string
	pf         string
	cache      string
	ass2bdnxml bool
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
		printLog(lcb, `Failed to get the mkv file info: "%s".`, file)
		return false
	}
	attachments := make([]string, 0)
	tracks := make([]string, 0)
	for _, _item := range obj.Attachments {
		attachments = append(attachments, fmt.Sprintf(`%d:%s`, _item.ID, path.Join(output, "fonts", _item.FileName)))
	}
	for _, _item := range obj.Tracks {
		if _item.Type == "subtitles" {
			s := fmt.Sprintf(`%d_%s_%s`, _item.ID, _item.Properties.Language, _item.Properties.TrackName)
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
		printLog(lcb, `This mkv file is not has the subtitles & attachments: "%s"`, file)
	}
	return ec == 0
}

func (self *mkvProcessor) CheckSubset(file string, lcb logCallback) (bool, bool) {
	obj := self.GetMKVInfo(file)
	if obj == nil {
		printLog(lcb, `Failed to get the mkv file info: "%s".`, file)
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
	if clean {
		args = append(args, "--no-subtitles", "--no-attachments")
	}
	if !self.mks {
		args = append(args, file)
	} else {
		d, _, _, ne := splitPath(output)
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
			_st = _arr[2]
		}
		if _sl != "" {
			args = append(args, "--language", "0:"+_sl)
		}
		if _st != "" {
			args = append(args, "--track-name", "0:"+_st)
		}
		args = append(args, _item)
	}
	if p, err := newProcess(nil, nil, nil, "", mkvmerge, args...); err == nil {
		s, err := p.Wait()
		return err == nil && s.ExitCode() != 2
	}
	return false
}

func (self *mkvProcessor) DumpMKVs(dir, output string, subset bool, lcb logCallback) bool {
	ec := 0
	files := findMKVs(dir)
	l := len(files)
	for i, item := range files {
		p := strings.TrimPrefix(item, dir)
		d, _, _, f := splitPath(p)
		p = path.Join(output, d, f)
		if !self.DumpMKV(item, p, subset, lcb) {
			ec++
			printLog(lcb, `Failed to dump the mkv file "%s".`, item)
		}
		printLog(lcb, "Dump (%d/%d) done.", i+1, l)
	}
	return ec == 0
}

func (self *mkvProcessor) QueryFolder(dir string, lcb logCallback) []string {
	ec := 0
	lines := make([]string, 0)
	files := findMKVs(dir)
	l := len(files)
	for i, file := range files {
		a, b := self.CheckSubset(file, lcb)
		if b {
			ec++
		} else if !a {
			lines = append(lines, file)
		}
		printLog(lcb, "Query (%d/%d) done.", i+1, l)
	}
	return lines
}

func (self *mkvProcessor) CreateMKVs(vDir, sDir, fDir, tDir, oDir, slang, stitle string, clean bool, lcb logCallback) bool {
	ec := 0
	if tDir == "" {
		tDir = os.TempDir()
	}
	tDir = path.Join(tDir, randomStr(8))
	files, _ := findPath(vDir, `\.\S+$`)
	l := len(files)
	for i, item := range files {
		_, _, _, _f := splitPath(item)
		tmp, _ := findPath(sDir, fmt.Sprintf(`%s\S*\.\S+$`, regexp.QuoteMeta(_f)))
		asses := make([]string, 0)
		subs := make([]string, 0)
		p := path.Join(tDir, _f)
		for _, sub := range tmp {
			if strings.HasSuffix(sub, ".ass") {
				_, _, _, __f := splitPath(sub)
				_s := path.Join(p, __f) + ".ass"
				_ = copyFileOrDir(sub, _s)
				asses = append(asses, _s)
			} else {
				subs = append(subs, sub)
			}
		}
		attachments := make([]string, 0)
		tracks := make([]string, 0)
		if len(asses) > 0 {
			if !self.ASSFontSubset(asses, fDir, "", false, lcb) {
				ec++
			} else {
				_tracks, _ := findPath(p, `\.pgs$`)
				__p := path.Join(p, "subsetted")
				attachments = findFonts(__p)
				tracks, _ = findPath(__p, `\.ass$`)
				tracks = append(tracks, _tracks...)
			}
		}
		tracks = append(tracks, subs...)
		fn := path.Join(oDir, _f) + ".mkv"
		if !self.CreateMKV(item, tracks, attachments, fn, slang, stitle, clean) {
			ec++
		}
		if ec > 0 {
			printLog(lcb, `Failed to create the mkv file: "%s".`, item)
		}
		printLog(lcb, "Create (%d/%d) done.", i+1, l)
	}
	_ = os.RemoveAll(tDir)
	return ec == 0
}

func (self *mkvProcessor) MakeMKVs(dir, data, output, slang, stitle string, lcb logCallback) bool {
	ec := 0
	files, _ := findPath(dir, `\.\S+$`)
	l := len(files)
	for i, item := range files {
		p := strings.TrimPrefix(item, dir)
		d, n, _, f := splitPath(p)
		p = path.Join(data, d, f)
		_p := path.Join(p, "subsetted")
		subs, _ := findPath(p, `\.(sub)|(pgs)`)
		asses, _ := findPath(_p, `\.ass$`)
		attachments := findFonts(_p)
		tracks := append(subs, asses...)
		fn := path.Join(output, d, n)
		if !self.CreateMKV(item, tracks, attachments, fn, slang, stitle, true) {
			ec++
			printLog(lcb, `Faild to make the mkv file: "%s".`, item)
		}
		printLog(lcb, "Make (%d/%d) done.", i+1, l)
	}
	return ec == 0
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
	d, _, _, _ := splitPath(obj.files[0])
	if obj._fonts == "" {
		obj._fonts += path.Join(d, "fonts")
	}
	if obj.output == "" {
		obj.output = d
		dirSafe = true
	}
	if dirSafe {
		obj.output = path.Join(obj.output, "subsetted")
	}
	obj.fonts = findFonts(obj._fonts)
	obj.loadCache(self.cache)
	r := obj.parse() && obj.matchFonts() && obj.createFontsSubset() && obj.changeFontsName() && obj.replaceFontNameInAss()
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

func (self *mkvProcessor) GetFontsList(input string, lcb logCallback) []string {
	files, _ := findPath(input, `\.ass$`)
	if len(files) > 0 {
		obj := new(assProcessor)
		obj.files = files
		obj.lcb = lcb
		if obj.parse() {
			return obj.getFontsList()
		}
	}
	return nil
}

func (self *mkvProcessor) CreateFontsCache(dir, output string, lcb logCallback) []string {
	obj := new(assProcessor)
	obj._fonts = dir
	obj.lcb = lcb
	return obj.createFontsCache(output)
}

func (self *mkvProcessor) CopyFontsFromCache(subs, dist string, lcb logCallback) bool {
	asses, _ := findPath(subs, `\.ass$`)
	obj := new(assProcessor)
	obj.lcb = lcb
	obj.files = asses
	obj.output = dist
	obj.loadCache(self.cache)
	return obj.copyFontsFromCache()
}

func (self *mkvProcessor) Cache(p string) {
	self.cache = p
}

func (self *mkvProcessor) MKS(mks bool) {
	self.mks = mks
}
