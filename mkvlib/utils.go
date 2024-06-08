package mkvlib

import (
	"errors"
	"fmt"
	"github.com/gogs/chardet"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func newDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func queryPath(path string, cb func(string) bool) error {
	return filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		if cb(path) {
			return nil
		}
		return errors.New("call cb return false")
	})
}

func findPath(path, expr string) (list []string, err error) {
	list = make([]string, 0)
	reg, e := regexp.Compile(expr)
	if e != nil {
		err = e
		return
	}
	err = queryPath(path, func(path string) bool {
		if expr == "" || reg.MatchString(path) {
			list = append(list, path)
		}
		return true
	})
	return
}

func copyFolder(src, dst string) error {
	e, f := isExists(src)
	if !e {
		return errors.New("src is not exists")
	}
	if !f {
		return errors.New("src is not folder")
	}
	if newDir(dst) != nil {
		return errors.New("faild to create dst folder")
	}
	s := len(src)
	if _, n, _, _ := splitPath(dst); n == "" {
		_, n, _, _ = splitPath(src)
		if n == "" {
			_, n, _, _ = splitPath(src[:len(src)-1])
		}
		dst = fmt.Sprintf("%s/%s", dst, n)
	}
	return filepath.Walk(src, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		return copyFile(path, dst+"/"+path[s:])
	})
}

func newFile(fp string) (file *os.File, err error) {
	dir, _ := filepath.Split(fp)
	if dir != "" {
		err = newDir(dir)
		if err != nil {
			return
		}
	}
	if err == nil {
		file, err = os.Create(fp)
	}
	return
}

func openFile(filepath string, readOnly, create bool) (file *os.File, err error) {
	f := os.O_RDWR | os.O_CREATE
	if readOnly {
		f = os.O_RDONLY
	}
	file, err = os.OpenFile(filepath, f, os.ModePerm)
	if err != nil && create {
		file, err = newFile(filepath)
	}
	return
}

func copyFile(src, dst string) error {
	e, f := isExists(src)
	if !e {
		return errors.New("src is not exists")
	}
	if f {
		return errors.New("src is not file")
	}
	if _, n, _, _ := splitPath(dst); n == "" {
		_, n, _, _ = splitPath(src)
		dst = path.Join(dst, n)
	}
	sf, err := openFile(src, true, false)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := openFile(dst, false, true)
	if err != nil {
		return err
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	return err
}

func splitPath(p string) (dir, name, ext, namewithoutext string) {
	dir, name = filepath.Split(p)
	ext = filepath.Ext(name)
	n := strings.LastIndex(name, ".")
	if n > 0 {
		namewithoutext = name[:n]
	}
	return
}

func isExists(path string) (exists bool, isFolder bool) {
	f, err := os.Stat(path)
	exists = err == nil || os.IsExist(err)
	if exists {
		isFolder = f.IsDir()
	}
	return
}

func copyFileOrDir(src, dst string) error {
	e, f := isExists(src)
	if !e {
		return errors.New("src is not exists")
	}
	if !f {
		return copyFile(src, dst)
	}
	return copyFolder(src, dst)
}

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomN(n int) int {
	return random.Intn(n)
}

func randomNumber(min, max int) int {
	sub := max - min + 1
	if sub <= 1 {
		return min
	}
	return min + randomN(sub)
}

func findMKVs(dir string) []string {
	list, _ := findPath(dir, `\.mkv$`)
	return list
}

func findFonts(dir string) []string {
	list, _ := findPath(dir, `\.((?i)(ttf)|(otf)|(ttc))$`)
	return list
}

func stringDeduplication(str string) string {
	_s := ""
	_m := make(map[rune]int)
	_rs := []rune(str)
	for i, r := range _rs {
		if _, ok := _m[r]; !ok {
			_m[r] = i
		}
	}
	for i, v := range _rs {
		if _m[v] == i {
			_s += string(_rs[i])
		}
	}
	return _s
}

func randomStr(l int) string {
	str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	var result []byte
	_l := len(str) - 1
	for i := 0; i < l; i++ {
		n := randomNumber(0, _l)
		result = append(result, bytes[n])
	}
	return string(result)
}

func toUTF8(data []byte) string {
	d := chardet.NewTextDetector()
	if r, err := d.DetectBest(data); err == nil {
		switch r.Charset {
		case "UTF-16LE":
			data, _, _ = transform.Bytes(unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder(), data)
			break
		case "UTF-16BE":
			data, _, _ = transform.Bytes(unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder(), data)
			break
		case "GBK", "GB2312":
			data, _, _ = transform.Bytes(simplifiedchinese.GBK.NewDecoder(), data)
			break
		case "GB18030":
			data, _, _ = transform.Bytes(simplifiedchinese.GB18030.NewDecoder(), data)
			break
		case "Big5":
			data, _, _ = transform.Bytes(traditionalchinese.Big5.NewDecoder(), data)
			break

		}
	}
	return string(data)
}
