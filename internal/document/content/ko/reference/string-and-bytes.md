---
translation_set_id: string-bytes
path: reference/string-and-bytes
locale: ko
group: reference
group_order: 3
order: 4
title: 문자열과 바이트
summary: 문자열 길이, 비교, 검색, 해시, ASCII와 엔디언 바이트 연산을 설명합니다.
---

## 문자열 모듈

```wave
import("std::string::len");
import("std::string::cmp");
import("std::string::find");
import("std::string::trim");

var size: i32 = len("Wave");
var empty: bool = is_empty("");
```

이 릴리스는 문자열 연산을 len, cmp, find, hash, trim, ascii 역할별 모듈로 나눕니다. 사용하는 함수를 선언한 구체 모듈을 가져오십시오.

## 바이트 순서

std::bytes::endian은 바이너리 형식과 프로토콜을 위한 엔디언 변환, 읽기, 쓰기 도우미를 제공합니다. 외부 형식이 요구하는 너비와 바이트 순서를 명시하십시오.

