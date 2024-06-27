/*----------------------------------------------------------------------------
 * sort.c - Generic heapsort
 * Copyright (C) 2011 Arne Bochem <heapsort at ps-auxw de>
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

/* Returns 1 if a > b, 0 otherwise. */
typedef int (*sort_func_t)(void *, void *);

static inline uint32_t depth_log (uint32_t i)
{
	uint32_t j;

#if defined(__i386__)
	asm __volatile__ ("bsr %1,%0\n":"=r"(j):"r"(i));
#else
	j = 0;
	while (i >>= 1)
		j++;
#endif

	return j;
}

static inline void swap (void **data, uint32_t i, uint32_t j)
{
	static void *tmp;

	tmp     = data[i];
	data[i] = data[j];
	data[j] = tmp;
}

static inline void reheap (sort_func_t cmp, void **data, uint32_t k, uint32_t hs)
{
	uint32_t max_son;
	uint32_t l, r;

	for (;;)
	{
		if (hs <= (l = 2 * k + 1))
			/* Return if k has no sons. */
			return;
		else
			if (hs > (r = l + 1) && cmp(data[r], data[l]))
				/* Set max_son to right son if k has a right son and it is greater than the left son. */
				max_son = r;
			else
				/* If there is no right son, or it is smaller, set max_son to left son. */
				max_son = l;

		/* Return if k is greater than max_son. */
		if (!cmp(data[max_son], data[k]))
			return;

		/* Otherwise swap k and max_son, and recurse. */
		swap(data, k, max_son);

		/* "Recurse." */
		k = max_son;
	}
}

static inline void bottom_up (sort_func_t cmp, void **data, uint32_t *path, uint32_t hs)
{
	uint32_t depth = depth_log(hs);
	uint32_t i, l, r;
	uint32_t k = 0;
	void *tmp;

	/* Check if there's anything left. */
	if (!depth)
		return;
	
	for (i = 1; i <= depth; i++)
	{
		/* Find maximum son. */
		if (hs <= (l = 2 * k + 1))
		{
			/* Terminate path early, if no son found. */
			depth--;
			break;
		}
		/* See reheap function. Get maximum son and add to path. */
		if (hs > (r = l + 1) && cmp(data[r], data[l]))
			k = path[i] = r;
		else
			k = path[i] = l;
	}

	/* Find new position for heap root. */
	for (i = depth; i > 0; i--)
		if (cmp(data[path[i]], data[0]))
			break;

	/* Move values along path. */
	tmp = data[0];
	for (k = 1; k <= i; k++)
		data[path[k - 1]] = data[path[k]];
	data[path[i]] = tmp;
}

void sort (sort_func_t cmp, void **data, uint32_t len)
{
	uint32_t path[33]; /* Path of maximum sons. */
	uint32_t hs = len; /* Heapsize. */
	int i;

	/* Small arrays are always sorted. */
	if (len < 2)
		return;

	/* Build heap. */
	for (i = len / 2 - 1; i >= 0; i--)
		reheap(cmp, data, i, hs);

	/* Initialize start point of the path as the heap root. */
	path[0] = 0;

	while (hs >= 2)
	{
		/* Swap last item with root. */
		swap(data, 0, hs - 1);

		/* Rebuild heap structure. */
		/*bottom_up(cmp, data, path, --hs);*/ /* Faster for bigger arrays. */
		reheap(cmp, data, 0, --hs); /* Faster for smaller arrays, such as here. */
	}
}
