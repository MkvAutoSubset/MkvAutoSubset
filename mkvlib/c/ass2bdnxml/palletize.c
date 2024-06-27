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
 *----------------------------------------------------------------------------
 * Inspired by the OctTree code by Jerry Huxtable and 0xdeadbeef, thank you.
 *----------------------------------------------------------------------------*/

#include <stdlib.h>
#include <stdint.h>
#include "abstract_lists.h"

#define LEVELS 5
#define COLORS 254 /* One reserved for 100% transparent */

typedef struct hexnode_s hexnode_t;
struct hexnode_s
{
	unsigned int v[4]; 
	hexnode_t *nodes[16];
	int children;
	int leaf;
	int count;
	int index;
};

static hexnode_t *new_hexnode ()
{
	hexnode_t *n = calloc(1, sizeof(hexnode_t));
	int i;

	for (i = 0; i < 4; i++)
	{
		n->v[i] = 0;
		n->nodes[i] = NULL;
	}
	for (i = 4; i < 16; i++)
		n->nodes[i] = NULL;
	n->children = 0;
	n->leaf = 0;
	n->count = 0;
	n->index = 0;

	return n;
}

static void tree_destroy (hexnode_t *n)
{
	int i;

	for (i = 0; i < 16; i++)
		if (n->nodes[i] != NULL)
			tree_destroy(n->nodes[i]);

	free(n);
}

STATIC_LIST(pal, hexnode_t)

typedef struct quantizer_s
{
	hexnode_t *root;
	pal_list_t *levels[LEVELS + 1];
	int colors;
	int nodes;
} quantizer_t;

static quantizer_t *new_quantizer ()
{
	quantizer_t *q = calloc(sizeof(quantizer_t), 1);
	int i;

	q->root = new_hexnode();
	for (i = 0; i <= LEVELS; i++)
		q->levels[i] = pal_list_new();

	return q;
}

static void destroy_quantizer (quantizer_t *q)
{
	int i;

	if (q->root != NULL)
		tree_destroy(q->root);
	for (i = 0; i <= LEVELS; i++)
		pal_list_destroy(q->levels[i]);

	free(q);
}

static int exec_find_node (hexnode_t *n, uint32_t color, hexnode_t **found, hexnode_t **last, int *index, int *level)
{
	int pows[4] = {1, 2, 4, 8};
	uint8_t *v = (uint8_t *)&color;
	int i;
	int idx;

	for (; *level <= LEVELS; (*level)++)
	{
		idx = 0;
		for (i = 0; i < 4; i++)
			idx += pows[i] * !!(v[i] & (0x80 >> *level));

		*last = n;
		*index = idx;
		if ((*found = n->nodes[idx]) == NULL)
			return 0;
		else if ((*found)->leaf)
			return 1;
		else
			n = *found;
	}
	return 0;
}

static int find_node (hexnode_t *n, uint32_t color, hexnode_t **found, hexnode_t **last, int *index, int *level)
{
	hexnode_t *f, *l;
	int i, r;
	int lv = level == NULL ? 0 : *level;

	r = exec_find_node(n, color, &f, &l, &i, &lv);
	if (found != NULL)
		*found = f;
	if (last != NULL)
		*last = l;
	if (index != NULL)
		*index = i;
	if (level != NULL)
		*level = lv;

	return r;
}

static int get_color_index (quantizer_t *q, uint32_t color)
{
	hexnode_t *f, *l;

	if (!color)
		return 0;

	if (find_node(q->root, color, &f, &l, NULL, NULL))
		return f->index;
	else
		return l->index;
}

static void reduce (quantizer_t *q)
{
	pal_list_t *l;
	hexnode_t *n, *c;
	int i, j, k;

	if (q->colors <= COLORS)
		return;

	for (i = LEVELS - 1; i >= 0; i--)
	{
		l = q->levels[i];
		if (pal_list_empty(l))
			continue;
		n = pal_list_first(l);
		do
		{
			if (!n->children)
				continue;
			for (j = 0; j < 16; j++)
				if ((c = n->nodes[j]) != NULL)
				{
					n->count += c->count;
					/* UHD compliance on 32bit arch */
					if (n->count >= 11480800)
					{
						n->count >>= 1;
						for (k = 0; k < 4; k++)
							n->v[k] >>= 1;
					}
					n->children--;
					for (k = 0; k < 4; k++)
						n->v[k] += c->v[k];
					n->nodes[j] = NULL;
					q->colors--;
					q->nodes--;
					pal_list_remove(q->levels[i+1], c);
					tree_destroy(c);
				}
			n->leaf = 1;
			q->colors++;
			if (q->colors <= COLORS)
				return;
		} while ((n = pal_list_next(l)) != NULL);
	}
}

static void insert_color (quantizer_t *q, uint32_t color)
{
	uint8_t *v = (uint8_t *)&color;
	hexnode_t *n = q->root;
	hexnode_t *f, *l;
	int i, j, level = 0;

	/* 100% transparent pixels will be ignored */
	if (!color)
		return;

	while (level <= LEVELS)
	{
		if (find_node(n, color, &f, &l, &i, &level))
		{
			f->count++;
			for (j = 0; j < 4; j++)
				f->v[j] += v[j];
			return;
		}
		else
		{
			f = new_hexnode();
			l->children++;
			l->nodes[i] = f;
			level++;

			q->nodes++;
			pal_list_insert(q->levels[level], f);

			if (level == LEVELS)
			{
				f->leaf = 1;
				f->count = 1;
				for (j = 0; j < 4; j++)
					f->v[j] = v[j];
				q->colors++;
				return;
			}

			n = f;
		}
	}

	if (q->colors > COLORS)
		reduce(q);
}

static int recursive_get_palette (hexnode_t *n, uint32_t pal[COLORS + 1], int index)
{
	uint8_t v[4];
	int i;

	if (n->leaf)
	{
		for (i = 0; i < 4; i++)
			v[i] = n->v[i] / n->count;
		pal[n->index = index++] = *(uint32_t *)v;
	}
	else
		for (i = 0; i < 16; i++)
			if (n->nodes[i] != NULL)
			{
				n->index = index;
				index = recursive_get_palette(n->nodes[i], pal, index);
			}

	return index;
}

static void get_palette (quantizer_t *q, uint32_t pal[COLORS + 1])
{
	int index;

	pal[0] = 0;
	if (q->colors > COLORS)
		reduce(q);
	index = recursive_get_palette(q->root, pal, 1);

	if (index <= COLORS && !pal[index])
		pal[index] = 0xc0decafe;
}

uint32_t *palletize (uint8_t *im, int w, int h)
{
	uint32_t *pal = calloc(256, sizeof(uint32_t));
	uint32_t *i = (uint32_t *)im;
	quantizer_t *q = new_quantizer();
	int x, y;

	for (y = 0; y < h; y++)
		for (x = 0; x < w; x++)
			insert_color(q, i[x + y * w]);

	get_palette(q, pal);

	for (y = 0; y < h; y++)
		for (x = 0; x < w; x++)
			im[x + y * w] = get_color_index(q, i[x + y * w]);

	destroy_quantizer(q);

	return pal;
}

