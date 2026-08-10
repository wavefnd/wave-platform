---
translation_set_id: data-types
path: language/structures-enums-and-aliases
locale: ko
group: language
group_order: 2
order: 6
title: 구조체, 열거형과 타입 별칭
summary: 구조체 필드·메서드, proto 확장, 명시적 표현 타입을 가진 enum과 타입 별칭을 설명합니다.
---

## 구조체

```wave
struct Point {
    x: f64;
    y: f64;
}

fun main() {
    var point: Point = Point { x: 1.0, y: 2.0 };
    println("{}", point.x);
}
```

구조체 필드는 `name: type;` 형태입니다. 필드 접근은 `value.field`을 사용합니다.

## 구조체 내부 메서드

파서는 구조체 본문 안에 `fun` 메서드를 둘 수 있습니다. `proto`는 이미 정의된 구조체에 별도 블록으로 메서드를 붙이는 형태도 제공합니다.

```wave
struct Counter {
    value: i32;
}

proto Counter {
    fun current(self: Counter) -> i32 {
        return self.value;
    }
}
```

호출은 일반적인 메서드 표기를 사용합니다.

```wave
var counter: Counter = Counter { value: 3 };
var value: i32 = counter.current();
```

## 열거형

이 릴리스의 enum 문법은 기반 표현 타입을 `->` 뒤에 명시합니다.

```wave
enum State -> i32 {
    Ready = 0,
    Running,
    Stopped,
}
```

variant에는 정수 값을 직접 지정할 수 있고, 생략된 값은 컴파일러가 앞 값에 이어 부여합니다.

## 타입 별칭

```wave
type FileHandle = i64;
```

별칭은 기존 타입을 다른 이름으로 참조할 수 있게 합니다.

## ABI와 레이아웃

Wave 구조체를 외부 ABI나 바이너리 파일 형식과 직접 공유할 때는 필드 순서만 보고 C 구조체와 같은 레이아웃을 가정하지 마십시오. ABI 경계에서 필요한 표현은 대상 ABI와 컴파일러가 제공하는 보장에 맞춰 명시적으로 설계하십시오.
