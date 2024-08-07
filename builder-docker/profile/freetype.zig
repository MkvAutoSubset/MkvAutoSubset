const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const lib = b.addStaticLibrary(.{
        .name = "freetype",
        .target = target,
        .optimize = optimize,
    });

    const srcs = &[_][]const u8{
        "src/autofit/autofit.c",
        "src/base/ftbase.c",
        "src/base/ftbbox.c",
        "src/base/ftbdf.c",
        "src/base/ftbitmap.c",
        "src/base/ftcid.c",
        "src/base/ftfstype.c",
        "src/base/ftgasp.c",
        "src/base/ftglyph.c",
        "src/base/ftgxval.c",
        "src/base/ftinit.c",
        "src/base/ftmm.c",
        "src/base/ftotval.c",
        "src/base/ftpatent.c",
        "src/base/ftpfr.c",
        "src/base/ftstroke.c",
        "src/base/ftsynth.c",
        "src/base/fttype1.c",
        "src/base/ftwinfnt.c",
        "src/bdf/bdf.c",
        "src/bzip2/ftbzip2.c",
        "src/cache/ftcache.c",
        "src/cff/cff.c",
        "src/cid/type1cid.c",
        "src/gzip/ftgzip.c",
        "src/lzw/ftlzw.c",
        "src/pcf/pcf.c",
        "src/pfr/pfr.c",
        "src/psaux/psaux.c",
        "src/pshinter/pshinter.c",
        "src/psnames/psnames.c",
        "src/raster/raster.c",
        "src/sdf/sdf.c",
        "src/sfnt/sfnt.c",
        "src/smooth/smooth.c",
        "src/svg/svg.c",
        "src/truetype/truetype.c",
        "src/type1/type1.c",
        "src/type42/type42.c",
        "src/winfonts/winfnt.c",
        "src/base/ftsystem.c",
        "src/base/ftdebug.c",
    };

    const hdrs = &[_][]const u8{
        "freetype.h",
        "ftadvanc.h",
        "ftbbox.h",
        "ftbdf.h",
        "ftbitmap.h",
        "ftbzip2.h",
        "ftcache.h",
        "ftchapters.h",
        "ftcid.h",
        "ftcolor.h",
        "ftdriver.h",
        "fterrdef.h",
        "fterrors.h",
        "ftfntfmt.h",
        "ftgasp.h",
        "ftglyph.h",
        "ftgxval.h",
        "ftgzip.h",
        "ftimage.h",
        "ftincrem.h",
        "ftlcdfil.h",
        "ftlist.h",
        "ftlogging.h",
        "ftlzw.h",
        "ftmac.h",
        "ftmm.h",
        "ftmodapi.h",
        "ftmoderr.h",
        "ftotval.h",
        "ftoutln.h",
        "ftparams.h",
        "ftpfr.h",
        "ftrender.h",
        "ftsizes.h",
        "ftsnames.h",
        "ftstroke.h",
        "ftsynth.h",
        "ftsystem.h",
        "fttrigon.h",
        "fttypes.h",
        "ftwinfnt.h",
        "otsvg.h",
        "t1tables.h",
        "ttnameid.h",
        "tttables.h",
        "tttags.h",
        "config/ftconfig.h",
        "config/ftheader.h",
        "config/ftmodule.h",
        "config/ftoption.h",
        "config/ftstdlib.h",
        "config/integer-types.h",
        "config/mac-support.h",
        "config/public-macros.h",
    };

    for (srcs) |item| {
        lib.addCSourceFile(.{ .file = b.path(item) });
    }

    for (hdrs) |item| {
        lib.installHeader(b.path(b.fmt("include/freetype/{s}", .{item})), b.fmt("freetype/{s}", .{item}));
    }
    lib.installHeader(b.path("include/ft2build.h"), "ft2build.h");

    lib.addIncludePath(b.path("include"));
    lib.defineCMacro("FT2_BUILD_LIBRARY", "1");
    lib.linkLibC();
    b.installArtifact(lib);
}
