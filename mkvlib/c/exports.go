package main

// #include "lcb.h"
import "C"
import (
	"encoding/json"
	"github.com/KurenaiRyu/MkvAutoSubset/mkvlib"
)

var processor = mkvlib.GetProcessorGetterInstance()

func _lcb(lcb C.logCallback) func(string) {
	return func(str string) {
		C.makeLogCallback(cs(str), lcb)
	}
}

//export CheckInstance
func CheckInstance() bool {
	return processor.GetProcessorInstance() != nil
}

//export GetMKVInfo
func GetMKVInfo(file *C.char) *C.char {
	if !CheckInstance() {
		return cs("")
	}
	obj := processor.GetProcessorInstance().GetMKVInfo(gs(file))
	data, _ := json.Marshal(obj)
	return cs(string(data))
}

//export DumpMKV
func DumpMKV(file, output *C.char, subset bool, lcb C.logCallback) bool {
	if !CheckInstance() {
		return false
	}
	return processor.GetProcessorInstance().DumpMKV(gs(file), gs(output), subset, _lcb(lcb))
}

type checkSubset_R struct {
	Subsetted bool `json:"subsetted"`
	Error     bool `json:"error"`
}

//export CheckSubset
func CheckSubset(file *C.char, lcb C.logCallback) *C.char {
	if !CheckInstance() {
		return cs("")
	}
	a, b := processor.GetProcessorInstance().CheckSubset(gs(file), _lcb(lcb))
	data, _ := json.Marshal(checkSubset_R{a, b})
	return cs(string(data))
}

//export CreateMKV
func CreateMKV(file, tracks, attachments, output, slang, stitle *C.char, clean bool) bool {
	if !CheckInstance() {
		return false
	}
	a := make([]string, 0)
	b := make([]string, 0)
	err := json.Unmarshal([]byte(gs(tracks)), &a)
	if err == nil {
		_tracks := a
		err = json.Unmarshal([]byte(gs(attachments)), &b)
		if err == nil {
			_attachments := b
			return processor.GetProcessorInstance().CreateMKV(gs(file), _tracks, _attachments, gs(output), gs(slang), gs(stitle), clean)
		}
	}
	return false
}

//export ASSFontSubset
func ASSFontSubset(files, fonts, output *C.char, dirSafe bool, lcb C.logCallback) bool {
	if !CheckInstance() {
		return false
	}
	obj := make([]string, 0)
	if json.Unmarshal([]byte(gs(files)), &obj) == nil {
		_files := obj
		return processor.GetProcessorInstance().ASSFontSubset(_files, gs(fonts), gs(output), dirSafe, _lcb(lcb))
	}
	return false
}

func cs(gs string) *C.char {
	return C.CString(gs)
}

func gs(cs *C.char) string {
	return C.GoString(cs)
}
