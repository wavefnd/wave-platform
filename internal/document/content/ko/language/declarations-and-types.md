---
translation_set_id: types
path: language/declarations-and-types
locale: ko
group: language
group_order: 2
order: 2
title: 선언과 타입
summary: 변수, 상수, 변경 가능성, 기본 타입, 배열과 별칭을 설명합니다.
---

## 변수와 제약 바인딩

```wave
var count: i32 = 1;
var total: i64 = 0;
const LIMIT: u64 = 64;
static requests: u64 = 0;
```

var는 Wave의 일반적인 변수 선언이며 보통의 프로그램에서 기본으로 사용합니다. const는 상수를, static은 정적 저장소를 선언합니다.

> **운영체제와 보안 코드**
> 
> let과 let mut는 불변성과 상태 전이를 특히 명확하게 드러내야 하는 운영체제 및 보안 중심 코드에서 사용하는 제약 바인딩 문법입니다. 사용할 수는 있지만 Wave의 기본 변수 작성 방식은 아닙니다.

## 기본 타입

| 분류 | 타입 |
| --- | --- |
| 부호 있는 정수 | i8부터 i1024 |
| 부호 없는 정수 | u8부터 u1024 |
| 부동소수점 | f32, f64 |
| 기타 | bool, char, byte, str, ptr<T>, array<T, N> |

## 타입 별칭

```wave
type UserId = u64;
var id: UserId = 7;
```

> **릴리스 제한**
> 
> isz와 usz는 예약된 철자이지만 0.2.0-pre-beta 파서가 올바르게 처리하지 못합니다. 이 릴리스에서는 고정 너비 정수 타입을 사용하십시오.

## 배열

```wave
var values: array<i32, 4> = [10, 20, 30, 40];
var first: i32 = values[0];
values[1] = 25;
```

array<T, N>은 요소 타입과 컴파일 시점 길이를 가집니다. 대괄호로 요소에 접근합니다.

