---
translation_set_id: build-link-targets
path: toolchain/build-link-targets
locale: ko
group: toolchain
group_order: 4
order: 2
title: 빌드, 링크와 대상 옵션
summary: emit 산출물, 입력 종류, 링크, target/CPU/ABI와 프리스탠딩 빌드 계획을 설명합니다.
---

## emit 산출물

```shell
wavec build main.wave --emit=check
wavec build main.wave --emit=ast
wavec build main.wave --emit=ir
wavec build main.wave --emit=bc
wavec build main.wave --emit=asm
wavec build main.wave --emit=obj -o main.o
wavec build main.wave --emit=bin -o app
```

v0.2.0-pre-beta의 artifact emit 종류는 `ast`, `ir`, `bc`, `asm`, `obj`, `bin`입니다. `check`는 산출물 종류가 아니라 검사 제어 모드이며 다른 artifact emit과 섞지 않는 것이 전제입니다.

```shell
wavec print supported-emit-kinds
```

## 입력 종류와 link-only

컴파일러는 Wave 소스 외에도 IR, bitcode, assembly, object, archive 입력 종류를 구분합니다. 현재 지원 목록은 다음으로 질의합니다.

```shell
wavec print supported-input-types
```

이미 만들어진 object나 archive만 링크하려면 `--input-type`과 `--link-only`를 사용할 수 있습니다.

```shell
wavec build module.o --input-type=obj --link-only --emit=bin -o app
```

## 네이티브 링크

```shell
wavec --link=m -L ./lib build main.wave
```

`--link`는 라이브러리를 추가하고 `-L`은 검색 경로를 추가합니다. FFI에서 심볼을 선언했더라도 해당 심볼을 제공하는 라이브러리가 자동으로 링크되는 것은 아닙니다.

## 대상 선택

주요 LLVM 대상 옵션은 다음과 같습니다.

- `--target <triple>`
- `--cpu <name>`
- `--features <csv>`
- `--abi <name>`
- `--sysroot <path>`

호스트 기본값과 지원 대상은 컴파일러에서 확인하십시오.

```shell
wavec print host-target
wavec print supported-targets
wavec print target-spec --target <triple>
wavec print cpu-list --target <triple>
wavec print target-features --target <triple>
```

## 프리스탠딩 링크

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files \
  -o kernel.elf
```

`--freestanding`은 기본 라이브러리를 사용하지 않는 쪽으로 빌드 설정을 조정합니다. `--entry`는 링커 엔트리를, `--linker-script`는 스크립트를, `--no-start-files`는 호스트 시작 파일 제외를 지정합니다.

실제 실행 전 링크 계획을 확인할 때는 `--dry-run`을 사용할 수 있습니다.

## 크로스 빌드에서 확인할 것

- target triple이 컴파일러의 지원 목록에 있는지
- sysroot와 링커가 대상 ABI에 맞는지
- 링크 라이브러리가 대상 아키텍처용인지
- CPU feature가 대상 CPU에서 유효한지
- 프리스탠딩이면 엔트리 심볼과 메모리 배치가 링커 스크립트와 일치하는지
