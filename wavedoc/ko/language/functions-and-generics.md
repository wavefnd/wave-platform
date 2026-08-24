---
translation_set_id: functions
path: language/functions-and-generics
locale: ko
group: language
group_order: 2
order: 5
title: 함수와 제네릭
summary: 함수 선언, 기본 매개변수, 반환과 명시적 제네릭 인스턴스화를 설명합니다.
---

## 함수 선언

```wave
fun add(left: i32, right: i32) -> i32 {
    return left + right;
}

fun log(message: str) {
    println("{}", message);
}
```

매개변수는 `name: type` 형태로 선언합니다. 반환값이 있는 함수는 `-> type`을 사용하고, 반환 타입을 생략한 함수는 값 없는 함수로 사용합니다.

## return

```wave
fun choose(flag: bool) -> i32 {
    if (flag) {
        return 1;
    }
    return 0;
}
```

값을 반환하지 않는 함수에서는 `return;`을 사용할 수 있습니다.

## 제네릭 함수

```wave
fun identity<T>(value: T) -> T {
    return value;
}

fun main() {
    var integer: i32 = identity<i32>(10);
    var decimal: f64 = identity<f64>(3.14);
}
```

제네릭 함수 호출은 **타입 인자를 명시해야 합니다**. `identity(10)`처럼 타입 인자를 생략하면 제네릭 함수에 대해 오류가 발생합니다.

## 제네릭 구조체

```wave
struct Pair<A, B> {
    first: A;
    second: B;
}

fun main() {
    var pair: Pair<i32, f64> = Pair<i32, f64> {
        first: 1,
        second: 2.5
    };
}
```

제네릭 함수와 구조체를 사용할 때 지정한 타입 조합마다 구체적인 형태가 만들어집니다.

## 기본 매개변수

기본 매개변수 값에는 정수, 부동소수점과 문자열 리터럴을 사용할 수 있습니다. 호출에서 해당 인자를 생략하면 선언에 적은 기본값을 사용합니다.

```wave
fun repeat(value: i32, count: i32 = 1) -> i32 {
    return value * count;
}

var result: i32 = repeat(7);
```

## 제네릭 규칙

- 제네릭 함수 호출에는 명시적 타입 인자가 필요합니다.
- `export(...)`로 내보내는 함수는 제네릭일 수 없습니다.
- `ptr<T>`와 `array<T, N>`은 언어가 제공하는 명시적인 메모리 타입이며 사용자 정의 제네릭 템플릿이 아닙니다.
