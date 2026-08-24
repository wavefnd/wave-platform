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

가져올 수 있는 선언 앞에는 `pub`을 붙입니다. 지역 `var` 선언은 함수나 블록 안에 둡니다.

## 일반 실행 프로그램

```wave
import("std::string::len")::{len};

const EXIT_OK: i32 = 0;

fun main() -> i32 {
    var message: str = "Wave";
    println("{} {}", message, len(message));
    return EXIT_OK;
}
```

일반적인 실행 파일은 `main`을 진입 함수로 사용합니다.

## 선언 순서와 import

`import`는 다른 소스 파일에 공개된 항목을 프로그램에서 사용할 수 있게 합니다. 표준 라이브러리는 `std::`, 로컬 파일은 `./`, 패키지는 패키지 이름으로 시작하는 경로를 사용합니다.

```wave
import("std::string::len");
import("./helpers" as helpers);
import("math")::{Vector};
import("math::ops");
```

모듈 전체를 가져오면 `helpers::function_name`처럼 경로를 붙여 접근할 수 있습니다. 선택 가져오기는 지정한 이름을 이 파일의 이름 공간에 놓습니다. `pub import("./module")::{Name};`은 선택한 공개 항목을 다시 내보냅니다.

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
