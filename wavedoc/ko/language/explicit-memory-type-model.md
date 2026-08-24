---
translation_set_id: memory-model
path: language/explicit-memory-type-model
locale: ko
group: language
group_order: 2
order: 7
title: 포인터와 명시적 메모리 접근
summary: ptr<T>와 array<T, N>, 주소 취득, deref, null, 포인터 변환과 수동 메모리 규칙을 설명합니다.
---

## **Wave Explicit Memory Type Model**

Wave의 포인터 설계는 **Wave Explicit Memory Type Model**을 기반으로 합니다. 이 모델은 포인터와 배열을 문법적 트릭이나 라이브러리 추상화가 아닌, 언어 차원의 명시적인 메모리 타입으로 정의합니다.

`ptr<T>`는 `T` 값을 저장한 메모리 주소를 가리키는 타입이고, `array<T, N>`은 `T` 값 `N`개를 연속해서 저장하는 고정 길이 메모리 타입입니다. 따라서 함수 인자, 반환값, 구조체 필드와 다른 타입 안에서도 포인터와 배열의 구조가 그대로 드러납니다.

## ptr<T>와 주소 취득

```wave
var value: i32 = 7;
var address: ptr<i32> = &value;
```

`&value`는 `value`의 주소를 얻습니다. `ptr<i32>`는 그 주소를 `i32` 값의 주소로 사용한다는 뜻입니다. 포인터 타입은 중첩할 수 있습니다.

```wave
var first: ptr<i32> = &value;
var second: ptr<ptr<i32>> = &first;
```

## null

```wave
var buffer: ptr<u8> = null;
if (buffer == null) {
    println("no buffer");
}
```

`null`은 유효한 메모리 주소를 가리키지 않는 포인터 값입니다. `null`은 `ptr<T>` 타입에만 대입할 수 있으며 정수, 불리언이나 배열 값으로 사용할 수 없습니다.

메모리 할당이나 검색처럼 결과가 없을 수 있는 함수는 `null`을 반환할 수 있습니다. 이런 결과는 `null`과 비교한 뒤에만 역참조합니다. `null` 포인터를 역참조하면 유효한 저장소에 접근할 수 없습니다.

## 명시적 역참조

```wave
var value: i32 = 7;
var p: ptr<i32> = &value;
var copy: i32 = deref p;
deref p = 42;
```

`deref`는 포인터가 가리키는 저장소를 실제로 읽거나 쓰는 연산입니다. 쓰기 대상은 대입 가능한 저장 위치여야 합니다.

인덱싱한 포인터에도 명시적으로 역참조할 수 있습니다.

```wave
deref bytes[index] = 0;
```

## 포인터 변환

주소나 다른 포인터 표현을 바꿔야 할 때는 `as`를 사용합니다.

```wave
var raw: i64 = 0;
var p: ptr<u8> = raw as ptr<u8>;
```

정수와 포인터 사이의 변환은 저수준 경계에서만 사용하고, 대상 플랫폼의 주소 너비와 ABI를 고려하십시오.

## 포인터 연산

`ptr<T>`에 정수를 더하거나 빼면 `T` 요소 단위로 주소가 이동합니다. 같은 메모리 영역을 가리키는 두 포인터를 빼면 주소 사이의 바이트 차이를 `i64`로 얻습니다.

```wave
var base: ptr<i32> = 0x1000 as ptr<i32>;
var next: ptr<i32> = base + 1;
var distance: i64 = next - base; // i32 한 개의 크기인 4
```

포인터 연산을 할 때는 계산한 주소가 원래 메모리 영역 안에 있고 `T`에 맞게 정렬되어 있는지 확인해야 합니다.

## 포인터와 배열

포인터를 요소 타입의 크기 단위로 이동하거나 인덱싱할 수 있습니다.

```wave
var values: array<i32, 4> = [10, 20, 30, 40];
var base: ptr<i32> = &values[0];
var third: i32 = deref base[2];
```

포인터를 요소로 가지는 배열과 배열 전체를 가리키는 포인터도 서로 다른 타입으로 표현합니다.

```wave
var left: i32 = 10;
var right: i32 = 20;
var addresses: array<ptr<i32>, 2> = [&left, &right];
var block: ptr<array<i32, 4>> = &values;
```

`array<ptr<i32>, 2>`는 포인터 두 개를 저장하는 배열이고, `ptr<array<i32, 4>>`는 정수 네 개로 이루어진 배열 하나를 가리키는 포인터입니다.

## 메모리 안전을 위한 책임

`ptr<T>`를 사용하는 코드는 다음 조건을 직접 지켜야 합니다.

- 할당의 소유자와 해제 책임
- 할당 크기와 인덱스 경계
- 포인터가 유효한 수명
- 정렬과 초기화 상태
- 동시에 접근할 때의 별칭 규칙

FFI나 수동 할당 API에는 포인터를 누가 만들고 해제하는지, 접근할 수 있는 크기는 얼마인지, 어느 타입과 정렬로 읽고 쓰는지를 함께 문서화하십시오.
