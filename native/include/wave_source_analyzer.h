#ifndef WAVE_SOURCE_ANALYZER_H
#define WAVE_SOURCE_ANALYZER_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

int32_t wave_source_analyzer_abi_version(void);

enum wave_source_token_kind {
    WAVE_SOURCE_TOKEN_PLAIN = 0,
    WAVE_SOURCE_TOKEN_KEYWORD = 1,
    WAVE_SOURCE_TOKEN_TYPE = 2,
    WAVE_SOURCE_TOKEN_STRING = 3,
    WAVE_SOURCE_TOKEN_COMMENT = 4,
    WAVE_SOURCE_TOKEN_NUMBER = 5
};

int32_t wave_source_analyzer_highlight(
    const uint8_t *source,
    int64_t source_length,
    uint8_t *out_kinds,
    int64_t out_length
);

#ifdef __cplusplus
}
#endif

#endif
