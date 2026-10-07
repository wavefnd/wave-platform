---
translation_set_id: whale-amd64-target
path: whale/amd64-target
locale: ko
group: whale
group_order: 1
order: 10
title: AMD64 타깃과 ABI
summary: Linux AMD64 타깃 속성, 호출 경계와 native 기능 범위를 설명합니다.
---

## 타깃 식별자

native 프로파일의 타깃 식별자는 `x86_64-whale-linux`입니다.

| 속성 | 값 |
| --- | --- |
| 운영체제 | Linux |
| 명령 집합 | AMD64 |
| 바이트 순서 | Little endian |
| Native 주소 폭 | 64비트 |
| 오브젝트 형식 | ELF64 |
| C 호출 규약 | SysV AMD64 ABI |
| 정적 실행 파일 형식 | ELF ET_EXEC |

빌드 호스트와 출력 타깃은 서로 다른 개념입니다. 다른 호스트에서 Whale을 실행할 수 있다고 해당 호스트의 명령 집합이나 오브젝트 형식을 출력할 수 있는 것은 아닙니다. 구현된 컴파일 경로는 [지원 현황](overview)을 참고하세요.

## 타깃 선택과 검증

실험적 `ir lower` 명령은 `x86_64-whale-linux`를 기본값이자 유일한 지원 타깃으로 사용합니다. 이 명령을 사용하려면 `--features socket-cli`로 빌드합니다.

```sh
whale ir lower program.json --target x86_64-whale-linux
```

출력 타깃은 빌드 호스트와 무관하게 64비트 little-endian 데이터 레이아웃을 제공합니다. 알 수 없는 식별자나 `aarch64-whale-linux`, `x86_64-whale-windows` 같은 미지원 조합은 입력을 읽거나 출력 파일을 교체하기 전에 지원 타깃을 안내하는 오류로 실패합니다. `--no-verify`로도 타깃 선택 검사를 끌 수 없습니다.

빈 AST(`{"format_version":2,"semantics_version":1,"features":[],"program":{"declarations":[],"globals":[],"functions":[]}}`)를 입력하면 다음 헤더와 모듈을 출력합니다.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

}
```

미지원 타깃을 지정하면 실패 상태로 종료합니다.

```sh
whale ir lower program.json --target banana
```

```text
Error: unsupported target "banana"; supported targets: x86_64-whale-linux
```

Rust에서는 `ir::Target::lookup("x86_64-whale-linux")`로 타깃을 선택하고, 그 `name()`과 `data_layout()`을 `lower_o0`에 전달할 수 있습니다. Lowering은 전달된 레이아웃이 선택한 타깃과 다르면 거부합니다. 직접 구성한 IR도 `verify_module`에서 미지원 타깃 이름과 레이아웃 불일치를 검사합니다.

오브젝트 모델은 `ObjectTarget`에 `format`, `machine`, `endian`, `address_bits`를 저장합니다. `ObjectFile::with_target`은 명시한 식별 정보를 보존하며, 직렬화는 AMD64·little-endian·64비트 ELF64 조합만 허용합니다. 다른 아키텍처의 machine 식별자가 있다는 사실은 해당 인코더의 지원을 뜻하지 않습니다. 두 ELF writer 진입점 모두 메타데이터를 검사하며, 링커 입력은 심볼 해석이나 링크 전에 검사합니다. 지원되는 오브젝트의 ELF 헤더는 `EM_X86_64`를 유지합니다.

`ObjectFile::new(ObjectFormat::ELF64)`는 기존처럼 이 AMD64 식별 정보를 사용하는 편의 생성자입니다. 기존에 `object.format`에 접근하던 코드는 `object.target.format`을 사용해야 합니다.

이 검사는 타깃 선택과 오브젝트 식별을 제공합니다. 구조체·배열 등의 크기, 필드 offset, stride는 IR 레이아웃 API로 조회할 수 있습니다. 오브젝트 파일 읽기, native ABI lowering, 실행 파일 링크는 아직 지원하지 않습니다. 스칼라 lowering은 명시적인 정렬을 계속 지정하며, 전체 복합 타입 레이아웃 규칙은 [메모리 참조 문서](memory-model)에 정의되어 있습니다.

## 호출과 서명

IR 선언·호출 검증기는 명시적인 `Whale`·`SysV64` 규약을 받습니다. 규약은 `fnptr` 타입에 포함되며 호출 위치와 일치해야 합니다. 가변 인자 서명과 SysV64 복합 서명은 거부합니다. [실행 가능한 호출 구성 예제](ir-reference)를 참고하세요. 이는 IR 계약 검증이며, 기계 ABI 분류·인자의 레지스터 및 스택 배치·native 호출 생성은 아직 미지원입니다.

지원되는 호출에는 명시적인 서명과 호출 규약이 필요합니다. 미지원 서명은 오류입니다. 백엔드는 인자를 누락하거나 다른 표현으로 대체해 비슷한 호출을 만들어서는 안 됩니다.

서명 지원은 기본 정수·포인터, f32/f64, 구조체·넓은 값, 가변 인자로 구분합니다. IR 타입 체계에 타입이 있다고 해당 타입의 인자 전달이나 반환까지 ABI에서 지원된다는 뜻은 아닙니다. 백엔드를 선택할 때 필요한 범주의 지원 여부를 각각 확인해야 합니다.

예를 들어 i32 반환을 처리하는 백엔드도 지원하지 않는 구조체 반환 서명은 거부해야 합니다. 구조체의 일부가 스칼라 레지스터에 들어간다는 이유로 스칼라 반환 규약을 적용할 수는 없습니다.

## 내부 호출과 C 경계

native 포인터 주소는 64비트입니다. 추적 포인터의 shadow metadata는 복사·저장·인자·반환을 거쳐도 함께 전달되어야 합니다.

C ABI와 내부 메타데이터 전달 규약은 구별됩니다. C 경계를 넘는 호출에는 명시적 어댑터가 필요합니다. 숫자 주소를 C ABI로 전달했다는 이유만으로 할당 정체성·수명·범위·접근 권한이 보존되었다고 간주해서는 안 됩니다.

## 스택 프레임

프레임 포인터를 유지하며 red zone을 사용하지 않습니다. O0에서는 서로 다른 지역변수의 저장 공간을 재사용하지 않습니다. 호출 프레임 디버그 정보는 native 프레임과 원래 프로그램의 대응을 설명합니다.

이 규칙은 디버깅을 위한 것이며 예외 unwinding 지원을 의미하지 않습니다. [O0와 디버깅](o0-debugging)을 참고하세요.

## 프로파일 제한

첫 O0 native 프로파일에는 다음 기능을 포함하지 않습니다.

- O1 이상 최적화, 벡터화, LTO.
- 공유 메모리와 atomic 연산.
- 예외 unwinding, async 함수, coroutine.
- GC와 언어별 소유권 검사.
- 임의 외부 메모리의 소유권 이전.
- 전체 인라인 어셈블리 제약 지원.
- 동적 링크, TLS, 추가 명령 집합, 추가 오브젝트 형식.

지원 범위를 벗어나는 요청은 다른 기능으로 조용히 대체하지 않고 거부해야 합니다. 이 제한은 native 컴파일 경로에 대한 것이며 독립적으로 제공되는 다른 툴체인 구성 요소를 제거한다는 뜻은 아닙니다.
