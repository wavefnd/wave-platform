---
translation_set_id: quick-reference
path: reference/syntax-quick-reference
locale: ko
group: reference
group_order: 3
order: 3
title: 문법 빠른 참조
summary: v0.2.0-pre-beta에서 자주 쓰는 선언, 제어 흐름, 타입, 포인터, FFI 문법을 한 페이지에 정리합니다.
---

## 선언

```wave
var value: i32 = 1;
let fixed: i32 = 2;
let mut mutable: i32 = 3;
const LIMIT: i32 = 64;
static total: i64 = 0;
type Identifier = u64;
```

`var`/`let`은 지역, `const`/`static`은 최상위 선언입니다.

## 함수

```wave
fun max(left: i32, right: i32) -> i32 {
    if (left > right) {
        return left;
    }
    return right;
}
```

## 제네릭

```wave
fun identity<T>(value: T) -> T {
    return value;
}

var value: i32 = identity<i32>(10);
```

제네릭 호출에는 타입 인자를 명시합니다.

## 구조체와 enum

```wave
struct Pair {
    left: i32;
    right: i32;
}

enum Result -> i32 {
    Ok = 0,
    Error,
}
```

## 조건과 반복

```wave
if (ready) {
    println("ready");
}

while (count < 10) {
    count += 1;
}

for (var i: i32 = 0; i < 10; i += 1) {
    println("{}", i);
}
```

`if`, `while`, `for`, `match`의 헤더는 이 릴리스에서 괄호를 사용합니다.

## 배열과 포인터

```wave
var values: array<i32, 4> = [1, 2, 3, 4];
var p: ptr<i32> = &values[0];
var first: i32 = deref p;
```

## import와 FFI

```wave
import("std::string::len");
extern(c) fun native_call(value: i32) -> i32;

export(c) fun wave_call(value: i32) -> i32 {
    return value + 1;
}
```

## 컴파일러 확인

```shell
wavec build main.wave --emit=check
wavec print supported-targets
wavec print supported-emit-kinds
```
