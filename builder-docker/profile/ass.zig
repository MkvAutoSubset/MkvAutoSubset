const std = @import("std");

pub fn build(b: *std.Build) !void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const i = b.option([]const u8, "i", "Path to libs include directory");

    const lib = b.addStaticLibrary(.{
        .name = "ass",
        .target = target,
        .optimize = optimize,
    });

    const srcs = &[_][]const u8{
        "libass/ass.c",
        "libass/ass_bitmap.c",
        "libass/ass_bitmap_engine.c",
        "libass/ass_blur.c",
        "libass/ass_cache.c",
        "libass/ass_drawing.c",
        "libass/ass_filesystem.c",
        "libass/ass_font.c",
        "libass/ass_fontselect.c",
        "libass/ass_library.c",
        "libass/ass_outline.c",
        "libass/ass_parse.c",
        "libass/ass_rasterizer.c",
        "libass/ass_render.c",
        "libass/ass_render_api.c",
        "libass/ass_shaper.c",
        "libass/ass_string.c",
        "libass/ass_strtod.c",
        "libass/ass_utils.c",
        "libass/c/c_be_blur.c",
        "libass/c/c_blend_bitmaps.c",
        "libass/c/c_blur.c",
        "libass/c/c_rasterizer.c",
    };

    const libs = &[_][]const u8{
        "freetype",
        "fribidi",
        "harfbuzz",
    };

    for (srcs) |item| {
        lib.addCSourceFile(.{ .file = b.path(item) });
    }

    lib.addIncludePath(b.path(""));
    lib.addIncludePath(b.path("libass"));

    if (i) |val| {
        lib.addIncludePath(b.path(val));
        for (libs) |item| {
            lib.addIncludePath(b.path(b.fmt("{s}/{s}", .{ val, item })));
        }
    }

    lib.installHeader(b.path("libass/ass.h"), "ass/ass.h");
    lib.installHeader(b.path("libass/ass_types.h"), "ass/ass_types.h");

    lib.linkLibC();
    b.installArtifact(lib);
}
