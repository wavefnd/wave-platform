---
translation_set_id: storage-duration
path: language/storage-duration
locale: ko
group: language
group_order: 2
order: 12
title: 저장 수명과 변경 가능성
summary: var, const와 static의 범위와 쓰기 가능성을 구분합니다.
---

## 선언별 의미

| 형식 | 허용 위치 | 재대입 | 용도 |
| --- | --- | --- | --- |
| `var` | 함수·블록 | 가능 | 일반적인 가변 지역 변수 |
| `const` | 최상위 | 불가 | 전역 상수 선언 |
| `static` | 최상위 | 가능 | 프로그램 수명 동안 존재하는 정적 저장 선언 |

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;

fun main() {
    var limit: i32 = 4;
    var current: i32 = 0;
    var retries: i32 = 0;

    current += 1;
    retries += 1;
    println("{} {} {}", limit, current, retries);
}
```

## 지역 선언 규칙

```wave
var value: i32 = 1;
value = 2;
```

위 재대입은 허용됩니다. Wave에는 별도의 불변 지역 바인딩 문법이
없습니다. `let`과 `let mut`은 폐지됐으며, 컴파일 시간 상수가 필요하면
최상위에서 `const`를 사용합니다.

## const와 static의 지역 사용

함수 파서는 함수 본문 안의 `const`와 `static`을 명시적으로 거부합니다. 같은 제한은 `for` 초기화에도 적용됩니다.

## 수명과 포인터

지역 변수의 주소를 `&`로 얻을 수 있지만, 포인터가 가리키는 저장소의 실제 유효 기간을 `ptr<T>` 타입이 추적하지는 않습니다. 지역 저장소의 주소를 함수 밖으로 넘길 때는 그 주소가 계속 유효한지 프로그램 구조에서 직접 보장해야 합니다.
