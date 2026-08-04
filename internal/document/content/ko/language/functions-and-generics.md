---
translation_set_id: functions
path: language/functions-and-generics
locale: ko
group: language
group_order: 2
order: 5
title: 함수와 제네릭
summary: 함수 선언, 매개변수, 반환값과 제네릭 코드를 설명합니다.
---

## 함수

```wave
fun add(left: i32, right: i32) -> i32 {
    return left + right;
}

fun log(message: str) {
    println(message);
}
```

매개변수와 반환 타입을 명시합니다. return은 값을 호출자에게 전달하며 결과가 없는 함수는 반환 타입을 생략합니다.

## 제네릭

```wave
fun identity<T>(value: T) -> T {
    return value;
}
```

```wave
var integer: i32 = identity<i32>(10);
var decimal: f64 = identity<f64>(3.14);

struct Pair<A, B> {
    first: A;
    second: B;
}

var pair: Pair<i32, f64> = Pair<i32, f64> { first: 1, second: 2.5 };
```

제네릭 매개변수를 사용하면 함수와 구조체가 명시적으로 전달한 구체 타입에서 동작할 수 있습니다. 이 릴리스의 호출은 생략된 타입 인자를 추론하지 않습니다. ABI 경계에서는 명시적으로 인스턴스화를 정의하지 않는 한 구체 타입을 사용하십시오.

> **포인터 모델과 구분**
> 
> ptr<T>의 꺾쇠 표기는 Wave Explicit Memory Type Model에 속합니다. ptr<T>는 내장 메모리 타입이며 여기서 설명하는 일반 제네릭의 인스턴스가 아닙니다.

