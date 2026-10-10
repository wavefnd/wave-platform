#include "wave_source_analyzer.h"

#include <stdint.h>
#include <stdio.h>
#include <string.h>

static int check(const char *source, const char *fragment, uint8_t expected) {
    uint8_t kinds[4096] = {0};
    size_t length = strlen(source);
    const char *match = strstr(source, fragment);
    if (!match || length > sizeof(kinds) ||
        wave_source_analyzer_highlight((const uint8_t *)source, (int64_t)length,
                                       kinds, sizeof(kinds)) != 0) return 1;
    for (size_t i = (size_t)(match - source); i < (size_t)(match - source) + strlen(fragment); i++) {
        if (kinds[i] != expected) {
            fprintf(stderr, "highlight: %s at byte %zu: expected %u, got %u\n", fragment, i, expected, kinds[i]);
            return 1;
        }
    }
    return 0;
}

int main(void) {
    if (wave_source_analyzer_abi_version() != 2) return 1;
    const char *source =
        "#[target(os = \"linux\")]\npub async fun work(value: ptr<array<u1024, 4>>) -> ! { await next(); }\n"
        "variant Result { Value(isz), Error(usz) }\n"
        "const size: byte = 0o7_55; static other: i512 = 0xFE-1; var value: f64 = 1.25e-3;\n"
        "/* outer /* inner */ var still_comment */ var name: str = \"\\\" // <tag>\"; // comment\rawait work();";
    const char *keywords[] = {"pub", "async", "fun", "await", "variant", "const", "static", "var name"};
    for (size_t i = 0; i < sizeof(keywords) / sizeof(*keywords); i++) {
        // Test the keyword alone for the disambiguated var occurrence.
        const char *word = keywords[i];
        if (strcmp(word, "var name") == 0) {
            if (check(strstr(source, word), "var", WAVE_SOURCE_TOKEN_KEYWORD)) return 2;
        } else if (check(source, word, WAVE_SOURCE_TOKEN_KEYWORD)) return 2;
    }
    const char *types[] = {"ptr", "array", "u1024", "isz", "usz", "byte", "i512", "f64", "str"};
    for (size_t i = 0; i < sizeof(types) / sizeof(*types); i++)
        if (check(source, types[i], WAVE_SOURCE_TOKEN_TYPE)) return 3;
    if (check(source, "0o7_55", WAVE_SOURCE_TOKEN_NUMBER) ||
        check(source, "0xFE", WAVE_SOURCE_TOKEN_NUMBER) ||
        check(source, "1.25e-3", WAVE_SOURCE_TOKEN_NUMBER) ||
        check("0xFE-1", "-", WAVE_SOURCE_TOKEN_PLAIN) ||
        check(source, "var still_comment", WAVE_SOURCE_TOKEN_COMMENT) ||
        check(source, "<tag>", WAVE_SOURCE_TOKEN_STRING) ||
        check("// comment\rawait work()", "await", WAVE_SOURCE_TOKEN_KEYWORD) ||
        check("\"unfinished\\\nvar next", "var", WAVE_SOURCE_TOKEN_KEYWORD) ||
        check("/* unfinished /* nested */", "nested */", WAVE_SOURCE_TOKEN_COMMENT)) return 4;
    const char *names = "async_task awaitable pub123 var한글 한글var item42 i1024_suffix";
    if (check(names, names, WAVE_SOURCE_TOKEN_PLAIN)) return 5;
    uint8_t output[1] = {99};
    if (wave_source_analyzer_highlight((const uint8_t *)"fun", 3, output, 1) != -1 || output[0] != 99) return 6;
    return 0;
}
