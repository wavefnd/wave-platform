---
translation_set_id: memory-model
path: language/explicit-memory-type-model
locale: ko
group: language
group_order: 2
order: 7
title: 포인터와 명시적 메모리 접근
summary: ptr<T>, 주소 취득, deref, null, 포인터 변환과 수동 메모리 계약을 설명합니다.
---

## ptr<T>

`ptr<T>`는 `T`를 대상으로 해석할 포인터 타입입니다.

```wave
var value: i32 = 7;
var address: ptr<i32> = &value;
```

포인터 타입은 주소가 가리키는 값의 의도한 타입을 표현하지만, 그 주소가 실제로 유효한지까지 보장하지는 않습니다.

## null

```wave
var buffer: ptr<u8> = null;
if (buffer == null) {
    println("no buffer");
}
```

코드 생성 단계는 `null`을 포인터 값으로 취급합니다. 포인터를 역참조하기 전에 실패를 나타내는 `null`인지 확인해야 하는 API가 많습니다.

## 명시적 역참조

```wave
var value: i32 = 7;
var p: ptr<i32> = &value;
var copy: i32 = deref p;
deref p = 42;
```

`deref`는 포인터가 가리키는 저장소를 실제로 읽거나 쓰는 연산입니다. 쓰기 대상은 대입 가능한 저장 위치여야 합니다.

인덱싱된 포인터에도 명시적으로 역참조할 수 있습니다.

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

## 컴파일러가 자동으로 보장하지 않는 것

`ptr<T>` 자체는 다음을 자동으로 추적하지 않습니다.

- 할당의 소유자와 해제 책임
- 할당 크기와 인덱스 경계
- 포인터가 유효한 수명
- 정렬과 초기화 상태
- 동시에 접근할 때의 별칭 규칙

따라서 FFI나 수동 할당 API에서는 “누가 만들고, 몇 바이트이며, 언제 해제하고, 어느 타입으로 접근하는지”를 함께 문서화하는 것이 중요합니다.
