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

구조체 본문에는 `fun` 메서드를 둘 수 있습니다. `proto`를 사용하면 이미 정의한 구조체의 메서드를 별도 블록에 작성할 수 있습니다.

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

열거형은 값을 표현할 정수 타입을 `->` 뒤에 명시합니다.

```wave
enum State -> i32 {
    Ready = 0,
    Running,
    Stopped,
}
```

variant에 정수 값을 직접 지정할 수 있습니다. 첫 variant의 값을 생략하면 `0`부터 시작하고, 이후 생략된 값은 바로 앞 값보다 `1` 큽니다.

## 타입 별칭

```wave
type FileHandle = i64;
```

타입 별칭은 같은 타입을 코드의 문맥에 맞는 이름으로 표현하는 문법입니다. `FileHandle`로 선언한 값은 `i64` 값처럼 그대로 사용할 수 있고 함수 인자와 반환 타입에도 쓸 수 있습니다.

## ABI와 레이아웃

Wave 구조체를 외부 ABI나 바이너리 파일 형식과 공유하려면 대상 ABI가 요구하는 크기, 정렬과 필드 배치를 명시적으로 맞춰야 합니다. 필드 순서만으로 다른 언어의 구조체와 같은 메모리 배치를 가정하지 마십시오.
