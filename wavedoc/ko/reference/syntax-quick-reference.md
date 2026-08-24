---
translation_set_id: quick-reference
path: reference/syntax-quick-reference
locale: ko
group: reference
group_order: 3
order: 3
title: 문법 빠른 참조
summary: 현재 자주 쓰는 선언, 제어 흐름, 타입, 포인터, FFI 문법을 한 페이지에 정리합니다.
---

## 선언

```wave
var value: i32 = 1;
var fixed: i32 = 2;
var mutable: i32 = 3;
const LIMIT: i32 = 64;
static total: i64 = 0;
type Identifier = u64;
```

`var`는 지역, `const`/`static`은 최상위 선언입니다.

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

match (status) {
    Ready => { println("ready"); }
    0 => { println("zero"); }
    _ => { println("other"); }
}
```

`if`, `while`, `for`, `match`의 헤더는 이 릴리스에서 괄호를 사용합니다.

## 배열과 포인터

```wave
var values: array<i32, 4> = [1, 2, 3, 4];
var p: ptr<i32> = &values[0];
var first: i32 = deref p;
```

## 콘솔 입출력

```wave
print("value = ");
println("{}", value);
input("{}", value);
```

첫 인자는 문자열 리터럴입니다. 정확한 `{}` 자리표시자마다 뒤따르는 식이 하나씩 필요하며 `input` 대상은 대입 가능해야 합니다.

## import와 FFI

```wave
import("std::string::len");
extern(c) fun native_call(value: i32) -> i32;

export(c) fun wave_call(value: i32) -> i32 {
    return value + 1;
}
```

## 대상 조건부 항목

```wave
#[target(os="linux", arch="riscv64")]
extern(c) fun platform_call(value: i32) -> i32;
```

지원 조건 키는 `arch`, `os`, `env`, `abi`입니다. 속성은 바로 다음 최상위 항목을 제어합니다.

## 인라인 어셈블리

```wave
var result: i64 = 0;
asm {
    "mv a0, a1"
    in("a1") 7
    out("a0") result
    clobber("memory")
}
```

명령 텍스트와 레지스터 이름은 대상에 종속됩니다. 블록에 필요한 모든 입력, 출력과 숨은 clobber를 선언하십시오.

## 컴파일러 확인

```shell
wavec build main.wave --emit=check
wavec print supported-targets
wavec print supported-emit-kinds
```
