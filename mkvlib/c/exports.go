package main

// #include "lcb.h"
import "C"
import (
	"encoding/json"
	"github.com/KurenaiRyu/MkvAutoSubset/mkvlib"
)

var getter = mkvlib.GetProcessorGetterInstance()

func checkInstance() bool {
	return getter.GetProcessorInstance() != nil
}

func _lcb(lcb C.logCallback) func(string) {
	return func(str string) {
		C.makeLogCallback(cs(str), lcb)
	}
}

//export InitInstance
func InitInstance(lcb C.logCallback) bool {
	return getter.InitProcessorInstance(_lcb(lcb))
}

//export GetMKVInfo
func GetMKVInfo(file *C.char) *C.char {
	if !checkInstance() {
		return cs("")
	}
	obj := getter.GetProcessorInstance().GetMKVInfo(gs(file))
	data, _ := json.Marshal(obj)
	return cs(string(data))
}

//export DumpMKV
func DumpMKV(file, output *C.char, subset bool, lcb C.logCallback) bool {
	if !checkInstance() {
		return false
	}
	return getter.GetProcessorInstance().DumpMKV(gs(file), gs(output), subset, _lcb(lcb))
}

type checkSubset_R struct {
	Subsetted bool `json:"subsetted"`
	Error     bool `json:"error"`
}

//export CheckSubset
func CheckSubset(file *C.char, lcb C.logCallback) *C.char {
	if !checkInstance() {
		return cs("")
	}
	a, b := getter.GetProcessorInstance().CheckSubset(gs(file), _lcb(lcb))
	data, _ := json.Marshal(checkSubset_R{a, b})
	return cs(string(data))
}

//export CreateMKV
func CreateMKV(file, tracks, attachments, output, slang, stitle *C.char, clean bool) bool {
	if !checkInstance() {
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
			return getter.GetProcessorInstance().CreateMKV(gs(file), _tracks, _attachments, gs(output), gs(slang), gs(stitle), clean)
		}
	}
	return false
}

//export ASSFontSubset
func ASSFontSubset(files, fonts, output *C.char, dirSafe bool, lcb C.logCallback) bool {
	if !checkInstance() {
		return false
	}
	obj := make([]string, 0)
	if json.Unmarshal([]byte(gs(files)), &obj) == nil {
		_files := obj
		return getter.GetProcessorInstance().ASSFontSubset(_files, gs(fonts), gs(output), dirSafe, _lcb(lcb))
	}
	return false
}

//export QueryFolder
func QueryFolder(dir *C.char, lcb C.logCallback) *C.char {
	if !checkInstance() {
		return cs("")
	}
	list := getter.GetProcessorInstance().QueryFolder(gs(dir), _lcb(lcb))
	data, _ := json.Marshal(list)
	return cs(string(data))
}

//export DumpMKVs
func DumpMKVs(dir, output *C.char, subset bool, lcb C.logCallback) bool {
	if !checkInstance() {
		return false
	}
	return getter.GetProcessorInstance().DumpMKVs(gs(dir), gs(output), subset, _lcb(lcb))
}

//export CreateMKVs
func CreateMKVs(vDir, sDir, fDir, tDir, oDir, slang, stitle *C.char, clean bool, lcb C.logCallback) bool {
	if !checkInstance() {
		return false
	}
	return getter.GetProcessorInstance().CreateMKVs(gs(vDir), gs(sDir), gs(fDir), gs(tDir), gs(oDir), gs(slang), gs(stitle), clean, _lcb(lcb))
}

//export MakeMKVs
func MakeMKVs(dir, data, output, slang, stitle *C.char, lcb C.logCallback) bool {
	if !checkInstance() {
		return false
	}
	return getter.GetProcessorInstance().MakeMKVs(gs(dir), gs(data), gs(output), gs(slang), gs(stitle), _lcb(lcb))
}

//export A2P
func A2P(a2p, apc bool, pr, pf string) {
	if !checkInstance() {
		return
	}
	getter.GetProcessorInstance().A2P(a2p, apc, pr, pf)
}

//export GetFontsList
func GetFontsList(dir *C.char, lcb C.logCallback) *C.char {
	if !checkInstance() {
		return cs("")
	}
	list := getter.GetProcessorInstance().GetFontsList(gs(dir), _lcb(lcb))
	data, _ := json.Marshal(list)
	return cs(string(data))
}

//export CreateFontsCache
func CreateFontsCache(dir, output *C.char, lcb C.logCallback) *C.char {
	if !checkInstance() {
		return cs("")
	}
	list := getter.GetProcessorInstance().CreateFontsCache(gs(dir), gs(output), _lcb(lcb))
	data, _ := json.Marshal(list)
	return cs(string(data))
}

//export CopyFontsFromCache
func CopyFontsFromCache(subs, dist *C.char, lcb C.logCallback) bool {
	if !checkInstance() {
		return false
	}
	return getter.GetProcessorInstance().CopyFontsFromCache(gs(subs), gs(dist), _lcb(lcb))
}

//export Cache
func Cache(p *C.char) {
	if !checkInstance() {
		return
	}
	getter.GetProcessorInstance().Cache(gs(p))
}

//export MKS
func MKS(mks bool) {
	if !checkInstance() {
		return
	}
	getter.GetProcessorInstance().MKS(mks)
}

func cs(gs string) *C.char {
	return C.CString(gs)
}

func gs(cs *C.char) string {
	return C.GoString(cs)
}
