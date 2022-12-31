using System;
using System.Runtime.InteropServices;
using System.Text.Json;

public static class mkvlib
{

    #region imports

    const string so = "mkvlib.so";

    [DllImport(so, EntryPoint = "Version")]
    static extern IntPtr _Version();

    [DllImport(so)]
    static extern bool InitInstance(logCallback lcb);

    [DllImport(so)]
    static extern IntPtr GetMKVInfo(IntPtr ptr);

    [DllImport(so)]
    static extern bool DumpMKV(IntPtr file, IntPtr output, bool subset, logCallback lcb);

    [DllImport(so)]
    static extern IntPtr CheckSubset(IntPtr file, logCallback lcb);

    [DllImport(so)]
    static extern bool CreateMKV(IntPtr file, IntPtr tracks, IntPtr attachments, IntPtr output, IntPtr slang, IntPtr stitle, bool clean);

    [DllImport(so)]
    static extern bool ASSFontSubset(IntPtr files, IntPtr fonts, IntPtr output, bool dirSafe, logCallback lcb);

    [DllImport(so)]
    static extern IntPtr QueryFolder(IntPtr dir, logCallback lcb);

    [DllImport(so)]
    static extern bool DumpMKVs(IntPtr dir, IntPtr output, bool subset, logCallback lcb);

    [DllImport(so)]
    static extern bool CreateMKVs(IntPtr vDir, IntPtr sDir, IntPtr fDir, IntPtr tDir, IntPtr oDir, IntPtr slang, IntPtr stitle, bool clean, logCallback lcb);

    [DllImport(so)]
    static extern bool MakeMKVs(IntPtr dir, IntPtr data, IntPtr output, IntPtr slang, IntPtr stitle, bool subset, logCallback lcb);

    [DllImport(so)]
    static extern bool CreateBlankOrBurnVideo(long t, IntPtr s, IntPtr enc, IntPtr ass, IntPtr fontdir, IntPtr output);

    [DllImport(so)]
    static extern bool CreateTestVideo(IntPtr asses, IntPtr s, IntPtr fontdir, IntPtr enc, bool burn, logCallback lcb);

    [DllImport(so)]
    static extern void A2P(bool a2p, bool apc, IntPtr pr, IntPtr pf);

    [DllImport(so)]
    static extern IntPtr GetFontsList(IntPtr files, IntPtr fonts, logCallback lcb);

    [DllImport(so)]
    static extern void Cache(IntPtr ccs);

    [DllImport(so, EntryPoint = "MKS")]
    static extern void _MKS(bool mks);

    [DllImport(so, EntryPoint = "NRename")]
    static extern void _NRename(bool n);

    [DllImport(so, EntryPoint = "NOverwrite")]
    static extern void _NOverwrite(bool n);

    [DllImport(so, EntryPoint = "Check")]
    static extern void _Check(bool check, bool strict);

    [DllImport(so)]
    static extern IntPtr CreateFontsCache(IntPtr dir, IntPtr output, logCallback lcb);

    [DllImport(so)]
    static extern bool CopyFontsFromCache(IntPtr asses, IntPtr dist, logCallback lcb);

    [DllImport(so)]
    static extern IntPtr GetFontInfo(IntPtr p);

    #endregion

    public static string Version()
    {
        return css(_Version());
    }

    public static bool InitInstance(Action<LogLevel, string> lcb)
    {
        return InitInstance(_lcb(lcb));
    }

    public static string GetMKVInfo(string file)
    {
        return css(GetMKVInfo(cs(file)));
    }

    public static bool DumpMKV(string file, string output, bool subset, Action<LogLevel, string> lcb)
    {
        return DumpMKV(cs(file), cs(output), subset, _lcb(lcb));
    }

    public static bool[] CheckSubset(string file, Action<LogLevel, string> lcb)
    {
        string json = css(CheckSubset(cs(file), _lcb(lcb)));
        return JsonSerializer.Deserialize<bool[]>(json);
    }

    public static bool CreateMKV(string file, string[] tracks, string[] attachments, string output, string slang, string stitle, bool clean)
    {
        string _tracks = JsonSerializer.Serialize(tracks);
        string _attachments = JsonSerializer.Serialize(attachments);
        return CreateMKV(cs(file), cs(_tracks), cs(_attachments), cs(output), cs(slang), cs(stitle), clean);
    }

    public static bool ASSFontSubset(string[] files, string fonts, string output, bool dirSafe, Action<LogLevel, string> lcb)
    {
        string _files = JsonSerializer.Serialize(files);
        return ASSFontSubset(cs(_files), cs(fonts), cs(output), dirSafe, _lcb(lcb));
    }

    public static string[] QueryFolder(string dir, Action<LogLevel, string> lcb)
    {
        string result = css(QueryFolder(cs(dir), _lcb(lcb)));
        return JsonSerializer.Deserialize<string[]>(result);
    }

    public static bool DumpMKVs(string dir, string output, bool subset, Action<LogLevel, string> lcb)
    {
        return DumpMKVs(cs(dir), cs(output), subset, _lcb(lcb));
    }

    public static bool CreateMKVs(string vDir, string sDir, string fDir, string tDir, string oDir, string slang, string stitle, bool clean, Action<LogLevel, string> lcb)
    {
        return CreateMKVs(cs(vDir), cs(sDir), cs(fDir), cs(tDir), cs(oDir), cs(slang), cs(stitle), clean, _lcb(lcb));
    }

    public static bool MakeMKVs(string dir, string data, string output, string slang, string stitle, bool subset, Action<LogLevel, string> lcb)
    {
        return MakeMKVs(cs(dir), cs(data), cs(output), cs(slang), cs(stitle), subset, _lcb(lcb));
    }

    public static bool CreateBlankOrBurnVideo(long t, string s, string enc, string ass, string fontdir, string output)
    {
        return CreateBlankOrBurnVideo(t, cs(s), cs(enc), cs(ass), cs(fontdir), cs(output));
    }

    public static bool CreateTestVideo(string[] asses, string s, string fontdir, string enc, bool burn, Action<LogLevel, string> lcb)
    {
        string _asses = JsonSerializer.Serialize(asses);
        return CreateTestVideo(cs(_asses), cs(s), cs(fontdir), cs(enc), burn, _lcb(lcb));
    }

    public static void A2P(bool a2p, bool apc, string pr, string pf)
    {
        A2P(a2p, apc, cs(pr), cs(pf));
    }

    public static string[][] GetFontsList(string[] files, string fonts, Action<LogLevel, string> lcb)
    {
        string _files = JsonSerializer.Serialize(files);
        string result = css(GetFontsList(cs(_files), cs(fonts), _lcb(lcb)));
        return JsonSerializer.Deserialize<string[][]>(result);
    }

    public static void Cache(string[] ccs)
    {
        string _ccs = JsonSerializer.Serialize(ccs);
        Cache(cs(_ccs));
    }

    public static void MKS(bool mks)
    {
        _MKS(mks);
    }

    public static void NRename(bool n)
    {
        _NRename(n);
    }

    public static void NOverwrite(bool n)
    {
        _NOverwrite(n);
    }

    public static void Check(bool check, bool strict)
    {
        _Check(check, strict);
    }

    public static string GetFontInfo(string p)
    {
        return css(GetFontInfo(cs(p)));
    }

    public static string[] CreateFontsCache(string dir, string output, Action<LogLevel, string> lcb)
    {
        string result = css(CreateFontsCache(cs(dir), cs(output), _lcb(lcb)));
        return JsonSerializer.Deserialize<string[]>(result);
    }

    public static bool CopyFontsFromCache(string[] asses, string dist, Action<LogLevel, string> lcb)
    {
        string _files = JsonSerializer.Serialize(asses);
        return CopyFontsFromCache(cs(_files), cs(dist), _lcb(lcb));
    }

    delegate void logCallback(byte l, IntPtr ptr);

    public enum LogLevel
    {
        Info,
        Warning,
        SWarning,
        Error,
        Progress
    }

    static logCallback _lcb(Action<LogLevel, string> lcb)
    {
        return (l, ptr) =>
        {
            if (lcb != null)
                lcb((LogLevel)l, css(ptr));
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