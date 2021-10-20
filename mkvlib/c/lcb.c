#include "lcb.h"

void makeLogCallback(char* s, logCallback lcb){ lcb(s); }