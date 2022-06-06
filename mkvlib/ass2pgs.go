package mkvlib

import (
	"os"
	"path"
)

func ass2Pgs(input []string, resolution, frameRate, fontsDir, output string, lcb logCallback) bool {
	r := false
	for _, item := range input {
		_, _, _, _f := splitPath(item)
		fn := path.Join(output, _f+".pgs")
		args := make([]string, 0)
		args = append(args, "-a1", "-p1")
		args = append(args, "-z0", "-u0", "-b0")
		args = append(args, "-g", fontsDir)
		args = append(args, "-v", resolution)
		args = append(args, "-f", frameRate)
		args = append(args, "-o", fn)
		args = append(args, item)
		if p, err := newProcess(nil, nil, nil, "", ass2bdnxml, args...); err == nil {
			s, err := p.Wait()
			r = err == nil && s.ExitCode() == 0
			if !r {
				printLog(lcb, LogError, `Failed to Ass2Pgs:"%s"`, item)
				_ = os.Remove(fn)
			}
		}
	}
	return r
}
