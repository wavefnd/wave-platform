---
translation_set_id: whale-cli
path: toolchain/whale-cli
locale: ko
group: toolchain
group_order: 4
order: 4
title: Whale 명령 참조
summary: Whale assembler, object wrapper, 진단 출력과 선택적 IR 명령을 설명합니다.
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
whale ir <subcommand> [options]
```

## AMD64 assembler

```shell
whale asm --amd64 input.asm -o output.o
```

AMD64 assembler는 `.o` 경로를 출력으로 받고, section, symbol과 relocation을 담은 ELF64 relocatable object를 만듭니다.

상세 진단 출력은 `--debug-whale`로 켭니다.

```shell
whale asm --amd64 input.asm -o output.o \
  --debug-whale --token --ast --bytes --dump-hex --stats
```

진단 플래그에는 `--token`, `--ast`, `--bytes`, `--dump-hex`, `--dump-bin`, `--dump-json`, `--stats`가 있습니다. `--trace`는 처리 과정을 출력합니다.

## Object wrapper

```shell
whale object input.bin -o output.o
```

`object` 명령은 raw byte를 ELF64 `.text` section에 넣고 offset 0에 전역 `start` symbol을 추가합니다. 이 명령은 raw code를 ELF object로 감싸는 용도입니다.

## 선택적 IR socket

`ir` 명령은 Whale을 `socket-cli` 기능과 함께 빌드했을 때만 포함됩니다.

```shell
cargo run -p whale --features socket-cli -- ir lower program.json
cargo run -p whale --features socket-cli -- ir lower program.json -o program.wir
```

`ir lower`는 Whale socket schema의 JSON을 읽고 Whale IR로 변환한 뒤 모듈을 검증합니다. 텍스트 IR은 stdout 또는 `-o` 경로에 출력됩니다. `--target <triple>`은 대상 문자열을 바꾸고 `--no-verify`는 검증을 생략합니다.

IR 명령을 사용하려면 Whale을 `socket-cli` 기능과 함께 빌드해야 합니다. Socket JSON 생산자와 Whale은 같은 socket schema version을 사용해야 합니다.
