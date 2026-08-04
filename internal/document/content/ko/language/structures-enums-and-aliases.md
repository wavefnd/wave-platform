---
translation_set_id: data-types
path: language/structures-enums-and-aliases
locale: ko
group: language
group_order: 2
order: 6
title: 구조체, 열거형과 별칭
summary: 복합 데이터, 이름 있는 선택지와 재사용할 타입 이름을 정의합니다.
---

## 구조체

```wave
struct Point {
    x: f64;
    y: f64;
}

var point: Point = Point { x: 1.0, y: 2.0 };
```

## 열거형

```wave
enum State -> i32 {
    Ready = 0,
    Running,
    Stopped,
}
```

구조체는 여러 필드를 하나의 타입으로 묶습니다. 열거형은 닫힌 이름 집합을 정의합니다. 타입 별칭은 기존 타입에 도메인 의미가 있는 이름을 부여합니다.

## 프로토콜 구현

```wave
proto Point {
    fun length_squared(self: Point) -> f64 {
        return self.x * self.x + self.y * self.y;
    }
}

var distance: f64 = point.length_squared();
```

proto 블록은 구조체에 타입이 명확한 동작을 붙입니다. self 매개변수는 수신 타입을 명시합니다.

> **레이아웃**
> 
> 컴파일러와 외부 ABI가 모두 보장하지 않는 한 FFI나 바이너리 데이터 경계에서 네이티브 레이아웃을 가정하지 마십시오.

