package c

//go:generate echo "Generating code..."

/*
#include <stdlib.h>

#include <subset.h>
*/
import "C"
import (
	"unsafe"
)

func Subset(oldpath string, idx int, newpath, newname, dest, txt string) bool {
	cOldpath := C.CString(oldpath)
	defer C.free(unsafe.Pointer(cOldpath))
	cNewpath := C.CString(newpath)
	defer C.free(unsafe.Pointer(cNewpath))
	cDest := C.CString(dest)
	defer C.free(unsafe.Pointer(cDest))
	cNewname := C.CString(newname)
	defer C.free(unsafe.Pointer(cNewname))
	cTxt := C.CString(txt)
	defer C.free(unsafe.Pointer(cTxt))

	return bool(C.subset(cOldpath, C.int(idx), cNewpath, cNewname, cDest, cTxt))
}
