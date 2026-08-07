---
translation_set_id: memory-buffer
path: reference/memory-and-buffer
locale: ko
group: reference
group_order: 3
order: 5
title: 메모리와 버퍼
summary: std::mem의 수동 할당·복사·정렬·페이지 도우미와 std::buffer 사용 원칙을 설명합니다.
---

## 기본 수동 할당

```wave
import("std::mem::alloc");
import("std::mem::ops");

fun main() {
    let size: i64 = 256;
    let mut memory: ptr<u8> = mem_alloc(size);

    if (memory != null) {
        mem_zero(memory, size);
        mem_free(memory, size);
    }
}
```

`mem_alloc(size)`는 `ptr<u8>`를 반환하고 실패 시 `null`을 반환할 수 있습니다. `mem_free`는 포인터와 원래 할당 크기를 함께 받습니다.

## zeroed와 realloc

`std::mem::alloc`에는 다음 계열이 구현되어 있습니다.

- `mem_alloc`
- `mem_alloc_zeroed`
- `mem_realloc`
- `mem_free`
- 제네릭 item 할당·재할당·해제 도우미
- 페이지 수와 페이지 정렬 크기 도우미
- 정렬 할당과 해제 도우미

`mem_realloc`은 새 저장소를 만들고 필요한 범위를 복사한 뒤 이전 저장소를 해제하는 구현이므로, 실패했을 때 기존 포인터의 처리 규칙을 함수 구현과 함께 확인하십시오.

## 크기 단위

메모리 함수의 주요 크기 인자는 `i64` 바이트 수입니다. 호출부에서 요소 개수와 바이트 수를 혼동하지 않도록 변수 이름에 단위를 드러내는 편이 좋습니다.

```wave
let count: i64 = 32;
let elem_size: i64 = 4;
let bytes: i64 = count * elem_size;
```

## 포인터 접근

수동 할당으로 얻은 `ptr<u8>`는 경계 정보를 포함하지 않습니다. `deref p[index]`로 접근하기 전에 `index`가 할당 크기 안에 있는지 호출자가 확인해야 합니다.

## std::buffer

`std::buffer`는 `std::mem` 위에 구성된 버퍼 도우미를 제공합니다. 버퍼 API를 사용할 때도 할당 실패, 길이와 capacity, 소유권과 해제 시점을 API 계약에 맞게 관리해야 합니다.

수동 메모리를 다루는 코드는 할당과 해제를 가능한 한 같은 추상화 계층에 두면 누수와 크기 불일치를 줄이기 쉽습니다.
