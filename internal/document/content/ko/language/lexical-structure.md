---
translation_set_id: lexical
path: language/lexical-structure
locale: ko
group: language
group_order: 2
order: 1
title: 어휘 구조
summary: 식별자, 리터럴, 구분자, 타입 철자와 예약 토큰을 구분해 설명합니다.
---

## 식별자

식별자는 변수, 함수, 타입, 필드 등에 이름을 붙입니다. 이름은 대소문자를 구분합니다. lexer는 식별자 내부에서 문자, 숫자와 `_`를 받아들이며 Unicode 문자 분류를 사용합니다.

```wave
var request_count: i64 = 0;
var 이름: str = "Wave";
```

실제 프로젝트에서는 도구 호환성과 검색 편의성을 위해 일관된 이름 규칙을 정해 사용하는 것이 좋습니다.

## 문장과 구분자

대부분의 선언과 표현식 문장은 `;`로 끝납니다. 함수·조건문·반복문·구조체처럼 본문을 가지는 구문은 `{ ... }` 블록을 사용합니다.

```wave
var answer: i32 = 42;

fun double(value: i32) -> i32 {
    return value * 2;
}
```

## 리터럴

```wave
var integer: i32 = 42;
var decimal: f64 = 3.14;
var text: str = "Wave";
var letter: char = 'W';
var enabled: bool = true;
var address: ptr<u8> = null;
```

정수·부동소수점·문자열·문자·불리언과 `null` 리터럴을 사용할 수 있습니다. `null`은 포인터 값에 사용하십시오.

## 예약 키워드와 타입 철자

lexer가 별도 토큰으로 인식하는 주요 키워드는 다음과 같습니다.

`fun`, `extern`, `export`, `type`, `enum`, `static`, `var`, `deref`, `let`, `mut`, `const`, `if`, `else`, `proto`, `struct`, `while`, `for`, `module`, `class`, `in`, `out`, `clobber`, `is`, `as`, `asm`, `xnand`, `import`, `return`, `continue`, `print`, `input`, `println`, `match`, `break`, `true`, `false`, `null`.

`char`, `byte`, 정수·부동소수점 타입 철자도 lexer에서 타입 토큰으로 처리됩니다. `ptr`과 `array`는 일반 식별자 토큰으로 읽힌 뒤 타입 파서가 `ptr<T>`, `array<T, N>` 형태를 해석합니다.

> **예약과 지원은 다름**
>
> `module`, `class`, `is`, `xnand`처럼 lexer에 예약된 철자가 모두 완성된 문법 기능을 의미하지는 않습니다. 문법 페이지에서 실제 사용 형태가 설명된 기능만 지원 기능으로 취급하십시오.
