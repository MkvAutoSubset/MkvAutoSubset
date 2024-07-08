const std = @import("std");

pub fn build(b: *std.Build) !void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const lib = b.addStaticLibrary(.{
        .name = "fribidi",
        .target = target,
        .optimize = optimize,
    });

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

    const _fmt = "lib/{s}";

    for (srcs) |item| {
        lib.addCSourceFile(.{ .file = b.path(b.fmt(_fmt, .{item})) });
    }
    for (hdrs) |item| {
        lib.installHeader(b.path(b.fmt(_fmt, .{item})), b.fmt("fribidi/{s}", .{item}));
    }

    lib.addIncludePath(b.path("lib"));
    lib.defineCMacro("HAVE_MEMMOVE", "1");
    lib.defineCMacro("HAVE_MEMORY_H", "1");
    lib.defineCMacro("HAVE_MEMSET", "1");
    lib.defineCMacro("HAVE_STDLIB_H", "1");
    lib.defineCMacro("HAVE_STRDUP", "1");
    lib.defineCMacro("HAVE_STRINGIZE", "1");
    lib.defineCMacro("HAVE_STRINGS_H", "1");
    lib.defineCMacro("HAVE_STRING_H", "1");
    lib.defineCMacro("HAVE_SYS_TIMES_H", "1");
    lib.defineCMacro("STDC_HEADERS", "1");
    lib.linkLibC();
    b.installArtifact(lib);
}
