---
translation_set_id: assembly
path: language/inline-assembly
locale: ko
group: language
group_order: 2
order: 9
title: 인라인 어셈블리
summary: 명시적인 저수준 경계에서 아키텍처별 명령을 사용합니다.
---

## 어셈블리 블록

asm은 아키텍처별 기능을 위한 탈출구입니다. 입력, 출력, 변경되는 상태는 명령 시퀀스를 정확하게 설명해야 합니다.

```wave
var result: i64;
asm {
    "mov rax, 123"
    in("rdi") 1
    out("rax") result
}
```

in은 Wave 값을 입력 레지스터에 연결하고 out은 출력 레지스터 값을 Wave 변수에 기록합니다. 필요하다면 clobber 선언으로 블록이 변경하는 추가 상태를 밝혀야 합니다.

> **이식성**
> 
> 인라인 어셈블리는 대상 아키텍처, ABI, 컴파일러 계약에 종속됩니다. 타입이 명확한 Wave 함수 뒤에 격리하고 가능하면 어셈블리를 사용하지 않는 구현도 제공하십시오.

