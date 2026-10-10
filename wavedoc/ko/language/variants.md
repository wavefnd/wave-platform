---
translation_set_id: language-variants
path: language/variants
locale: ko
group: language
group_order: 2
order: 17
title: variant와 패턴 매칭
summary: 경우별 데이터를 함께 저장하고 match에서 안전하게 분리합니다.
---

## enum과 구분

enum은 이름 붙은 정수 값을 나타내고, variant는 경우마다 다른 payload를 담습니다. 아래 Result는 성공 시 i32 값, 실패 시 i32 오류 번호를 저장합니다. 오류와 결과가 같은 정수 타입이어도 경우 이름으로 의미를 구분할 수 있습니다.

## 선언·생성·검사

`main.wave`에 저장해 실행합니다.

<!-- wave-example: variant-api -->
```wave playground
variant Result {
    Value(i32), Error(i32)
}

fun calculate(valid: bool) -> Result {
    if (!valid) {
        return Result::Error(1);
    }

    return Result::Value(42);
}

fun main() {
    var result: Result = calculate(true);
    match (result) {
        Result::Value(value) => {
            println("value={}", value);
        }
        Result::Error(code) => {
            println("error={}", code);
        }
    }
}
```

실행 결과:

```text
value=42
```

`Result::Value(42)`가 payload를 가진 경우를 만듭니다. `match`의 해당 패턴 안에서만 value라는 이름으로 payload를 사용합니다. 다른 경우의 payload를 강제로 읽지 않습니다. arm 본문은 블록으로 작성합니다. 모든 경우를 처리하거나 `_`로 나머지를 처리합니다. 새로운 경우를 추가했을 때 처리가 필요한지 드러내려면 각 경우를 명시적으로 나누는 편이 좋습니다.

## 제네릭과 수명

`variant Optional<T> { Some(T), None }`처럼 타입 매개변수를 사용할 수 있습니다. 지역 변수 타입에는 `Optional<i32>`처럼 실제 타입을 지정합니다. variant가 포인터를 담아도 메모리 소유권이 자동 관리되지는 않습니다. 값 복사와 할당 복사는 다릅니다.

variant의 메모리 표현을 임의의 C union과 동일하다고 가정하지 마십시오. 외부 ABI로 보낼 데이터는 별도 표현을 정하고 허용된 FFI 타입으로 전달합니다.

## 연습

calculate(false)로 바꾸면 `error=1`이 나와야 합니다. payload가 없는 Empty 경우를 추가한 뒤 match에서도 해당 경우를 처리해 보십시오.

[구조체와 enum](/docs/ko/language/structures-enums-and-aliases) · [오류 처리 수업](/docs/ko/language/errors)
