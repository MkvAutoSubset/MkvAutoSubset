from ctypes import *
from json import *

libpath="./mkvlib.so"
lib=CDLL(libpath)

@CFUNCTYPE(None, c_char_p)
def _lcb(s):
    print(s.decode())

def initInstance(lcb):
    call=lib.InitInstance
    return call(lcb)

def checkInstance():
    call=lib.CheckInstance
    return call()

def getMKVInfo(file):
    call=lib.GetMKVInfo
    call.restype=c_char_p
    return call(file.encode())

def dumpMKV(file,output,subset,dirSafe,lcb):
    call=lib.DumpMKV
    return call(file.encode(),output.encode(),subset,dirSafe,lcb)

def checkSubset(file,lcb):
    call=lib.CheckSubset
    call.restype=c_char_p
    return call(file.encode(),lcb)

def createMKV(file,tracks,attachments,output,slang,stitle,clean):
    call=lib.CreateMKV
    _tracks=dumps(tracks)
    _attachments=dumps(attachments)
    return call(file.encode(),_tracks.encode(),_attachments.encode(),output.encode(),slang.encode(),stitle.encode(),clean)

def assFontSubset(files,fonts,output,dirSafe,lcb):
    call=lib.ASSFontSubset
    _files=dumps(files)
    return call(_files.encode(),fonts.encode(),output.encode(),dirSafe,lcb)
