---
translation_set_id: string-bytes
path: reference/string-and-bytes
locale: ko
group: reference
group_order: 3
order: 4
title: 문자열과 바이트
summary: string 하위 모듈, len/is_empty, 비교·검색·trim·ASCII와 endian 바이트 도우미를 설명합니다.
---

## 문자열 모듈 구성

`std/string`은 다음 소스 단위로 구성됩니다.

- `ascii.wave`
- `cmp.wave`
- `find.wave`
- `hash.wave`
- `len.wave`
- `trim.wave`

필요한 기능이 들어 있는 파일을 import합니다.

## 길이와 빈 문자열 확인

`std::string::len`은 `len`과 `is_empty`를 함께 정의합니다.

```wave
import("std::string::len")::{len, is_empty};

fun main() {
    var size: i32 = len("Wave");
    var empty: bool = is_empty("");
    println("{} {}", size, empty);
}
```

`len`은 문자열의 끝을 나타내는 0 바이트 전까지의 길이를 `i32`로 반환합니다. `is_empty`는 첫 바이트가 문자열 끝인지 확인합니다.

## 비교·검색·정리

```wave
import("std::string::cmp")::{eq, cmp, starts_with, ends_with};
import("std::string::find")::{find, contains};
import("std::string::trim")::{trim_range};
```

| 모듈 | 주요 함수 |
| --- | --- |
| `std::string::cmp` | `eq`, `cmp`, `starts_with`, `ends_with` |
| `std::string::find` | `find`, `rfind_char`, `contains`, `count` |
| `std::string::trim` | 공백을 제외한 범위를 찾는 `trim_left_index`, `trim_right_index`, `trim_range` |
| `std::string::hash` | `djb2_32`, `fnv1a_64` |

검색 함수는 찾은 위치를 `i32`로 반환하고, 일치하는 항목이 없으면 `-1`을 반환합니다. `wavec print std-path`로 표준 라이브러리 소스 경로를 확인할 수 있습니다.

## ASCII 도우미

`std::string::ascii`는 `is_digit`, `is_alpha`, `is_alnum`, `is_space`, `to_lower`, `to_upper` 같은 바이트 단위 ASCII 도우미를 제공합니다. 이 함수들은 Unicode 문자 전체가 아니라 ASCII 바이트를 처리합니다.

## 바이트 순서

`std::bytes::endian`은 16·32·64비트 정수의 바이트 순서 변환과 big-endian·little-endian 읽기/쓰기를 제공합니다. 파일 형식이나 네트워크 프로토콜을 처리할 때는 호스트의 바이트 순서와 외부 형식이 요구하는 바이트 순서를 구분하십시오.

정수 너비와 offset을 명시적으로 유지하면 플랫폼이 달라져도 바이너리 형식 코드를 검토하기 쉽습니다.
