---
translation_set_id: whale-overview
path: whale/overview
locale: ko
group: whale
group_order: 1
order: 1
title: Whale 문서
summary: 툴체인 구성 요소, 사용 가능한 명령과 참조 문서를 안내합니다.
---

## 소개

Whale은 프로그래밍 언어 구현과 컴파일러 도구를 위한 범용 컴파일러 툴체인입니다. 타입을 갖는 중간 표현(IR), AMD64 어셈블러, 오브젝트 파일 라이브러리, 링커 기반 기능을 제공합니다. 각 구성 요소는 Rust 라이브러리와 `whale` 명령을 통해 사용합니다.

IR은 기계어 인코딩과 독립적으로 계산의 의미를 표현합니다. 어셈블러는 기계 명령을 인코딩해 재배치 가능한 오브젝트를 생성합니다. 오브젝트 라이브러리는 섹션·심볼·재배치를 표현합니다. 링커는 오브젝트 사이의 참조를 해결하고 실행 파일을 배치합니다.

## 빌드와 도구 사용

| 작업 | 문서 |
| --- | --- |
| 각 도구의 역할 | [툴체인 구성](/docs/ko/whale/ecosystem) |
| Wave 프로그램 빌드·링크·대상 선택 | [빌드와 링크](/docs/ko/whale/build-link-targets) |
| 패키지·의존성 관리 | [Vex](/docs/ko/whale/vex-package-manager) |
| Whale 명령 실행 | [Whale CLI](/docs/ko/whale/whale-cli) |

## 문서 안내

| 참조 문서 | 내용 |
| --- | --- |
| [IR 참조](ir-reference) | 타입, 값, 함수, 제어 흐름, 검증, 교환 형식 |
| [수치 연산](numeric-operations) | 정수 연산, 시프트, 형변환, 부동소수점 |
| [메모리 모델](memory-model) | 초기화, 포인터 유효성, 주소 계산, 레이아웃, 문자열 |
| [O0와 디버깅](o0-debugging) | 계산 보존, 변수 저장 공간, 디버그 정보 |
| [AMD64 타깃](amd64-target) | 타깃 식별자, 호출 규약, native 기능의 범위 |
| [어셈블러와 링커](assembler-linker) | 어셈블리 피연산자, 섹션, 심볼, 정적 링크 |

참조 문서는 Whale의 의미 규칙을 정의합니다. 사용 가능한 기능은 아래 표를 따릅니다. 참조 문서에 설명된 연산이 모든 빌드에서 제공되는 것은 아닙니다.

## 지원 현황

| 구성 요소 | 제공되는 인터페이스 | 제한 |
| --- | --- | --- |
| 어셈블러 | AMD64 어셈블리에서 ELF64 재배치 가능 오브젝트 생성 | 명령과 지시어 지원 범위가 완전하지 않음 |
| 오브젝트 라이브러리 | 오브젝트 구성, payload 없는 BSS, 타깃 검증, 크기를 검사하는 AMD64 ELF64 직렬화와 선택 가능한 Wave 레코드 구현 | 재배치 가능한 오브젝트는 실행 파일이 아님 |
| IR | 구성·출력·검증, 서명을 검사하는 직접·간접 호출, 버전이 명시된 AST lowering, 타깃 검증과 타입 레이아웃 검사 | AST 형식 2·typed IR 형식 4, 정확한 비트 상수, 텍스트 파싱·검증·왕복 출력을 지원하며 기계 호출 생성은 미지원 |
| 링커 | 심볼 해석, 입력 타깃 검증, 파일·메모리 섹션 배치 검사 | 전체 재배치 적용과 실행 파일 출력은 미지원 |
| 실행과 디버깅 | 정수·Bool, 제어 흐름, 추적 스택 메모리와 초기화 검사 | float·호출·전역 주소·native 코드 생성·DWARF 미지원 |

정의되지 않은 동작을 금지하는 규칙은 검증된 IR과 추적 메모리에 적용됩니다. 실험적 구현에는 메모리·실행 참조 문서의 런타임 검사가 아직 모두 구현되어 있지 않습니다.

## 빌드와 어셈블

Rust 1.86.0 이상으로 빌드합니다.

```sh
git clone https://github.com/wavefnd/Whale.git
cd Whale
cargo build --release --locked
```

다음 어셈블리를 `answer.asm`으로 저장합니다.

```asm
section .text
global answer

answer:
    mov eax, 42
    ret
```

ELF64 오브젝트를 생성합니다.

```sh
./target/release/whale asm --amd64 answer.asm -o answer.o
```

Whale의 자체 어셈블러를 사용하므로 외부 어셈블러가 필요하지 않습니다. 결과물에는 호출 가능한 함수가 들어 있으며 프로세스 시작 코드는 포함하지 않습니다.

실험적 AST→IR 명령은 `--features socket-cli`로 빌드하면 사용할 수 있습니다. CLI의 명령은 [명령 참조](/docs/ko/whale/whale-cli)를 참고하세요.

## 라이브러리 사용과 진단

실행이나 코드 생성 전에 IR을 검증해야 합니다. 입력 오류와 builder 오용은 구조적 오류로 반환합니다. 라이브러리의 입력 오류가 호스트 프로세스를 종료하거나 기존 내용을 덮어써서는 안 됩니다. 신뢰할 수 없는 입력을 받는 도구에서는 자원 한도를 조정할 수 있어야 합니다.

같은 툴체인 버전·입력·타깃·설정은 결정적인 산출물을 생성해야 합니다. 배포물의 메타데이터는 버전·commit·사용 가능한 기능을 식별합니다. CI는 각 인터페이스에 대해 정상 입력, 거부할 입력, trap, O0 보존, round-trip, native 실행 의미를 검사합니다. 개발·검증 명령은 [Whale 기여 안내](https://github.com/wavefnd/Whale/blob/master/CONTRIBUTING.md)를 참고하세요.
