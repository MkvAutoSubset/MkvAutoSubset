const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const lib = b.addStaticLibrary(.{
        .name = "zlib",
        .target = target,
        .optimize = optimize,
    });

    const srcs = &[_][]const u8{
        "zutil.c",
        "inffast.c",
        "gzlib.c",
        "adler32.c",
        "uncompr.c",
        "infback.c",
        "compress.c",
        "crc32.c",
        "gzclose.c",
        "trees.c",
        "inflate.c",
        "inftrees.c",
        "deflate.c",
        "gzread.c",
        "gzwrite.c",
    };

    for (srcs) |item| {
        lib.addCSourceFile(.{ .file = b.path(item) });
    }

    lib.installHeader(b.path("zlib.h"), "zlib.h");
    lib.installHeader(b.path("zconf.h"), "zconf.h");
    lib.defineCMacro("_LARGEFILE64_SOURCE", "1");
    lib.defineCMacro("HAVE_SYS_TYPES_H", "1");
    lib.defineCMacro("HAVE_STDINT_H", "1");
    lib.defineCMacro("HAVE_STDDEF_H", "1");
    lib.linkLibC();
    b.installArtifact(lib);
}
