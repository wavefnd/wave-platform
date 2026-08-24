---
translation_set_id: assembly
path: language/inline-assembly
locale: ko
group: language
group_order: 2
order: 9
title: 인라인 어셈블리
summary: asm 블록의 명령 문자열, in/out 피연산자와 clobber 계약을 설명합니다.
---

## asm 블록

`asm`은 대상 아키텍처의 명령을 직접 삽입하기 위한 저수준 문법입니다.

```wave
fun read_value() -> i64 {
    var result: i64 = 0;
    asm {
        "mov rax, 123"
        out("rax") result
    }
    return result;
}
```

블록 안의 문자열 리터럴은 어셈블리 명령 목록으로 전달됩니다.

## 입력과 출력

```wave
var result: i64 = 0;
asm {
    "mov rax, rdi"
    in("rdi") 123
    out("rax") result
}
```

- `in("reg") expression`은 Wave 값을 입력 피연산자에 연결합니다.
- `out("reg") target`은 출력 값을 대입 가능한 Wave 대상에 기록합니다.
- 레지스터 이름은 문자열 또는 식별자 형태로 적을 수 있습니다.

입력 피연산자에는 변수, 정수·문자열 리터럴, `&identifier`, `deref identifier`와 음수 숫자를 사용할 수 있습니다.

## clobber

블록이 명시적 출력 외의 레지스터나 메모리 상태를 바꾸면 `clobber(...)`에 기록합니다.

```wave
asm {
    "nop"
    clobber("rax", "rcx", "memory")
}
```

## 사용할 때 확인할 것

- 명령어 문법은 대상 아키텍처와 LLVM 인라인 어셈블리 계약에 맞아야 합니다.
- 호출 규약상 보존해야 하는 레지스터를 임의로 파괴하지 마십시오.
- 메모리를 읽거나 쓰는 블록은 `memory`를 포함하여 필요한 clobber를 선언하십시오.
- 가능한 경우 아키텍처별 asm을 작은 함수 뒤에 격리하십시오.

인라인 어셈블리의 동작과 이식성은 언어 타입만으로 보장되지 않습니다.
