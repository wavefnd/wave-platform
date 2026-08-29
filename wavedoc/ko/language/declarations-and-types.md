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

`var`는 지역 변수를 선언하는 문법입니다. 지역 변수는 `var name: Type = value;` 형식으로 타입을 명시적으로 선언합니다. 초깃값 없이 저장 공간만 선언할 때는 `var name: Type;` 형식을 사용합니다.

## 최상위 저장 선언

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;
```

`const`와 `static`은 최상위에서 사용합니다. 함수 블록 안의 지역 선언은 `var`를 사용하십시오.

## 정수와 부동소수점 타입

정수 타입은 다음과 같습니다.

- 부호 있음: `i8`, `i16`, `i32`, `i64`, `i128`
- 부호 없음: `u8`, `u16`, `u32`, `u64`, `u128`
- 주소 크기 정수: `isz`, `usz`
- 부동소수점: `f32`, `f64`

`isz`는 주소 크기에 맞는 부호 있는 정수 타입이고, `usz`는 주소 크기에 맞는 부호 없는 정수 타입입니다.

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

타입 별칭은 같은 타입을 코드의 문맥에 맞는 이름으로 표현하는 문법입니다. `UserId`로 선언한 값은 `u64` 값처럼 그대로 사용할 수 있습니다.
