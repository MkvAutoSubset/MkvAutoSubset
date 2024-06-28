#include "ass2bdnxml/ass2bdnxml.c"
#include "ass2bdnxml/sup.c"
#include "ass2bdnxml/auto_split.c"
#include "ass2bdnxml/palletize.c"
#include "ass2bdnxml/sort.c"

#include <stdbool.h>

bool ass2pgs(const char *ass, const char *resolution, const char *rate, const char *fontdir, const char *output)
{
    // Calculate the number of arguments
    int argc = 15;

    // Allocate memory for argv array
    char **argv = (char **)malloc(argc * sizeof(char *));
    if (!argv) {
        return false;
    }

    // Fill the argv array with arguments
    argv[0] = NULL; // Program name
    argv[1] = "-a1";
    argv[2] = "-p1";
    argv[3] = "-z0";
    argv[4] = "-u0";
    argv[5] = "-b0";
    argv[6] = "-g";
    argv[7] = (char *)fontdir;
    argv[8] = "-v";
    argv[9] = (char *)resolution;
    argv[10] = "-f";
    argv[11] = (char *)rate;
    argv[12] = "-o";
    argv[13] = (char *)output;
    argv[14] = (char *)ass;
    // Call the main function of ass2bdnxml
    int result = app(argc, argv);

    // Free the allocated memory
    free(argv);

    // Check the result
    return result == 0;
}