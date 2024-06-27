/*----------------------------------------------------------------------------
 * avs2bdnxml - Generates BluRay subtitle stuff from RGBA AviSynth scripts
 * Copyright (C) 2008-2010 Arne Bochem <avs2bdnxml at ps-auxw de>
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

#ifndef SUP_H
#define SUP_H

#include "auto_split.h"
#include "abstract_lists.h"

typedef struct subtitle_info_s
{
	int start;
	int end;
	int num_crop;
	rect_t crops[2];
	int rle_len[2];
	uint8_t *rle[2];
	uint32_t pal[256];
} subtitle_info_t;

DECLARE_LIST(si, subtitle_info_t)

typedef struct sup_writer_s
{
	FILE *fh;
	int non_new;
	int im_w;
	int im_h;
	int colorspace;
	int fps_num;
	int fps_den;
	int fps_id;
	uint16_t comp_num;
	unsigned int end;
	unsigned int follower_end;
	int buffer;
	int objects;
	int palettes;
	int palette_offset;
	int picture_offset;
	int last_end_ts;
	int last_window_ts;
	int window_num;
	rect_t windows[2];
	si_list_t *sil;
} sup_writer_t;

/* Create a new sup writer state */
sup_writer_t *new_sup_writer (char *filename, int im_w, int im_h, int fps_num, int fps_den);

/* Write sup data for subtitle */
void write_sup (sup_writer_t *sw, uint8_t *im, int num_crop, rect_t *crops, uint32_t *pal, int start, int end, int strict);

/* Call this once at the end */
void close_sup_writer (sup_writer_t *sw);

#endif

