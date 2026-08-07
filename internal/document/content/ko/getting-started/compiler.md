---
translation_set_id: compiler
path: getting-started/compiler
locale: ko
group: getting-started
group_order: 1
order: 3
title: 컴파일러 명령 참조
summary: wavec의 build, run, emit, 진단, 대상·의존성·링크 옵션을 빠르게 설명합니다.
---

## 기본 명령

```shell
wavec build main.wave
wavec run main.wave
wavec build main.wave --emit=check
```

- `build`는 입력을 컴파일하고 기본적으로 실행 파일을 만듭니다.
- `run`은 빌드한 뒤 결과 프로그램을 실행합니다.
- `--emit=check`는 실행 파일을 만들지 않고 Wave 소스의 프런트엔드 검사를 수행합니다.

정확한 옵션 목록은 설치된 릴리스의 `wavec --help`를 기준으로 하십시오.

## 출력 제어

```shell
wavec build main.wave -o app
wavec build main.wave --emit=ast
wavec build main.wave --emit=ir
wavec build main.wave --emit=obj -o main.o
```

v0.2.0-pre-beta의 산출물 emit 종류는 `ast`, `ir`, `bc`, `asm`, `obj`, `bin`이며 `check`는 별도의 검사 모드입니다.

```shell
wavec print supported-emit-kinds
```

## 최적화와 디버그 출력

```shell
wavec -O2 build main.wave
wavec --debug-wave=tokens,ast build main.wave
```

최적화 플래그는 `-O0`, `-O1`, `-O2`, `-O3`, `-Os`, `-Oz`, `-Ofast` 계열을 사용합니다. 내부 단계 확인에는 `--debug-wave`를 사용할 수 있습니다.

## 의존성과 링크

```shell
wavec --dep-root .vex/dep build main.wave
wavec --dep math=/opt/wave-deps/math build main.wave
wavec --link=m -L ./lib build main.wave
```

- `--dep-root <dir>`는 외부 `package::module` import를 찾을 루트를 추가합니다.
- `--dep <name>=<path>`는 패키지 이름을 특정 경로에 고정합니다.
- `--link <lib>`와 `-L <path>`는 네이티브 링크 입력을 추가합니다.

## 기계가 읽을 수 있는 정보

툴링은 `wavec print`를 사용해 컴파일러의 실제 지원 정보를 질의할 수 있습니다.

```shell
wavec print supported-targets
wavec print supported-input-types
wavec print supported-emit-kinds
wavec print std-path
wavec print target-spec --format=json
```

진단을 자동화 도구에서 처리하려면 `--error-format=json`을 사용할 수 있습니다.

## Whale 옵션

`--whale`은 v0.2.0-pre-beta에서 예약되어 있지만 구현된 백엔드가 아닙니다. 이 옵션을 사용하면 컴파일러가 구현되지 않았다는 사용 오류를 반환합니다.
