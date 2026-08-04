---
translation_set_id: quick-reference
path: reference/syntax-quick-reference
locale: ko
group: reference
group_order: 3
order: 3
title: 문법 빠른 참조
summary: 이 설명서가 다루는 Wave 문법을 짧게 다시 확인합니다.
---

## 선언

```wave
var value: i32 = 1;
var mutable: i32 = 2;
let fixed: i32 = 3;       // 제약된 불변 바인딩
let mut guarded: i32 = 4; // 엄격한 코드용 명시적 가변 바인딩
const LIMIT: u64 = 64;
type Identifier = u64;
```

## 함수

```wave
fun max(left: i32, right: i32) -> i32 {
    if left > right { return left; }
    return right;
}
```

## 복합 타입

```wave
struct Pair { left: i32; right: i32; }
enum Result -> i32 { Ok = 0, Error }
```

## 메모리와 네이티브 경계

```wave
var address: ptr<i32> = raw as ptr<i32>;
var item: i32 = deref address;
extern(c) fun native_call(value: i32) -> i32;
```

