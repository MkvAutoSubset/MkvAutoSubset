const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const lib = b.addStaticLibrary(.{
        .name = "zlib",
        .target = target,
        .optimize = optimize,
    });

    const cflags = &[_][]const u8{
        "-std=c89",
    };
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
        lib.addCSourceFile(.{ .file = b.path(item), .flags = cflags });
    }

    lib.installHeader(b.path("zlib.h"), "zlib.h");
    lib.installHeader(b.path("zconf.h"), "zconf.h");
    lib.linkLibC();
    b.installArtifact(lib);
}
