/*----------------------------------------------------------------------------
 * ass2bdnxml - Generates BluRay subtitle stuff from ass/ssa subtitles
 * based on avs2bdnxml 2.08
 * Copyright (C) 2008-2013 Arne Bochem <avs2bdnxml at ps-auxw de>
 * Copyright (C) 2022-2022 Masaiki <mydarer@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *----------------------------------------------------------------------------*/

#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <limits.h>
#include <time.h>
#include <png.h>
#include <getopt.h>
#include <assert.h>
#include <stdbool.h>
#include "auto_split.h"
#include "palletize.h"
#include "sup.h"
#include "abstract_lists.h"

#include <ass/ass.h>

#define MAX_PATH 1024
#define max(a,b) ((a)>(b)?(a):(b))

static char *read_file_bytes(FILE *fp, size_t *bufsize)
{
    int res;
    long sz;
    long bytes_read;
    char *buf;
    res = fseek(fp, 0, SEEK_END);
    if (res == -1) {
        fclose(fp);
        return 0;
    }

    sz = ftell(fp);
    rewind(fp);

    buf = sz < SIZE_MAX ? malloc(sz + 1) : NULL;
    if (!buf) {
        fclose(fp);
        return NULL;
    }
    bytes_read = 0;
    do {
        res = fread(buf + bytes_read, 1, sz - bytes_read, fp);
        if (res <= 0) {
            fclose(fp);
            free(buf);
            return 0;
        }
        bytes_read += res;
    } while (sz - bytes_read > 0);
    buf[sz] = '\0';
    fclose(fp);

    if (bufsize)
        *bufsize = sz;
    return buf;
}

static const char *detect_bom(const char *buf, const size_t bufsize) {
    if (bufsize >= 4) {
        if (!strncmp(buf, "\xef\xbb\xbf", 3))
            return "UTF-8";
        if (!strncmp(buf, "\x00\x00\xfe\xff", 4))
            return "UTF-32BE";
        if (!strncmp(buf, "\xff\xfe\x00\x00", 4))
            return "UTF-32LE";
        if (!strncmp(buf, "\xfe\xff", 2))
            return "UTF-16BE";
        if (!strncmp(buf, "\xff\xfe", 2))
            return "UTF-16LE";
    }
    return "UTF-8";
}

#if defined(_WIN32)
char *rindex(char *s, int c) {
    int len = strlen(s);
    for (int i = len - 1; i > -1; --i)
        if (s[i] == c) return s + i;
    return NULL;
}
#else
char *_fullpath(char *absPath, const char *relPath, size_t maxLength) {
    return realpath(relPath, absPath);
}
#include <libgen.h>
void _splitpath(char *path, char *drive, char *dir, char *fname, char *ext) {
    if (drive) drive[0] = 0;
    if (dir) {
        char *tmp = dirname(path);
        strcpy(dir, tmp);
    }
}
#include <errno.h>
#endif

typedef struct {
    ASS_Renderer *ass_renderer;
    ASS_Library *ass_library;
    ASS_Track *ass;
} ass_input_t;

typedef struct {
    int i_width;
    int i_height;
    int i_fps_den;
    int i_fps_num;
} stream_info_t;

void msg_callback(int level, const char *fmt, va_list va, void *data)
{
    if (level > (intptr_t)data)
        return;
}

int open_file_ass( char *psz_filename, ass_input_t **p_handle, stream_info_t *p_param)
{
    *p_handle = malloc(sizeof(ass_input_t));
    memset(*p_handle, 0, sizeof(ass_input_t));

    ASS_Renderer *ass_renderer;
    ASS_Library *ass_library = ass_library_init();

    if (!ass_library)
        return 1;

    ass_set_message_cb(ass_library, msg_callback, (void *)(intptr_t)0);

    ass_renderer = ass_renderer_init(ass_library);

    if (!ass_renderer)
        return 1;

    FILE *fp = fopen(psz_filename, "rb");
    if (!fp)
        return 1;
    size_t bufsize;
    char *buf = read_file_bytes(fp, &bufsize);
    const char* cs = detect_bom(buf, bufsize);
    ASS_Track *ass = ass_read_memory(ass_library, buf, bufsize, (char *)cs);

    if (!ass)
        return 1;

    (*p_handle)->ass_library = ass_library;
    (*p_handle)->ass_renderer = ass_renderer;
    (*p_handle)->ass = ass;

    p_param->i_height = ass->PlayResY;
    p_param->i_width = ass->PlayResX;

    return 0;
}

int get_frame_total_ass( ass_input_t *handle, stream_info_t *p_param)
{
    long long max_time_point = -1;
    for (int i = 0; i < handle->ass->n_events; ++i) {
        ASS_Event *evt = handle->ass->events + i;
        max_time_point = max(max_time_point, evt->Start + evt->Duration);
    }
    return (long double)max_time_point / p_param->i_fps_den * p_param->i_fps_num / 1000;
}

int close_file_ass( ass_input_t *handle )
{
    ass_renderer_done(handle->ass_renderer);
    ass_library_done(handle->ass_library);
    ass_free_track(handle->ass);
}

bool get_dir_path(char *filename, char *dir_path)
{
    char abs_path[MAX_PATH + 1] = {0};
    char drive[3] = {0};
    char dir[MAX_PATH + 1] = {0};

    /* Get absolute path of output XML file */
    if (_fullpath(abs_path, filename, MAX_PATH) == NULL)
    {
        return false;
    }

    /* Split absolute path into components */
    _splitpath(abs_path, drive, dir, NULL, NULL);
    strncpy(dir_path, drive, 2);
    strncat(dir_path, dir, MAX_PATH - 2);

    if (strlen(dir_path) > MAX_PATH - 16)
    {
        return false;
    }

    return true;
}

bool write_png(char *dir, int file_id, uint8_t *image, int w, int h, int graphic, uint32_t *pal, crop_t c)
{
    FILE *fh;
    png_structp png_ptr;
    png_infop info_ptr;
    png_bytep *row_pointers;
    png_colorp palette = NULL;
    png_bytep trans = NULL;
    char tmp[16] = {0};
    char filename[MAX_PATH + 1] = {0};
    char *col;
    int step = pal == NULL ? 4 : 1;
    int colors = 0;
    int i;

    snprintf(tmp, 15, "%08d_%d.png", file_id, graphic);
    strncpy(filename, dir, MAX_PATH);
    strncat(filename, tmp, 15);

    if ((fh = fopen(filename, "wb")) == NULL)
    {
        return false;
    }

    /* Initialize png struct */
    png_ptr = png_create_write_struct(PNG_LIBPNG_VER_STRING, NULL, NULL, NULL);
    if (png_ptr == NULL)
    {
        return false;
    }

    /* Initialize info struct */
    info_ptr = png_create_info_struct(png_ptr);
    if (info_ptr == NULL)
    {
        png_destroy_write_struct(&png_ptr, (png_infopp)NULL);
        return false;
    }

    /* Set long jump stuff (weird..?) */
    if (setjmp(png_jmpbuf(png_ptr)))
    {
        png_destroy_write_struct(&png_ptr, &info_ptr);
        fclose(fh);
        return false;
    }

    /* Initialize IO */
    png_init_io(png_ptr, fh);

    /* Set file info */
    if (pal == NULL)
        png_set_IHDR(png_ptr, info_ptr, c.w, c.h, 8, PNG_COLOR_TYPE_RGB_ALPHA, PNG_INTERLACE_NONE, PNG_COMPRESSION_TYPE_DEFAULT, PNG_FILTER_TYPE_DEFAULT);
    else
    {
        png_set_IHDR(png_ptr, info_ptr, c.w, c.h, 8, PNG_COLOR_TYPE_PALETTE, PNG_INTERLACE_NONE, PNG_COMPRESSION_TYPE_DEFAULT, PNG_FILTER_TYPE_DEFAULT);
        palette = calloc(256, sizeof(png_color));
        trans = calloc(256, sizeof(png_byte));
        colors = 1;
        for (i = 1; i < 256 && pal[i]; i++)
        {
            col = (char *)&(pal[i]);
            palette[i].red = col[0];
            palette[i].green = col[1];
            palette[i].blue = col[2];
            trans[i] = col[3];
            colors++;
        }
        png_set_PLTE(png_ptr, info_ptr, palette, colors);
        png_set_tRNS(png_ptr, info_ptr, trans, colors, NULL);
    }

    /* Allocate row pointer memory */
    row_pointers = calloc(c.h, sizeof(png_bytep));

    /* Set row pointers */
    image = image + step * (c.x + w * c.y);
    for (i = 0; i < c.h; i++)
    {
        row_pointers[i] = image + i * w * step;
    }
    png_set_rows(png_ptr, info_ptr, row_pointers);

    /* Set compression */
    png_set_filter(png_ptr, 0, PNG_FILTER_VALUE_SUB);
    png_set_compression_level(png_ptr, 5);

    /* Write image */
    png_write_png(png_ptr, info_ptr, PNG_TRANSFORM_IDENTITY, NULL);

    /* Free memory */
    png_destroy_write_struct(&png_ptr, &info_ptr);
    free(row_pointers);
    if (palette != NULL)
        free(palette);
    if (trans != NULL)
        free(trans);

    /* Close file handle */
    fclose(fh);

    return true;
}

//extern int asm_is_identical_sse2 (stream_info_t *s_info, char *img, char *img_old);
//extern int asm_is_empty_sse2 (stream_info_t *s_info, char *img);
//extern void asm_zero_transparent_sse2 (stream_info_t *s_info, char volatile *img);
//extern void asm_swap_rb_sse2 (stream_info_t *s_info, char volatile *img, char volatile *out);

int is_identical (stream_info_t *s_info, char *img, char *img_old)
{
    uint32_t *max = (uint32_t *)(img + s_info->i_width * s_info->i_height * 4);
    uint32_t *im = (uint32_t *)img;
    uint32_t *im_old = (uint32_t *)img_old;

    while (im < max)
    {
        if (!((char *)im)[3])
            *im = 0;
        if (*(im++) ^ *(im_old++))
            return 0;
    }

    return 1;
}

int is_empty (stream_info_t *s_info, char *img)
{
    char *max = img + s_info->i_width * s_info->i_height * 4;
    char *im = img;

    while (im < max)
    {
        if (im[3])
            return 0;
        im += 4;
    }

    return 1;
}

void zero_transparent (stream_info_t *s_info, char *img)
{
    char *max = img + s_info->i_width * s_info->i_height * 4;
    char *im = img;

    while (im < max)
    {
        if (!im[3])
            *(uint32_t *)img = 0;
        im += 4;
    }
}

void swap_rb (stream_info_t *s_info, char *img, char *out)
{
    char *max = img + s_info->i_width * s_info->i_height * 4;

    while (img < max)
    {
        out[0] = img[2];
        out[1] = img[1];
        out[2] = img[0];
        out[3] = img[3];
        img += 4;
        out += 4;
    }
}

//int detect_sse2 ()
//{
//	static int detection = -1;
//	unsigned int func = 0x00000001;
//	unsigned int eax, ebx, ecx, edx;
//
//	if (detection != -1)
//		return detection;
//
//	asm volatile
//	(
//		"cpuid\n"
//		: "=a" (eax), "=b" (ebx), "=c" (ecx), "=d" (edx)
//		: "a" (func)
//	);
//
//	/* SSE2:  edx & 0x04000000
//	 * SSSE3: ecx & 0x00000200
//	 */
//	detection = (edx & 0x04000000) ? 1 : 0;
//
//	if (detection)
//		fprintf(stderr, "CPU: Using SSE2 optimized functions.\n");
//	else
//		fprintf(stderr, "CPU: Using pure C functions.\n");
//
//	return detection;
//}

//int is_identical (stream_info_t *s_info, char *img, char *img_old)
//{
//	if (detect_sse2())
//		return asm_is_identical_sse2(s_info, img, img_old);
//	else
//		return is_identical_c(s_info, img, img_old);
//}
//
//int is_empty (stream_info_t *s_info, char *img)
//{
//	if (detect_sse2())
//		return asm_is_empty_sse2(s_info, img);
//	else
//		return is_empty_c(s_info, img);
//}
//
//void zero_transparent (stream_info_t *s_info, char *img)
//{
//	if (detect_sse2())
//		return asm_zero_transparent_sse2(s_info, img);
//	else
//		return zero_transparent_c(s_info, img);
//}
//
//void swap_rb (stream_info_t *s_info, char *img, char *out)
//{
//	if (detect_sse2())
//		return asm_swap_rb_sse2(s_info, img, out);
//	else
//		return swap_rb_c(s_info, img, out);
//}

/* SMPTE non-drop time code */
bool mk_timecode (int frame, int fps, char *buf) /* buf must have length 12 (incl. trailing \0) */
{
    int frames, s, m, h;
    int tc = frame;

    tc = frame;
    frames = tc % fps;
    tc /= fps;
    s = tc % 60;
    tc /= 60;
    m = tc % 60;
    tc /= 60;
    h = tc;

    if (h > 99)
    {
        return false;
    }

    if (snprintf(buf, 12, "%02d:%02d:%02d:%02d", h, m, s, frames) != 11)
    {
        return false;
    }
}

void print_usage ()
{
}

int is_extension(char *filename, char *check_ext)
{
    char *ext = rindex(filename, '.');

    if (ext == NULL)
        return 0;

    ext++;
    if (!strcasecmp(ext, check_ext))
        return 1;

    return 0;
}

int parse_int(char *in, char *name, int *error)
{
    char *end;
    int r;
    errno = 0;
    if (error != NULL)
        *error = 0;
    r = strtol(in, &end, 10);
    if (errno || end == in || end != in + strlen(in))
    {
        if (error != NULL)
            *error = 1;
        if (name != NULL)
        {
            return false;
        }
    }
    return r;
}

int parse_tc(char *in, int fps)
{
    int r = 0;
    int e;
    int h, m, s, f;

    /* Test for raw frame number. */
    r = parse_int(in, NULL, &e);
    if (!e)
        return r;

    if (strlen(in) != 2 * 4 + 3 || in[2] != ':' || in[5] != ':' || in[8] != ':')
    {
        return false;
    }
    in[2] = 0;
    in[5] = 0;
    in[8] = 0;
    h = parse_int(in,     "t-offset hours",   NULL);
    m = parse_int(in + 3, "t-offset minutes", NULL);
    s = parse_int(in + 6, "t-offset seconds", NULL);
    f = parse_int(in + 9, "t-offset frames",  NULL);
    r = f;
    r += s * fps;
    fps *= 60;
    r += m * fps;
    fps *= 60;
    r += h * fps;
    return r;
}

typedef struct event_s
{
    int image_number;
    int start_frame;
    int end_frame;
    int graphics;
    crop_t c[2];
} event_t;

STATIC_LIST(event, event_t)

void add_event_xml_real (event_list_t *events, int image, int start, int end, int graphics, crop_t *crops)
{
    event_t *new = calloc(1, sizeof(event_t));
    new->image_number = image;
    new->start_frame = start;
    new->end_frame = end;
    new->graphics = graphics;
    new->c[0] = crops[0];
    new->c[1] = crops[1];
    event_list_insert_after(events, new);
}

void add_event_xml (event_list_t *events, int split_at, int min_split, int start, int end, int graphics, crop_t *crops)
{
    int image = start;
    int d = end - start;

    if (!split_at)
        add_event_xml_real(events, image, start, end, graphics, crops);
    else
    {
        while (d >= split_at + min_split)
        {
            d -= split_at;
            add_event_xml_real(events, image, start, start + split_at, graphics, crops);
            start += split_at;
        }
        if (d)
            add_event_xml_real(events, image, start, start + d, graphics, crops);
    }
}

void write_sup_wrapper (sup_writer_t *sw, uint8_t *im, int num_crop, crop_t *crops, uint32_t *pal, int start, int end, int split_at, int min_split, int stricter)
{
    int d = end - start;

    if (!split_at)
        write_sup(sw, im, num_crop, crops, pal, start, end, stricter);
    else
    {
        while (d >= split_at + min_split)
        {
            d -= split_at;
            write_sup(sw, im, num_crop, crops, pal, start, start + split_at, stricter);
            start += split_at;
        }
        if (d)
            write_sup(sw, im, num_crop, crops, pal, start, start + d, stricter);
    }
}

struct framerate_entry_s
{
    char *name;
    char *out_name;
    int rate;
    int drop;
    int fps_num;
    int fps_den;
};

// codes from assrender, modified

#define _r(c) (( (c) >> 24))
#define _g(c) ((((c) >> 16) & 0xFF))
#define _b(c) ((((c) >> 8)  & 0xFF))
#define _a(c) (( (c)        & 0xFF))

#define div256(x)   (((x + 128)   >> 8))
#define div255(x)   ((div256(x + div256(x))))

#define scale(srcA, srcC, dstC) \
	((srcA * srcC + (255 - srcA) * dstC))
#define dblend(srcA, srcC, dstA, dstC, outA) \
	(((srcA * srcC * 255 + dstA * dstC * (255 - srcA) + (outA >> 1)) / outA))

void col2rgb(uint32_t *c, uint8_t *r, uint8_t *g, uint8_t *b)
{
    *r = _r(*c);
    *g = _g(*c);
    *b = _b(*c);
}

void make_sub_img(ASS_Image *img, uint8_t *sub_img, uint32_t width)
{
    uint8_t c1, c2, c3, a, a1;
    uint8_t *src;
    uint8_t *dstC1, *dstC2, *dstC3, *dstA, *dst;
    uint32_t dsta;

    while (img) {
        if (img->w == 0 || img->h == 0) {
            // nothing to render
            img = img->next;
            continue;
        }
        col2rgb(&img->color, &c1, &c2, &c3);
        a1 = 255 - _a(img->color); // transparency

        src = img->bitmap;
        dst = sub_img + (img->dst_y * width + img->dst_x)*4;
        dstC1 = dst+2;
        dstC2 = dst+1;
        dstC3 = dst+0;
        dstA = dst+3;

        for (int i = 0; i < img->h; i++) {
            for (int j = 0; j < img->w*4; j+=4) {
                a = div255(src[j/4] * a1);
                if (a) {
                    if (dstA[j]) {
                        dsta = scale(a, 255, dstA[j]);
                        dstC1[j] = dblend(a, c1, dstA[j], dstC1[j], dsta);
                        dstC2[j] = dblend(a, c2, dstA[j], dstC2[j], dsta);
                        dstC3[j] = dblend(a, c3, dstA[j], dstC3[j], dsta);
                        dstA[j] = div255(dsta);
                    }
                    else {
                        dstC1[j] = c1;
                        dstC2[j] = c2;
                        dstC3[j] = c3;
                        dstA[j] = a;
                    }
                }
            }

            src += img->stride;
            dstC1 += width * 4;
            dstC2 += width * 4;
            dstC3 += width * 4;
            dstA += width * 4;
        }

        img = img->next;
    }
}

// codes from assrender end here

int app (int argc, char *argv[])
{
    detect_endianness();
    struct framerate_entry_s framerates[] = { {"23.976", "23.976", 24, 0, 24000, 1001}
        /*, {"23.976d", "23.976", 24000/1001.0, 1}*/
        , {"24", "24", 24, 0, 24, 1}
        , {"25", "25", 25, 0, 25, 1}
        , {"29.97", "29.97", 30, 0, 30000, 1001}
        , {"30", "30", 30, 0, 30, 1}
        /*, {"29.97d", "29.97", 30000/1001.0, 1}*/
        , {"50", "50", 50, 0, 50, 1}
        , {"59.94", "59.94", 60, 0, 60000, 1001}
        , {"60", "60", 60, 0, 60, 1}
        /*, {"59.94d", "59.94", 60000/1001.0, 1}*/
        , {NULL, NULL, 0, 0, 0, 0}
    };
    char *ass_filename = NULL;
    char *track_name = "Undefined";
    char *language = "und";
    char *video_format = "1080p";
    char *frame_rate = "23.976";
    char *out_filename[2] = {NULL, NULL};
    char *sup_output_fn = NULL;
    char *xml_output_fn = NULL;
    char *x_offset = "0";
    char *y_offset = "0";
    char *t_offset = "0";
    char *buffer_optimize = "0";
    char *split_after = "0";
    char *minimum_split = "3";
    char *palletize_png = "1";
    char *even_y_string = "0";
    char *auto_crop_image = "1";
    char *ugly_option = "0";
    char *seek_string = "0";
    char *allow_empty_string = "0";
    char *stricter_string = "0";
    char *count_string = "2147483647";
    char *in_img = NULL, *old_img = NULL, *tmp = NULL, *out_buf = NULL;
    char *intc_buf = NULL, *outtc_buf = NULL;
    char *drop_frame = NULL;
    char png_dir[MAX_PATH + 1] = {0};
    const char *additional_font_dir = NULL;
    crop_t crops[2];
    pic_t pic;
    uint32_t *pal = NULL;
    int out_filename_idx = 0;
    int have_fps = 0;
    int n_crop = 1;
    int split_at = 0;
    int min_split = 3;
    int autocrop = 0;
    int xo, yo, to;
    int fps = 25;
    int count_frames = INT_MAX, last_frame;
    int init_frame = 0;
    int frames;
    int first_frame = -1, start_frame = -1, end_frame = -1;
    int num_of_events = 0;
    int i, c, j;
    int have_line = 0;
    int must_zero = 0;
    int checked_empty;
    int even_y = 0;
    int auto_cut = 0;
    int pal_png = 1;
    int ugly = 0;
    int progress_step = 1000;
    int buffer_opt;
    long long bench_start = time(NULL);
    int fps_num = 25, fps_den = 1;
    int sup_output = 0;
    int xml_output = 0;
    int allow_empty = 0;
    int stricter = 0;
    sup_writer_t *sw = NULL;
    ass_input_t *ass_context;
    stream_info_t *s_info = malloc(sizeof(stream_info_t));
    event_list_t *events = event_list_new();
    event_t *event;
    FILE *fh;

    /* Get args */
    if (argc < 2)
    {
        print_usage();
        return 0;
    }
    while (1)
    {
        static struct option long_options[] =
        {   {"output",       required_argument, 0, 'o'}
            , {"seek",         required_argument, 0, 'j'}
            , {"count",        required_argument, 0, 'c'}
            , {"trackname",    required_argument, 0, 't'}
            , {"language",     required_argument, 0, 'l'}
            , {"video-format", required_argument, 0, 'v'}
            , {"fps",          required_argument, 0, 'f'}
            , {"x-offset",     required_argument, 0, 'x'}
            , {"y-offset",     required_argument, 0, 'y'}
            , {"t-offset",     required_argument, 0, 'd'}
            , {"split-at",     required_argument, 0, 's'}
            , {"min-split",    required_argument, 0, 'm'}
            , {"autocrop",     required_argument, 0, 'a'}
            , {"even-y",       required_argument, 0, 'e'}
            , {"palette",      required_argument, 0, 'p'}
            , {"buffer-opt",   required_argument, 0, 'b'}
            , {"ugly",         required_argument, 0, 'u'}
            , {"null-xml",     required_argument, 0, 'n'}
            , {"stricter",     required_argument, 0, 'z'}
            , {"font-dir",     required_argument, 0, 'g'}
            , {0, 0, 0, 0}
        };
        int option_index = 0;

        c = getopt_long(argc, argv, "o:j:c:t:l:v:f:x:y:d:b:s:m:e:p:a:u:n:z:g:", long_options, &option_index);
        if (c == -1)
            break;
        switch (c)
        {
        case 'o':
            if (out_filename_idx < 2)
                out_filename[out_filename_idx++] = optarg;
            else
            {
                return false;
            }
            break;
        case 'j':
            seek_string = optarg;
            break;
        case 'c':
            count_string = optarg;
            break;
        case 't':
            track_name = optarg;
            break;
        case 'l':
            language = optarg;
            break;
        case 'v':
            video_format = optarg;
            break;
        case 'f':
            frame_rate = optarg;
            break;
        case 'x':
            x_offset = optarg;
            break;
        case 'y':
            y_offset = optarg;
            break;
        case 'd':
            t_offset = optarg;
            break;
        case 'e':
            even_y_string = optarg;
            break;
        case 'p':
            palletize_png = optarg;
            break;
        case 'a':
            auto_crop_image = optarg;
            break;
        case 'b':
            buffer_optimize = optarg;
            break;
        case 's':
            split_after = optarg;
            break;
        case 'm':
            minimum_split = optarg;
            break;
        case 'u':
            ugly_option = optarg;
            break;
        case 'n':
            allow_empty_string = optarg;
            break;
        case 'z':
            stricter_string = optarg;
            break;
        case 'g':
            additional_font_dir = optarg;
            break;
        default:
            print_usage();
            return 0;
            break;
        }
    }
    if (argc - optind == 1)
        ass_filename = argv[optind];
    else
    {
        return 1;
    }

    /* Both input and output filenames are required */
    if (ass_filename == NULL)
    {
        print_usage();
        return 0;
    }
    if (out_filename[0] == NULL)
    {
        print_usage();
        return 0;
    }

    memset(s_info, 0, sizeof(stream_info_t));

    /* Get target output format */
    for (i = 0; i < out_filename_idx; i++)
    {
        if (is_extension(out_filename[i], "xml"))
        {
            xml_output_fn = out_filename[i];
            xml_output++;
            if (!get_dir_path(xml_output_fn, png_dir)) {
                return 1;
            }
        }
        else if (is_extension(out_filename[i], "sup") || is_extension(out_filename[i], "pgs"))
        {
            sup_output_fn = out_filename[i];
            sup_output++;
            pal_png = 1;
        }
        else
        {
            return 1;
        }
    }
    if (sup_output > 1 || xml_output > 1)
    {
        return false;
    }

    /* Set X and Y offsets, and split value */
    xo = parse_int(x_offset, "x-offset", NULL);
    yo = parse_int(y_offset, "y-offset", NULL);
    pal_png = parse_int(palletize_png, "palette", NULL);
    even_y = parse_int(even_y_string, "even-y", NULL);
    autocrop = parse_int(auto_crop_image, "autocrop", NULL);
    split_at = parse_int(split_after, "split-at", NULL);
    ugly = parse_int(ugly_option, "ugly", NULL);
    allow_empty = parse_int(allow_empty_string, "null-xml", NULL);
    stricter = parse_int(stricter_string, "stricter", NULL);
    init_frame = parse_int(seek_string, "seek", NULL);
    count_frames = parse_int(count_string, "count", NULL);
    min_split = parse_int(minimum_split, "min-split", NULL);
    if (!min_split)
        min_split = 1;

    /* TODO: Sanity check video_format and frame_rate. */

    /* Get frame rate */
    i = 0;
    while (framerates[i].name != NULL)
    {
        if (!strcasecmp(framerates[i].name, frame_rate))
        {
            frame_rate = framerates[i].out_name;
            fps = framerates[i].rate;
            drop_frame = framerates[i].drop ? "true" : "false";
            s_info->i_fps_num = fps_num = framerates[i].fps_num;
            s_info->i_fps_den = fps_den = framerates[i].fps_den;
            have_fps = 1;
        }
        i++;
    }
    if (!have_fps)
    {
        if (sscanf(frame_rate, "%d/%d", &fps_num, &fps_den) == 2) {
            drop_frame = "false";
            s_info->i_fps_num = fps_num;
            s_info->i_fps_den = fps_den;
            have_fps = 1;
        }
        return 1;
    }

    /* Get timecode offset. */
    to = parse_tc(t_offset, fps);

    /* Detect CPU features
    detect_sse2();*/

    /* Get video info and allocate buffer */
    if (open_file_ass(ass_filename, &ass_context, s_info))
    {
        print_usage();
        return 1;
    }

    char *video_formats[] = { "4k","2160p","2k","1440p","fhd","1080p","hd","720p" };
    int video_format_widths[] = { 3840,3840,2560,2560,1920,1920,1280,1280 };
    int video_format_heights[] = { 2160,2160,1440,1440,1080,1080,720,720 };
    int video_format_matched = 0;
    for (int ii = 0; ii < 8; ++ii) {
        if (!strcasecmp(video_format, video_formats[ii])) {
            s_info->i_width = video_format_widths[ii];
            s_info->i_height = video_format_heights[ii];
            video_format_matched = 1;
        }
    }
    if (video_format && !video_format_matched)
    {
        if (sscanf(video_format,"%d*%d", &s_info->i_width, &s_info->i_height) != 2) {
            return 1;
        }
    }

    ass_set_storage_size(ass_context->ass_renderer, s_info->i_width, s_info->i_height);
    ass_set_frame_size(ass_context->ass_renderer, s_info->i_width, s_info->i_height);

    if (additional_font_dir)
        ass_set_fonts_dir(ass_context->ass_library, additional_font_dir);

    ass_set_fonts(ass_context->ass_renderer, NULL, NULL, ASS_FONTPROVIDER_AUTODETECT, NULL, 1);

    in_img  = calloc(s_info->i_width * s_info->i_height * 4 + 16 * 2, sizeof(char)); /* allocate + 16 for alignment, and + n * 16 for over read/write */
    old_img = calloc(s_info->i_width * s_info->i_height * 4 + 16 * 2, sizeof(char)); /* see above */
    out_buf = calloc(s_info->i_width * s_info->i_height * 4 + 16 * 2, sizeof(char));

    /* Check minimum size */
    if (s_info->i_width < 8 || s_info->i_height < 8)
    {
        return 1;
    }

    /* Align buffers */
    in_img  = in_img + (short)(16 - ((long)in_img % 16));
    old_img = old_img + (short)(16 - ((long)old_img % 16));
    out_buf = out_buf + (short)(16 - ((long)out_buf % 16));

    /* Set up buffer (non-)optimization */
    buffer_opt = parse_int(buffer_optimize, "buffer-opt", NULL);
    pic.b = out_buf;
    pic.w = s_info->i_width;
    pic.h = s_info->i_height;
    pic.s = s_info->i_width;
    n_crop = 1;
    crops[0].x = 0;
    crops[0].y = 0;
    crops[0].w = pic.w;
    crops[0].h = pic.h;

    /* Get frame number */
    frames = get_frame_total_ass(ass_context, s_info);
    if (count_frames + init_frame > frames)
    {
        count_frames = frames - init_frame;
    }
    last_frame = count_frames + init_frame;

    /* No frames mean nothing to do */
    if (count_frames < 1)
    {
        return 0;
    }

    /* Set progress step */
    if (count_frames < 1000)
    {
        if (count_frames > 200)
            progress_step = 50;
        else if (count_frames > 50)
            progress_step = 10;
        else
            progress_step = 1;
    }

    /* Open SUP writer, if applicable */
    if (sup_output)
    {
        sw = new_sup_writer(sup_output_fn, pic.w, pic.h, fps_num, fps_den);
        if(sw==NULL) {
            return 1;
        }
    }

    int changed = 1;

    /* Process frames */
    for (i = init_frame; i < last_frame; i++)
    {
        long long ts = (long double)i * s_info->i_fps_den / s_info->i_fps_num * 1000;

        ASS_Image *img = ass_render_frame(ass_context->ass_renderer, ass_context->ass, ts, &changed);
        memset(in_img, 0, s_info->i_width *s_info->i_height * 4);
        make_sub_img(img, in_img, s_info->i_width);

        checked_empty = 0;

        /* Progress indicator */
        if (i % (count_frames / progress_step) == 0)
        {
        }

        /* If we are outside any lines, check for empty frames first */
        if (!have_line)
        {
            if (is_empty(s_info, in_img))
                continue;
            else
                checked_empty = 1;
        }

        /* Check for duplicate, unless first frame */
        if ((i != init_frame) && have_line && !changed)
            continue;
        /* Mark frames that were not used as new image in comparison to have transparent pixels zeroed */
        else if (!(i && have_line))
            must_zero = 1;

        /* Not a dup, write end-of-line, if we had a line before */

        if (have_line)
        {
            if (sup_output)
            {
                assert(pal != NULL);
                write_sup_wrapper(sw, (uint8_t *)out_buf, n_crop, crops, pal, start_frame + to, i + to, split_at, min_split, stricter);
                if (!xml_output)
                    free(pal);
                pal = NULL;
            }
            if (xml_output)
                add_event_xml(events, split_at, min_split, start_frame + to, i + to, n_crop, crops);
            end_frame = i;
            have_line = 0;
        }

        /* Check for empty frame, if we didn't before */
        if (!checked_empty && is_empty(s_info, in_img))
            continue;

        /* Zero transparent pixels, if needed */
        if (must_zero)
            zero_transparent(s_info, in_img);
        must_zero = 0;

        /* Not an empty frame, start line */
        have_line = 1;
        start_frame = i;
        swap_rb(s_info, in_img, out_buf);
        if (buffer_opt)
            n_crop = auto_split(pic, crops, ugly, even_y);
        else if (autocrop)
        {
            crops[0].x = 0;
            crops[0].y = 0;
            crops[0].w = pic.w;
            crops[0].h = pic.h;
            auto_crop(pic, crops);
        }
        if ((buffer_opt || autocrop) && even_y)
            enforce_even_y(crops, n_crop);
        if ((pal_png || sup_output) && pal == NULL)
            pal = palletize(out_buf, s_info->i_width, s_info->i_height);
        if (xml_output)
            for (j = 0; j < n_crop; j++)
            {
                if (!write_png(png_dir, start_frame, (uint8_t *)out_buf, s_info->i_width, s_info->i_height, j, pal, crops[j])) {
                    return 1;
                }
            }
        if (pal_png && xml_output && !sup_output)
        {
            free(pal);
            pal = NULL;
        }
        num_of_events++;
        if (first_frame == -1)
            first_frame = i;

        /* Save image for next comparison. */
        tmp = in_img;
        in_img = old_img;
        old_img = tmp;
    }

    /* Add last event, if available */
    if (have_line)
    {
        if (sup_output)
        {
            assert(pal != NULL);
            write_sup_wrapper(sw, (uint8_t *)out_buf, n_crop, crops, pal, start_frame + to, i - 1 + to, split_at, min_split, stricter);
            if (!xml_output)
                free(pal);
            pal = NULL;
        }
        if (xml_output)
        {
            add_event_xml(events, split_at, min_split, start_frame + to, i - 1 + to, n_crop, crops);
            free(pal);
            pal = NULL;
        }
        auto_cut = 1;
        end_frame = i - 1;
    }

    if (sup_output)
    {
        close_sup_writer(sw);
    }

    if (xml_output)
    {
        /* Check if we actually have any events */
        if (first_frame == -1)
        {
            if (!allow_empty)
            {
                return 0;
            }
            else
            {
                first_frame = 0;
                end_frame = 0;
            }
        }

        /* Initialize timecode buffers */
        intc_buf = calloc(12, 1);
        outtc_buf = calloc(12, 1);

        /* Creating output file */
        if ((fh = fopen(xml_output_fn, "w")) == 0)
        {
            return 1;
        }

        /* Write XML header */
        if (!mk_timecode(first_frame + to, fps, intc_buf)) {
            return 1;
        }
        if (!mk_timecode(end_frame + to + auto_cut, fps, outtc_buf)) {
            return 1;
        }
        fprintf(fh, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
                "<BDN Version=\"0.93\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n"
                "xsi:noNamespaceSchemaLocation=\"BD-03-006-0093b BDN File Format.xsd\">\n"
                "<Description>\n"
                "<Name Title=\"%s\" Content=\"\"/>\n"
                "<Language Code=\"%s\"/>\n"
                "<Format VideoFormat=\"%s\" FrameRate=\"%s\" DropFrame=\"%s\"/>\n"
                "<Events LastEventOutTC=\"%s\" FirstEventInTC=\"%s\"\n", track_name, language, video_format, frame_rate, drop_frame, outtc_buf, intc_buf);

        if(!mk_timecode(0, fps, intc_buf)) {
            return 1;
        }
        if(! mk_timecode(frames + to, fps, outtc_buf)) {
            return 1;
        }
        fprintf(fh, "ContentInTC=\"%s\" ContentOutTC=\"%s\" NumberofEvents=\"%d\" Type=\"Graphic\"/>\n"
                "</Description>\n"
                "<Events>\n", intc_buf, outtc_buf, num_of_events);

        /* Write XML events */
        if (!event_list_empty(events))
        {
            event = event_list_first(events);
            do
            {
                if (!mk_timecode(event->start_frame, fps, intc_buf)) {
                    return 1;
                }
                if (!mk_timecode(event->end_frame, fps, outtc_buf)) {
                    return 1;
                }

                if (auto_cut && event->end_frame == frames - 1)
                {
                    if (!mk_timecode(event->end_frame + 1, fps, outtc_buf)) {
                        return 1;
                    }
                }

                fprintf(fh, "<Event Forced=\"False\" InTC=\"%s\" OutTC=\"%s\">\n", intc_buf, outtc_buf);
                for (i = 0; i < event->graphics; i++)
                {
                    fprintf(fh, "<Graphic Width=\"%d\" Height=\"%d\" X=\"%d\" Y=\"%d\">%08d_%d.png</Graphic>\n", event->c[i].w, event->c[i].h, xo + event->c[i].x, yo + event->c[i].y, event->image_number - to, i);
                }
                fprintf(fh, "</Event>\n");
                event = event_list_next(events);
            }
            while (event != NULL);
        }

        /* Write XML footer */
        fprintf(fh, "</Events>\n</BDN>\n");

        /* Close XML file */
        fclose(fh);
    }

    /* Cleanup */
    close_file_ass(ass_context);

    return 0;
}