---
translation_set_id: compiler
path: getting-started/compiler
locale: ko
group: getting-started
group_order: 1
order: 3
title: 컴파일러 명령 참조
summary: wavec 명령, 빌드 파이프라인, 출력, 대상, 진단, 의존성 연결과 도구 질의를 설명합니다.
---

## 명령 모델

`wavec`는 컴파일러 CLI입니다. 개별 입력을 직접 컴파일하고, 도구에 컴파일러 지원 정보를 제공하며, 설치된 표준 라이브러리 소스를 관리합니다.

```text
wavec [global-options] <command> [command-options]
```

| 명령 | 용도 |
| --- | --- |
| `wavec build <input...>` | 플래그에 따라 검사, 코드 생성, 링크 또는 실행 파이프라인을 수행합니다. |
| `wavec check <file>` | `build <file> --emit=check`의 별칭입니다. |
| `wavec run <file> [-- <args...>]` | `build <file> --run`의 별칭이며 `--` 뒤 인자를 프로그램에 전달합니다. |
| `wavec print <item>` | 대상과 툴체인 지원 정보를 질의합니다. |
| `wavec install std` | 표준 라이브러리를 설치합니다. |
| `wavec update std` | 설치된 표준 라이브러리를 업데이트합니다. |
| `wavec --version` | 컴파일러와 LLVM 백엔드 버전을 출력합니다. |

`wavec --help`에서 명령과 옵션의 전체 목록을 확인할 수 있습니다.

## 빌드, 검사와 실행

```shell
wavec build main.wave
wavec check main.wave
wavec run main.wave -- first-argument second-argument
```

`build`는 기본적으로 실행 파일을 만듭니다. `check`는 프런트엔드 검사를 마친 뒤 멈춥니다. `run`은 바이너리 출력이 필요하며 공유 라이브러리 빌드와 함께 사용할 수 없습니다.

컴파일·링크·실행 없이 요청을 검증하고 실행할 단계를 확인하려면 `--dry-run`을 사용합니다.

```shell
wavec build main.wave --target riscv64-unknown-linux-gnu --dry-run
wavec build main.wave --dry-run --error-format=json
```

JSON 형식은 Vex 같은 빌드 도구가 사용하는 안정적인 통합 인터페이스입니다.

## emit과 입력 종류

```shell
wavec build main.wave --emit=ast
wavec build main.wave --emit=ir
wavec build main.wave --emit=bc
wavec build main.wave --emit=asm
wavec build main.wave --emit=obj -o main.o
wavec build main.wave --emit=bin -o app
```

산출물 emit 종류는 `ast`, `ir`, `bc`, `asm`, `obj`, `bin`입니다. `check`는 제어 모드이므로 단독으로 사용해야 합니다. 파이프라인이 허용하는 산출물 종류는 쉼표로 여러 개 지정할 수 있습니다.

입력 종류는 `wave`, `ir`, `bc`, `asm`, `obj`, `archive`입니다. `--input-type=<kind>`는 모든 입력의 종류를 강제로 지정합니다. object 또는 archive 입력만 링크할 때는 바이너리 emit과 `--link-only`를 사용합니다.

```shell
wavec build module.o --input-type=obj --link-only --emit=bin -o app
```

## 출력 위치

| 옵션 | 효과 |
| --- | --- |
| `-o <file>` | 주 출력 경로를 지정합니다. |
| `--out-dir <dir>` | emit 산출물을 지정한 디렉터리에 둡니다. |
| `--target-dir <dir>` | 중간 산출물과 기본 산출물 루트를 지정합니다. |

## 최적화와 진단 출력

```shell
wavec -O2 build main.wave
wavec --debug-wave=tokens,ast build main.wave
```

최적화 단계는 `-O0`, `-O1`, `-O2`, `-O3`, `-Os`, `-Oz`, `-Ofast`입니다. `--debug-wave`에는 `tokens`, `ast`, `ir`, `mc`, `hex`, `all`을 사용할 수 있고 여러 단계는 쉼표로 결합할 수 있습니다.

## 네이티브 링크

```shell
wavec --link=m -L ./lib build main.wave
wavec build main.wave --shared -o libexample.so
wavec build main.wave --static -o app
wavec build main.wave --pie -o app
```

`--link=<lib>`는 네이티브 라이브러리를 추가하고 `-L <path>`는 검색 경로를 추가합니다. 링크 모드는 호환 규칙에 따라 `--shared`, `--static`, `--pie`, `--no-pie`를 사용합니다.

백엔드와 링커 제어 옵션은 다음과 같습니다.

- `--target`, `--cpu`, `--features`, `--abi`, `--sysroot`
- `-C linker=<path>`와 `-C link-arg=<arg>`
- `-C link-sysroot=<path>`와 `-C relocation-model=<model>`
- `-C no-default-libs`

커널 같은 프리스탠딩 출력은 `--freestanding`과 함께 환경에 맞는 `--entry`, `--linker-script`, `--no-start-files` 설정을 사용합니다.

## 외부 패키지 해석

```shell
wavec --dep-root .vex/deps build main.wave
wavec --dep math=/opt/wave-deps/math build main.wave
```

`--dep-root <dir>`는 외부 `package::module` import를 찾을 루트를 추가합니다. `--dep <name>=<path>`는 패키지 이름을 한 디렉터리에 고정합니다. 이것은 컴파일러 통합 지점이며 프로젝트 manifest, 의존성 다운로드와 lockfile은 Vex가 담당합니다.

## 지원 기능 질의

대상이나 산출물 종류를 사용하는 도구는 `wavec print`로 지원 정보를 질의할 수 있습니다.

```shell
wavec print host-target
wavec print target-spec --format=json
wavec print supported-targets
wavec print supported-input-types
wavec print supported-emit-kinds
wavec print supported-print-items
wavec print cpu-list --target riscv64-unknown-linux-gnu
wavec print target-features --target riscv64-unknown-linux-gnu
wavec print default-linker
wavec print sysroot
wavec print std-path
wavec print dep-search-paths
```

`host`, `default-target`, `target-list` 같은 항목도 질의할 수 있습니다. 구조화된 출력을 지원하는 항목은 `--format=json`을 받습니다.

## 컴파일러와 툴체인의 경계

`wavec`는 소스 검사, 코드 생성과 링크를 담당합니다. Vex는 패키지 manifest, 의존성 그래프, lockfile과 재현 가능한 패키지 빌드를 담당합니다. Whale은 독립적으로 실행하는 저수준 툴체인입니다.
