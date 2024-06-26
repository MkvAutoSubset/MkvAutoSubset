/*
add to build.rs:
    let so = "mkvlib";
    println!("cargo:rustc-link-search=native=.");
    println!("{}{}", "cargo:rustc-link-lib=dylib=", so);
require crate
    json
*/

use {
    json,
    std::ffi,
    std::os::raw,
};

pub type c_char = *const raw::c_char;
pub type c_uchar = raw::c_uchar;
pub type logCallback = Option<fn(c_uchar, c_char)>;

extern {
    fn A2P(a2p: bool, apc: bool, pr: c_char, pf: c_char);
    fn ASSFontSubset(files: c_char, fonts: c_char, output: c_char, dirSafe: bool, lcb: logCallback) -> bool;
    fn Cache(ccs: c_char);
    fn Check(check: bool, strict: bool);
    fn CheckSubset(file: c_char, lcb: logCallback) -> c_char;
    fn CopyFontsFromCache(asses: c_char, dist: c_char, lcb: logCallback) -> bool;
    fn CreateBlankOrBurnVideo(t: u64, s: c_char, enc: c_char, ass: c_char, fontdir: c_char, output: c_char) -> bool;
    fn CreateFontsCache(dir: c_char, output: c_char, lcb: logCallback) -> c_char;
    fn CreateMKV(file: c_char, tracks: c_char, attachments: c_char, output: c_char, slang: c_char, stitle: c_char, clean: bool) -> bool;
    fn CreateMKVs(vDir: c_char, sDir: c_char, fDir: c_char, tDir: c_char, oDir: c_char, slang: c_char, stitle: c_char, clean: bool, lcb: logCallback) -> bool;
    fn CreateTestVideo(asses: c_char, s: c_char, fontdir: c_char, enc: c_char, burn: bool, lcb: logCallback) -> bool;
    fn DumpMKV(file: c_char, output: c_char, subset: bool, lcb: logCallback) -> bool;
    fn DumpMKVs(dir: c_char, output: c_char, subset: bool, lcb: logCallback) -> bool;
    fn GetFontInfo(p: c_char) -> c_char;
    fn GetFontsList(files: c_char, fonts: c_char, lcb: logCallback) -> c_char;
    fn GetMKVInfo(file: c_char) -> c_char;
    fn InitInstance(lcb: logCallback) -> bool;
    fn MKS(mks: bool);
    fn MakeMKVs(dir: c_char, data: c_char, output: c_char, slang: c_char, stitle: c_char, subset: bool, lcb: logCallback) -> bool;
    fn NOverwrite(n: bool);
    fn NRename(n: bool);
    fn QueryFolder(dir: c_char, lcb: logCallback) -> c_char;
    fn Version() -> c_char;
}

pub fn rs(cs: c_char) -> &'static str {
    unsafe {
        return ffi::CStr::from_ptr(cs).to_str().unwrap();
    }
}

fn cs(rs: &str) -> c_char {
    let s = ffi::CString::new(rs).unwrap();
    return s.into_raw();
}

fn jtoa(str: c_char) -> Vec<String> {
    let mut arr: Vec<String> = vec![];
    let str = rs(str);
    if let json::JsonValue::Array(_arr) = json::parse(str).unwrap() {
        for x in _arr {
            if let json::JsonValue::String(x) = x {
                arr.push(x.to_string());
            }
        }
    }
    return arr;
}

pub fn a2p(a2p: bool, apc: bool, pr: &str, pf: &str) {
    unsafe {
        A2P(a2p, apc, cs(pr), cs(pf));
    }
}

pub fn assFontSubset(files: &[&str], fonts: &str, output: &str, dirSafe: bool, lcb: logCallback) -> bool {
    unsafe {
        let files = json::stringify(files);
        return ASSFontSubset(cs(files.as_str()), cs(fonts), cs(output), dirSafe, lcb);
    }
}

pub fn cache(ccs: &str) {
    unsafe {
        Cache(cs(ccs));
    }
}

pub fn check(check: bool, strict: bool) {
    unsafe {
        Check(check, strict);
    }
}

pub fn checkSubset(file: &str, lcb: logCallback) -> [bool; 2] {
    unsafe {
        let str = rs(CheckSubset(cs(file), lcb));
        let arr = [str.contains("[true"), str.contains("true]")];
        return arr;
    }
}

pub fn copyFontsFromCache(asses: &[&str], dist: &str, lcb: logCallback) -> bool {
    unsafe {
        let files = json::stringify(asses);
        return CopyFontsFromCache(cs(files.as_str()), cs(dist), lcb);
    }
}

pub fn createBlankOrBurnVideo(t: u64, s: &str, enc: &str, ass: &str, fontdir: &str, output: &str) -> bool {
    unsafe {
        return CreateBlankOrBurnVideo(t, cs(s), cs(enc), cs(ass), cs(fontdir), cs(output));
    }
}

pub fn createFontsCache(dir: &str, output: &str, lcb: logCallback) -> Vec<String> {
    unsafe {
        let str = CreateFontsCache(cs(dir), cs(output), lcb);
        return jtoa(str);
    }
}

pub fn createMKV(file: &str, tracks: &[&str], attachments: &[&str], output: &str, slang: &str, stitle: &str, clean: bool) -> bool {
    unsafe {
        let tracks = json::stringify(tracks);
        let attachments = json::stringify(attachments);
        return CreateMKV(cs(file), cs(tracks.as_str()), cs(attachments.as_str()), cs(output), cs(slang), cs(stitle), clean);
    }
}

pub fn createMKVs(vDir: &str, sDir: &str, fDir: &str, tDir: &str, oDir: &str, slang: &str, stitle: &str, clean: bool, lcb: logCallback) -> bool {
    unsafe {
        return CreateMKVs(cs(vDir), cs(sDir), cs(fDir), cs(tDir), cs(oDir), cs(slang), cs(stitle), clean, lcb);
    }
}

pub fn createTestVideo(asses: &[&str], s: &str, fontdir: &str, enc: &str, burn: bool, lcb: logCallback) -> bool {
    unsafe {
        let asses = json::stringify(asses);
        return CreateTestVideo(cs(asses.as_str()), cs(s), cs(fontdir), cs(enc), burn, lcb);
    }
}

pub fn dumpMKV(file: &str, output: &str, subset: bool, lcb: logCallback) -> bool {
    unsafe {
        return DumpMKV(cs(file), cs(output), subset, lcb);
    }
}

pub fn dumpMKVs(dir: &str, output: &str, subset: bool, lcb: logCallback) -> bool {
    unsafe {
        return DumpMKVs(cs(dir), cs(output), subset, lcb);
    }
}

pub fn getFontInfo(p: &str) -> &str {
    unsafe {
        return rs(GetFontInfo(cs(p)));
    }
}

pub fn getFontsList(files: &[&str], fonts: &str, lcb: logCallback) -> Vec<Vec<String>> {
    unsafe {
        let files = json::stringify(files);
        let str = rs(GetFontsList(cs(files.as_str()), cs(fonts), lcb));
        let j = json::parse(str).unwrap();
        let mut arr: Vec<Vec<String>> = vec![];
        let mut c = 0;
        if let json::JsonValue::Array(j) = j {
            for x in j {
                arr.push(vec![]);
                arr[c] = vec![];
                if let json::JsonValue::Array(y) = x {
                    for y in y {
                        if let json::JsonValue::Short(y) = y {
                            arr[c].push(y.to_string());
                        }
                    }
                }
                c += 1;
            }
        }
        return arr;
    }
}

pub fn getMKVInfo(file: &str) -> &str {
    unsafe {
        return rs(GetMKVInfo(cs(file)));
    }
}

pub fn initInstance(lcb: logCallback) -> bool {
    unsafe {
        return InitInstance(lcb);
    };
}

pub fn mks(mks: bool) {
    unsafe {
        MKS(mks);
    }
}

pub fn makeMKVs(dir: &str, data: &str, output: &str, slang: &str, stitle: &str, subset: bool, lcb: logCallback) -> bool {
    unsafe {
        return MakeMKVs(cs(dir), cs(data), cs(output), cs(slang), cs(stitle), subset, lcb);
    }
}

pub fn noverwrite(n: bool) {
    unsafe {
        NOverwrite(n);
    }
}

pub fn nrename(n: bool) {
    unsafe {
        NRename(n);
    }
}

pub fn queryFolder(dir: &str, lcb: logCallback) -> Vec<String> {
    unsafe {
        let str = QueryFolder(cs(dir), lcb);
        return jtoa(str);
    }
}

pub fn version() -> &'static str {
    unsafe {
        return rs(Version());
    };
}