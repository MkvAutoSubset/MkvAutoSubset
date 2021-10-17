package main

import (
	"C"
	"encoding/json"
	"mkvlib"
)

var _instance = mkvlib.GetInstance()

//export CheckInstance
func CheckInstance() bool {
	return _instance != nil
}

//export GetMKVInfo
func GetMKVInfo(file *C.char) *C.char {
	if !CheckInstance() {
		return cs("")
	}
	obj := _instance.GetMKVInfo(gs(file))
	data, _ := json.Marshal(obj)
	return cs(string(data))
}

//export DumpMKV
func DumpMKV(file, output *C.char, subset bool) bool {
	if !CheckInstance() {
		return false
	}
	return _instance.DumpMKV(gs(file), gs(output), subset)
}

type checkSubset_R struct {
	Subseted bool `json:"subseted"`
	Error    bool `json:"error"`
}

//export CheckSubset
func CheckSubset(file *C.char) *C.char {
	if !CheckInstance() {
		return cs("")
	}
	a, b := _instance.CheckSubset(gs(file))
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
			return _instance.CreateMKV(gs(file), _tracks, _attachments, gs(output), gs(slang), gs(stitle), clean)
		}
	}
	return false
}

//export ASSFontSubset
func ASSFontSubset(files, fonts, output *C.char, dirSafe bool) bool {
	if !CheckInstance() {
		return false
	}
	obj := make([]string, 0)
	if json.Unmarshal([]byte(gs(files)), &obj) == nil {
		_files := obj
		return _instance.ASSFontSubset(_files, gs(fonts), gs(output), dirSafe)
	}
	return false
}

func cs(gs string) *C.char {
	return C.CString(gs)
}

func gs(cs *C.char) string {
	return C.GoString(cs)
}
