---
translation_set_id: control-flow
path: language/control-flow
locale: ko
group: language
group_order: 2
order: 4
title: 제어 흐름
summary: 괄호가 필요한 조건문과 반복문, C형 for, match, break와 continue를 설명합니다.
---

## 조건문

`if`와 `else if` 조건은 **괄호가 필수**입니다.

```wave
if (score >= 90) {
    println("A");
} else if (score >= 80) {
    println("B");
} else {
    println("C");
}
```

`if score >= 90 { ... }`처럼 괄호를 생략한 형태는 파서가 받지 않습니다.

## 조건식에서는 값을 변경할 수 없음

`if`, `else if`, `while`, `for` 조건식 안에서는 대입, 복합 대입,
`++`, `--`를 어느 위치에서도 사용할 수 없습니다. 비교하려면 `==`를
사용하고, 값을 변경해야 한다면 조건식 앞의 일반 문장으로 옮깁니다.

```wave
value = read_value();
if (value == expected) {
    println("matched");
}
```

일반 대입문과 `for` 문의 증감식은 계속 사용할 수 있습니다.

## while 반복문

`while` 조건도 괄호로 감쌉니다.

```wave
var index: i32 = 0;
while (index < 10) {
    index += 1;
}
```

## for 반복문

`for`는 초기화, 조건, 증감식의 세 부분을 가지는 형태를 사용합니다.

```wave
for (var i: i32 = 0; i < 10; i += 1) {
    println("{}", i);
}
```

초기화에는 `var`, 타입이 명시된 바인딩 또는 일반 표현식을 사용할 수 있습니다. `const`와 `static`은 지역 for 초기화에 사용할 수 없습니다.

## break와 continue

```wave
var i: i32 = 0;
while (i < 20) {
    i += 1;
    if (i == 5) {
        continue;
    }
    if (i == 10) {
        break;
    }
}
```

`break`는 가장 가까운 반복문을 끝내고, `continue`는 다음 반복으로 진행합니다.

## match

`match`의 대상 식도 괄호 안에 둡니다. 현재 패턴은 정수 리터럴, 식별자 형태의 이름, `_` 와일드카드를 처리합니다.

```wave
match (status) {
    200 => { println("ok"); }
    404 => { println("not found"); }
    _ => { println("other"); }
}
```

와일드카드 `_`는 한 `match` 안에 중복해서 둘 수 없습니다. 각 arm 본문은 `{ ... }` 블록이어야 합니다.
