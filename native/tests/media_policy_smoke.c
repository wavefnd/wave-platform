#include "wave_media_policy.h"

int main(void) {
    int64_t width = 0;
    int64_t height = 0;
    if (wave_media_policy_abi_version() != 1) {
        return 1;
    }
    if (wave_media_policy_plan(4000, 2000, 1048576, 1, &width, &height) != 0) {
        return 2;
    }
    if (width != 1920 || height != 960) {
        return 3;
    }
    if (wave_media_policy_plan(12000, 12000, 1048576, 2, &width, &height) != -4) {
        return 4;
    }
    return 0;
}
