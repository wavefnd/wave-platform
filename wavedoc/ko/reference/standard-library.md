---
translation_set_id: standard-library
path: reference/standard-library
locale: ko
group: reference
group_order: 3
order: 1
title: 표준 라이브러리 색인
summary: Wave 표준 라이브러리의 최상위 모듈과 import 방법을 정리합니다.
---

## 최상위 모듈

`std/manifest.json`은 다음 최상위 모듈을 선언합니다.

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
import("std::string::len")::{len, is_empty};
import("std::fs::file")::{fs_open_read, fs_close};
import("std::mem::alloc")::{mem_alloc, mem_free};
```

선택 가져오기는 지정한 공개 항목을 이름 공간 접두사 없이 사용할 수 있게 합니다. 일반 `import("std::string::len");`은 같은 항목을 `std::string::len` 이름 공간을 통해 제공합니다.

## API를 찾는 방법

표준 라이브러리 경로는 다음 명령으로 찾을 수 있습니다.

```shell
wavec print std-path
```

이 경로에는 모듈별 `.wave` 파일과 공개 함수 시그니처가 들어 있습니다.

## 플랫폼 경계

`sys`, `libc`, `fs`, `net`, `process`처럼 운영체제 기능과 연결되는 API는 실패를 음수 값, `null` 또는 플랫폼 상태 값으로 전달합니다. 각 함수가 정한 성공값과 실패값을 구분하고, 포인터를 반환하는 함수는 역참조하기 전에 `null`을 확인하십시오.
