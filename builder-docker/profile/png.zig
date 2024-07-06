const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const i = b.option([]const u8, "i", "Path to zlib include directory");

    const lib = b.addStaticLibrary(.{
        .name = "png",
        .target = target,
        .optimize = optimize,
    });

    const srcs = &[_][]const u8{
        "png.c",
        "pngerror.c",
        "pngget.c",
        "pngmem.c",
        "pngpread.c",
        "pngread.c",
        "pngrio.c",
        "pngrtran.c",
        "pngrutil.c",
        "pngset.c",
        "pngtrans.c",
        "pngwio.c",
        "pngwrite.c",
        "pngwtran.c",
        "pngwutil.c",
    };

    for (srcs) |item| {
        lib.addCSourceFile(.{ .file = b.path(item) });
    }

    if (i) |val| {
        lib.addIncludePath(b.path(val));
    }

    lib.installHeader(b.path("png.h"), "png.h");
    lib.installHeader(b.path("pnglibconf.h"), "pnglibconf.h");
    lib.installHeader(b.path("pngconf.h"), "pngconf.h");
    lib.linkLibC();
    b.installArtifact(lib);
}
