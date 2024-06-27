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

#ifndef AUTO_SPLIT_H
#define AUTO_SPLIT_H

#define MAX(x,y) ((x) > (y) ? (x) : (y))
#define MIN(x,y) ((x) < (y) ? (x) : (y))

typedef struct pic_s
{
	char *b; /* Buffer */
	int w;
	int h;
	int s;   /* Stride */
} pic_t;

typedef struct crop_s
{
	int x;
	int y;
	int w;
	int h;
} crop_t;

typedef crop_t rect_t;

void auto_crop (pic_t p, crop_t *c);
int find_windows (crop_t *rects, int n_rects, crop_t *windows);
int auto_split (pic_t p, crop_t *c, int ugly, int even_y);
rect_t merge_rects (rect_t r1, rect_t r2);
int score_rect (rect_t r);
void enforce_even_y (crop_t *c, int n);

#endif

