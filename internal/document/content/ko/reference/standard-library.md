---
translation_set_id: standard-library
path: reference/standard-library
locale: ko
group: reference
group_order: 3
order: 1
title: 표준 라이브러리 색인
summary: 현재 std manifest의 최상위 모듈과 선택 기준을 정리합니다.
---

## 릴리스에 포함된 최상위 모듈

현재 `std/manifest.json`에는 다음 최상위 모듈이 선언되어 있습니다.

| 영역 | 모듈 | 대표 역할 |
| --- | --- | --- |
| 문자열·데이터 | `string`, `bytes`, `buffer` | 문자열 처리, 엔디언/바이트 연산, 가변 버퍼 |
| 수학 | `math` | 정수·수학 도우미 |
| 메모리·C 경계 | `mem`, `libc` | 수동 메모리, C 런타임 바인딩 |
| 파일·입출력 | `io`, `fs` | 파일 디스크립터, 파일·디렉터리 연산 |
| 네트워크 | `net` | 주소, 소켓, poll, TCP, UDP |
| 환경·경로·시간 | `env`, `path`, `time` | 환경 변수, 경로, 시간 처리 |
| 시스템·프로세스 | `sys`, `process` | OS 경계와 프로세스 생성·대기 |

## import 단위

최상위 모듈 이름만 import하는 것이 아니라 필요한 `.wave` 소스 단위까지 경로로 지정합니다.

```wave
import("std::string::len");
import("std::fs::file");
import("std::mem::alloc");
```

예를 들어 `std::string::len` 파일은 `len`과 `is_empty`를 함께 정의합니다.

## API를 찾는 방법

표준 라이브러리는 Wave 소스로 제공되므로 현재 설치된 릴리스의 소스를 직접 확인하는 것이 가장 정확합니다.

```shell
wavec print std-path
```

그 경로에서 모듈별 `.wave` 파일을 열면 함수 시그니처, 반환값과 내부적으로 사용하는 하위 모듈을 확인할 수 있습니다.

## 플랫폼 경계

`sys`, `libc`, `fs`, `net`, `process`처럼 운영체제 기능과 연결되는 API는 실패를 음수 값이나 플랫폼 방식 결과로 전달하는 경우가 많습니다. 성공값만 가정하지 말고 각 함수의 구현과 반환 계약을 확인하십시오.

표준 라이브러리는 언어 문법과 별도로 발전할 수 있으므로 다른 Wave 버전의 `std` 소스를 섞어 사용하는 것은 피하는 것이 좋습니다.
