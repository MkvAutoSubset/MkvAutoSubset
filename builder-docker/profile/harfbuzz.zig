const std = @import("std");

pub fn build(b: *std.Build) !void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const lib1 = b.addStaticLibrary(.{
        .name = "harfbuzz",
        .target = target,
        .optimize = optimize,
    });
    const lib2 = b.addStaticLibrary(.{
        .name = "harfbuzz-subset",
        .target = target,
        .optimize = optimize,
    });

    const srcPrefix = "src/";

    const srcs1 = &[_][]const u8{
        "OT/Var/VARC/VARC.cc",
        "hb-aat-layout.cc",
        "hb-aat-map.cc",
        "hb-blob.cc",
        "hb-buffer-serialize.cc",
        "hb-buffer-verify.cc",
        "hb-buffer.cc",
        "hb-common.cc",
        "hb-draw.cc",
        "hb-face-builder.cc",
        "hb-face.cc",
        "hb-fallback-shape.cc",
        "hb-font.cc",
        "hb-map.cc",
        "hb-number.cc",
        "hb-ot-cff1-table.cc",
        "hb-ot-cff2-table.cc",
        "hb-ot-color.cc",
        "hb-ot-face.cc",
        "hb-ot-font.cc",
        "hb-ot-layout.cc",
        "hb-ot-map.cc",
        "hb-ot-math.cc",
        "hb-ot-meta.cc",
        "hb-ot-metrics.cc",
        "hb-ot-name.cc",
        "hb-ot-shape-fallback.cc",
        "hb-ot-shape-normalize.cc",
        "hb-ot-shape.cc",
        "hb-ot-shaper-arabic.cc",
        "hb-ot-shaper-default.cc",
        "hb-ot-shaper-hangul.cc",
        "hb-ot-shaper-hebrew.cc",
        "hb-ot-shaper-indic-table.cc",
        "hb-ot-shaper-indic.cc",
        "hb-ot-shaper-khmer.cc",
        "hb-ot-shaper-myanmar.cc",
        "hb-ot-shaper-syllabic.cc",
        "hb-ot-shaper-thai.cc",
        "hb-ot-shaper-use.cc",
        "hb-ot-shaper-vowel-constraints.cc",
        "hb-ot-tag.cc",
        "hb-ot-var.cc",
        "hb-outline.cc",
        "hb-paint-extents.cc",
        "hb-paint.cc",
        "hb-set.cc",
        "hb-shape-plan.cc",
        "hb-shape.cc",
        "hb-shaper.cc",
        "hb-static.cc",
        "hb-style.cc",
        "hb-ucd.cc",
        "hb-unicode.cc",
    };
    const srcs2 = &[_][]const u8{
        "graph/gsubgpos-context.cc",
        "hb-number.cc",
        "hb-ot-cff1-table.cc",
        "hb-ot-cff2-table.cc",
        "hb-static.cc",
        "hb-subset-cff-common.cc",
        "hb-subset-cff1.cc",
        "hb-subset-cff2.cc",
        "hb-subset-input.cc",
        "hb-subset-instancer-iup.cc",
        "hb-subset-instancer-solver.cc",
        "hb-subset-plan.cc",
        "hb-subset-repacker.cc",
        "hb-subset.cc",
    };

    for (srcs1) |item| {
        var path = std.ArrayList(u8).init(b.allocator);
        defer path.deinit();
        try path.appendSlice(srcPrefix);
        try path.appendSlice(item);
        lib1.addCSourceFile(.{ .file = b.path(path.items) });
    }
    for (srcs2) |item| {
        var path = std.ArrayList(u8).init(b.allocator);
        defer path.deinit();
        try path.appendSlice(srcPrefix);
        try path.appendSlice(item);
        lib2.addCSourceFile(.{ .file = b.path(path.items) });
    }
    const hdrs = &[_][]const u8{
        "hb-aat-layout.h",
        "hb-aat.h",
        "hb-blob.h",
        "hb-buffer.h",
        "hb-common.h",
        "hb-cplusplus.hh",
        "hb-deprecated.h",
        "hb-draw.h",
        "hb-face.h",
        "hb-font.h",
        "hb-ft.h",
        "hb-map.h",
        "hb-ot-color.h",
        "hb-ot-deprecated.h",
        "hb-ot-font.h",
        "hb-ot-layout.h",
        "hb-ot-math.h",
        "hb-ot-meta.h",
        "hb-ot-metrics.h",
        "hb-ot-name.h",
        "hb-ot-shape.h",
        "hb-ot-var.h",
        "hb-ot.h",
        "hb-paint.h",
        "hb-set.h",
        "hb-shape-plan.h",
        "hb-shape.h",
        "hb-style.h",
        "hb-subset-repacker.h",
        "hb-subset.h",
        "hb-unicode.h",
        "hb-version.h",
        "hb.h",
    };

    const defs = [_][]const u8{
        "HAVE_ATEXIT",
        "HAVE_GETPAGESIZE",
        "HAVE_ISATTY",
        "HAVE_MPROTECT",
        "HAVE_NEWLOCALE",
        "HAVE_STDBOOL_H",
        "HAVE_SYSCONF",
        "HAVE_UNISTD_H",
        "HB_EXPERIMENTAL_API",
    };

    for (hdrs) |item| {
        var path1 = std.ArrayList(u8).init(b.allocator);
        defer path1.deinit();
        try path1.appendSlice(srcPrefix);
        try path1.appendSlice(item);

        var path2 = std.ArrayList(u8).init(b.allocator);
        defer path2.deinit();
        try path2.appendSlice("harfbuzz/");
        try path2.appendSlice(item);

        lib1.installHeader(b.path(path1.items), path2.items);
    }

    lib1.linkLibCpp();
    b.installArtifact(lib1);

    for (defs) |item| {
        lib1.defineCMacro(item, "1");
        lib2.defineCMacro(item, "1");
    }

    lib2.linkLibCpp();
    b.installArtifact(lib2);
}
