#include "wave_source_analyzer.h"

#include <stdint.h>
#include <string.h>

int main(void) {
    static const char source[] =
        "import(\"std::io\");\n"
        "// fun ignored() {}\n"
        "export(c) fun main() {}\n"
        "struct Result {}\n";
    uint8_t kinds[sizeof(source)] = {0};

    if (wave_source_analyzer_abi_version() != 2) {
        return 1;
    }
    if (wave_source_analyzer_highlight(
            (const uint8_t *)source,
            (int64_t)strlen(source),
            kinds,
            (int64_t)sizeof(kinds)) != 0) {
        return 2;
    }
    if (kinds[0] != WAVE_SOURCE_TOKEN_KEYWORD ||
        kinds[7] != WAVE_SOURCE_TOKEN_STRING ||
        kinds[19] != WAVE_SOURCE_TOKEN_COMMENT ||
        kinds[39] != WAVE_SOURCE_TOKEN_KEYWORD ||
        kinds[49] != WAVE_SOURCE_TOKEN_KEYWORD) {
        return 3;
    }
    return 0;
}
