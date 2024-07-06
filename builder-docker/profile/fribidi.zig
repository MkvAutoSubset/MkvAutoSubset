const std = @import("std");

pub fn build(b: *std.Build) !void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const lib = b.addStaticLibrary(.{
        .name = "fribidi",
        .target = target,
        .optimize = optimize,
    });

    const srcPrefix = "lib/";

    const srcs = &[_][]const u8{
        "fribidi-arabic.c",
        "fribidi-bidi-types.c",
        "fribidi-bidi.c",
        "fribidi-brackets.c",
        "fribidi-char-sets-cap-rtl.c",
        "fribidi-char-sets-cp1255.c",
        "fribidi-char-sets-cp1256.c",
        "fribidi-char-sets-iso8859-6.c",
        "fribidi-char-sets-iso8859-8.c",
        "fribidi-char-sets-utf8.c",
        "fribidi-char-sets.c",
        "fribidi-deprecated.c",
        "fribidi-joining-types.c",
        "fribidi-joining.c",
        "fribidi-mirroring.c",
        "fribidi-run.c",
        "fribidi-shape.c",
        "fribidi.c",
    };

    const cflags = &[_][]const u8{
        "-std=c89",
    };
    for (srcs) |item| {
        var path = std.ArrayList(u8).init(b.allocator);
        defer path.deinit();
        try path.appendSlice(srcPrefix);
        try path.appendSlice(item);
        lib.addCSourceFile(.{ .file = b.path(path.items), .flags = cflags });
    }

    const hdrs = &[_][]const u8{
        "fribidi-arabic.h",
        "fribidi-begindecls.h",
        "fribidi-bidi-types-list.h",
        "fribidi-bidi-types.h",
        "fribidi-bidi.h",
        "fribidi-brackets.h",
        "fribidi-char-sets-list.h",
        "fribidi-char-sets.h",
        "fribidi-common.h",
        "fribidi-config.h",
        "fribidi-deprecated.h",
        "fribidi-enddecls.h",
        "fribidi-flags.h",
        "fribidi-joining-types-list.h",
        "fribidi-joining-types.h",
        "fribidi-joining.h",
        "fribidi-mirroring.h",
        "fribidi-shape.h",
        "fribidi-types.h",
        "fribidi-unicode-version.h",
        "fribidi-unicode.h",
        "fribidi.h",
    };

    for (hdrs) |item| {
        var path1 = std.ArrayList(u8).init(b.allocator);
        defer path1.deinit();
        try path1.appendSlice(srcPrefix);
        try path1.appendSlice(item);

        var path2 = std.ArrayList(u8).init(b.allocator);
        defer path2.deinit();
        try path2.appendSlice("fribidi/");
        try path2.appendSlice(item);

        lib.installHeader(b.path(path1.items), path2.items);
    }

    lib.addIncludePath(b.path(srcPrefix));
    lib.defineCMacro("HAVE_STRINGIZE", "1");
    lib.linkLibC();
    b.installArtifact(lib);
}
