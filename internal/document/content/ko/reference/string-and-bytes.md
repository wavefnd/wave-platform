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

현재 `std/string` 트리에는 다음 소스 단위가 있습니다.

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
import("std::string::len");

fun main() {
    let size: i32 = len("Wave");
    let empty: bool = is_empty("");
    println("{} {}", size, empty);
}
```

현재 구현의 `len`은 문자열을 인덱싱하며 0 값이 나올 때까지 세어 `i32` 길이를 반환합니다. 따라서 이 API를 사용할 때는 해당 릴리스의 `str` 표현 계약을 따르십시오.

## 비교·검색·정리

```wave
import("std::string::cmp");
import("std::string::find");
import("std::string::trim");
```

함수 이름과 정확한 반환 규칙은 설치된 표준 라이브러리 소스를 기준으로 하십시오. `wavec print std-path`로 현재 `std` 위치를 확인할 수 있습니다.

## ASCII 도우미

`std::string::ascii`는 ASCII 문자 분류와 변환처럼 바이트 수준 문자 처리가 필요한 코드를 위한 도우미를 제공합니다. Unicode 전체 문자 처리를 자동으로 제공한다고 가정하지 마십시오.

## 바이트 순서

`std::bytes`는 바이너리 데이터의 읽기·쓰기와 엔디언 변환을 위한 모듈을 제공합니다. 파일 형식이나 네트워크 프로토콜을 처리할 때는 “호스트의 바이트 순서”와 “외부 형식이 요구하는 바이트 순서”를 구분하십시오.

정수 너비와 offset을 명시적으로 유지하면 플랫폼이 달라져도 바이너리 형식 코드를 검토하기 쉽습니다.
