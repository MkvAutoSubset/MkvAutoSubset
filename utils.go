package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"unsafe"
)

func newProcess(stdin io.Reader, stdout, stderr io.Writer, dir, prog string, args ...string) (p *os.Process, err error) {
	cmd := exec.Command(prog, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if stdin != nil {
		cmd.Stdin = stdin
	}
	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}
	err = cmd.Start()
	if err == nil {
		p = cmd.Process
	}
	return
}

func setWindowTitle(title string) {
	kernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err == nil {
		defer syscall.FreeLibrary(kernel32)
		setConsoleTitle, err := syscall.GetProcAddress(kernel32, "SetConsoleTitleW")
		if err == nil {
			ptr, err := syscall.UTF16PtrFromString(title)
			if err == nil {
				syscall.Syscall(setConsoleTitle, 1, uintptr(unsafe.Pointer(ptr)), 0, 0)
			}
		}
	}
}

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

func openFile(filepath string, readOnly bool) (file *os.File, err error) {
	f := os.O_RDWR
	if readOnly {
		f = os.O_RDONLY
	}
	file, err = os.OpenFile(filepath, f, os.ModePerm)
	if err != nil {
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
		dst = fmt.Sprintf("%s/%s", dst, n)
	}
	sf, err := openFile(src, true)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := openFile(dst, false)
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
