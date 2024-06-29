/*----------------------------------------------------------------------------
 * avs2bdnxml - Generates BluRay subtitle stuff from RGBA AviSynth scripts
 * Copyright (C) 2008-2013 Arne Bochem <avs2bdnxml at ps-auxw de>
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
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <math.h>
#include "auto_split.h"
#include "sup.h"
#include "abstract_lists.h"

#include <stdbool.h>  // Include the stdbool.h header for bool type

// Global variable to store system endianness, default to true (little-endian)
static bool is_little_endian = true;

// Function to detect system endianness
void detect_endianness() {
    uint16_t test = 0x1;
    is_little_endian = *((uint8_t*)&test) == 0x1;
}

// Modify macro definitions
#define SWAP32(x) (is_little_endian ? (int32_t)(((x & 0xff000000) >> 24) | ((x & 0xff0000) >> 8) | ((x & 0xff00) << 8) | ((x & 0xff) << 24)) : (x))
#define SWAP16(x) (is_little_endian ? (int16_t)(((x & 0xff) << 8) | ((x & 0xff00) >> 8)) : (x))

#ifndef DEBUG
#define DEBUG 0
#endif

static int count (uint8_t *im, int x, int w, int col)
{
    int c = 0;
    for (; x < w && im[x] == col && c <= 16383; x++) /* 14bit max */
        c++;
    return c;
}

/* Output buffer will contain malloced RLE data */
#define PUSH(x) {*(b++)=(x);(*len)++;}
#define FLAG_COLOR 0x80
#define FLAG_LONG 0x40
static uint8_t *rl_encode (uint8_t *im, int w, int h, rect_t crop, int *len)
{
    uint8_t *b = calloc(crop.w * crop.h, 4); /* Over-allocation */
    uint8_t *rle = b;
    uint8_t col, o;
    int x, y, c;

    *len = 0;
    for (y = crop.y; y < crop.y + crop.h && y < h; y++)
    {
        for (x = crop.x; x < crop.x + crop.w && x < w; x += c)
        {
            col = im[x + y * w];
            c = count(im + y * w, x, MIN(crop.x + crop.w, w), col);

            /* Shorter than shortest range encoding */
            if (c < 3 && col)
            {
                for (o = 0; o < c; o++)
                    PUSH(col);
                continue;
            }

            /* Range marker */
            PUSH (0);

            /* Set color flag, if necessary */
            o = 0;
            if (col)
                o |= FLAG_COLOR;

            /* Encode length */
            if (c >= 64) /* 6bit max */
            {
                o |= FLAG_LONG;
                o |= c >> 8;
                PUSH(o);
                PUSH(c & 0xFF);
            }
            else
                PUSH(o | c);

            /* Output color, if necessary */
            if (col)
                PUSH(col);
        }

        /* End of line marker */
        PUSH(0);
        PUSH(0);
    }

    return rle;
}

typedef struct sup_header_s
{
    uint8_t m1;          /* 'P' */
    uint8_t m2;          /* 'G' */
    uint32_t start_time;
    uint32_t dts;
    uint8_t packet_type; /* 0x16 = pcs_start/end, 0x17 = wds, 0x14 = palette, 0x15 = ods_first, 0x80 = null */
    uint16_t packet_len;
} __attribute ((packed)) sup_header_t;

void conv_sup_header (sup_header_t *h)
{
    h->start_time = SWAP32(h->start_time);
    h->dts = SWAP32(h->dts);
    h->packet_len = SWAP16(h->packet_len);
}

static void write_header (FILE *fh, int start_time, int dts, int packet_type, int packet_len)
{
    sup_header_t h;

    h.m1 = 80;
    h.m2 = 71;
    h.start_time = start_time;
    h.dts = dts;
    h.packet_type = packet_type;
    h.packet_len = packet_len;

    conv_sup_header(&h);
    fwrite(&h, sizeof(h), 1, fh);
}

typedef struct sup_pcs_start_s
{
    uint16_t width;
    uint16_t height; /* height - 2 * Core.getCropOfsY */
    uint8_t fps_id; /* getFpsId() */
    uint16_t comp_num;
    uint8_t follower;  /* 0x80 if first or single, 0x40 if follows directly (end = start) or the frame after */
    uint16_t m; /* 0 */
    uint8_t objects; /* 1 */
} __attribute ((packed)) sup_pcs_start_t;

void conv_sup_pcs_start (sup_pcs_start_t *pcss)
{
    pcss->width = SWAP16(pcss->width);
    pcss->height = SWAP16(pcss->height);
    pcss->comp_num = SWAP16(pcss->comp_num);
    pcss->m = SWAP16(pcss->m);
}

typedef struct sup_pcs_start_obj_s
{
    uint16_t picture;
    uint8_t window;
    uint8_t forced; /* forced ? 64 : 0 */
    uint16_t x_off;
    uint16_t y_off;
} __attribute ((packed)) sup_pcs_start_obj_t;

void conv_sup_pcs_start_obj (sup_pcs_start_obj_t *pcsso)
{
    pcsso->picture = SWAP16(pcsso->picture);
    pcsso->x_off = SWAP16(pcsso->x_off);
    pcsso->y_off = SWAP16(pcsso->y_off);
}

static void write_pcs_start (FILE *fh, int start_time, int dts, int follower, int objects, int vid_w, int vid_h, int fps_id, int comp_num)
{
    sup_pcs_start_t pcss;

    write_header(fh, start_time, dts, 22, sizeof(pcss) + objects * sizeof(sup_pcs_start_obj_t));

    pcss.m = 0;
    pcss.width = vid_w;
    pcss.height = vid_h;
    pcss.fps_id = fps_id;
    pcss.comp_num = comp_num;
    pcss.follower = !follower ? 0x80 : 0x40; /* 0x80 for single lines, and first lines, 0x40 for following directly or with one frame between */
    pcss.objects = objects;

    conv_sup_pcs_start(&pcss);
    fwrite(&pcss, sizeof(pcss), 1, fh);
}

static void write_pcs_start_obj (FILE *fh, int picture, int window, int x_off, int y_off)
{
    sup_pcs_start_obj_t pcsso;

    pcsso.picture = picture;
    pcsso.window = window;
    pcsso.forced = 0;
    pcsso.x_off = x_off;
    pcsso.y_off = y_off;

    conv_sup_pcs_start_obj(&pcsso);
    fwrite(&pcsso, sizeof(pcsso), 1, fh);
}

typedef struct sup_wds_s
{
    uint8_t windows; /* 1 or 2 */
} __attribute ((packed)) sup_wds_t;

void conv_sup_wds (sup_wds_t *wds)
{
    /* Do nothing. */
}

typedef struct sup_wds_obj_s
{
    uint8_t window; /* 0 or 1 */
    uint16_t x_off;
    uint16_t y_off;
    uint16_t width;
    uint16_t height;
} __attribute ((packed)) sup_wds_obj_t;

void conv_sup_wds_obj (sup_wds_obj_t *wdso)
{
    wdso->x_off = SWAP16(wdso->x_off);
    wdso->y_off = SWAP16(wdso->y_off);
    wdso->width = SWAP16(wdso->width);
    wdso->height = SWAP16(wdso->height);
}

static void write_wds (FILE *fh, int timestamp, int dts, int windows)
{
    sup_wds_t wds;

    write_header(fh, timestamp, dts, 23, sizeof(wds) + windows * sizeof(sup_wds_obj_t));

    wds.windows = windows;

    conv_sup_wds(&wds);
    fwrite(&wds, sizeof(wds), 1, fh);
}

static void write_wds_obj (FILE *fh, int window, int w, int h, int x_off, int y_off)
{
    sup_wds_obj_t wdso;

    wdso.window = window;
    wdso.x_off = x_off;
    wdso.y_off = y_off;
    wdso.width = w;
    wdso.height = h;

    conv_sup_wds_obj(&wdso);
    fwrite(&wdso, sizeof(wdso), 1, fh);
}

#define CLAMP(x,min,max) (MAX(MIN(x,max),min))
static uint8_t get_y (uint32_t c, int s)
{
    uint8_t *v = (uint8_t *)&c;
    uint8_t r = v[0], g = v[1], b = v[2];
    if (s)
        return (uint8_t)CLAMP(16 + (int)floor(0.5 + (double)(r * 0.299 * 219.0 / 255.0 + g * 0.587 * 219.0 / 255.0 + b * 0.114 * 219.0 / 255.0)), 16, 235);
    else
        return (uint8_t)CLAMP(16 + (int)floor(0.5 + (double)(r * 0.2126 * 219.0 / 255.0 + g * 0.7152 * 219.0 / 255.0 + b * 0.0722 * 219.0 / 255.0)), 16, 235);
}
static uint8_t get_u (uint32_t c, int s)
{
    uint8_t *v = (uint8_t *)&c;
    uint8_t r = v[0], g = v[1], b = v[2];
    if (s)
        return (uint8_t)CLAMP(128 + (int)floor(0.5 + (double)(r * 0.5 * 224.0 / 255.0 - (g * 0.418688 * 224.0 / 255.0) - (b * 0.081312 * 224.0 / 255.0))), 16, 240);
    else
        return (uint8_t)CLAMP(128 + (int)floor(0.5 + (double)(r * 0.5 * 224.0 / 255.0 - (g * 0.7152 / 1.5748 * 224.0 / 255.0) - (b * 0.0722 / 1.5748 * 224.0 / 255.0))), 16, 240);
}
static uint8_t get_v (uint32_t c, int s)
{
    uint8_t *v = (uint8_t *)&c;
    uint8_t r = v[0], g = v[1], b = v[2];
    if (s)
        return (uint8_t)CLAMP(128 + (int)floor(0.5 + (double)(-r * 0.168736 * 224.0 / 255.0 - (g * 0.331264 * 224.0 / 255.0) + b * 0.5 * 224.0 / 255.0)), 16, 240);
    else
        return (uint8_t)CLAMP(128 + (int)floor(0.5 + (double)(-r * 0.2126 / 1.8556 * 224.0 / 255.0 - (g * 0.7152 / 1.8556 * 224.0 / 255.0) + b * 0.5 * 224.0 / 255.0)), 16, 240);
}

typedef struct sup_palette_s
{
    uint16_t palette;
} __attribute ((packed)) sup_palette_t;

void conv_sup_palette (sup_palette_t *p)
{
    p->palette = SWAP16(p->palette);
}

/* Colorspace = 1 for 480p/576p, 0 otherwise */
#define PUT(x) { t = (uint8_t)(x); fwrite(&t, 1, 1, fh); }
static void write_palette (FILE *fh, int dts, int palette, uint32_t *pal, int colorspace)
{
    sup_palette_t p;
    int entries = 1, i;
    uint8_t t;

    for (i = 1; i < 256 && pal[i]; i++)
        entries++;
    write_header(fh, dts, 0, 20, sizeof(p) + entries * 5);

    p.palette = palette;
    conv_sup_palette(&p);
    fwrite(&p, sizeof(p), 1, fh);

    for (i = 0; i < entries; i++)
    {
        PUT(i)
        PUT(get_y(pal[i], colorspace))
        PUT(get_u(pal[i], colorspace))
        PUT(get_v(pal[i], colorspace))
        PUT(((uint8_t *)&(pal[i]))[3])
    }
}

typedef struct sup_ods_first_s
{
    uint16_t picture;
    uint8_t m; /* 0 */
    uint32_t magic_len; /* (single_packet ? 0xc0000000 : 0x80000000) | (length + 4) */
    uint16_t width;
    uint16_t height;
} __attribute ((packed)) sup_ods_first_t;

void conv_sup_ods_first (sup_ods_first_t *odsf)
{
    odsf->picture = SWAP16(odsf->picture);
    odsf->magic_len = SWAP32(odsf->magic_len);
    odsf->width = SWAP16(odsf->width);
    odsf->height = SWAP16(odsf->height);
}

typedef struct sup_ods_next_s
{
    uint16_t picture;
    uint8_t m; /* 0 */
    uint8_t last; /* 0 if not, if 64 yes */
} __attribute ((packed)) sup_ods_next_t;

void conv_sup_ods_next (sup_ods_next_t *odsn)
{
    odsn->picture = SWAP16(odsn->picture);
}

static void write_image (FILE *fh, int timestamp, int dts, int picture, int w, int h, uint8_t *rle, int rle_len)
{
    sup_ods_first_t odsf;
    sup_ods_next_t odsn = {picture, 0, 0};
    int length = 0x80000000 | (rle_len + 4);
    int size;

    if (rle_len > 65508)
        size = 65508;
    else
    {
        size = rle_len;
        length |= 0x40000000;
    }
    rle_len -= size;

    write_header(fh, timestamp, dts, 21, sizeof(odsf) + size);

    odsf.picture = picture;
    odsf.m = 0;
    odsf.magic_len = length;
    odsf.width = w;
    odsf.height = h;

    conv_sup_ods_first(&odsf);
    fwrite(&odsf, sizeof(odsf), 1, fh);
    fwrite(rle, size, 1, fh);
    rle += size;

    while (rle_len)
    {
        if (rle_len > 65515)
            size = 65515;
        else
            size = rle_len;
        rle_len -= size;

        if (!rle_len)
            odsn.last = 64;

        write_header(fh, timestamp, dts, 21, sizeof(odsn) + size);
        conv_sup_ods_next(&odsn);
        fwrite(&odsn, sizeof(odsn), 1, fh);
        fwrite(rle, size, 1, fh);
        rle += size;
    }
}

static void write_marker (FILE *fh, int time)
{
    write_header(fh, time, 0, 0x80, 0);
}

typedef struct sup_pcs_end_s
{
    uint16_t width;
    uint16_t height;
    uint8_t fps_id;
    uint16_t comp_num;
    uint32_t m;
} __attribute ((packed)) sup_pcs_end_t;

void conv_sup_pcs_end (sup_pcs_end_t *pcse)
{
    pcse->width = SWAP16(pcse->width);
    pcse->height = SWAP16(pcse->height);
    pcse->comp_num = SWAP16(pcse->comp_num);
    pcse->m = SWAP32(pcse->m);
}

static void write_pcs_end (FILE *fh, int end_time, int dts, int w, int h, int fps_id, int comp_num)
{
    sup_pcs_end_t pcse;

    write_header(fh, end_time, dts, 22, sizeof(pcse));

    pcse.width = w;
    pcse.height = h;
    pcse.fps_id = fps_id;
    pcse.comp_num = comp_num;
    pcse.m = 0;

    conv_sup_pcs_end(&pcse);
    fwrite(&pcse, sizeof(pcse), 1, fh);
}

typedef struct fps_id_s
{
    int num;
    int den;
    int id;
} fps_id_t;

static int get_id (int fps_num, int fps_den)
{
    fps_id_t ids[] = {{24000, 1001, 16}, {24, 1, 32}, {25, 1, 48}, {30000, 1001, 64}, {50, 1, 96}, {60000, 1001, 112}, {0, 0, 0}};
    int i = 0;

    while (ids[i].den)
    {
        if (ids[i].num == fps_num && ids[i].den == fps_den)
            return ids[i].id;
        i++;
    }

    return 16;
}

sup_writer_t *new_sup_writer (char *filename, int im_w, int im_h, int fps_num, int fps_den)
{
    sup_writer_t *sw = malloc(sizeof(sup_writer_t));

    if ((sw->fh = fopen(filename, "wb")) == NULL)
    {
        return NULL;
    }

    sw->non_new = 0;
    sw->im_w = im_w;
    sw->im_h = im_h;

    if (im_h == 480 || im_h == 576)
        sw->colorspace = 1;
    else
        sw->colorspace = 0;

    sw->fps_num = fps_num;
    sw->fps_den = fps_den;
    sw->fps_id = get_id(fps_num, fps_den);
    sw->comp_num = 0;
    sw->end = -2;
    sw->follower_end = -2;
    sw->objects = 0;
    sw->palettes = 0;
    sw->buffer = 0;
    sw->palette_offset = 0;
    sw->picture_offset = 0;
    sw->last_end_ts = 0;
    sw->last_window_ts = 0;
    sw->window_num = 0;
    sw->sil = si_list_new();

    memset(sw->windows, 0, 2 * sizeof(rect_t));

    return sw;
}

void destroy_si (subtitle_info_t *si)
{
    int i;

    for (i = 0; i < si->num_crop; i++)
        free(si->rle[i]);

    free(si);
}

void write_subtitle (sup_writer_t *sw, uint8_t **rle, int *rle_len, int num_crop, rect_t *crops, uint32_t *pal, int start, int end, int new_composition)
{
    uint32_t frame_ts, window_ts, decode_ts;
    uint32_t window_ts_list[2], decode_ts_list[2];
    uint32_t later_window;
    int in_window[2];
    uint32_t dts;
    uint32_t start_ts, end_ts, ts;
    int follower = 0;
    uint32_t im_ts = 0;
    int i, j;
    double tick_fac = 90000;

    /* For stuff following frame by frame (endts = startts):
     *   - Write: PCSS, WDS, PAL, IMG, MARK, [PCSS, WDS, PAL, IMG, MARK...], PCSE, WDS, MARK
     *
     * Behaviour in reference file:
     *   During an epoch of followers, if number of objects/position/size changes, increase
     *   composition by one, increase all further picture values by previous number of
     *   composition objects, increase palette by one. Object id in PCSS is 1 for all.
     *   Do some stuff to WDS too. Keep it to first value, if stuff still fits in the window?
     *   It seems WDS has to stay constant within a single composition. What a pain.
     */
    tick_fac *= ((double)sw->fps_den) / ((double)sw->fps_num);

    start_ts = (int)floor((double)start * tick_fac + 0.5);
    end_ts = (int)floor((double)end * tick_fac + 0.5);

    /* Calculate some timestamps/modifiers */
    frame_ts = (sw->im_w * sw->im_h * 9 + 3199) / 3200;
    window_ts = 0;

    if (sw->non_new && ((start == sw->follower_end) || (start == sw->follower_end + 1)))
        follower = 1;
    sw->follower_end = end;

    for (i = 0; i < sw->window_num; i++)
    {
        window_ts_list[i] = (sw->windows[i].w * sw->windows[i].h * 9 + 3199) / 3200;
        window_ts += window_ts_list[i];
    }
    decode_ts = 0;
    for (i = 0; i < num_crop; i++)
    {
        decode_ts_list[i] = (crops[i].w * crops[i].h * 9 + 1599) / 1600;
        decode_ts += decode_ts_list[i];
    }

    if (sw->window_num > 1)
        later_window = window_ts_list[1];
    else
        later_window = window_ts_list[0];

    /* Calculate dts */
    if (num_crop == 1)
    {
        if (new_composition)
            dts = start_ts - frame_ts - window_ts;
        else
            dts = start_ts - window_ts - decode_ts;
    }
    else
    {
        if (new_composition)
            dts = start_ts - frame_ts - window_ts;
        else
            dts = start_ts - decode_ts - later_window;
    }

    /* Determine windows */
    for (i = 0; i < num_crop; i++)
        for (j = 0; j < sw->window_num; j++)
            if (crops[i].x >= sw->windows[j].x && crops[i].x + crops[i].w <= sw->windows[j].x + sw->windows[j].w && crops[i].y >= sw->windows[j].y && crops[i].y + crops[i].h <= sw->windows[j].y + sw->windows[j].h)
            {
                in_window[i] = j;
                break;
            }

    /* Write PCSS */
    write_pcs_start(sw->fh, start_ts, dts, follower, num_crop, sw->im_w, sw->im_h, sw->fps_id, sw->comp_num);
    for (i = 0; i < num_crop; i++)
        write_pcs_start_obj(sw->fh, sw->picture_offset + i, in_window[i], crops[i].x, crops[i].y);

    /* Write WDS */
    ts = start_ts - window_ts; /* Can be very slightly off, possible rounding error (FIXME: fixed?) */
    write_wds(sw->fh, ts, dts, sw->window_num);
    for (i = 0; i < sw->window_num; i++)
        write_wds_obj(sw->fh, i, sw->windows[i].w, sw->windows[i].h, sw->windows[i].x, sw->windows[i].y);

    /* Write palette */
    write_palette(sw->fh, dts, sw->palette_offset, pal, sw->colorspace);

    /* Write image data */
    for (i = 0; i < num_crop; i++)
    {
        if (num_crop == 1)
        {
            if (new_composition)
                im_ts = start_ts - frame_ts + window_ts; /* This one can be off a bit. (FIXME) */
            else
                im_ts = start_ts - window_ts;
        }
        else if (i == 0)
        {
            if (new_composition)
                im_ts = start_ts - frame_ts + window_ts_list[0] - later_window; /* This one can be off a bit, ~5/90000s was observed (FIXME: fixed?) */
            else
                im_ts = start_ts - later_window - decode_ts_list[1];
        }
        else
        {
            if (new_composition)
            {
                im_ts = start_ts - frame_ts + window_ts; /* Can be slightly off, possible rounding error (FIXME: fixed?) */
                dts = start_ts - frame_ts + window_ts_list[0] - later_window;
            }
            else
            {
                im_ts = start_ts - later_window;
                dts = start_ts - later_window - decode_ts_list[1];
            }
        }
        write_image(sw->fh, im_ts, dts, sw->picture_offset + i, crops[i].w, crops[i].h, rle[i], rle_len[i]);
    }

    /* Write marker */
    write_marker(sw->fh, im_ts);

    /* Remember data for creation of composition end */
    sw->last_end_ts = end_ts;
    sw->last_window_ts = window_ts;
}

void write_composition (sup_writer_t *sw)
{
    rect_t *rects;
    subtitle_info_t *si;
    int last_num_crop = 0;
    rect_t last_crops[2];
    int new_composition = 1;
    int si_rects = 0;
    int ts, dts;
    int i;

    /* Only write anything if there is a non-empty composition */
    if (!sw->non_new)
        return;

    /* Count subtitles */
    si = si_list_first(sw->sil);
    while (si != NULL)
    {
        si_rects += si->num_crop;
        si = si_list_next(sw->sil);
    }

    /* Gather crop rects. */
    rects = malloc(si_rects * sizeof(rect_t));
    si = si_list_first(sw->sil);
    si_rects = 0;
    while (si != NULL)
    {
        for (i = 0; i < si->num_crop; i++)
            rects[si_rects++] = si->crops[i];
        si = si_list_next(sw->sil);
    }

    /* Calculate windows */
    sw->window_num = find_windows(rects, si_rects, sw->windows);

    if (!sw->window_num)
    {
        return;
    }

    /* Write subtitles */
    si = si_list_first(sw->sil);
    while (si != NULL)
    {
        if (!new_composition && (last_num_crop != si->num_crop || memcmp(last_crops, si->crops, MIN(last_num_crop, si->num_crop) * sizeof(rect_t))))
        {
            (sw->comp_num)++;
            (sw->palette_offset)++;
            (sw->picture_offset) += last_num_crop;
        }
        last_num_crop = si->num_crop;
        memcpy(last_crops, si->crops, si->num_crop * sizeof(rect_t));
        write_subtitle(sw, si->rle, si->rle_len, si->num_crop, si->crops, si->pal, si->start, si->end, new_composition);
        new_composition = 0;
        si_list_delete(sw->sil);
        destroy_si(si);
        si = si_list_get(sw->sil);
    }

    /* Write PCSE */
    dts = sw->last_end_ts - sw->last_window_ts - 1;
    write_pcs_end(sw->fh, sw->last_end_ts, dts, sw->im_w, sw->im_h, sw->fps_id, ++(sw->comp_num));

    /* Write WDS */
    ts = sw->last_end_ts - sw->last_window_ts;
    write_wds(sw->fh, ts, dts, sw->window_num);
    for (i = 0; i < sw->window_num; i++)
        write_wds_obj(sw->fh, i, sw->windows[i].w, sw->windows[i].h, sw->windows[i].x, sw->windows[i].y);

    /* Write marker */
    write_marker(sw->fh, dts);

    /* Cleanup */
    free(rects);

    /* New composition */
    (sw->comp_num)++;

    /* Reset picture and palette count, new buffer. */
    sw->objects = 0;
    sw->palettes = 0;
    sw->palette_offset = 0;
    sw->picture_offset = 0;
    sw->buffer = 0;
}

subtitle_info_t *collect_si (sup_writer_t *sw, uint8_t *im, int num_crop, rect_t *crops, uint32_t *pal, int start, int end)
{
    subtitle_info_t *si = malloc(sizeof(subtitle_info_t));
    int i;

    si->start = start;
    si->end = end;
    si->num_crop = num_crop;
    for (i = 0; i < num_crop; i++)
    {
        si->crops[i].w = crops[i].w;
        si->crops[i].h = crops[i].h;
        si->crops[i].x = crops[i].x;
        si->crops[i].y = crops[i].y;
        si->rle[i] = rl_encode(im, sw->im_w, sw->im_h, si->crops[i], &(si->rle_len[i]));
    }
    memcpy(si->pal, pal, 256 * sizeof(uint32_t));

    return si;
}

void close_sup_writer (sup_writer_t *sw)
{
    write_composition(sw);

    si_list_first(sw->sil);
    while(!si_list_empty(sw->sil))
    {
        destroy_si(si_list_first(sw->sil));
        si_list_delete(sw->sil);
    }
    si_list_destroy(sw->sil);

    fclose(sw->fh);
    free(sw);
}

IMPLEMENT_LIST(si, subtitle_info_t)

void write_sup (sup_writer_t *sw, uint8_t *im, int num_crop, rect_t *crops, uint32_t *pal, int start, int end, int strict)
{
    rect_t tmp;
    int buffer_increase;
    int i;

    buffer_increase = 0;
    for (i = 0; i < num_crop; i++)
        buffer_increase += crops[i].w * crops[i].h + 16;
    /* Disabled some conditions for now. */
    if (sw->non_new && ((start > sw->end + 1) || (sw->objects + num_crop > 64) || (strict && ((sw->buffer + buffer_increase >= 4 * 1024 * 1024) || (sw->palettes + 1 > 8)))))
    {
        write_composition(sw);
    }
    sw->non_new = 1;
    sw->end = end;
    sw->buffer += buffer_increase;
    sw->objects += num_crop;
    (sw->palettes)++; /* FIXME: It's probably okay not to increase this if the palette is identical to the last, but this would require further testing and possibly additional code so identical palettes are not written to the stream multiple times? */

    if (num_crop > 1)
    {
        /* Let's order them, so the one closer to 0/0 is the second. */
        if (crops[0].y < crops[1].y || (crops[0].y == crops[1].y && crops[0].x < crops[1].x))
        {
            tmp = crops[0];
            crops[0] = crops[1];
            crops[1] = tmp;
        }
    }

    si_list_insert_after(sw->sil, collect_si(sw, im, num_crop, crops, pal, start, end));
}

