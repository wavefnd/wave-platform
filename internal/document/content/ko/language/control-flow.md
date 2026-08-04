---
translation_set_id: control-flow
path: language/control-flow
locale: ko
group: language
group_order: 2
order: 4
title: 제어 흐름
summary: 조건, 반복, match와 흐름 제어를 설명합니다.
---

## 조건문

```wave
if score >= 90 {
    println("A");
} else if score >= 80 {
    println("B");
} else {
    println("C");
}
```

## 반복문

```wave
var index: i32 = 0;
while index < 10 {
    index += 1;
}

for (item: i32 = 0; item < 10; item += 1) {
    println("{}", item);
}
```

break는 가장 가까운 반복문을 종료하고 continue는 다음 반복을 시작합니다.

## 패턴 선택

```wave
match (status) {
    200 => { println("ok"); }
    404 => { println("not found"); }
    _ => { println("other"); }
}
```

