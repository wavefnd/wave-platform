---
translation_set_id: whale-ir-reference
path: whale/ir-reference
locale: ko
group: whale
group_order: 1
order: 6
title: Whale IR 참조
summary: 타입, 식별자, 함수의 유효성, 평가 순서와 교환 형식을 설명합니다.
---

## 모듈과 식별자

모듈은 타깃 정보, 전역 정의, 함수로 구성됩니다. 값에는 명시적인 타입이 있습니다. 프런트엔드는 소스 언어의 이름·타입·오버로드·제네릭을 해결한 뒤 typed IR을 생성합니다.

함수와 전역 변수는 서로 다른 내부 이름 공간을 사용합니다. 따라서 함수와 변수가 같은 이름을 가질 수 있습니다. 내부 식별자는 외부 연결 이름인 `link_name`과 구별되며, 외부 이름은 프런트엔드가 명시합니다. Whale은 새로운 이름을 자동 생성해 외부 충돌을 해결하지 않습니다. [심볼과 링크](assembler-linker)를 참고하세요.

각 값 정의는 식별자를 갖습니다. 정의는 중복될 수 없고, 타입 메타데이터는 정의에 명시된 타입과 일치해야 합니다. 이름이 같은 선언이 서로를 가리는 경우에도 이름만으로 정의를 식별하지 않습니다.

## IR 구성과 읽기

다음은 `ir` 크레이트로 함수를 구성하고 검증한 뒤 typed IR을 출력하는 완전한 Rust 예제입니다.

```rust
use ir::{ModuleBuilder, Target, Type};

fn main() {
    let target = Target::lookup("x86_64-whale-linux").unwrap();
    let mut module = ModuleBuilder::new(target.name(), target.data_layout());
    let mut function = module.begin_function("answer", vec![], Type::I32);
    let left = function.const_i32(40);
    let right = function.const_i32(2);
    let answer = function.add(Type::I32, left, right);
    function.ret(Some(answer));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    print!("{}", ir::print_module(&module));
}
```

프린터는 다음 IR을 출력합니다.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "answer": whale () -> i32, linkage internal

  fn @f0 "answer"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 40
    %v1: i32 = const i32 2
    %v2: i32 = add i32 %v0, %v1
    ret i32 %v2
  }

}
```

위 출력을 `answer.wir`로 저장하면 텍스트 파서와 CLI에서 읽고 검증할 수 있습니다. 스칼라 정수 실행은 아래에서 설명합니다.

위 출력을 `answer.wir`로 저장하면 텍스트 파서와 CLI에서 읽고 검증할 수 있습니다. IR 실행은 아직 제공하지 않습니다.

## 타입

| 타입 | 의미 |
| --- | --- |
| `bool` | 논리값 false 또는 true |
| `i1`, `i8`, `i16`, `i32`, `i64`, `i128` | 지정된 비트 폭의 부호 있는 정수 |
| `u1`, `u8`, `u16`, `u32`, `u64`, `u128` | 지정된 비트 폭의 부호 없는 정수 |
| `f16`, `f32`, `f64` | 지정된 비트 폭의 부동소수점 값 |
| `ptr<T>` | T 타입 값을 가리키는 포인터 |
| `fnptr<signature>` | 정확한 매개변수·반환 타입과 호출 규약을 가진 함수 포인터 |
| `array<T, N>` | 같은 타입의 원소 N개 |
| `struct{T, ...}` | 순서가 있는 구조체 필드 |
| `tuple<T, ...>` | 순서가 있는 튜플 원소 |
| `void` | 결과 없음 |

`bool`, `i1`, `u1`은 서로 다른 타입입니다. signed `i1`은 −1과 0, unsigned `u1`은 0과 1을 표현합니다. 정수 1은 암묵적으로 논리 조건이 되지 않습니다. 조건 분기, Select의 조건, `trap_if`는 Bool 피연산자를 요구합니다.

저장 크기는 값의 비트 수만으로 정해지지 않고 [타깃 레이아웃](memory-model)을 따릅니다. 예를 들어 `i1`의 값은 1비트지만 메모리에서는 최소 1바이트를 차지합니다.

## 함수와 호출

함수에는 전체 매개변수·결과 타입, 호출 규약, linkage를 명시합니다. 직접 호출과 간접 호출은 호출 대상의 서명과 일치해야 합니다. void 호출은 결과 ID가 없습니다. nonvoid 호출은 O0에서 결과를 사용하지 않더라도 결과 정의를 유지합니다.

반환은 함수의 결과 타입과 일치해야 합니다. void 반환은 값을 전달하지 않으며, nonvoid 반환은 선언된 결과 타입의 값을 전달합니다.

### 선언, 식별자와 호출

`Module.declarations`는 함수의 `FunctionId`, 이름, 전체 서명, linkage와 외부 연결 이름을 보관합니다. 정의는 이 식별자를 참조하며 매개변수·반환 타입이 선언과 일치해야 합니다. `declare_function`은 동일한 반복 선언을 같은 ID로 해석하고, 충돌하는 선언과 중복 정의는 오류로 처리합니다. 내부 선언에는 같은 모듈의 본문이 필요합니다. 외부 선언은 링크할 때까지 미해결 상태이거나 공개할 본문을 가질 수 있습니다. 내부 함수에는 `link_name`이 없고, 외부 함수는 NUL이 없는 비어 있지 않은 이름을 명시해야 합니다. 서로 다른 함수 선언이 같은 외부 이름을 차지하면 오류입니다. 전역 변수와 함수의 내부 이름 공간은 구분됩니다.

전방 호출·재귀 호출은 모든 선언을 먼저 등록한 뒤 `begin_declared_function`으로 본문을 만듭니다. `begin_function`은 새 내부 Whale 함수를 만드는 편의 API입니다. 검사형 API인 `declare_function`, `begin_declared_function`, `function_addr`, `null_function`, `call`은 `Result`를 반환하며, 거부된 호출은 명령이나 결과 ID를 추가하지 않습니다.

다음 완전한 Rust 예제는 외부 함수를 선언하고 typed 주소를 얻어 직접 호출과 간접 호출을 모두 생성합니다.

```rust
use ir::{Callee, CallingConvention, DataLayout, FunctionSignature, Linkage, ModuleBuilder, Type};

fn main() {
    let mut module = ModuleBuilder::new("x86_64-whale-linux", DataLayout::default_64bit_le());
    let signature = FunctionSignature {
        params: vec![Type::I32], ret: Type::I32,
        convention: CallingConvention::SysV64, variadic: false,
    };
    let identity = module.declare_function(
        "identity", signature, Linkage::External, Some("identity_i32".into()),
    ).unwrap();
    let mut function = module.begin_function("answer", vec![], Type::I32);
    let input = function.const_i32(42);
    let callback = function.function_addr(identity).unwrap();
    // The direct call's result remains defined even though it is unused.
    function.call(Callee::Direct(identity), vec![input]).unwrap();
    let result = function.call(Callee::Indirect(callback), vec![input]).unwrap().unwrap();
    function.ret(Some(result));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    print!("{}", ir::print_module(&module));
}
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "identity": sysv64 (i32) -> i32, linkage external, link_name "identity_i32"
  declare @f1 "answer": whale () -> i32, linkage internal

  fn @f1 "answer"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 42
    %v1: fnptr<sysv64 (i32) -> i32> = function_addr @f0
    %v2: i32 = call sysv64 i32 @f0(%v0)
    %v3: i32 = call sysv64 i32 indirect %v1(%v0)
    ret i32 %v3
  }

}
```

`Callee::Direct(FunctionId)`는 선언 표를 참조하고, `Callee::Indirect(ValueId)`는 `Type::FnPtr(FunctionSignature)` 값을 요구합니다. 서명에는 모든 매개변수·반환 타입과 `CallingConvention::{Whale, SysV64}`가 포함됩니다. 복사, 저장, 매개변수, 반환, phi와 select를 거쳐도 서명을 보존합니다. 데이터 포인터와 정수는 호출할 수 없습니다. 함수 포인터 타입이 관련된 cast는 거부하며 타입 표기를 바꿔 호출 가능한 서명을 변경할 수 없습니다. 이 타깃에서 함수 포인터의 주소 저장 크기는 64비트지만, 이것만으로 실행 시 shadow metadata가 구현되지는 않습니다.

인자 수, 정확한 인자·결과 타입, 결과 ID 유무와 호출 규약이 일치해야 하며 암묵적 변환은 없습니다. 간접 호출 대상도 인자와 마찬가지로 호출을 지배해야 합니다. `variadic: true`, void 매개변수, SysV64의 복합 매개변수·반환 서명은 거부합니다. Whale 복합 서명은 IR에서 표현할 수 있지만 두 규약 모두 native ABI 분류와 기계 호출 생성은 아직 제공하지 않습니다.

`null_function(signature)`는 서명을 가진 null 함수 포인터입니다. 이를 호출하는 IR은 타입상 유효하며 실행 시 callee 진입 전에 반드시 trap해야 합니다. null이 아니어도 잘못되었거나 수명이 끝났거나 검사한 서명과 호환되지 않는 대상은 trap해야 합니다. 이 실행 검사와 외부 콜백 수명 관리는 interpreter·native 실행 계층의 후속 기능입니다. 검증을 통과했다고 임의의 외부 주소가 안전해지는 것은 아닙니다.

### AST 호출 형태

다음은 AST 형식 2 프로그램 안에 넣는 표현식 조각입니다.

```json
{"Call":{"callee":{"Direct":"increment"},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

```json
{"Call":{"callee":{"Indirect":{"FunctionRef":"increment"}},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

`Direct`와 `FunctionRef`는 같은 이름의 변수가 있어도 함수 이름 공간에서 찾습니다. `Indirect`는 대상 표현식을 먼저 평가하고 인자를 왼쪽부터 평가합니다. void 호출은 `ExprStmt`에 사용할 수 있지만 변수 초기값·인자·피연산자·반환값으로 사용할 수 없습니다. 호출과 함수 참조는 컴파일 타임 수치 상수식이 아닙니다. `NullFunction`은 `params`, `ret`, `convention`, `variadic` 필드를 가진 서명 객체를 받습니다.

[완전한 JSON 예제](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/ast-v2-calls.json)는 콜백을 저장한 뒤 간접 호출하고 외부 함수를 호출합니다. 다음 명령으로 lowering합니다.

```sh
cargo run --locked --features socket-cli -- ir lower ir/tests/fixtures/ast-v2-calls.json
```

[예상 IR](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/calls-v3.wir)은 lowering 테스트에서 비교합니다. 함수 식별자와 연결 이름은 IR 경계에서 표현하며, native 오브젝트 생성·링크까지 보존하는 작업은 별도입니다.

## 블록과 값의 사용 가능성

각 블록에는 고유한 식별자와 정확히 하나의 terminator가 있습니다. 분기 대상은 같은 함수에 속해야 합니다. 진입 블록은 반드시 존재하며 선행 간선과 phi를 가질 수 없습니다. 루프를 만들 때는 진입 블록에서 별도의 루프 헤더로 분기합니다.

실행 가능한 경로에서 값의 정의는 일반적인 사용 지점을 지배해야 합니다. 즉, 진입점에서 사용 지점으로 가는 모든 경로가 그 정의를 지나야 합니다. 같은 블록에서는 정의가 사용보다 앞에 있어야 합니다. 블록의 저장 순서는 지배 관계를 결정하지 않습니다.

진입점에서 left 또는 right로 분기한 뒤 join에서 합류한다고 가정합니다. left에서만 정의한 값은 join에서 일반 값으로 사용할 수 없습니다. right를 지나는 경로에는 정의가 없기 때문입니다. 각 선행 블록의 값을 입력으로 받는 phi로 합쳐야 합니다.

도달 불가능한 블록도 모듈에 보존됩니다. 검증기는 해당 블록의 식별자·타입·피연산자·분기 구조를 계속 검사합니다. 도달 불가능한 블록의 정의가 도달 가능한 경로의 일반 사용에 값을 공급할 수는 없습니다.

## Phi 명령

phi는 블록 안의 모든 일반 명령보다 앞에 놓입니다. 서로 다른 선행 블록마다 정확히 하나의 입력이 필요합니다. 입력 값은 phi와 같은 타입이어야 하며, 해당 선행 블록의 끝에서 사용할 수 있어야 합니다.

하나의 선행 블록에서 간선이 여러 개 들어와도 입력은 하나입니다. 루프 phi는 모듈 저장 순서상 나중에 나오는 블록의 값이라도 반복 간선에서 계산되는 값이면 참조할 수 있습니다. 누락·중복·무관한 선행 블록과 잘못된 타입의 입력은 검증 오류입니다.

## 평가와 선택

Whale AST는 호출 대상과 하위 표현식을 명시된 왼쪽부터 평가합니다. 프런트엔드는 단락 평가를 제어 흐름 분기로 표현합니다.

Select는 이미 계산된 값 중 하나를 선택합니다. 어느 쪽 입력의 계산도 생략하지 않습니다. 예를 들어 안전한 값을 선택하더라도 다른 입력을 계산하면서 발생한 trap을 피할 수 없습니다. 특정 경로에서만 실행해야 하는 계산은 조건부 블록 안에 배치해야 합니다.

## 검증과 trap

잘못된 IR은 검증 오류입니다. legacy `undef`는 검증기에서 거부하며 기존 AST를 다시 lowering해야 합니다. 초기화 없는 선언은 format 4의 `uninit`과 검사된 실제 읽기로 표현합니다. 자동 0 초기화나 임의 값으로 대체하지 않습니다. 실행 조건 위반은 IR 위치를 포함한 정의된 trap입니다.

현재 `InterpreterTrap`은 이유·실행 단계 수·`ExecutionSite`를 제공합니다. 위치는 함수 ID, 블록 ID, 0부터 시작하는 명령 인덱스, 선택적인 결과 값 ID입니다. terminator 인덱스는 마지막 명령 다음입니다. CLI 진단에는 입력 파일도 표시합니다. typed IR에는 아직 소스 span이 없으므로 이 위치는 소스 줄 번호가 아닌 IR 위치입니다. trap은 오류를 반환하고 이후 실행을 중단하며 라이브러리는 호스트 프로세스를 종료하지 않습니다.

이 보장은 검증된 IR과 추적 메모리에 적용됩니다. 외부 C·원시 주소·인라인 어셈블리는 별도 계약을 가지며, 그 경계 밖의 위반까지 항상 검출하지는 않습니다. [메모리 모델](memory-model)을 참고하세요.

## 교환 형식과 텍스트 표현

AST와 typed IR은 각각의 format version과 공통 semantics version을 사용합니다. 읽는 쪽은 무버전·알 수 없는 버전·필드·기능·중복 JSON 키를 거부해야 합니다. 생성자는 지원하지 않는 속성이 조용히 무시될 것이라고 가정해서는 안 됩니다.

정수는 비트 폭·signedness·문자열 숫자로 전달합니다. 부동소수점 상수는 폭과 정확한 비트열로 전달합니다. 텍스트 IR의 round-trip은 이름·ID·타입·상수·순서·속성·메타데이터를 보존해야 합니다. 공백과 주석 배치는 보존 대상이 아닙니다.

아래 AST JSON 계약과 typed IR 형식 4의 읽기·검증·왕복 출력을 사용할 수 있습니다.

### 출력 식별자와 인용된 이름

typed IR format 4은 함수 식별자를 `@fN`, 전역을 `@gN`, 값을 `%vN`, 블록을 `%bN`으로 출력합니다. 함수·전역 ID는 모듈에, 값·블록 ID는 해당 함수에 속합니다. 번호에 빈 구간이 있어도 입력 ID를 보존합니다. 인용된 이름은 설명용 표기이며 참조 해석에 사용하지 않습니다. 함수는 블록 저장 순서와 독립적으로 `entry %bN`을 명시합니다.

다음 전체 모듈은 Rust IR API에서 검증하고 출력했습니다. 두 분기 블록의 이름이 모두 `"branch"`이지만, ID가 정의와 phi 입력을 구분합니다.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "choose": whale (bool) -> i32, linkage internal

  fn @f0 "choose"(%v0 "condition": bool) -> i32, entry %b0 {
  %b0 "entry":
    cbr bool %v0, label %b1, label %b2
  %b1 "branch":
    %v1: i32 = const i32 1
    br label %b3
  %b2 "branch":
    %v2: i32 = const i32 2
    br label %b3
  %b3 "join":
    %v99: i32 = phi i32 [ %v1, %b1 ], [ %v2, %b2 ]
    ret i32 %v99
  }

}
```

`%v0`은 매개변수 목록에서 정의됩니다. `%b1`과 `%b2`는 이름이 같아도 별개이며, phi는 ID로 각 선행 블록을 지정합니다. 분기와 switch 목적지도 같은 블록 ID 문법을 사용합니다. 프린터는 명시적으로 지정한 `%v99`의 번호를 바꾸지 않습니다.

모든 이름·문자열 필드는 큰따옴표를 사용합니다. 타깃, 함수·전역·매개변수·블록 이름, 외부 연결 이름, 상수 선언 이름, trap 이유에 같은 규칙을 적용합니다. 출력 가능한 Unicode 문자는 그대로 남깁니다. escape는 `\"`, `\\`, `\n`, `\r`, `\t`, `\0`이며, 그 밖의 제어 문자와 U+2028/U+2029에는 소문자 16진수의 `\u{hex}`를 사용합니다. 예를 들어 개행·탭·따옴표·역슬래시·한글이 포함된 이름도 한 레코드로 출력됩니다.

```text
"line\ncolumn\tquote\"slash\\한글"
```

파서는 typed IR format 3과 4, semantics version 1을 읽고 항상 format 4를 출력합니다. `uninit`은 format 4에서만 사용할 수 있습니다. format 3의 유효한 기존 명령은 유지되지만 legacy `undef`는 어느 형식에서도 검증 오류이며 AST에서 재생성해야 합니다. format 2 텍스트는 명시적 ID·인용 이름·진입 참조로 직접 이관해야 합니다. AST JSON은 별도 format 2입니다.

### 텍스트 IR 읽기와 검증

`ir::parse_module`은 typed IR을 읽은 뒤 검증하여 `Module`을 반환합니다. 현재 프린터의 스칼라·제어 흐름·메모리·직접/간접 호출·상수식 구문을 받습니다. ID·이름·타입·정확한 정수와 float 비트·상수식 트리와 평가 결과·블록 순서·진입 블록·정렬·서명·link_name을 보존합니다. 공백과 `//` 줄 주석은 표준 출력으로 정리합니다. 위의 전체 IR을 `answer.wir`로 저장하고 기본 빌드에서 실행하세요.

```sh
cargo run --locked -- ir verify answer.wir
cargo run --locked -- ir print answer.wir -o canonical.wir
```

알 수 없는 버전·필드·명령·이스케이프, 범위를 벗어난 리터럴, 중복 ID, 상충하는 타입 표기, 뒤따르는 입력은 거부합니다. `ParseError`는 바이트 위치와 1부터 시작하는 행·Unicode scalar 기준 열을 제공합니다. 검증 오류는 가능한 경우 해당 함수나 전역 선언에 연결됩니다. `print`도 검증을 수행하며 실패하면 기존 출력 파일을 보존합니다. 새 구문이나 의미는 해당 format/semantics version을 변경해야 하며 알 수 없는 버전은 오류입니다.

### 검증 자원 한도

`IrLimits` 기본값은 입력 8 MiB, 토큰 1,000,000개, 순회 노드 1,000,000개, 타입과 상수식 깊이 각각 128입니다. 루트 깊이는 0이며, 자식마다 1 증가합니다. 깊이는 낮추거나 `MAX_IR_NESTING`인 256까지 높일 수 있습니다. 깊이·작업량 초과는 `LimitError`, `VerifyError::ResourceLimit`, `ConstEvalError::ResourceLimit` 또는 `CallError::ResourceLimit`으로 반환합니다. 노드 한도는 타입 표기와 상수식 노드를 포함한 각 입력 경계의 순회량이며 실행 시간 한도가 아닙니다.

```rust
use ir::{parse_module_with_limits, print_module, IrLimits};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let source = std::fs::read_to_string("answer.wir")?;
    let limits = IrLimits {
        max_type_depth: 32,
        max_const_depth: 32,
        max_nodes: 10_000,
        ..IrLimits::default()
    };
    let module = parse_module_with_limits(&source, limits)?;
    let canonical = print_module(&module);
    let reread = parse_module_with_limits(&canonical, limits)?;
    assert_eq!(print_module(&reread), canonical);
    let invalid = source.replacen("format_version 4", "format_version 99", 1);
    assert!(parse_module_with_limits(&invalid, limits).is_err());
    Ok(())
}
```

`verify_module_with_limits`, `ConstExpr::evaluate_with_limits`, `validate_signature_with_limits`, `ModuleBuilder::declare_function_with_limits`에도 한도를 전달할 수 있습니다. 타입은 재귀 clone·비교·진단 전에 반복 순회로 검사하며 상수식은 작업 스택으로 평가합니다. 차용한 Rust 트리의 소유권과 Drop은 호출자에게 있습니다. 임의로 구성한 미검증 트리의 clone/Drop까지 보호하지 않으며, checked 선언 API가 소유한 거부된 서명은 반복적으로 해제합니다. 잘못된 IR은 검증 오류입니다. legacy `undef`는 검증기에서 거부하며 기존 AST를 다시 lowering해야 합니다. 초기화 없는 선언은 format 4의 `uninit`과 검사된 실제 읽기로 표현합니다. 자동 0 초기화나 임의 값으로 대체하지 않습니다. 실행 조건 위반은 IR 위치를 포함한 정의된 trap입니다.

### 버전이 명시된 AST JSON

다음을 `program.json`으로 저장합니다. Envelope의 네 필드는 모두 필수입니다. `program`의 `declarations`, `globals`, `functions` 배열도 필수이며 빈 배열을 허용합니다. 함수의 이름·매개변수·반환 타입·본문과 `convention`, `linkage`는 필수입니다. 내부 함수의 `link_name`은 생략하거나 null로 지정하며, 외부 함수는 NUL이 없는 비어 있지 않은 문자열로 명시합니다. 각 enum은 unit 이름 또는 variant 키 하나를 가진 객체로 표현합니다. Unit variant는 `{"Void":null}`처럼 null 값을 가진 객체도 허용하며, encoder는 unit 이름 `"Void"`로 출력합니다. `VarDecl.init`은 생략하거나 null로 지정할 수 있으며 다른 필수 필드는 생략할 수 없습니다.

```json
{
  "format_version": 2,
  "semantics_version": 1,
  "features": [],
  "program": {
    "globals": [],
    "functions": [
      {
        "name": "answer",
        "parameters": [],
        "return_type": {
          "Int": {
            "bits": 128,
            "signed": false
          }
        },
        "body": [
          {
            "Return": {
              "Lit": {
                "Int": {
                  "bits": 128,
                  "signed": false,
                  "value": "340282366920938463463374607431768211455"
                }
              }
            }
          }
        ],
        "convention": "Whale",
        "linkage": "Internal",
        "link_name": null
      }
    ],
    "declarations": []
  }
}
```

```sh
cargo run --locked --features socket-cli -- ir lower program.json
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "answer": whale () -> u128, linkage internal

  fn @f0 "answer"() -> u128, entry %b0 {
  %b0 "entry":
    %v0: u128 = const u128 340282366920938463463374607431768211455
    ret u128 %v0
  }

}
```

정수 `value`는 10진 문자열입니다. Signed 정수의 선택적인 마이너스 뒤에 숫자를 쓰며 공백·플러스·지수·구분자는 허용하지 않습니다. 허용 범위는 선언한 폭과 signedness로 결정합니다. 위의 `u128::MAX`는 JSON과 lowering을 거치면서 그대로 보존됩니다. 음수 unsigned 값이나 범위를 벗어난 값은 wrap하지 않고 오류입니다. Float 값에는 [수치 연산](numeric-operations)에서 설명하는 정확한 폭의 16진 비트 문자열을 사용합니다.

이 AST의 `format_version`은 2이고 `semantics_version`은 1입니다. `features`는 빈 배열이어야 합니다. 알 수 없는 필드·버전·기능, 중복된 원본 JSON 키(escape를 풀면 같은 키인 경우 포함), 뒤따르는 추가 값은 `--no-verify`에서도 오류입니다. 라이브러리 진입점은 `ir::lower_ast::interchange::decode`이며 `encode`는 envelope를 출력합니다. `decode`의 기본 원본 크기 한도는 8 MiB이고 `decode_with_limit`으로 한도를 지정합니다. JSON 중첩에도 한도가 있습니다. 먼저 일반 map으로 읽으면 중복 키가 사라질 수 있으므로 원본 decoder를 사용합니다.

[전체 JSON Schema](https://github.com/wavefnd/Whale/blob/master/ir/schema/ast-v2.schema.json)는 형태·필수 필드·variant를 정의합니다. 범위·타입 검사와 중복 키 검사가 추가로 적용됩니다. 스칼라 lowering은 리터럴, 변수·상수, add/sub/mul, 비교, 대입, if/while, return과 break/continue를 지원합니다. 함수 참조·직접 호출·간접 호출을 지원하며, 복합 값 표현식은 미지원입니다. `Opaque`는 스키마에서 표현할 수 있지만 lowering에서 거부합니다.

기존의 bare Program에는 envelope를 추가하고 JSON 숫자 리터럴을 정수의 10진 문자열 또는 float 비트 문자열로 바꿔야 합니다. 무버전 입력은 거부합니다. 형식 1 입력은 형식 2로 옮기면서 `program.declarations` 배열(사용하지 않으면 빈 배열)과 함수 정의의 명시적인 `convention`·`linkage`를 추가해야 합니다. AST와 typed IR은 독립적으로 버전을 관리합니다. 현재 AST 형식은 2, typed IR 형식은 3이며 의미 버전은 1입니다.

### 거부되는 입력과 CLI 복구

다음 완전한 입력을 `invalid.json`으로 저장합니다.

```json
{
  "format_version": 99,
  "semantics_version": 1,
  "features": [],
  "program": {
    "globals": [],
    "functions": [],
    "declarations": []
  }
}
```

```sh
cargo run --locked --features socket-cli -- ir lower invalid.json -o rejected.wir
```

```text
Failed to parse socket JSON: unsupported AST format_version 99; expected 2
```

명령은 0이 아닌 상태로 종료하며 새 출력을 만들거나 기존 파일을 덮어쓰지 않습니다. 타입 불일치도 출력 게시 전에 실패합니다. `socket-cli` 없이 빌드한 바이너리는 상태 2로 종료하고 `--features socket-cli`가 포함된 복구 명령을 출력합니다.


## 스칼라 정수 인터프리터

기본 인터프리터는 정수·Bool, 제어 흐름, 스택 할당, 데이터 포인터 저장·읽기, typed GEP, memcpy·memset을 실행합니다. checked 쌍의 값 읽기는 패딩을 제외합니다. 주소는 호스트 메모리를 가리키지 않는 합성 64비트 주소입니다. 함수 인자·반환은 정수·Bool 또는 void로 제한되며 float, 호출, 함수 포인터, 일반 aggregate 값, 전역 주소, native 실행은 미지원입니다. 스택 할당은 함수 반환까지 살아 있습니다. 블록별 수명 종료, 호출·반환의 포인터 전달, 외부 메모리 어댑터와 native shadow metadata는 후속 구현이 필요합니다.

`InterpreterOptions::max_steps` 기본값은 1,000,000입니다. phi를 포함한 실행 명령 하나와 terminator 하나가 각각 한 단계를 소비합니다. 한도가 0이면 첫 연산 전에 중단하며 무한 분기 루프는 `InterpreterError::StepLimit`을 반환합니다. `ir_limits`는 검증량을 별도로 제한합니다. 미사용 정수 연산도 실행하고 trap할 수 있습니다. checked overflow는 Bool 결과이며, 명시적인 trap_if가 있어야 trap으로 처리합니다.

다음 전체 모듈을 `swap-loop.wir`로 저장하세요. 종료 블록이 먼저 저장되어 있어도 진입 블록은 명시되어 있습니다. 블록에 진입하면 모든 phi 입력을 이전 블록의 값에서 읽은 뒤 phi 결과를 한꺼번에 기록합니다. 세 번 반복하면 11과 22를 세 번 교환하여 22를 반환합니다.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f7 "swap_loop": whale (u32) -> i32, linkage internal

  fn @f7 "swap_loop"(%v0 "iterations": u32) -> i32, entry %b11 {
  %b90 "exit":
    ret i32 %v5
  %b11 "entry":
    %v1: i32 = const i32 11
    %v2: i32 = const i32 22
    %v3: u32 = const u32 0
    %v4: u32 = const u32 1
    br label %b20
  %b20 "loop":
    %v5: i32 = phi i32 [ %v1, %b11 ], [ %v6, %b30 ]
    %v6: i32 = phi i32 [ %v2, %b11 ], [ %v5, %b30 ]
    %v7: u32 = phi u32 [ %v3, %b11 ], [ %v9, %b30 ]
    %v8: bool = icmp ult u32 %v7, %v0
    cbr bool %v8, label %b30, label %b90
  %b30 "next":
    %v9: u32 = add u32 %v7, %v4
    br label %b20
  }

}
```

```shell
whale ir run swap-loop.wir --function @f7 --arg 3 --max-steps 100
```

```text
i32 22
```

Rust API는 값과 단계 수 또는 구조적인 검증·미지원 연산·인자·단계 한도·trap 오류를 반환합니다. 다음 전체 프로그램은 같은 `swap-loop.wir` 파일을 읽습니다.

```rust
use ir::{interpret_with_options, parse_module, ConstValue, FunctionId,
         InterpreterError, InterpreterOptions};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let source = std::fs::read_to_string("swap-loop.wir")?;
    let module = parse_module(&source)?;
    let options = InterpreterOptions {
        max_steps: 100,
        ..InterpreterOptions::default()
    };
    let result = interpret_with_options(
        &module, FunctionId(7), &[ConstValue::U(3)], options,
    )?;
    assert_eq!(result.value, Some(ConstValue::I(22)));
    assert_eq!(result.steps, 32);
    let stopped = interpret_with_options(
        &module, FunctionId(7), &[ConstValue::U(3)],
        InterpreterOptions { max_steps: 0, ..options },
    );
    assert!(matches!(stopped, Err(InterpreterError::StepLimit { .. })));
    println!("i32 22");
    Ok(())
}
```
