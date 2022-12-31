from ctypes import *
from json import *

libpath = "./mkvlib.so"
lib = CDLL(libpath)


def _lcb(lcb):
    @CFUNCTYPE(None, c_byte, c_char_p)
    def logcallback(l, s):
        if lcb:
            lcb(l, s.decode())

    return logcallback


def version():
    call = lib.Version
    call.restype = c_char_p
    return call().decode()


def initInstance(lcb):
    call = lib.InitInstance
    return call(_lcb(lcb))


def getMKVInfo(file):
    call = lib.GetMKVInfo
    call.restype = c_char_p
    return loads(call(file.encode()).decode())


def dumpMKV(file, output, subset, lcb):
    call = lib.DumpMKV
    return call(file.encode(), output.encode(), subset, _lcb(lcb))


def checkSubset(file, lcb):
    call = lib.CheckSubset
    call.restype = c_char_p
    return loads(call(file.encode(), _lcb(lcb)).decode())


def createMKV(file, tracks, attachments, output, slang, stitle, clean, lcb):
    call = lib.CreateMKV
    _tracks = dumps(tracks)
    _attachments = dumps(attachments)
    return call(file.encode(), _tracks.encode(), _attachments.encode(), output.encode(), slang.encode(),
                stitle.encode(), clean, _lcb(lcb))


def assFontSubset(files, fonts, output, dirSafe, lcb):
    call = lib.ASSFontSubset
    _files = dumps(files)
    return call(_files.encode(), fonts.encode(), output.encode(), dirSafe, _lcb(lcb))


def queryFolder(dir, lcb):
    call = lib.QueryFolder
    call.restype = c_char_p
    return call(dir.encode(), _lcb(lcb))


def dumpMKVs(dir, output, subset, lcb):
    call = lib.DumpMKVs
    return call(dir.encode(), output.encode(), subset, _lcb(lcb))


def createMKVs(vDir, sDir, fDir, tDir, oDir, slang, stitle, clean, lcb):
    call = lib.CreateMKVs
    return call(vDir.encode(), sDir.encode(), fDir.encode(), tDir.encode(), oDir.encode(), slang.encode(),
                stitle.encode(), clean, _lcb(lcb))


def makeMKVs(dir, data, output, slang, stitle, subset, lcb):
    call = lib.MakeMKVs
    return call(dir.encode(), data.encode(), output.encode(), slang.encode(), stitle.encode(), subset, _lcb(lcb))


def createBlankOrBurnVideo(t, s, enc, ass, fontdir, output):
    call = lib.CreateBlankOrBurnVideo
    call(t.encode(), s.encode(), enc.encode(), ass.encode(), fontdir.encode(), output.encode())


def createTestVideo(asses, s, fontdir, enc, burn, lcb):
    call = lib.CreateTestVideo
    _files = dumps(asses)
    call(_files.encode(), s.encode(), fontdir.encode(), enc.encode(), burn, _lcb(lcb))


def a2p(en, apc, pr, pf):
    call = lib.A2P
    call(en, apc, pr.encode(), pf.encode())


def getFontsList(files, fonts, lcb):
    call = lib.GetFontsList
    call.restype = c_char_p
    _files = dumps(files)
    return loads(call(_files.encode(), fonts.encode(), _lcb(lcb)).decode())


def cache(ccs):
    call = lib.Cache
    _ccs = dumps(ccs)
    call(_ccs.encode())


def getFontInfo(p):
    call = lib.GetFontInfo
    call.restype = c_char_p
    return loads(call(p.encode()).decode())


def createFontsCache(dir, output, lcb):
    call = lib.CreateFontsCache
    call.restype = c_char_p
    return loads(call(dir.encode(), output.encode(), _lcb(lcb)).decode())


def copyFontsFromCache(asses, dist, lcb):
    call = lib.CopyFontsFromCache
    _files = dumps(asses)
    return call(_files.encode(), dist.encode(), _lcb(lcb))


def mks(en):
    call = lib.MKS
    call(en)


def nrename(n):
    call = lib.NRename
    call(n)


def noverwrite(n):
    call = lib.NOverwrite
    call(n)


def check(c, s):
    call = lib.Check
    call(c, s)
