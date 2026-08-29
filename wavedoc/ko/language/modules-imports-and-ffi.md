---
translation_set_id: modules-ffi
path: language/modules-imports-and-ffi
locale: ko
group: language
group_order: 2
order: 8
title: 모듈, import와 FFI
summary: 로컬·표준·패키지 import, 공개 항목과 C ABI의 extern/export 선언을 설명합니다.
---

## import 문법

```wave
import("std::string::len");
```

`import`는 모듈 경로를 문자열로 받고 `;`로 끝납니다. 모듈 전체, 별칭 또는 선택 가져오기 형태를 사용할 수 있습니다.

```wave
import("add");
import("./helpers" as helpers);
import("add")::{sum, Point};
```

## 표준 라이브러리 import

`std::`로 시작하는 경로는 설치된 Wave 표준 라이브러리에서 해석됩니다.

```wave
import("std::fs::file");
import("std::io::fd");
```

`wavec print std-path`로 설치된 표준 라이브러리 위치를 확인할 수 있습니다.

## 로컬 파일 import

로컬 파일 경로는 `./`로 시작하며 import 문장을 작성한 소스 파일의 디렉터리를 기준으로 합니다. `.wave` 확장자는 생략할 수 있습니다.

```wave
import("./math");
import("./helpers" as helpers);
```

첫 문장은 같은 디렉터리의 `math.wave`를 가져옵니다. 별칭을 사용하면 `helpers::function_name`처럼 모듈 이름을 명시해서 공개 항목에 접근할 수 있습니다. 로컬 경로에는 `..`, 역슬래시 또는 절대 경로를 사용할 수 없습니다.

## 패키지 import

`./`와 `std::`로 시작하지 않는 경로의 첫 부분은 패키지 이름입니다. 패키지 이름만 적으면 패키지의 `src/lib.wave`를 가져오고, `::` 뒤에 경로를 붙이면 패키지 안의 모듈을 가져옵니다.

```wave
import("add");
import("add::math");
```

외부 패키지 위치는 다음 옵션으로 제공할 수 있습니다.

```shell
wavec --dep-root .vex/deps build main.wave
wavec --dep math=/absolute/path/to/math build main.wave
```

여러 dependency root에서 같은 패키지가 발견되면 해석이 모호하므로 `--dep name=path`로 경로를 고정합니다.

## 선택 가져오기와 공개 항목

선택 가져오기는 모듈에서 필요한 공개 항목만 이 파일의 이름 공간으로 가져옵니다.

```wave
import("add")::{sum, Point};

fun main() {
    var total: i32 = sum(2, 3);
    var point: Point = Point { x: 0, y: 0 };
}
```

함수, 구조체, 열거형, 타입 별칭, 상수와 정적 선언 앞에 `pub`을 붙이면 다른 모듈에서 가져올 수 있습니다.

```wave
pub struct Point {
    x: i32;
    y: i32;
}

pub fun sum(left: i32, right: i32) -> i32 {
    return left + right;
}
```

`pub import`는 선택한 공개 항목을 다시 내보냅니다.

```wave
pub import("./extra")::{increment};
```

별칭 가져오기와 선택 가져오기는 한 import 문장에서 함께 사용할 수 없습니다. `pub`은 Wave 모듈 사이의 공개 범위를 정하며 C ABI 심볼을 만드는 `export(c)`와는 별개의 기능입니다.

## C 함수 가져오기

```wave
extern(c) fun puts(text: ptr<i8>) -> i32;
```

ABI 이름 뒤에는 실제 심볼 이름을 문자열로 지정할 수 있습니다.

```wave
extern(c, "native_symbol") fun local_name(value: i32) -> i32;
```

## Wave 함수 내보내기

```wave
export(c) fun wave_add(left: i32, right: i32) -> i32 {
    return left + right;
}
```

`extern`과 `export`는 단일 함수와 블록 형태로 사용할 수 있습니다. 내보내는 함수는 구체적인 ABI 시그니처를 가져야 하므로 제네릭일 수 없습니다.

## ABI에서 직접 확인할 항목

- 정수와 포인터 너비
- 호출 규약과 대상 ABI 이름
- 외부 심볼 이름
- 문자열의 실제 표현
- 포인터의 유효 기간과 소유권
- 링크할 라이브러리와 검색 경로

링크에 성공했다는 사실만으로 함수 시그니처와 메모리 계약까지 일치한다는 뜻은 아닙니다.

## 대상 조건 속성

최상위 항목에는 대상 조건 속성을 붙일 수 있습니다.

```wave
#[target(os="linux", arch="x86_64")]
extern(c) fun platform_call(value: i32) -> i32;
```

조건 키는 `arch`, `os`, `env`, `abi`이며 속성은 바로 다음 최상위 항목에 적용됩니다.
