---
translation_set_id: modules-ffi
path: language/modules-imports-and-ffi
locale: ko
group: language
group_order: 2
order: 8
title: 모듈, import와 FFI
summary: 로컬·표준·외부 패키지 import 해석과 C ABI의 extern/export 선언을 설명합니다.
---

## import 문법

```wave
import("std::string::len");
```

`import`는 문자열 리터럴 하나를 받고 `);`까지 포함한 문장으로 끝납니다.

## 표준 라이브러리 import

`std::`로 시작하는 경로는 설치된 Wave 표준 라이브러리에서 해석됩니다.

```wave
import("std::fs::file");
import("std::io::fd");
```

`wavec print std-path`로 현재 컴파일러가 사용하는 표준 라이브러리 위치를 확인할 수 있습니다.

## 로컬 파일 import

`::`가 없는 경로는 현재 소스 파일의 디렉터리를 기준으로 로컬 파일을 찾습니다. `.wave` 확장자를 생략하면 컴파일러가 붙여서 찾습니다.

```wave
import("math");
```

위 형태는 같은 기준 디렉터리의 `math.wave`를 찾습니다.

## 외부 패키지 import

`std::`가 아니면서 `::`를 포함하는 경로는 첫 번째 부분을 패키지 이름으로 해석합니다.

```wave
import("math::vector::ops");
```

외부 패키지 위치는 다음 옵션으로 제공할 수 있습니다.

```shell
wavec --dep-root .vex/deps build main.wave
wavec --dep math=/absolute/path/to/math build main.wave
```

여러 dependency root에서 같은 패키지가 발견되면 해석이 모호하므로 `--dep name=path`로 고정하는 편이 안전합니다.

## C 함수 가져오기

```wave
extern(c) fun puts(text: ptr<i8>) -> i32;
```

ABI 이름 뒤에 실제 심볼 이름을 문자열로 지정하는 형태도 파서가 지원합니다.

```wave
extern(c, "native_symbol") fun local_name(value: i32) -> i32;
```

## Wave 함수 내보내기

```wave
export(c) fun wave_add(left: i32, right: i32) -> i32 {
    return left + right;
}
```

`extern`과 `export`는 단일 함수뿐 아니라 블록 형태도 지원합니다. 내보내는 함수는 이 릴리스에서 제네릭일 수 없습니다.

## ABI에서 직접 확인할 항목

- 정수와 포인터 너비
- 호출 규약과 대상 ABI 이름
- 외부 심볼 이름
- 문자열의 실제 표현
- 포인터의 유효 기간과 소유권
- 링크할 라이브러리와 검색 경로

링크에 성공했다는 사실만으로 함수 시그니처와 메모리 계약까지 일치한다는 뜻은 아닙니다.

## 대상 조건 속성

import 전처리기는 다음과 같은 최상위 대상 조건 속성을 처리할 수 있습니다.

```wave
#[target(os="linux", arch="x86_64")]
extern(c) fun platform_call(value: i32) -> i32;
```

조건 키는 `arch`, `os`, `env`, `abi`이며 속성은 바로 다음 최상위 항목에 적용됩니다.
