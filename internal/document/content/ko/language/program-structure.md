---
translation_set_id: program-structure
path: language/program-structure
locale: ko
group: language
group_order: 2
order: 10
title: 프로그램 구조
summary: 최상위 선언, main 진입점, import 순서와 프리스탠딩 엔트리를 설명합니다.
---

## 소스 파일의 최상위 항목

Wave 소스는 다음과 같은 최상위 항목을 구성할 수 있습니다.

- `import(...)`
- `const`, `static`
- `type`
- `struct`, `enum`, `proto`
- `extern(...)`, `export(...)`
- `fun`
- 지원되는 항목 앞의 `#[target(...)]` 조건

지역 `var`와 `let`은 함수나 블록 안에 둡니다.

## 일반 실행 프로그램

```wave
import("std::string::len");

const EXIT_OK: i32 = 0;

fun main() -> i32 {
    let message: str = "Wave";
    println("{} {}", message, len(message));
    return EXIT_OK;
}
```

일반적인 실행 파일은 `main`을 진입 함수로 사용합니다.

## 선언 순서와 import

`import`는 가져온 파일의 AST를 현재 프로그램에 결합하는 방식으로 처리됩니다. 동일한 모듈을 반복해서 가져오는 상황을 컴파일러가 추적하며, 표준·외부 패키지·로컬 import는 각각 다른 해석 규칙을 사용합니다.

소스 파일을 작은 책임 단위로 분리하고, 각 파일이 직접 사용하는 모듈을 명시적으로 import하면 의존성 관계를 읽기 쉽습니다.

## 프리스탠딩 프로그램

커널, 부트 코드 또는 런타임이 없는 대상은 프리스탠딩 빌드 옵션을 사용할 수 있습니다.

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files
```

`--freestanding`은 기본 라이브러리 의존을 끄는 빌드 계획과 연결되며, `--entry`는 링커 엔트리 심볼을 설정합니다. 실제 부팅 가능한 산출물을 만들려면 대상 아키텍처, 링커 스크립트와 오브젝트 형식까지 함께 설계해야 합니다.
