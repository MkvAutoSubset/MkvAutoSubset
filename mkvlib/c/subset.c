#ifdef _WIN32
#include <windows.h>
#endif

#include <harfbuzz/hb.h>
#include <harfbuzz/hb-subset.h>
#include <stdlib.h>
#include <stdio.h>
#include <stdbool.h>

void s(char**);
bool subset(char *oldpath, int idx, char *newpath, const char *newname, const char *dest, const char *txt) {
    s(&oldpath);
    s(&newpath);

    FILE *oldfile = fopen(oldpath, "rb");

    if (!oldfile) return false;
    fseek(oldfile, 0, SEEK_END);
    size_t oldsize = ftell(oldfile);
    fseek(oldfile, 0, SEEK_SET);

    char *olddata = (char *)malloc(oldsize);
    if (fread(olddata, 1, oldsize, oldfile) != oldsize) {
        fclose(oldfile);
        free(olddata);
        return false;
    }
    fclose(oldfile);

    hb_blob_t *blob = hb_blob_create(olddata, oldsize, HB_MEMORY_MODE_READONLY, olddata, free);
    hb_face_t *face = hb_face_create(blob, idx);
    hb_blob_destroy(blob);

    if (idx >= hb_face_count(blob)) {
        free(olddata);
        return false;
    }

    hb_subset_input_t *input = hb_subset_input_create_or_fail();
    if (!input) {
        hb_face_destroy(face);
        return false;
    }

    hb_buffer_t *buf = hb_buffer_create();
    hb_buffer_add_utf8(buf, txt, -1, 0, -1);
    hb_buffer_guess_segment_properties(buf);

    unsigned int glyph_cnt;
    hb_glyph_info_t *glyph_info = hb_buffer_get_glyph_infos(buf, &glyph_cnt);

    hb_set_t *codepoints = hb_subset_input_unicode_set(input);
    for (unsigned int i = 0; i < glyph_cnt; i++) {
        hb_set_add(codepoints, glyph_info[i].codepoint);
    }
    hb_buffer_destroy(buf);

    // Override name tables
    hb_subset_input_override_name_table(input, HB_OT_NAME_ID_COPYRIGHT, 3, 1, 0x409, dest, -1);
    hb_subset_input_override_name_table(input, HB_OT_NAME_ID_FONT_FAMILY, 3, 1, 0x409, newname, -1);
    hb_subset_input_override_name_table(input, HB_OT_NAME_ID_FULL_NAME, 3, 1, 0x409, newname, -1);
    hb_subset_input_override_name_table(input, HB_OT_NAME_ID_POSTSCRIPT_NAME , 3, 1, 0x409, newname, -1);

    // Set flags to drop hints and other non-essential data
    hb_subset_input_set_flags(input, HB_SUBSET_FLAGS_NO_HINTING | HB_SUBSET_FLAGS_DESUBROUTINIZE | HB_SUBSET_FLAGS_OPTIMIZE_IUP_DELTAS);

    hb_face_t *subset_face = hb_subset_or_fail(face, input);
    hb_subset_input_destroy(input);
    hb_face_destroy(face);

    if (!subset_face) return false;

    hb_blob_t *subset_blob = hb_face_reference_blob(subset_face);
    hb_face_destroy(subset_face);

    unsigned int length;
    const char *subset_data = hb_blob_get_data(subset_blob, &length);

    FILE *newfile = fopen(newpath, "wb");

    if (!newfile) {
        hb_blob_destroy(subset_blob);
        return false;
    }

    if (fwrite(subset_data, 1, length, newfile) != length) {
        fclose(newfile);
        hb_blob_destroy(subset_blob);
        return false;
    }
    fclose(newfile);
    hb_blob_destroy(subset_blob);

    return true;
}
