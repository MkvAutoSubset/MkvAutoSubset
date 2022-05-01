using System;
using System.Runtime.InteropServices;
using System.Text.Json;

public static class mkvlib
{

    #region imports

    [DllImport("mkvlib.so")]
    static extern bool InitInstance(logCallback lcb);

    [DllImport("mkvlib.so")]
    static extern IntPtr GetMKVInfo(IntPtr ptr);

    [DllImport("mkvlib.so")]
    static extern bool DumpMKV(IntPtr file, IntPtr output, bool subset, logCallback lcb);

    [DllImport("mkvlib.so")]
    static extern IntPtr CheckSubset(IntPtr file, logCallback lcb);

    [DllImport("mkvlib.so")]
    static extern bool CreateMKV(IntPtr file, IntPtr tracks, IntPtr attachments, IntPtr output, IntPtr slang, IntPtr stitle, bool clean);

    [DllImport("mkvlib.so")]
    static extern bool ASSFontSubset(IntPtr files, IntPtr fonts, IntPtr output, bool dirSafe, logCallback lcb);

    [DllImport("mkvlib.so")]
    static extern IntPtr QueryFolder(IntPtr dir, logCallback lcb);

    [DllImport("mkvlib.so")]
    static extern bool DumpMKVs(IntPtr dir, IntPtr output, bool subset, logCallback lcb);

    [DllImport("mkvlib.so")]
    static extern bool CreateMKVs(IntPtr vDir, IntPtr sDir, IntPtr fDir, IntPtr tDir, IntPtr oDir, IntPtr slang, IntPtr stitle, bool clean, logCallback lcb);

    [DllImport("mkvlib.so")]
    static extern bool MakeMKVs(IntPtr dir, IntPtr data, IntPtr output, IntPtr slang, IntPtr stitle, logCallback lcb);

    [DllImport("mkvlib.so")]
    static extern void A2P(bool a2p, bool apc, IntPtr pr, IntPtr pf);

    [DllImport("mkvlib.so")]
    static extern IntPtr GetFontsList(IntPtr dir, logCallback lcb);

    [DllImport("mkvlib.so")]
    static extern void Cache(IntPtr ccs);

    [DllImport("mkvlib.so")]
    static extern void MKS(bool mks);

    [DllImport("mkvlib.so")]
    static extern void NRename(bool n);

    [DllImport("mkvlib.so")]
    static extern void Check(bool check, bool strict);

    [DllImport("mkvlib.so")]
    static extern IntPtr CreateFontsCache(IntPtr dir, IntPtr output, logCallback lcb);

    [DllImport("mkvlib.so")]
    static extern bool CopyFontsFromCache(IntPtr subs, IntPtr dist, logCallback lcb);

    [DllImport("mkvlib.so")]
    static extern IntPtr GetFontInfo(IntPtr p);

    #endregion

    public static bool InitInstance(Action<string> lcb)
    {
        return InitInstance(_lcb(lcb));
    }

    public static string GetMKVInfo(string file)
    {
        return css(GetMKVInfo(cs(file)));
    }

    public static bool DumpMKV(string file, string output, bool subset, Action<string> lcb)
    {
        return DumpMKV(cs(file), cs(output), subset, _lcb(lcb));
    }

    public static bool[] CheckSubset(string file, Action<string> lcb)
    {
        string json = css(CheckSubset(cs(file), _lcb(lcb)));
        JsonDocument doc = JsonDocument.Parse(json);
        bool[] result = new bool[2];
        result[0] = doc.RootElement.GetProperty("subsetted").GetBoolean();
        result[1] = doc.RootElement.GetProperty("error").GetBoolean();
        return result;
    }

    public static bool CreateMKV(string file, string[] tracks, string[] attachments, string output, string slang, string stitle, bool clean)
    {
        string _tracks = JsonSerializer.Serialize<string[]>(tracks);
        string _attachments = JsonSerializer.Serialize<string[]>(attachments);
        return CreateMKV(cs(file), cs(_tracks), cs(_attachments), cs(output), cs(slang), cs(stitle), clean);
    }

    public static bool ASSFontSubset(string[] files, string fonts, string output, bool dirSafe, Action<string> lcb)
    {
        string _files = JsonSerializer.Serialize<string[]>(files);
        return ASSFontSubset(cs(_files), cs(fonts), cs(output), dirSafe, _lcb(lcb));
    }

    public static string[] QueryFolder(string dir, Action<string> lcb)
    {
        string result = css(QueryFolder(cs(dir), _lcb(lcb)));
        return JsonSerializer.Deserialize<string[]>(result);
    }

    public static bool DumpMKVs(string dir, string output, bool subset, Action<string> lcb)
    {
        return DumpMKVs(cs(dir), cs(output), subset, _lcb(lcb));
    }

    public static bool CreateMKVs(string vDir, string sDir, string fDir, string tDir, string oDir, string slang, string stitle, bool clean, Action<string> lcb)
    {
        return CreateMKVs(cs(vDir), cs(sDir), cs(fDir), cs(tDir), cs(oDir), cs(slang), cs(stitle), clean, _lcb(lcb));
    }

    public static bool MakeMKVs(string dir, string data, string output, string slang, string stitle, Action<string> lcb)
    {
        return MakeMKVs(cs(dir), cs(data), cs(output), cs(slang), cs(stitle), _lcb(lcb));
    }

    public static void A2P(bool a2p, bool apc, string pr, string pf)
    {
        A2P(a2p, apc, cs(pr), cs(pf));
    }

    public static string[] GetFontsList(string dir, Action<string> lcb)
    {
        string result = css(GetFontsList(cs(dir), _lcb(lcb)));
        return JsonSerializer.Deserialize<string[]>(result);
    }

    public static void Cache(string[] ccs)
    {
        string _ccs = JsonSerializer.Serialize<string[]>(ccs);
        Cache(cs(_ccs));
    }

    public static void MKS(bool mks)
    {
        MKS(mks);
    }

    public static void NRename(bool n)
    {
        NRename(n);
    }

    public static void Check(bool check, bool strict)
    {
        Check(check, strict);
    }

    public static string GetFontInfo(string p)
    {
        return css(GetFontInfo(cs(p)));
    }

    public static string[] CreateFontsCache(string dir, string output, Action<string> lcb)
    {
        string result = css(CreateFontsCache(cs(dir), cs(output), _lcb(lcb)));
        return JsonSerializer.Deserialize<string[]>(result);
    }

    public static bool CopyFontsFromCache(string subs, string dist, Action<string> lcb)
    {
        return CopyFontsFromCache(cs(subs), cs(dist), _lcb(lcb));
    }

    delegate void logCallback(IntPtr ptr);
    static logCallback _lcb(Action<string> lcb)
    {
        return (ptr) =>
        {
            if (lcb != null)
                lcb(css(ptr));
        };
    }

    static IntPtr cs(string str)
    {
        return Marshal.StringToCoTaskMemUTF8(str);
    }

    static string css(IntPtr ptr)
    {
        return Marshal.PtrToStringUTF8(ptr);
    }

}