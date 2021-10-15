package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

const (
	mkvmerge   = `mkvmerge`
	mkvextract = `mkvextract`
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

func getMKVInfo(file string) *mkv {
	buf := bytes.NewBufferString("")
	if p, err := newProcess(nil, buf, nil, "", mkvmerge, "-J", file); err == nil {
		if s, err := p.Wait(); err == nil && s.ExitCode() == 0 {
			obj := new(mkv)
			_ = json.Unmarshal(buf.Bytes(), obj)
			return obj
		}
	}
	return nil
}

func dumpMKVs(dir string, subset bool) bool {
	ec := 0
	files, _ := findPath(dir, `\.mkv$`)
	arr := strings.Split(dir, string(os.PathSeparator))
	p := path.Join(`data`, arr[len(arr)-1])
	l := len(files)
	for i, item := range files {
		tmp := strings.Replace(item, dir, p, 1)
		obj := getMKVInfo(item)
		if obj == nil {
			ec++
			log.Printf(`Failed to get the mkv file info: "%s".`, item)
			break
		}
		attachments := make([]string, 0)
		tracks := make([]string, 0)
		for _, _item := range obj.Attachments {
			d, _, _, f := splitPath(tmp)
			attachments = append(attachments, fmt.Sprintf(`%d:%s`, _item.ID, path.Join(d, f, "fonts"+
				"", _item.FileName)))
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
				tracks = append(tracks, fmt.Sprintf(`%d:%s`, _item.ID, path.Join(d, f, s)))
			}
		}
		args := make([]string, 0)
		args = append(args, item)
		args = append(args, "attachments")
		args = append(args, attachments...)
		args = append(args, "tracks")
		args = append(args, tracks...)
		if p, err := newProcess(nil, nil, nil, "", mkvextract, args...); err == nil {
			s, err := p.Wait()
			ok := err == nil && s.ExitCode() == 0
			if ok {
				if subset {
					asses := make([]string, 0)
					for _, _item := range tracks {
						_arr := strings.Split(_item, ":")
						f := _arr[len(_arr)-1]
						if strings.HasSuffix(f, ".ass") {
							asses = append(asses, f)
						}
						if len(asses) > 0 {
							if !genASSes(asses, "", "") {
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
		if ec > 0 {
			log.Printf(`Failed to dump the mkv file "%s".`, item)
		}
		log.Printf("Dump (%d/%d) done.", i+1, l)
	}
	return ec == 0
}

func makeMKVs(dir string) bool {
	ec := 0
	_arr := strings.Split(dir, string(os.PathSeparator))
	p := _arr[len(_arr)-1]
	dir2 := path.Join(`data`, p)
	files, _ := findPath(dir, `\.mkv$`)
	l := len(files)
	for i, item := range files {
		tmp := strings.Replace(item, dir, p, 1)
		d, _, _, f := splitPath(tmp)
		d = strings.Replace(d, p, "", 1)
		_p := path.Join(dir2, d, f)
		__p := path.Join(_p, "output")
		attachments := findFonts(__p)
		subs, _ := findPath(_p, `\.sub`)
		asses, _ := findPath(__p, `\.ass$`)
		tracks := append(subs, asses...)
		args := make([]string, 0)
		args = append(args, "--output", path.Join("dist", tmp))
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
		if p, err := newProcess(nil, nil, nil, "", mkvmerge, args...); err == nil {
			s, err := p.Wait()
			ok := err == nil && s.ExitCode() == 0
			if !ok {
				ec++
			}
		} else {
			ec++
		}
		if ec > 0 {
			log.Printf(`Faild to make the mkv file: "%s".`, item)
		}
		log.Printf("Make (%d/%d) done.", i+1, l)
	}
	return ec == 0
}

func createMKVs(dir string, slang, stitle string) bool {
	ec := 0
	v := path.Join(dir, "v")
	s := path.Join(dir, "s")
	f := path.Join(dir, "f")
	t := path.Join(dir, "t")
	o := path.Join(dir, "o")
	files, _ := findPath(v, fmt.Sprintf(`\.\S+$`))
	l := len(files)
	_ = os.RemoveAll(t)
	for i, item := range files {
		_, _, _, _f := splitPath(item)
		tmp, _ := findPath(s, fmt.Sprintf(`%s\S*\.\S+$`, regexp.QuoteMeta(_f)))
		asses := make([]string, 0)
		subs := make([]string, 0)
		p := path.Join(t, _f)
		for _, sub := range tmp {
			if strings.HasSuffix(sub, ".ass") {
				_, _, _, __f := splitPath(sub)
				__s := path.Join(p, __f) + ".ass"
				_ = copyFileOrDir(sub, __s)
				asses = append(asses, __s)
			} else {
				subs = append(subs, sub)
			}
		}
		if len(asses) > 0 {
			if !genASSes(asses, f, "") {
				ec++
			}
		}
		__p := path.Join(p, "output")
		attachments := findFonts(__p)
		tracks, _ := findPath(__p, `\.ass$`)
		tracks = append(tracks, subs...)
		args := make([]string, 0)
		args = append(args, "--output", path.Join(o, _f)+".mkv")
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
		if p, err := newProcess(nil, nil, nil, "", mkvmerge, args...); err == nil {
			s, err := p.Wait()
			ok := err == nil && s.ExitCode() == 0
			if !ok {
				ec++
			}
		} else {
			ec++
		}
		if ec > 0 {
			log.Printf(`Failed to create the mkv file: "%s".`, item)
		}
		log.Printf("Create (%d/%d) done.", i+1, l)
	}
	return ec == 0
}

func checkSubset(path string) (bool, bool) {
	obj := getMKVInfo(path)
	if obj == nil {
		log.Printf(`Failed to get the mkv file info: "%s".`, path)
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

func queryFolder(dir string) bool {
	ec := 0
	lines := make([]string, 0)
	files, _ := findPath(dir, `\.mkv$`)
	l := len(files)
	for i, file := range files {
		a, b := checkSubset(file)
		if b {
			ec++
		} else if !a {
			lines = append(lines, file)
		}
		log.Printf("Query (%d/%d) done.", i+1, l)
	}
	if len(lines) > 0 {
		fmt.Print("Has item(s).")
		data := []byte(strings.Join(lines, "\n"))
		if os.WriteFile("list.txt", data, os.ModePerm) != nil {
			log.Printf(`Faild to write the dir result file: "%s".`, dir)
			ec++
		}
	} else {
		fmt.Print("No item.")
	}
	return ec == 0
}
