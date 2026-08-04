---
translation_set_id: storage-duration
path: language/storage-duration
locale: ko
group: language
group_order: 2
order: 12
title: 저장 수명과 변경 가능성
summary: 범위와 상태 요구에 따라 var, let, const, static을 선택합니다.
---

| 형식 | 범위와 용도 |
| --- | --- |
| var | 일반 지역 변수이며 Wave의 기본 변수 문법 |
| let | 엄격한 OS·보안 코드의 제약된 불변 지역 바인딩 |
| let mut | 엄격한 OS·보안 코드의 제약된 명시적 가변 지역 바인딩 |
| const | 최상위 컴파일 시점 상수 |
| static | 최상위 저장 변수 |

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;

fun main() {
    var current: i32 = 1;
    current += 1;
}
```

> **범위 규칙**
> 
> 함수와 블록 안에서는 var와 let 계열을 사용합니다. 최상위 저장 선언에는 const와 static을 사용합니다.

