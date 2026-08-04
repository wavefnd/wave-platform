---
translation_set_id: memory-buffer
path: reference/memory-and-buffer
locale: ko
group: reference
group_order: 3
order: 5
title: 메모리와 버퍼
summary: 수동 할당, 복사, 정렬, 페이지와 가변 바이트 버퍼를 설명합니다.
---

## 수동 할당

```wave
import("std::mem::alloc");
import("std::mem::ops");

var memory: ptr<u8> = mem_alloc(256);
if (memory != null) {
    mem_zero(memory, 256);
    mem_free(memory, 256);
}
```

mem_alloc, mem_alloc_zeroed, mem_realloc, mem_free, 페이지 도우미, 정렬 할당은 명시적 네이티브 메모리를 제공합니다. 호출자는 할당 크기를 유지하고 대응하는 함수로 저장소를 해제해야 합니다.

## 가변 버퍼

std::buffer는 std::mem 위에 구성되며 할당, 읽기, 쓰기, 버퍼 타입을 분리합니다. 반환된 저장소를 사용하기 전에 null과 오류 결과를 확인하십시오.

> **메모리 모델**
> 
> 이 API의 ptr<T>는 전용 Wave Explicit Memory Type Model입니다. 제네릭 소유권, 자동 경계 검사, 가비지 컬렉션을 뜻하지 않습니다.

