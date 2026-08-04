---
translation_set_id: program-structure
path: language/program-structure
locale: ko
group: language
group_order: 2
order: 10
title: 프로그램 구조와 가져오기
summary: 소스 파일, 진입점, 선언과 의존성 가져오기를 구성합니다.
---

## 소스 파일

Wave 소스 파일에는 import, 최상위 const와 static 선언, 타입 선언, 구조체, 열거형, proto 블록, extern·export 선언과 함수를 둘 수 있습니다.

```wave
import("std::string::len");

const EXIT_OK: i32 = 0;

fun main() -> i32 {
    var message: str = "Wave";
    println("{} {}", message, len(message));
    return EXIT_OK;
}
```

## 가져오기 해석

import는 정규 문자열 경로를 받습니다. 표준 라이브러리 경로는 std::로 시작하며 의존성 경로는 wavec 또는 패키지 도구에 전달한 루트에서 해석합니다.

## 진입점

일반 실행 파일은 보통 main을 정의합니다. 프리스탠딩 빌드는 --entry와 대응하는 링커 설정으로 다른 진입 심볼을 선택할 수 있습니다.

