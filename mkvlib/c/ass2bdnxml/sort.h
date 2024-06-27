/*----------------------------------------------------------------------------
 * sort.h - Generic heapsort header
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

#ifndef SORT_H
#define SORT_H

#include <stdint.h>

/* Returns 1 if a > b, 0 otherwise. */
typedef int (*sort_func_t)(void *, void *);

void sort (sort_func_t cmp, void **data, uint32_t len);

#endif
