#ifndef WAVE_MEDIA_POLICY_H
#define WAVE_MEDIA_POLICY_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

int32_t wave_media_policy_abi_version(void);

int32_t wave_media_policy_plan(
    int64_t width,
    int64_t height,
    int64_t input_bytes,
    int32_t format,
    int64_t *out_width,
    int64_t *out_height
);

#ifdef __cplusplus
}
#endif

#endif
