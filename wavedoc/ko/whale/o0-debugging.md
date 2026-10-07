---
translation_set_id: whale-o0-debugging
path: whale/o0-debugging
locale: ko
group: whale
group_order: 1
order: 9
title: O0와 디버깅
summary: 계산, 상수식, 저장 공간과 디버그 대응 정보를 보존하는 규칙입니다.
---

## 보존 모델

O0는 디버깅을 위해 원본 typed IR의 계산·변수·제어 흐름을 보존합니다. 사용하지 않는 결과와 구조적으로 도달 불가능한 블록도 유지합니다. 검증은 잘못된 IR을 진단하며 블록을 삭제하거나 연산을 단순화하지 않습니다.

기계어를 생성하는 데 필요한 변환은 별도의 하위 IR에서 수행합니다. 변환 이후에도 원본 ID와의 대응을 유지해야 합니다. typed IR 보존이 모든 IR 연산과 기계 명령의 일대일 대응을 뜻하지는 않습니다.

## O0에서 수행하지 않는 변환

| 변환 | O0 동작 |
| --- | --- |
| 인라이닝 | 호출과 함수 경계를 유지 |
| Tail-call 변환 | 일반 호출·반환 구조를 유지 |
| 죽은 코드 제거 | 미사용 계산과 도달 불가능한 블록을 유지 |
| 런타임 상수 접기 | 원래 연산을 유지 |
| 공통 계산 합치기 | 별개의 계산을 유지 |
| 지역변수 저장 공간 재사용 | 각각의 저장 공간을 유지 |
| 프레임 포인터 생략 | 프레임 포인터를 유지 |
| 문자열 자동 병합 | 별개의 문자열 객체를 유지 |

예를 들어 두 상수를 더하는 런타임 연산은 결과를 사용하지 않더라도 덧셈으로 남습니다. 조건이 상수인 분기도 원래 제어 흐름 구조를 유지합니다.

## 컴파일 시점 상수

컴파일 시점 상수 선언은 typed 초기화 식과 평가 결과를 함께 보존합니다. 이는 일반 런타임 명령의 상수 접기와 구별됩니다.

초기화 식이 `1 + 2`인 선언에는 덧셈 식과 결과 `3`이 함께 남습니다. 이 예제는 식의 의미를 설명하며 특정 소스 언어의 선언 문법이 아닙니다. 결과는 정적 데이터를 구성할 때 사용할 수 있으며, 초기화 식을 런타임 덧셈으로 바꾸지는 않습니다.

미사용·도달 불가능한 선언도 식별할 수 있어야 합니다. 같은 이름을 가진 선언이 서로를 가리더라도 이름·선언 ID·참조로 구분되어야 합니다. 검증은 잘못된 참조, 의존성 순환, 잘못된 타입, 원래 식과 일치하지 않는 저장 결과를 거부합니다.

## 도달 불가능한 소스 문장

return·break·continue 뒤의 문장도 연결되지 않은 블록에 남습니다. 이 문장 때문에 앞의 terminator를 바꾸거나 새로운 실행 경로를 만들어서는 안 됩니다. 도달 불가능한 문장 안의 잘못된 표현식도 진단합니다.

이 보존 규칙으로 원래 프로그램 구조를 조사할 수 있습니다. terminator 뒤의 문장이 실제로 실행된다는 뜻은 아닙니다.

## 보존되는 IR 예제

다음은 검증기를 통과한 모듈의 현재 프린터 출력입니다. 사용하지 않는 계산, 컴파일 시점 선언, 연결되지 않은 블록을 함께 보여 줍니다.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "preserved": whale () -> void, linkage internal

  fn @f0 "preserved"() -> void, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 1
    %v1: i32 = const i32 2
    %v2: i32 = add i32 %v0, %v1
    %v3: i32 = const_decl "count" add(i32 1, i32 2) => const i32 3
    ret void
  %b1 "unreachable.cont":
    %v4: i32 = const i32 4
    %v5: i32 = const i32 5
    %v6: i32 = add i32 %v4, %v5
    ret void
  }

}
```

`%v2`는 사용처가 없어도 `add`로 남습니다. `%v3`는 별도의 `const_decl`이며, `add(i32 1, i32 2)` 식과 평가 결과 3을 함께 보존합니다. `unreachable.cont`에는 들어오는 간선이 없으므로 `ret void` 뒤에 실행 경로를 추가하지 않으면서 `%v6` 덧셈을 확인할 수 있습니다.

검증은 이 명령과 블록을 보존합니다. 이 예제는 IR 구성·검증·출력을 보여 주며 native 실행이나 DWARF 출력이 제공된다는 뜻은 아닙니다.

## 소스와 스택 정보

디버그 인터페이스는 DWARF 5의 함수 정보, 소스 행, 기본 지역변수, 호출 프레임 정보를 사용합니다. 하위 표현으로 변환하더라도 함수·지역변수 정보와 원본 IR 식별자의 대응을 유지해야 합니다.

AMD64 프로파일은 프레임 포인터를 유지하고 red zone을 사용하지 않습니다. 호출 프레임 정보는 스택 조사에 사용하며 예외 unwinding 지원을 뜻하지 않습니다. trap은 소멸자나 unwinding을 보장하지 않고 실행을 종료합니다.

DWARF 출력과 native 실행의 제공 여부는 [툴체인 개요](overview)를 참고하세요. O1 이상 최적화의 동작은 이 O0 참조의 범위 밖입니다.

## 선언이 반복되는 루프

이 전체 모듈을 `initialization-loop.wir`로 저장하세요. 첫 반복에서 42를 저장해도 두 번째 선언의 uninit이 초기화 상태를 다시 지웁니다. 따라서 두 번째 반복의 읽기는 trap입니다. O0 lowering은 alloca를 진입 블록에 두어도 uninit을 실제 선언 위치에 유지합니다. 사용하지 않는 읽기도 실행되면 검사하고, 보존된 도달 불가능 블록은 실행하지 않습니다.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "redeclaration": whale () -> u32, linkage internal

  fn @f0 "redeclaration"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    %v1: u32 = const u32 0
    %v2: u32 = const u32 1
    %v3: u32 = const u32 42
    br label %b1
  %b1 "declaration":
    %v4: u32 = phi u32 [ %v1, %b0 ], [ %v6, %b2 ]
    uninit u32, ptr<u32> %v0, align 4
    %v5: bool = icmp eq u32 %v4, %v1
    cbr bool %v5, label %b2, label %b3
  %b2 "first_iteration":
    store u32 %v3, ptr<u32> %v0, align 4
    %v6: u32 = add u32 %v4, %v2
    br label %b1
  %b3 "second_iteration":
    %v7: u32 = load u32, ptr<u32> %v0, align 4
    ret u32 %v7
  }

}
```

```shell
whale ir run initialization-loop.wir --function @f0
```

```text
Error: initialization-loop.wir: trap at @f0 %b3 instruction 0 (%v7): uninitialized byte at allocation offset 0 (after 17 steps)
```
