#include "lcb.h"

void makeLogCallback(char* s, logCallback lcb)
{
    if(lcb)
        lcb(s);
}