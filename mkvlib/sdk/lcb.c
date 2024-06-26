#include "lcb.h"

void makeLogCallback(unsigned char l, char* s, logCallback lcb)
{
    if(lcb)
        lcb(l, s);
}