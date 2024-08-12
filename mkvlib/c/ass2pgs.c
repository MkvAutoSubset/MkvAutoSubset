#include "ass2bdnxml/ass2bdnxml.c"
#include "ass2bdnxml/sup.c"
#include "ass2bdnxml/auto_split.c"
#include "ass2bdnxml/palletize.c"
#include "ass2bdnxml/sort.c"

#include <stdlib.h>
#include <stdbool.h>

void s(char**);
bool ass2pgs(char *ass, const char *resolution, const char *rate, char *fontdir, char *output) {
    s(&ass);
    s(&fontdir);
    s(&output);

    int argc = 15;

    char **argv = (char **)malloc(argc * sizeof(char *));
    if (!argv) {
        return false;
    }

    // Set constant arguments
    argv[0] = NULL;
    argv[1] = "-a0";
    argv[2] = "-p1";
    argv[3] = "-z0";
    argv[4] = "-u0";
    argv[5] = "-b0";
    argv[6] = "-g";
    argv[7] = fontdir;
    argv[8] = "-v";
    argv[9] = resolution;
    argv[10] = "-f";
    argv[11] = rate;
    argv[12] = "-o";
    argv[13] = output;
    argv[14] = ass;

    bool success = (app(argc, argv) == 0);

    // Free the argv array itself
    free(argv);

    return success;
}

