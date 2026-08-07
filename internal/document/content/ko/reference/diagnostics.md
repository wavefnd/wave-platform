---
translation_set_id: diagnostics
path: reference/diagnostics
locale: ko
group: reference
group_order: 3
order: 2
title: 진단과 문제 해결
summary: 사람용·JSON 진단, check 모드, 디버그 출력과 재현 가능한 버그 보고 절차를 설명합니다.
---

## 먼저 버전 확인

문법 오류를 조사하기 전에 현재 컴파일러가 문서 기준과 같은지 확인합니다.

```shell
wavec --version
```

이 문서는 v0.2.0-pre-beta를 기준으로 합니다.

## 프런트엔드만 검사

링크나 실행 단계와 분리해 Wave 소스 자체를 검사하려면 다음을 사용합니다.

```shell
wavec build main.wave --emit=check
```

이 모드는 일반 실행 파일을 만들지 않고 Wave 입력을 검사합니다.

## JSON 진단

IDE, CI, 빌드 도구처럼 진단을 프로그램에서 처리해야 한다면:

```shell
wavec --error-format=json build main.wave --emit=check
```

사람이 읽는 기본 출력과 자동화용 JSON 출력을 구분하면 진단 파싱이 안정적입니다.

## 내부 단계 확인

```shell
wavec --debug-wave=tokens build main.wave --emit=check
wavec --debug-wave=ast build main.wave --emit=check
```

`--debug-wave`는 lexer 토큰, AST, IR 등 선택한 내부 단계 확인에 사용할 수 있습니다. 일반 사용자 오류를 해결할 때는 먼저 첫 번째 실제 진단과 해당 소스 위치를 읽고, 내부 덤프는 필요할 때만 사용하십시오.

## 흔한 문제 분리

1. **파싱/타입 문제**: `--emit=check`에서도 실패합니다.
2. **import 문제**: `std-path`, `--dep-root`, `--dep` 설정과 실제 파일 경로를 확인합니다.
3. **링크 문제**: `--link`, `-L`, 대상 ABI, 심볼 이름을 확인합니다.
4. **실행 문제**: 빌드는 성공하지만 실행 시 종료 코드나 런타임 환경에서 실패합니다.
5. **FFI 문제**: 타입 너비, 문자열 표현, 포인터 수명, 호출 규약을 C 선언과 다시 대조합니다.

## 좋은 버그 보고서

- `wavec --version` 출력
- 운영체제와 대상 triple
- 실행한 전체 명령
- 문제를 재현하는 최소 `.wave` 소스
- 전체 진단 출력
- 기대 결과와 실제 결과

개인 경로, 토큰, 비밀 키 등 불필요한 민감 정보는 제거하십시오.
