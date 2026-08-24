---
translation_set_id: lexical
path: language/lexical-structure
locale: ko
group: language
group_order: 2
order: 1
title: 어휘 구조
summary: 식별자, 리터럴, 구분자, 키워드와 타입 이름을 설명합니다.
---

## 식별자

식별자는 변수, 함수, 타입과 필드 등에 이름을 붙입니다. 이름은 대소문자를 구분하며 문자, 숫자와 `_`를 조합할 수 있습니다. 첫 글자에는 숫자를 사용할 수 없습니다. Unicode 문자도 식별자에 사용할 수 있습니다.

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

## 키워드와 타입 이름

Wave 문법에서 사용하는 주요 키워드는 다음과 같습니다.

`pub`, `fun`, `extern`, `export`, `type`, `enum`, `static`, `var`, `deref`, `const`, `if`, `else`, `proto`, `struct`, `while`, `for`, `in`, `out`, `clobber`, `as`, `asm`, `import`, `return`, `continue`, `print`, `input`, `println`, `match`, `break`, `true`, `false`, `null`.

내장 타입 이름에는 `bool`, `char`, `byte`, `str`, 정수·부동소수점 타입, `ptr`과 `array`가 있습니다. 포인터는 `ptr<T>`, 고정 길이 배열은 `array<T, N>` 형태로 적습니다.
