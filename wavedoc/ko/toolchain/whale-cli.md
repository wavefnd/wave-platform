---
translation_set_id: whale-cli
path: toolchain/whale-cli
locale: ko
group: toolchain
group_order: 4
order: 4
title: Whale 명령 참조
summary: 현재 Whale assembler, object, linker와 선택적 IR 명령 및 구현 한계를 설명합니다.
---

## Whale 빌드

Whale 저장소에서 다음을 실행합니다.

```shell
cargo build --release
```

최상위 실행 파일에는 네 가지 명령 계열이 있습니다.

```text
whale asm [--amd64 | --aarch64] <input> -o <output>
whale object <input> -o <output>
whale link <...>
whale ir <subcommand> [options]
```

이 명령들은 개발 중입니다. 아래 완성도 설명도 명령 계약의 일부로 보아야 합니다.

## AMD64 assembler

```shell
whale asm --amd64 input.asm -o output.o
```

구현된 assembler 경로는 AMD64이며 현재 `.o` 출력을 요구합니다. 조립된 section, symbol과 지원하는 relocation을 보존한 ELF64 relocatable object를 만듭니다. 최상위 명령 형태에는 `--aarch64`가 표시되지만 현재 assembler는 지원하지 않는다고 거부합니다.

개발자용 내부 확인은 `--debug-whale`로 켭니다.

```shell
whale asm --amd64 input.asm -o output.o \
  --debug-whale --token --ast --bytes --dump-hex --stats
```

내부 확인 플래그에는 `--token`, `--ast`, `--bytes`, `--dump-hex`, `--dump-bin`, `--dump-json`, `--stats`가 있습니다. `--trace`는 파이프라인 진행 과정을 출력합니다.

## Object wrapper

```shell
whale object input.bin -o output.o
```

현재 `object` 명령은 raw byte를 읽어 ELF64 `.text` section에 넣고 offset 0에 전역 `start` symbol을 추가합니다. 제한된 object 생성 경로이며 아직 범용 object 편집기나 IR-to-object frontend는 아닙니다.

## Linker 상태

```shell
whale link object.o -o executable
```

현재 CLI의 `link` 명령은 placeholder이며 실행 파일을 만들지 않습니다. 아직 빌드 파이프라인에 사용하지 마십시오.

## 선택적 IR socket

`ir` 명령은 Whale을 `socket-cli` 기능과 함께 빌드했을 때만 포함됩니다.

```shell
cargo run -p whale --features socket-cli -- ir lower program.json
cargo run -p whale --features socket-cli -- ir lower program.json -o program.wir
```

`ir lower`는 현재 Whale frontend 스키마와 일치하는 socket JSON을 읽고 Whale IR로 낮춘 뒤 기본적으로 모듈을 검증합니다. 텍스트 IR은 stdout 또는 `-o` 경로에 출력됩니다. `--target <triple>`은 대상 문자열을 바꾸고 `--no-verify`는 검증을 생략합니다.

`socket-cli` 없이 실행한 `whale ir`은 다시 빌드하는 방법을 안내하며 오류로 종료합니다. Socket JSON은 버전이 있는 내부 교환 형식이므로 생산자는 Whale 빌드가 사용하는 socket version과 일치해야 합니다.
