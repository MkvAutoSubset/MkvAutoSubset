/*----------------------------------------------------------------------------
 * abstract_lists.h - Simple, typesafe, doubly linked lists for C
 * Copyright (C) 2010 Arne Bochem <abstract.lists at ps-auxw de>
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

#ifndef ABSTRACT_LISTS_H
#define ABSTRACT_LISTS_H

#include <stdlib.h>

#define DECLARE_LIST_BACKEND(kind, prefix, type) \
typedef struct prefix##_list_s prefix##_list_t;\
typedef struct prefix##_list_node_s prefix##_list_node_t;\
struct prefix##_list_s\
{\
	prefix##_list_node_t *current, *first, *last;\
};\
struct prefix##_list_node_s\
{\
	type *v;\
	prefix##_list_node_t *prev, *next;\
};\
kind prefix##_list_t *prefix##_list_new ();\
kind type *prefix##_list_get (prefix##_list_t *l) __attribute__ ((unused));\
kind type *prefix##_list_next (prefix##_list_t *l) __attribute__ ((unused));\
kind type *prefix##_list_prev (prefix##_list_t *l) __attribute__ ((unused));\
kind type *prefix##_list_first (prefix##_list_t *l) __attribute__ ((unused));\
kind type *prefix##_list_last (prefix##_list_t *l) __attribute__ ((unused));\
kind int prefix##_list_empty (prefix##_list_t *l) __attribute__ ((unused));\
kind void prefix##_list_set (prefix##_list_t *l, type *v) __attribute__ ((unused));\
kind void prefix##_list_insert (prefix##_list_t *l, type *v) __attribute__ ((unused));\
kind void prefix##_list_insert_after (prefix##_list_t *l, type *v) __attribute__ ((unused));\
kind void prefix##_list_delete (prefix##_list_t *l) __attribute__ ((unused));\
kind void prefix##_list_remove (prefix##_list_t *l, type *v) __attribute__ ((unused));\
kind void prefix##_list_destroy (prefix##_list_t *l) __attribute__ ((unused));\
kind void prefix##_list_destroy_deep (prefix##_list_t *l) __attribute__ ((unused));

#define IMPLEMENT_LIST_BACKEND(kind, prefix, type) \
kind prefix##_list_t *prefix##_list_new ()\
{\
	prefix##_list_t *l = malloc(sizeof(prefix##_list_t));\
	l->first = NULL;\
	l->last = NULL;\
	l->current = NULL;\
	return l;\
}\
kind type *prefix##_list_get (prefix##_list_t *l)\
{\
	if (l->current == NULL)\
		return NULL;\
	return l->current->v;\
}\
kind type *prefix##_list_next (prefix##_list_t *l)\
{\
	if (l->current != NULL)\
		l->current = l->current->next;\
	if (l->current != NULL)\
		return l->current->v;\
	return NULL;\
}\
kind type *prefix##_list_prev (prefix##_list_t *l)\
{\
	if (l->current != NULL)\
		l->current = l->current->prev;\
	if (l->current != NULL)\
		return l->current->v;\
	return NULL;\
}\
kind type *prefix##_list_first (prefix##_list_t *l)\
{\
	l->current = l->first;\
	if (l->current == NULL)\
		return NULL;\
	return l->current->v;\
}\
kind type *prefix##_list_last (prefix##_list_t *l)\
{\
	l->current = l->last;\
	if (l->current == NULL)\
		return NULL;\
	return l->current->v;\
}\
kind int prefix##_list_empty (prefix##_list_t *l)\
{\
	return (l->first == NULL && l->last == NULL);\
}\
kind void prefix##_list_set (prefix##_list_t *l, type *v)\
{\
	if (l->current != NULL)\
		l->current->v = v;\
}\
kind void prefix##_list_insert (prefix##_list_t *l, type *v)\
{\
	prefix##_list_node_t *new;\
	new = calloc(1, sizeof(prefix##_list_node_t));\
	new->v = v;\
	new->next = l->current;\
	new->prev = NULL;\
	if (l->first == l->current)\
		l->first = new;\
	if (l->current != NULL)\
	{\
		new->prev = l->current->prev;\
		if (l->current->prev != NULL)\
			l->current->prev->next = new;\
		l->current->prev = new;\
	}\
	l->current = new;\
	if (l->last == NULL)\
		l->last = new;\
}\
kind void prefix##_list_insert_after (prefix##_list_t *l, type *v)\
{\
	prefix##_list_node_t *new;\
	new = calloc(1, sizeof(prefix##_list_node_t));\
	new->v = v;\
	new->next = NULL;\
	new->prev = l->current;\
	if (l->last == l->current)\
		l->last = new;\
	if (l->current != NULL)\
	{\
		new->next = l->current->next;\
		if (l->current->next != NULL)\
			l->current->next->prev = new;\
		l->current->next = new;\
	}\
	l->current = new;\
	if (l->first == NULL)\
		l->first = new;\
}\
kind void prefix##_list_delete (prefix##_list_t *l)\
{\
	prefix##_list_node_t *node = l->current;\
	if (node == NULL)\
		return;\
	if (node->prev != NULL)\
		node->prev->next = node->next;\
	if (node->next != NULL)\
		node->next->prev = node->prev;\
	l->current = node->next;\
	if (l->first == node)\
		l->first = node->next;\
	if (l->last == node)\
		l->last = node->prev;\
	free(node);\
}\
kind void prefix##_list_remove (prefix##_list_t *l, type *v)\
{\
	type *c = prefix##_list_first(l);\
	do\
	{\
		if (c == v)\
		{\
			prefix##_list_delete(l);\
			l->current = l->first;\
			return;\
		}\
		c = prefix##_list_next(l);\
	}\
	while (c != NULL);\
}\
kind void prefix##_list_destroy (prefix##_list_t *l)\
{\
	prefix##_list_first(l);\
	while (!prefix##_list_empty(l))\
		prefix##_list_delete(l);\
	free(l);\
}\
kind void prefix##_list_destroy_deep (prefix##_list_t *l)\
{\
	prefix##_list_first(l);\
	while (!prefix##_list_empty(l))\
	{\
		free(l->current->v);\
		prefix##_list_delete(l);\
	}\
	free(l);\
}

#define DECLARE_LIST(prefix, type) DECLARE_LIST_BACKEND(;, prefix, type)
#define IMPLEMENT_LIST(prefix, type) IMPLEMENT_LIST_BACKEND(;, prefix, type)
#define DECLARE_STATIC_LIST(prefix, type) DECLARE_LIST_BACKEND(static, prefix, type)
#define IMPLEMENT_STATIC_LIST(prefix, type) IMPLEMENT_LIST_BACKEND(static, prefix, type)
#define STATIC_LIST(prefix, type) DECLARE_STATIC_LIST(prefix, type) IMPLEMENT_STATIC_LIST(prefix, type)

#endif
