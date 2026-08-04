---
translation_set_id: build-link-targets
path: toolchain/build-link-targets
locale: ko
group: toolchain
group_order: 4
order: 2
title: 빌드, 링크와 대상 옵션
summary: 출력 산출물, 링크, 대상 선택과 프리스탠딩 빌드를 제어합니다.
---

## 산출물 출력

```shell
wavec build main.wave --emit=check
wavec build main.wave --emit=ast
wavec build main.wave --emit=ir
wavec build main.wave --emit=obj -o main.o
wavec build main.wave --emit=bin -o app
```

지원 emit 종류는 check, ast, ir, bc, asm, obj, bin입니다. check는 단독으로 사용합니다. 도구에서 기능을 확인하려면 wavec print supported-emit-kinds를 사용합니다.

## 링크

```shell
wavec build main.wave --link=m -L ./lib
wavec build module.o --input-type=obj --link-only --emit=bin -o app
```

## 대상과 프리스탠딩 모드

--target, --cpu, --features, --abi, --sysroot는 컴파일 대상을 설명합니다. --freestanding, --entry, --linker-script, --no-start-files, -C no-default-libs는 커널과 OS 방식 링크 계획을 지원합니다. 실행 전 --dry-run으로 계획을 검증하십시오.

