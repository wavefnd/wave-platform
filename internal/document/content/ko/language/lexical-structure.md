---
translation_set_id: lexical
path: language/lexical-structure
locale: ko
group: language
group_order: 2
order: 1
title: 어휘 구조
summary: 식별자, 리터럴, 주석, 구분자와 예약어를 설명합니다.
---

## 식별자와 문장

식별자는 선언에 이름을 부여하며 대소문자를 구분합니다. 문장은 보통 세미콜론으로 끝나고 중괄호는 블록을 구분합니다.

```wave
var answer: i32 = 42;
var greeting: str = "hello";
var enabled: bool = true;
```

## 리터럴

- 정수와 부동소수점 리터럴은 수 값을 표현합니다.
- 문자열은 큰따옴표, 문자는 작은따옴표를 사용합니다.
- true, false, null은 내장 리터럴입니다.

## 주석

```wave
// 한 줄 주석
/* 블록 주석 */
```

## 예약어

fun, extern, export, type, enum, static, var, deref, let, mut, const, if, else, proto, struct, while, for, module, class, in, out, clobber, is, as, asm, xnand, import, return, continue, print, input, println, match, break, true, false, null은 이 릴리스의 예약어입니다.

