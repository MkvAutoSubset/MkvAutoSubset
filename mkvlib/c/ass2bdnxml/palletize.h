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

#ifndef QUANTIZE_H
#define QUANTIZE_H

/* Return malloced palette and overwrite im with 8bpp data */
uint32_t *palletize(uint8_t *im, int w, int h);

#endif

