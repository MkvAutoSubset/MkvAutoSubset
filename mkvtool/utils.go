package main

import (
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

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

func splitPath(p string) (dir, name, ext, namewithoutext string) {
	dir, name = filepath.Split(p)
	ext = filepath.Ext(name)
	n := strings.LastIndex(name, ".")
	if n > 0 {
		namewithoutext = name[:n]
	}
	return
}

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomN(n int) int {
	return r.Intn(n)
}

func randomNumber(min, max int) int {
	sub := max - min + 1
	if sub <= 1 {
		return min
	}
	return min + randomN(sub)
}

func randomStr(l int) string {
	str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	var result []byte
	lstr := len(str) - 1
	for i := 0; i < l; i++ {
		n := randomNumber(0, lstr)
		result = append(result, bytes[n])
	}
	return string(result)
}
