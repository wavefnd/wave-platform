---
translation_set_id: expressions
path: language/expressions-and-operators
locale: ko
group: language
group_order: 2
order: 3
title: 표현식과 연산자
summary: 산술, 비교, 논리, 대입, 형 변환과 우선순위를 설명합니다.
---

## 연산자 분류

| 용도 | 연산자 |
| --- | --- |
| 산술 | + - * / % |
| 비교 | == != < <= > >= |
| 논리 | && || ! |
| 비트 | & | ^ ~ << >> |
| 대입 | = 및 복합 대입 |
| 명시적 형 변환 | as |

```wave
var width: i32 = 12;
var area: i32 = width * 8;
var large: bool = area >= 64;
var widened: i64 = area as i64;
```

서로 다른 연산자 분류를 섞을 때 평가 순서가 즉시 분명하지 않다면 괄호를 사용하십시오. 명시적 형 변환은 as를 사용합니다.

> **예약 연산**
> 
> is와 xnand를 포함한 일부 토큰은 lexer에 예약되어 있지만 0.2.0-pre-beta의 검증된 표현식 연산자는 아닙니다.

