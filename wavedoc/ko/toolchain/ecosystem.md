---
translation_set_id: ecosystem
path: toolchain/ecosystem
locale: ko
group: toolchain
group_order: 4
order: 1
title: Whale 툴체인
summary: 별도 저수준 툴체인인 Whale의 역할과 Wave 생태계의 구성 요소 경계를 설명합니다.
---

## Whale이란

Whale은 Rust로 작성된 독립적인 저수준 툴체인입니다. 어셈블리, 오브젝트, 링크와 중간 표현을 다루는 구성 요소를 Wave와 다른 네이티브 코드 생성 도구에서 재사용할 수 있게 설계되었습니다.

Whale은 Wave 개발 환경 전체를 부르는 이름이 아닙니다. 각 프로젝트의 책임은 다음처럼 구분됩니다.

| 프로젝트 | 책임 |
| --- | --- |
| `wavec` | Wave 소스를 검사·컴파일하고 LLVM 백엔드로 네이티브 코드를 생성합니다. |
| Vex | Wave 패키지, manifest, 의존성 그래프, lockfile과 패키지 빌드를 관리합니다. |
| Whale | 독립적인 assembler, object, linker와 IR 구성 요소를 제공합니다. |
| Wave `std` | 런타임과 시스템 API를 Wave 소스 모듈로 제공합니다. |

## 구성 요소

Whale workspace는 네 가지 주요 라이브러리 영역으로 구성됩니다.

- `assembler`: 토큰화, AMD64 파싱·인코딩, section, symbol과 relocation
- `object`: 오브젝트 파일 모델과 ELF64 writer
- `linker`: 링크 계층
- `ir`: Whale IR 타입, builder, 출력, 검증과 선택적 frontend socket

`whale` 실행 파일은 이 영역을 `asm`, `object`, `link`, `ir` 명령으로 제공합니다.

## 도구 경계

`wavec` 빌드는 LLVM 백엔드를 사용하고 Whale은 별도 명령으로 실행합니다. 두 도구의 옵션과 산출물은 각각의 명령 참조를 따릅니다.

Whale을 설치해도 `wavec`의 빌드 방식은 바뀌지 않습니다. Vex는 `wavec`를 사용해 Wave 패키지를 빌드하고, Whale은 저수준 산출물을 다루는 작업에서 직접 실행합니다.

## 산출물 검증

Whale 산출물을 빌드 과정에 연결할 때는 object format과 대상 architecture가 일치하는지 확인하십시오. symbol과 relocation은 `readelf`, `objdump` 같은 독립 도구로 검사할 수 있습니다. IR socket을 사용하는 빌드는 생산자와 Whale이 같은 socket schema를 사용해야 합니다.
