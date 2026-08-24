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

artifact emit 종류는 `ast`, `ir`, `bc`, `asm`, `obj`, `bin`입니다. `check`는 산출물 종류가 아니라 검사 제어 모드이며 다른 artifact emit과 함께 사용하지 않습니다.

```shell
wavec print supported-emit-kinds
```

## 입력 종류와 link-only

컴파일러는 Wave 소스 외에도 IR, bitcode, assembly, object와 archive 입력을 구분합니다. 지원 목록은 다음 명령으로 질의합니다.

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

호스트 기본값과 지원 대상은 다음 명령으로 확인합니다.

```shell
wavec print host-target
wavec print supported-targets
wavec print target-spec --target <triple>
wavec print cpu-list --target <triple>
wavec print target-features --target <triple>
```

## 지원 대상 계열

Wave 툴체인이 정의하는 대상 계열은 다음과 같습니다. 설치된 빌드에서 사용할 수 있는 대상은 `wavec print supported-targets`로 확인합니다.

| 대상 | 환경 | object format |
| --- | --- | --- |
| `x86_64-unknown-linux-gnu` | Hosted Linux GNU | ELF |
| `x86_64-apple-darwin` | Hosted macOS | Mach-O |
| `x86_64-w64-windows-gnu` | Hosted Windows GNU | COFF |
| `x86_64-pc-windows-gnu` | Hosted Windows GNU alias | COFF |
| `x86_64-unknown-none-elf` | Freestanding | ELF |
| `aarch64-unknown-linux-gnu` | Hosted Linux GNU | ELF |
| `aarch64-apple-darwin` | Hosted macOS | Mach-O |
| `aarch64-unknown-none-elf` | Freestanding | ELF |
| `riscv64-unknown-linux-gnu` | Hosted Linux GNU | ELF |
| `riscv64-unknown-none-elf` | Freestanding | ELF |

## RISC-V 64 계약

Hosted RISC-V 대상의 기본값은 `generic-rv64`, RV64GC, `lp64d` ABI입니다. Freestanding 대상의 기본값은 `generic-rv64`, RV64IMAC, `lp64`입니다.

```shell
wavec print target-spec --target riscv64-unknown-linux-gnu --format=json
wavec print target-spec --target riscv64-unknown-none-elf --format=json
```

지원 RISC-V CPU는 `generic`, `generic-rv64`, `rocket-rv64`, `sifive-u74`입니다. Feature override는 `m`, `a`, `f`, `d`, `c`, `zicsr`, `zifencei` 이름 앞에 부호를 붙여 쉼표로 구분합니다.

```shell
wavec build main.wave \
  --target riscv64-unknown-linux-gnu \
  --features=+m,+a,+f,-d,+c,+zicsr \
  --abi=lp64f
```

RISC-V 검증은 일관되지 않은 조합을 거부합니다. `d`에는 `f`가 필요하고 `f`에는 `zicsr`가 필요합니다. `lp64`, `lp64f`, `lp64d`는 활성화한 부동소수점 feature와 일치해야 합니다. ABI를 직접 지정하지 않으면 컴파일러가 feature에서 ABI를 유도합니다.

## 프리스탠딩 링크

```shell
wavec build kernel.wave \
  --target riscv64-unknown-none-elf \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files \
  -o kernel.elf
```

`--freestanding`은 기본 라이브러리를 사용하지 않는 쪽으로 빌드 설정을 조정합니다. `--entry`는 링커 엔트리를, `--linker-script`는 스크립트를, `--no-start-files`는 호스트 시작 파일 제외를 지정합니다.

실제 실행 전 링크 계획을 확인할 때는 `--dry-run`을 사용할 수 있습니다.

## Hosted 크로스 링크

대상 코드를 생성할 수 있다는 사실이 대상의 C runtime, 시작 object와 라이브러리까지 제공한다는 뜻은 아닙니다. Hosted 크로스 빌드에는 호환되는 sysroot가 필요하며 상황에 따라 링커도 명시해야 합니다.

```shell
wavec build main.wave \
  --target riscv64-unknown-linux-gnu \
  --sysroot /path/to/riscv64-sysroot \
  -C linker=/path/to/target-linker \
  -o app-riscv64
```

sysroot에는 선택한 ABI용 파일이 들어 있어야 합니다. 이름이 같은 호스트 라이브러리는 대상 라이브러리를 대신할 수 없습니다.

## 크로스 빌드에서 확인할 것

- target triple이 컴파일러의 지원 목록에 있는지
- sysroot와 링커가 대상 ABI에 맞는지
- 링크 라이브러리가 대상 아키텍처용인지
- CPU feature가 대상 CPU에서 유효한지
- 프리스탠딩이면 엔트리 심볼과 메모리 배치가 링커 스크립트와 일치하는지
