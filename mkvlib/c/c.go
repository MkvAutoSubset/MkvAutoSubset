package c

/*
#include <stdlib.h>
#include <stdbool.h>

bool subset(const char *oldpath, int idx, const char *newpath, const char *newname, const char *dest, const char *txt);

#include "ass2bdnxml/ass2bdnxml.c"
#include "ass2bdnxml/sup.c"
#include "ass2bdnxml/auto_split.c"
#include "ass2bdnxml/palletize.c"
#include "ass2bdnxml/sort.c"
#include "ass2bdnxml/ass2bdnxml.h"

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

func Ass2Pgs(input string, resolution, frameRate, fontsDir, output string) bool {
	cInput := C.CString(input)
	defer C.free(unsafe.Pointer(cInput))
	cResolution := C.CString(resolution)
	defer C.free(unsafe.Pointer(cResolution))
	cFrameRate := C.CString(frameRate)
	defer C.free(unsafe.Pointer(cFrameRate))
	cFontsDir := C.CString(fontsDir)
	defer C.free(unsafe.Pointer(cFontsDir))
	cOutput := C.CString(output)
	defer C.free(unsafe.Pointer(cOutput))

	return bool(C.ass2pgs(cInput, cResolution, cFrameRate, cFontsDir, cOutput))
}
