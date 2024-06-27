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
#include <string.h>
#include <limits.h>
#include "auto_split.h"
#include "abstract_lists.h"
#include "sort.h"

/* Transparent pixels are assumed to be set to zero */

void auto_crop (pic_t p, crop_t *c)
{
	uint32_t *b = (uint32_t *)p.b;
	int min_x = INT_MAX, max_x = INT_MIN, min_y = INT_MAX, max_y = INT_MIN;
	int seen_pixel = 0;
	int x, y;

	for (y = c->y; y < c->y + c->h && y < p.h; y++)
		for (x = c->x; x < c->x + c->w && x < p.w; x++)
			if (b[x + p.s * y])
			{
				seen_pixel = 1;
				if (x < min_x)
					min_x = x;
				if (x > max_x)
					max_x = x;
				if (y < min_y)
					min_y = y;
				if (y > max_y)
					max_y = y;
			}

	if (!seen_pixel)
	{
		c->w = 0;
		c->h = 0;
	}
	else
	{
		c->x = min_x;
		c->y = min_y;
		c->w = max_x - min_x + 1;
		c->h = max_y - min_y + 1;
	}

	/* Ensure no forbidden/tiny results are produced */
	if (c->w < 8)
	{
		if (c->x + 8 > p.w)
			c->x -= c->x + 8 - p.w;
		c->w = 8;
	}
	if (c->h < 8)
	{
		if (c->y + 8 > p.h)
			c->y -= c->y + 8 - p.h;
		c->h = 8;
	}
}

static int block_state (pic_t p, crop_t c)
{
	uint32_t *b = (uint32_t *)p.b;
	int x, y;

	for (y = c.y; y < c.y + c.h && y < p.h; y++)
		for (x = c.x; x < c.x + c.w && x < p.w; x++)
			if (b[x + p.s * y])
				return -1;

	return 0;
}

#define GRID_BLOCKS 24 /* GCD of 480, 576, 720, 1080 */

static int line_ok (int grid[GRID_BLOCKS + 1][GRID_BLOCKS + 1], int x, int y, int w)
{
	int i;

	if (grid[y][MAX(0, x - 1)] == -1)
		return 0;

	for (i = 0; i < w; i++)
		if (grid[y][x + i] != -1)
			return 0;

	if (grid[y][MIN(GRID_BLOCKS, x + i)] == -1)
		return 0;

	return 1;
}

static void set_line (int grid[GRID_BLOCKS + 1][GRID_BLOCKS + 1], int x, int y, int w, int n)
{
	int i;

	for (i = 0; i < w; i++)
		grid[y][x + i] = n;
}

static rect_t make_rect (int grid[GRID_BLOCKS + 1][GRID_BLOCKS + 1], int x, int y, int n)
{
	rect_t r = {x, y, 1, 1};
	int line_length = 1;

	/* Get length of first rectangle line, and assign rect number */
	grid[y][x] = n;
	while (line_length + x < GRID_BLOCKS + 1 && grid[y][x + line_length] == -1)
	{
		grid[y][x + line_length] = n;
		line_length++;
	}
	r.w = line_length;

	/* Add lines while available */
	while (line_ok(grid, x, r.y + r.h, r.w))
		set_line(grid, x, r.y + r.h++, r.w, n);

	return r;
}

rect_t merge_rects (rect_t r1, rect_t r2)
{
	rect_t r;

	/* Set rectangle that covers both input rectangles */
	r.x = MIN(r1.x, r2.x);
	r.y = MIN(r1.y, r2.y);
	r.w = MAX(r1.x + r1.w, r2.x + r2.w) - r.x;
	r.h = MAX(r1.y + r1.h, r2.y + r2.h) - r.y;

	return r;
}

int score_rect (rect_t r)
{
	return r.w * r.h;
}

static int check_close (rect_t r1, rect_t r2, int distance)
{
	rect_t a, b;

	if (r1.x <= r2.x)
	{
		a = r1;
		b = r2;
	}
	else
	{
		a = r2;
		b = r1;
	}

	/* The rectangles are not too close, if the left border of of the right-most
	 * rectangle has a distance of at least distance to the right border of the
	 * left-most rectangle.
	 */
	if (b.x > a.x + a.w + distance)
		return 0;

	if (r1.y <= r2.y)
	{
		a = r1;
		b = r2;
	}
	else
	{
		a = r2;
		b = r1;
	}

	/* Same as before, except vertically */
	if (b.y > a.y + a.h + distance)
		return 0;

	/* Too close! */
	return 1;
}

/* crop_t *c - Array of length 2 */
int auto_split (pic_t p, crop_t *c, int ugly, int even_y)
{
	crop_t c1 = {0, 0, 0, 0};
	crop_t c2 = {0, 0, 0, 0};
	crop_t null = {0, 0, 0, 0};
	crop_t t;
	rect_t rects[(GRID_BLOCKS + 1) * (GRID_BLOCKS + 1)];
	rect_t r1 = {0};
	rect_t r2 = {0};
	rect_t rt1, rt2;
	int grid[GRID_BLOCKS + 1][GRID_BLOCKS + 1];
	int score_t1, score_t2, score_r1 = 0, score_r2 = 0;
	int n_rect = 0;
	int score = 0;
	int bw, bh;
	int x, y;
	int i, j;
	int n_res;

	/* Initialize grid */
	memset(grid, 0, sizeof(int) * (GRID_BLOCKS + 1) * (GRID_BLOCKS + 1));
	bw = p.w / GRID_BLOCKS;
	bh = p.h / GRID_BLOCKS;

	/* Ensure block height is even, if even_y is enabled */
	if (even_y && (bh % 2))
		bh--;

	/* Ensure block dimensions are not zero */
	if (!bw)
		bw = 1;
	if (!bh && even_y)
		bh = 2;
	else if (!bh)
		bh = 1;

	/* Determine state of blocks */
	t.w = bw;
	t.h = bh;
	for (y = 0; (t.y = y * bh) < p.h; y++)
		for (x = 0; (t.x = x * bw) < p.w; x++)
			grid[y][x] = block_state(p, t);

	/* Create rectangles */
	for (y = 0; y < GRID_BLOCKS + 1; y++)
		for (x = 0; x < GRID_BLOCKS + 1; x++)
			if (grid[y][x] == -1)
			{
				rects[n_rect] = make_rect(grid, x, y, n_rect);
				n_rect++;
			}

	/* Shouldn't happen, empty frame */
	if (!n_rect)
	{
		c[0] = c1;
		c[1] = c2;
		return 0;
	}

	/* Single rectangle */
	if (n_rect == 1)
	{
		c1.x = rects[0].x * bw;
		c1.y = rects[0].y * bh;
		c1.w = rects[0].w * bw;
		c1.h = rects[0].h * bh;
		auto_crop(p, &c1);
		c[0] = c1;
		c[1] = c2;
		return 1;
	}

	/* Two rectangles */
	n_res = 2;

	/* Any other number of rectangles, first find most "distant" ones */
	for (i = 0; i < n_rect; i++)
		for (j = 0; j < n_rect; j++)
			if (i == j)
				continue;
			else
			{
				rt1 = merge_rects(rects[i], rects[j]);
				score_t1 = score_rect(rt1);
				if (score <= score_t1)
				{
					score = score_t1;
					r1 = rects[i];
					r2 = rects[j];
					score_r1 = score_rect(r1);
					score_r2 = score_rect(r2);
				}
			}

	/* Merge all other rectangles with the "nearest" one */
	for (i = 0; i < n_rect; i++)
	{
		rt1 = merge_rects(r1, rects[i]);
		rt2 = merge_rects(r2, rects[i]);
		score_t1 = score_rect(rt1);
		score_t2 = score_rect(rt2);
		if (score_t1 - score_r1 < score_t2 - score_r2)
		{
			r1 = rt1;
			score_r1 = score_t1;
		}
		else
		{
			r2 = rt2;
			score_r2 = score_t2;
		}
	}

	/* Turn rectangles into crops */
	c1.x = r1.x * bw;
	c1.y = r1.y * bh;
	c1.w = r1.w * bw;
	c1.h = r1.h * bh;
	c2.x = r2.x * bw;
	c2.y = r2.y * bh;
	c2.w = r2.w * bw;
	c2.h = r2.h * bh;

	/* Minimize surfaces and return them */
	auto_crop(p, &c1);
	auto_crop(p, &c2);

	/* Merge in rare cases of closeness or overlap */
	if ((!ugly && check_close(c1, c2, 0)) || (ugly && check_close(c1, c2, -1)))
	{
		c1 = merge_rects(c1, c2);
		c2 = null;
		auto_crop(p, &c1);
		n_res = 1;
	}
	else if (!ugly)
	{
		/* Check whether split is ugly due to small gains */
		rt1 = merge_rects(c1, c2);
		score_t1 = score_rect(rt1);
		score_t2 = score_rect(c1) + score_rect(c2);

		/* Merge if area taken by the merged rectangle is less than 1.5 * sum of
		 * split rectangles and the difference is below a hard limit.
		 */
		if ((score_t1 < 3 * score_t2 / 2) && (score_t1 - score_t2 < 500 * 300))
		{
			c1 = merge_rects(c1, c2);
			c2 = null;
			auto_crop(p, &c1);
			n_res = 1;
		}
	}

	c[0] = c1;
	c[1] = c2;
	return n_res;
}

/* Find two minimal non-overlapping windows covering the given rectangles. */

typedef struct window_s
{
	rect_t r;  /* Position and size. */
	int s;     /* Score. */
} window_t;

typedef struct interval_s
{
	rect_t r;
	int start, end;
} interval_t;

STATIC_LIST(interval, interval_t)

int cmp_rect_x (rect_t *a, rect_t *b)
{
	return a->x > b->x;
}

int cmp_rect_y (rect_t *a, rect_t *b)
{
	return a->y > b->y;
}

void rect_bounds_x (rect_t *c, int *x, int *w)
{
	*x = c->x;
	*w = c->w;
}

void rect_bounds_y (rect_t *c, int *y, int *h)
{
	*y = c->y;
	*h = c->h;
}

/* The windows argument must point to 2 * sizeof(rect_t) allocated memory. */
int find_windows (rect_t *rects, int n_rects, rect_t *windows)
{
	void (*rect_bounds)(rect_t *, int *, int *);
	rect_t **sorted     = malloc(sizeof(void *) * n_rects);
	void *work[2][2]    = { {rect_bounds_x, cmp_rect_x}
	                      , {rect_bounds_y, cmp_rect_y}
	                      };
	interval_list_t *segs;
	interval_t *iv = NULL;
	window_t *fwd, *bwd;
	rect_t best[2], tmp;
	int i, dir, edge, a, b, n_ivs, found;
	int score = -1;

	if (!n_rects)
	{
		free(sorted);
		return 0;
	}

	memset(best, 0, 2 * sizeof(rect_t));
	found = 0;

	for (dir = 0; dir < 2; dir++)
	{
		/* Sort rectangles from left/top to right/bottom. */
		for (i = 0; i < n_rects; i++)
			sorted[i] = &(rects[i]);
		sort((sort_func_t)work[dir][1], (void **)sorted, n_rects);

		/* Group overlapping rectangles. */
		segs  = interval_list_new();
		n_ivs = 0;
		edge  = -1;

		rect_bounds = work[dir][0];
		for (i = 0; i < n_rects; i++)
		{
			rect_bounds(sorted[i], &a, &b);
			if (a < edge)
			{
				iv->end = MAX(iv->end, a + b);
				iv->r   = merge_rects(iv->r, *(sorted[i]));
				edge    = iv->end;
			}
			else
			{
				iv = malloc(sizeof(interval_t));
				iv->start = a;
				iv->end   = a + b;
				iv->r     = *(sorted[i]);
				interval_list_insert_after(segs, iv);
				n_ivs++;
				edge = iv->end;
			}
		}

		if (!n_ivs)
		{
			interval_list_destroy_deep(segs);
			continue;
		}

		/* Sum up scores for interval 0-n and vice versa. */
		fwd = malloc(sizeof(window_t) * n_ivs);
		bwd = malloc(sizeof(window_t) * n_ivs);

		iv = interval_list_first(segs);
		/* Initialize first candidate. */
		fwd[0].r = iv->r;
		fwd[0].s = score_rect(fwd[0].r);
		/* Merge/sum up the rest. */
		for (i = 1; i < n_ivs && iv != NULL; i++)
		{
			iv = interval_list_next(segs);
			fwd[i].r = merge_rects(fwd[i - 1].r, iv->r);
			fwd[i].s = score_rect(fwd[i].r);
		}

		iv = interval_list_last(segs);
		/* Corresponding to a "merged all" forward window is a null window. */
		memset(&(bwd[n_ivs - 1]), 0, sizeof(window_t));
		/* Initialize first real candidate if it exists. */
		if (n_ivs > 1)
		{
			bwd[n_ivs - 2].r = iv->r;
			bwd[n_ivs - 2].s = score_rect(bwd[n_ivs - 2].r);
		}
		/* Merge/sum up the rest. */
		for (i = n_ivs - 3; i >= 0 && iv != NULL; i--)
		{
			iv = interval_list_prev(segs);
			bwd[i].r = merge_rects(bwd[i + 1].r, iv->r);
			bwd[i].s = score_rect(bwd[i].r);
		}

		/* Find best pair of two windows. */
		for (i = 0; i < n_ivs; i++)
			if (fwd[i].s + bwd[i].s < score || score == -1)
			{
				score = fwd[i].s + bwd[i].s;
				best[0] = fwd[i].r;
				best[1] = bwd[i].r;
			}

		/* Cleanup. */
		free(bwd);
		free(fwd);
		interval_list_destroy_deep(segs);
	}

	/* Is any of the best rectangles not null? */
	if ((best[0].w != 0 && best[0].h != 0) || (best[1].w != 0 && best[1].h != 0))
	{
		/* We have at least one. */
		found = 1;

		/* If the first one is null, set it to the rect of the second and null that one. */
		if (best[0].w == 0 || best[0].h == 0)
		{
			best[0] = best[1];
			memset(&(best[1]), 0, sizeof(rect_t));
		}
		/* If the second rect is still not null... */
		if (best[1].w != 0 && best[1].h != 0)
		{
			/* We found two. */
			found = 2;
			/* Let's order them, so the one closer to 0/0 is the second. */
			if (best[1].y > best[0].y || (best[1].y == best[0].y && best[1].x > best[0].x))
			{
				tmp = best[0];
				best[0] = best[1];
				best[1] = tmp;
			}
		}
		memcpy(windows, best, 2 * sizeof(rect_t));
	}

	/* Cleanup. */
	free(sorted);

	return found;
}

void enforce_even_y (crop_t *c, int n)
{
	int mod;

	if (!n)
		return;
	/* Since block borders used to auto split always lie on even rows, y will
	 * only be odd when it was shrunk to below a block border, meaning expanding it
	 * back up by one row should be harmless and never lead to overlap.
	 */
	mod = c[0].y % 2;
	c[0].y -= mod;
	c[0].h += mod;
	if (n > 1)
	{
		mod = c[1].y % 2;
		c[1].y -= mod;
		c[1].h += mod;
	}
}

