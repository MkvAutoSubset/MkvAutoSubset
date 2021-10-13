package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	mkvmerge      = `mkvmerge`
	mkvextract    = `mkvextract`
	assfontsubset = `AssFontSubset`
)

type mkv struct {
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

func main() {
	setWindowTitle("MKV Tool v2.1.1")
	s := ""
	c := false
	d := false
	m := false
	n := false
	q := false
	sl, st := "", ""
	flag.StringVar(&s, "s", "", "Source folder.")
	flag.BoolVar(&c, "c", false, "Create mode.")
	flag.BoolVar(&d, "d", false, "Dump mode.")
	flag.BoolVar(&m, "m", false, "Make mode.")
	flag.BoolVar(&q, "q", false, "Query mode.")
	flag.BoolVar(&n, "n", false, "Not do ass font subset. (dump mode only)")
	flag.StringVar(&sl, "sl", "chi", " Subtitle language. (create mode only)")
	flag.StringVar(&st, "st", "", " Subtitle title. (create mode only)")
	flag.Parse()
	if s != "" {
		if q {
			queryFolder(s)
			return
		}
		if c {
			if sl != "" {
				createMKVs(s, sl, st)
				return
			}
		}
		if d {
			dumpMKVs(s, !n)
			return
		}
		arr := strings.Split(s, `\`)
		p := fmt.Sprintf(`data\%s`, arr[len(arr)-1])
		if m {
			makeMKVs(s, p)
			return
		}
		dumpMKVs(s, true)
		makeMKVs(s, p)
		return
	}
	flag.PrintDefaults()
}

func getMKVInfo(path string) *mkv {
	buf := bytes.NewBufferString("")
	p, _ := newProcess(nil, buf, nil, "", mkvmerge, "-J", path)
	_, _ = p.Wait()
	obj := new(mkv)
	_ = json.Unmarshal(buf.Bytes(), obj)
	return obj
}

func dumpMKVs(dir string, subset bool) {
	files, _ := findPath(dir, `\.mkv$`)
	arr := strings.Split(dir, `\`)
	p := fmt.Sprintf(`data\%s`, arr[len(arr)-1])
	l := len(files)
	for i, item := range files {
		tmp := strings.Replace(item, dir, p, 1)
		obj := getMKVInfo(item)
		attachments := make([]string, 0)
		tracks := make([]string, 0)
		for _, _item := range obj.Attachments {
			d, _, _, f := splitPath(tmp)
			attachments = append(attachments, fmt.Sprintf(`%d:%s`, _item.ID, fmt.Sprintf(`%s%s\fonts\%s`, d, f, _item.FileName)))
		}
		for _, _item := range obj.Tracks {
			if _item.Type == "subtitles" {
				d, _, _, f := splitPath(tmp)
				s := fmt.Sprintf(`%d_%s_%s`, _item.ID, _item.Properties.Language, _item.Properties.TrackName)
				if _item.Codec == "SubStationAlpha" {
					s += ".ass"
				} else {
					s += ".sub"
				}
				tracks = append(tracks, fmt.Sprintf(`%d:%s`, _item.ID, fmt.Sprintf(`%s%s\%s`, d, f, s)))
			}
		}
		args := make([]string, 0)
		args = append(args, item)
		args = append(args, "attachments")
		args = append(args, attachments...)
		args = append(args, "tracks")
		args = append(args, tracks...)
		p, _ := newProcess(nil, nil, nil, "", mkvextract, args...)
		_, _ = p.Wait()
		if subset {
			asses := make([]string, 0)
			for _, _item := range tracks {
				_arr := strings.Split(_item, ":")
				f := _arr[len(_arr)-1]
				if strings.HasSuffix(f, ".ass") {
					asses = append(asses, f)
				}
				if len(asses) > 0 {
					p, _ = newProcess(nil, nil, nil, "", assfontsubset, asses...)
					_, _ = p.Wait()
				}
			}
		}
		fmt.Printf("\rDump (%d/%d) done.", i+1, l)
	}
}

func makeMKVs(dir, dir2 string) {
	files, _ := findPath(dir, `\.mkv$`)
	arr := strings.Split(dir2, `\`)
	p := arr[len(arr)-1]
	l := len(files)
	for i, item := range files {
		tmp := strings.Replace(item, dir, p, 1)
		d, _, _, f := splitPath(tmp)
		d = strings.Replace(d, p, "", 1)
		_p := fmt.Sprintf(`%s%s%s\`, dir2, d, f)
		__p := _p + "output"
		attachments, _ := findPath(__p, `\.(ttf)|(otf)|(ttc)|(fon)$`)
		subs, _ := findPath(_p, `\.sub`)
		asses, _ := findPath(__p, `\.ass$`)
		tracks := append(subs, asses...)
		args := make([]string, 0)
		args = append(args, "--output", fmt.Sprintf(`dist\%s`, tmp))
		args = append(args, "--no-subtitles", "--no-attachments")
		args = append(args, item)
		for _, _item := range attachments {
			args = append(args, "--attach-file", _item)
		}
		for _, _item := range tracks {
			_, _, _, f = splitPath(_item)
			_arr := strings.Split(f, "_")
			args = append(args, "--language", "0:"+_arr[1])
			if len(_arr) > 2 {
				args = append(args, "--track-name", "0:"+_arr[2])
			}
			args = append(args, _item)
		}
		p, _ := newProcess(nil, nil, nil, "", mkvmerge, args...)
		_, _ = p.Wait()
		fmt.Printf("\rMake (%d/%d) done.", i+1, l)
	}
}

func createMKVs(dir string, slang, stitle string) {
	v := dir + `\v`
	s := dir + `\s`
	f := dir + `\f`
	t := dir + `\t`
	o := dir + `\o`
	files, _ := findPath(v, fmt.Sprintf(`\.\S+$`))
	l := len(files)
	_ = os.RemoveAll(t)
	reg, _ := regexp.Compile(`[\*\.\?\+\$\^\[\]\(\)\{\}\|\\\/]`)
	for i, item := range files {
		_, _, _, _f := splitPath(item)
		_tf := reg.ReplaceAllString(_f, `\$0`)
		tmp, _ := findPath(s, fmt.Sprintf(`%s\S*\.\S+$`, _tf))
		asses := make([]string, 0)
		subs := make([]string, 0)
		p := fmt.Sprintf(`%s\%s\`, t, _f)
		for _, sub := range tmp {
			if strings.HasSuffix(sub, ".ass") {
				_, _, _, __f := splitPath(sub)
				__s := fmt.Sprintf(`%s%s.ass`, p, __f)
				_ = copyFileOrDir(sub, __s)
				asses = append(asses, __s)
			} else {
				subs = append(subs, sub)
			}
		}
		if len(asses) > 0 {
			asses = append([]string{"|f" + f}, asses...)
			_p, _ := newProcess(nil, nil, nil, "", assfontsubset, asses...)
			_, _ = _p.Wait()
		}
		__p := fmt.Sprintf(`%s\output`, p)
		attachments, _ := findPath(__p, `\.(ttf)|(otf)|(ttc)|(fon)$`)
		tracks, _ := findPath(__p, `\.ass$`)
		tracks = append(tracks, subs...)
		args := make([]string, 0)
		args = append(args, "--output", fmt.Sprintf(`%s\%s.mkv`, o, _f))
		args = append(args, item)
		for _, _item := range attachments {
			args = append(args, "--attach-file", _item)
		}
		for _, _item := range tracks {
			_, _, _, _f = splitPath(_item)
			_arr := strings.Split(_f, "_")
			_l := len(_arr)
			_sl := slang
			_st := stitle
			if _l > 1 {
				_sl = _arr[1]
			}
			if _l > 2 {
				_st = _arr[2]
			}
			args = append(args, "--language", "0:"+_sl)
			args = append(args, "--track-name", "0:"+_st)
			args = append(args, _item)
		}
		_p, _ := newProcess(nil, nil, nil, "", mkvmerge, args...)
		_, _ = _p.Wait()
		fmt.Printf("\rCreate (%d/%d) done.", i+1, l)
	}
}

func checkSubset(path string) bool {
	obj := getMKVInfo(path)
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
	return !ass || (ass && ok)
}

func queryFolder(dir string) {
	lines := make([]string, 0)
	files, _ := findPath(dir, `\.mkv$`)
	l := len(files)
	for i, file := range files {
		if !checkSubset(file) {
			lines = append(lines, file)
		}
		fmt.Printf("\rQuery (%d/%d) done.", i+1, l)
	}
	if len(lines) > 0 {
		fmt.Print("\rHas item(s).")
		data := []byte(strings.Join(lines, "\n"))
		_ = os.WriteFile("list.txt", data, os.ModePerm)
	} else {
		fmt.Print("\rNo item.")
	}
}
