#ifdef _WIN32
#include <windows.h>
#endif

#include <stdio.h>
#include <stdlib.h>

void s(char **utf8) {
#ifdef _WIN32
    int wlen = MultiByteToWideChar(CP_UTF8, 0, *utf8, -1, NULL, 0);
    if (wlen == 0) {
        return;
    }

    wchar_t *wstr = (wchar_t *)malloc(wlen * sizeof(wchar_t));
    if (wstr == NULL) {
        return;
    }

    MultiByteToWideChar(CP_UTF8, 0, *utf8, -1, wstr, wlen);

    int clen = WideCharToMultiByte(CP_ACP, 0, wstr, -1, NULL, 0, NULL, NULL);
    if (clen == 0) {
        free(wstr);
        return;
    }

    char *cstr = (char *)malloc(clen * sizeof(char));
    if (cstr == NULL) {
        free(wstr);
        return;
    }

    WideCharToMultiByte(CP_ACP, 0, wstr, -1, cstr, clen, NULL, NULL);
    free(wstr);
    *utf8 = cstr;
#endif
}