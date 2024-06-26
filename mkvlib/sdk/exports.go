package main

// #include "lcb.h"
import "C"
import (
	"encoding/json"
	"github.com/MkvAutoSubset/MkvAutoSubset/mkvlib"
)

var getter = mkvlib.GetProcessorGetterInstance()

func checkInstance() bool {
	return getter.GetProcessorInstance() != nil
}

func _lcb(lcb C.logCallback) func(byte, string) {
	return func(l byte, str string) {
		C.makeLogCallback(C.uchar(l), cs(str), lcb)
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

//export CheckSubset
func CheckSubset(file *C.char, lcb C.logCallback) *C.char {
	if !checkInstance() {
		return cs("")
	}
	a, b := getter.GetProcessorInstance().CheckSubset(gs(file), _lcb(lcb))
	data, _ := json.Marshal([]bool{a, b})
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
func MakeMKVs(dir, data, output, slang, stitle *C.char, subset bool, lcb C.logCallback) bool {
	if !checkInstance() {
		return false
	}
	return getter.GetProcessorInstance().MakeMKVs(gs(dir), gs(data), gs(output), gs(slang), gs(stitle), subset, _lcb(lcb))
}

//export A2P
func A2P(a2p, apc bool, pr, pf *C.char) {
	if !checkInstance() {
		return
	}
	getter.GetProcessorInstance().A2P(a2p, apc, gs(pr), gs(pf))
}

//export GetFontsList
func GetFontsList(files, fonts *C.char, lcb C.logCallback) *C.char {
	if !checkInstance() {
		return cs("")
	}
	obj := make([]string, 0)
	if json.Unmarshal([]byte(gs(files)), &obj) == nil {
		_files := obj
		list := getter.GetProcessorInstance().GetFontsList(_files, gs(fonts), _lcb(lcb))
		data, _ := json.Marshal(list)
		return cs(string(data))
	}
	return cs("")
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
func CopyFontsFromCache(asses, dist *C.char, lcb C.logCallback) bool {
	if !checkInstance() {
		return false
	}
	obj := make([]string, 0)
	if json.Unmarshal([]byte(gs(asses)), &obj) == nil {
		_files := obj
		return getter.GetProcessorInstance().CopyFontsFromCache(_files, gs(dist), _lcb(lcb))
	}
	return false
}

//export Cache
func Cache(ccs *C.char) {
	if !checkInstance() {
		return
	}
	obj := make([]string, 0)
	if json.Unmarshal([]byte(gs(ccs)), &obj) == nil {
		_ccs := obj
		getter.GetProcessorInstance().Cache(_ccs)
	}
}

//export MKS
func MKS(mks bool) {
	if !checkInstance() {
		return
	}
	getter.GetProcessorInstance().MKS(mks)
}

//export NRename
func NRename(n bool) {
	if !checkInstance() {
		return
	}
	getter.GetProcessorInstance().NRename(n)
}

//export Check
func Check(check, strict bool) {
	if !checkInstance() {
		return
	}
	getter.GetProcessorInstance().Check(check, strict)
}

//export GetFontInfo
func GetFontInfo(p *C.char) *C.char {
	if !checkInstance() {
		return cs("")
	}
	info := getter.GetProcessorInstance().GetFontInfo(gs(p))
	data, _ := json.Marshal(info)
	return cs(string(data))
}

//export NOverwrite
func NOverwrite(n bool) {
	if !checkInstance() {
		return
	}
	getter.GetProcessorInstance().NOverwrite(n)
}

//export Version
func Version() *C.char {
	return cs(mkvlib.Version())
}

//export CreateBlankOrBurnVideo
func CreateBlankOrBurnVideo(t int64, s, enc, ass, fontdir, output *C.char) bool {
	if !checkInstance() {
		return false
	}
	return getter.GetProcessorInstance().CreateBlankOrBurnVideo(t, gs(s), gs(enc), gs(ass), gs(fontdir), gs(output))
}

//export CreateTestVideo
func CreateTestVideo(asses, s, fontdir, enc *C.char, burn bool, lcb C.logCallback) bool {
	if !checkInstance() {
		return false
	}
	obj := make([]string, 0)
	if json.Unmarshal([]byte(gs(asses)), &obj) == nil {
		_asses := obj
		return getter.GetProcessorInstance().CreateTestVideo(_asses, gs(s), gs(fontdir), gs(enc), burn, _lcb(lcb))
	}
	return false
}

func cs(gs string) *C.char {
	return C.CString(gs)
}

func gs(cs *C.char) string {
	return C.GoString(cs)
}
