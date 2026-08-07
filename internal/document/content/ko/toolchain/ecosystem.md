---
translation_set_id: ecosystem
path: toolchain/ecosystem
locale: ko
group: toolchain
group_order: 4
order: 1
title: Wave 툴체인
summary: wavec, 표준 라이브러리, Vex, Whale 상태와 네이티브 상호 운용의 역할을 구분합니다.
---

## wavec

`wavec`는 v0.2.0-pre-beta의 중심 컴파일러 명령행 도구입니다. Wave 소스를 검사·컴파일·링크·실행하고, 대상과 지원 기능을 질의하는 `print` 인터페이스를 제공합니다.

```shell
wavec --version
wavec build main.wave
wavec run main.wave
wavec print supported-targets
```

## 표준 라이브러리

`std`는 컴파일러 저장소에 Wave 소스로 포함되어 있으며 문자열, 메모리, 파일, 네트워크, 프로세스 등 런타임 기능을 제공합니다.

```shell
wavec print std-path
```

현재 컴파일러가 사용하는 표준 라이브러리 경로를 확인한 뒤 실제 `.wave` 파일을 API 참조로 사용할 수 있습니다.

## Vex

Vex는 Wave 의존성과 패키지를 다루는 별도 패키지 관리 프로젝트입니다. `wavec` 자체는 외부 패키지 import를 위해 `--dep-root`와 `--dep name=path` 같은 안정적인 연결 지점을 제공합니다.

따라서 패키지 매니저가 dependency tree를 준비하고, 컴파일러에는 최종 해석 경로를 전달하는 방식으로 역할을 나눌 수 있습니다.

## Whale

`wavec`에는 `--whale` 플래그가 남아 있지만 v0.2.0-pre-beta에서는 예약 상태이며 구현된 백엔드가 아닙니다. 컴파일 작업에는 사용하지 않습니다.

## 네이티브 상호 운용

Wave는 `extern(c)`, `export(c)`, 링크 라이브러리 옵션과 인라인 어셈블리를 통해 C/C++ 및 플랫폼 기능과 연결할 수 있습니다. 표준 라이브러리에서 제공하지 않는 기능도 좁은 네이티브 경계를 만들어 사용할 수 있습니다.

## 도구가 기능을 확인하는 방법

하드코딩된 대상 목록 대신 컴파일러에 직접 질의할 수 있습니다.

```shell
wavec print supported-targets
wavec print supported-input-types
wavec print supported-emit-kinds
wavec print supported-print-items
```

JSON 출력이 필요한 도구는 `wavec print ... --format=json` 형태를 사용할 수 있습니다.
