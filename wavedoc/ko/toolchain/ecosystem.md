---
translation_set_id: ecosystem
path: toolchain/ecosystem
locale: ko
group: toolchain
group_order: 4
order: 1
title: Whale 툴체인
summary: 별도 저수준 툴체인인 Whale의 역할, 구성 요소 경계와 현재 완성도를 설명합니다.
---

## Whale이란

Whale은 Rust로 작성된 별도의 저수준 툴체인입니다. 장기적으로 Wave와 다른 네이티브 코드 생성기가 재사용할 수 있는 어셈블리, 오브젝트, 링크와 중간 표현 구성 요소를 제공하는 것이 목적입니다.

Whale은 Wave 개발 환경 전체를 부르는 이름이 아닙니다. 각 프로젝트의 책임은 다음처럼 구분됩니다.

| 프로젝트 | 책임 |
| --- | --- |
| `wavec` | Wave 소스를 파싱·검증·컴파일하며 현재 LLVM을 통해 네이티브 코드를 생성합니다. |
| Vex | Wave 패키지, manifest, 의존성 그래프, lockfile과 패키지 빌드를 관리합니다. |
| Whale | 독립적인 assembler, object, linker와 IR 구성 요소를 개발합니다. |
| Wave `std` | 런타임과 시스템 API를 Wave 소스 모듈로 제공합니다. |

## 구성 요소

현재 Whale workspace는 네 가지 주요 라이브러리 영역으로 구성됩니다.

- `assembler`: 토큰화, AMD64 파싱·인코딩, section, symbol과 relocation
- `object`: 오브젝트 파일 모델과 ELF64 writer
- `linker`: 개발 중인 링크 계층
- `ir`: Whale IR 타입, builder, 출력, 검증과 선택적 frontend socket

`whale` 실행 파일은 이 영역을 `asm`, `object`, `link`, `ir` 명령으로 제공합니다.

## 현재 통합 상태

현재 일반적인 `wavec` 빌드는 LLVM 백엔드를 사용합니다. Whale은 독립 툴체인으로 개발되고 있으며 `wavec --whale` 옵션으로 선택하는 기능이 아닙니다.

사용자와 도구 개발자는 이 경계를 지켜야 합니다. Whale을 설치해도 `wavec`의 동작이 자동으로 바뀌지 않으며 Whale 옵션을 Vex에 전달해서도 안 됩니다. 각 도구는 자신의 명령행 계약으로 직접 실행합니다.

## 완성도 경계

Whale은 활발히 개발 중입니다. AMD64 assembler와 ELF64 object 경로는 제한된 실험에 사용할 수 있지만 AArch64 assembler 경로와 linker CLI는 아직 완성되지 않았습니다. IR socket 명령은 빌드할 때 기능을 명시적으로 켜야 합니다. 자동화에 사용하기 전 Whale 명령 참조에서 해당 하위 명령의 상태를 확인하십시오.

구성 요소가 안정적인 계약을 선언하기 전에는 소스 수준에서 산출물을 재현할 수 있게 유지하고, object format, architecture, symbol과 relocation을 독립 도구로 검증하는 것이 좋습니다.
