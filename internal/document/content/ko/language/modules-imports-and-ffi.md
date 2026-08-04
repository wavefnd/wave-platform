---
translation_set_id: modules-ffi
path: language/modules-imports-and-ffi
locale: ko
group: language
group_order: 2
order: 8
title: 모듈, 가져오기와 FFI
summary: Wave 코드를 구성하고 네이티브 C 경계를 선언합니다.
---

## 모듈과 가져오기

```wave
import("std::io::fd");
import("std::math::int");
```

import는 정규 문자열 경로로 모듈을 불러옵니다. 정확한 경로는 설치된 표준 라이브러리나 의존성 트리를 따릅니다.

## C 상호 운용

```wave
extern(c) fun puts(text: ptr<i8>) -> i32;

export(c) fun wave_add(left: i32, right: i32) -> i32 {
    return left + right;
}
```

extern(c)는 C 호환 라이브러리가 구현한 함수를 선언합니다. export(c)는 Wave 함수를 C 호환 호출 경계로 공개합니다.

> **ABI 계약**
> 
> 정수 너비, 호출 규약, 심볼 이름, 포인터 수명, 문자열 표현, 소유권을 네이티브 선언과 일치시키십시오. 링크 성공만으로 ABI가 올바르다는 뜻은 아닙니다.

