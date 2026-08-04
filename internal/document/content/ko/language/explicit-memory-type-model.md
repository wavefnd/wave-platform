---
translation_set_id: memory-model
path: language/explicit-memory-type-model
locale: ko
group: language
group_order: 2
order: 7
title: Wave Explicit Memory Type Model
summary: 포인터, 명시적 역참조, 주소 연산과 네이티브 메모리 안전 규칙을 설명합니다.
---

Wave는 저수준 메모리 접근이 소스 코드에 드러나도록 합니다. 포인터는 대상 타입의 값이 아니며, 포인터 생성, 주소 이동, 변환, 역참조는 서로 구분되는 연산입니다. 이 분리를 Wave Explicit Memory Type Model이라고 합니다.

> **전용 메모리 타입 문법**
> 
> ptr<T>는 Wave Explicit Memory Type Model이 직접 정의합니다. T 자리는 메모리 요소 타입을 설명하며 일반 제네릭 타입 인자가 아닙니다. ptr 자체도 사용자 정의 제네릭이 아닙니다.

## 포인터 타입

```wave
var raw: u64 = 0;
var typed: ptr<i32> = raw as ptr<i32>;
```

ptr<T>는 네이티브 주소가 가리킬 의도한 대상 타입을 기록합니다. 타입 포인터도 할당 수명, 정렬, 초기화, 소유권을 보장하지 않습니다.

```wave
var value: i32 = 7;
var address: ptr<i32> = &value;
```

주소 연산자 &는 기존 저장소를 가리키는 포인터를 만듭니다. 포인터가 그 저장소보다 오래 살아서는 안 됩니다.

## 명시적 역참조

```wave
var value: i32 = deref typed;
deref typed = 42;
```

deref는 주소를 가진 값에서 실제 메모리 접근으로 넘어가는 명시적 연산입니다. 역참조 전에는 포인터가 null이 아니고, T에 맞게 정렬되며, 초기화된 저장소를 가리키고, 접근 동안 유효하며, 요청한 읽기나 쓰기를 허용하는지 확인해야 합니다.

## 포인터 이동

포인터 산술은 주소를 이동하며 외부 할당 계약이 설명하는 저장소 범위를 벗어나면 안 됩니다. ptr<T>만으로 소유권이나 경계를 추론하지 마십시오.

> **네이티브 메모리**
> 
> FFI 경계에서 사라진 할당 계약을 컴파일러가 복원할 수는 없습니다. 할당, 크기, 소유권, 해제 규칙을 좁은 래퍼 안에 함께 두십시오.

## 경계 사용 절차

- 문서화된 네이티브 함수에서 메모리를 받거나 할당합니다.
- 변환 전에 실패와 null을 확인합니다.
- 필요한 가장 좁은 타입 포인터로 변환합니다.
- 수명이 유효한 동안에만 명시적으로 역참조합니다.
- 메모리를 만든 할당자에 대응하는 함수로 해제합니다.
