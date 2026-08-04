---
translation_set_id: compiler
path: getting-started/compiler
locale: ko
group: getting-started
group_order: 1
order: 3
title: 컴파일러 명령 참조
summary: wavec로 Wave 프로그램을 컴파일하고 검사하고 실행합니다.
---

## 기본 작업 흐름

```shell
wavec build main.wave
wavec run main.wave
wavec build --emit=check main.wave
```

설치된 릴리스가 지원하는 정확한 옵션은 wavec --help에서 확인합니다. 진단에는 문제가 발생한 소스 위치와 거부 이유가 포함됩니다.

- build는 입력 프로그램을 컴파일합니다.
- run은 프로그램을 컴파일한 뒤 실행합니다.
- --emit=check는 일반 실행 파일을 만들지 않고 프런트엔드 검사를 수행합니다.

> **Whale 상태**
> 
> 0.2.0-pre-beta에서 --whale 옵션은 예약되어 있으며 Whale 백엔드가 구현되지 않았다는 오류를 냅니다. 운영 빌드에서 사용하지 마십시오.

