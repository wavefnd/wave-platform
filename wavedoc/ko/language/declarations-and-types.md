---
translation_set_id: types
path: language/declarations-and-types
locale: ko
group: language
group_order: 2
order: 2
title: 선언과 타입
summary: 지역 바인딩, 최상위 상수·정적 저장소, 내장 타입, 배열과 타입 별칭을 설명합니다.
---

## 지역 변수

```wave
var count: i32 = 1;
var limit: i32 = 10;
var index: i32 = 0;
```

| 선언 | 의미 |
| --- | --- |
| `var` | 값을 다시 대입할 수 있는 지역 변수 |

`var`는 유일한 지역 선언 형식입니다. 폐지된 `let`과 `let mut` 표기는
문법 오류이며 새 Wave 소스에서 사용하면 안 됩니다.

## 최상위 저장 선언

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;
```

`const`와 `static`은 최상위에서 사용합니다. 함수 블록 안의 지역 선언은 `var`를 사용하십시오.

## 정수와 부동소수점 타입

문서화된 고정 너비 정수 타입은 다음과 같습니다.

- 부호 있음: `i8`, `i16`, `i32`, `i64`, `i128`, `i256`, `i512`, `i1024`
- 부호 없음: `u8`, `u16`, `u32`, `u64`, `u128`, `u256`, `u512`, `u1024`
- 부동소수점: `f32`, `f64`

`isz`와 `usz`는 lexer가 인식하지만 현재 타입 변환 경로에서 처리되지 않습니다. 컴파일러 지원이 문서화되기 전에는 사용하지 마십시오.

## 기타 내장 타입

| 타입 | 용도 |
| --- | --- |
| `bool` | `true` 또는 `false` |
| `char` | 문자 값 |
| `byte` | 바이트 값 |
| `str` | 문자열 값 |
| `ptr<T>` | `T`를 대상으로 하는 포인터 |
| `array<T, N>` | 요소 타입 `T`, 길이 `N`인 고정 길이 배열 |

사용자 정의 구조체와 열거형, 타입 별칭도 타입 위치에 사용할 수 있습니다.

## 배열

```wave
var values: array<i32, 4> = [10, 20, 30, 40];
var first: i32 = values[0];
values[1] = 25;
```

배열 리터럴로 초기화할 때 선언된 길이와 요소 수가 일치해야 합니다. 인덱싱은 `[]`를 사용합니다.

## 타입 별칭

```wave
type UserId = u64;
var id: UserId = 7;
```

타입 별칭은 기존 타입에 새로운 이름을 부여합니다. 별칭 자체가 새로운 런타임 표현이나 별도의 저장소를 추가하는 것은 아닙니다.
