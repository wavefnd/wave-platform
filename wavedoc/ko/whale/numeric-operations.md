---
translation_set_id: whale-numeric-operations
path: whale/numeric-operations
locale: ko
group: whale
group_order: 1
order: 7
title: 수치 연산
summary: 정수 wrap, checked 연산, 시프트, 형변환 오류와 부동소수점 결과를 설명합니다.
---

## 정수 표현

N비트 정수는 N개의 값 비트를 갖습니다. 부호 없는 정수의 범위는 0부터 2^N − 1까지이고, 부호 있는 정수의 범위는 −2^(N−1)부터 2^(N−1) − 1까지입니다. 연산의 signedness가 비트열의 해석을 결정합니다.

아래 표는 연산 결과를 설명합니다. IR 코드 예제는 현재 프린터의 표현을 사용합니다.

## 덧셈·뺄셈·곱셈

기본 정수 add·sub·mul은 결과의 하위 N비트를 유지합니다. overflow는 trap을 발생시키지 않습니다. checked 연산은 같은 wrap 결과와 함께, 수학적 결과가 해당 signed 또는 unsigned 범위를 벗어났는지를 나타내는 Bool을 반환합니다.

| 연산 | Wrap 결과 | Checked overflow |
| --- | --- | --- |
| u8: 255 + 1 | 0 | true |
| i8: 127 + 1 | −128 | true |
| u8: 0 − 1 | 255 | true |
| i8: 12 × 3 | 36 | false |

overflow에서 실행을 중단하는 언어의 프런트엔드는 checked 연산의 overflow 결과에 명시적인 `trap_if`를 사용해야 합니다. 기본 연산이 소스 언어의 overflow 정책을 암묵적으로 적용하지는 않습니다.

### Wrap과 명시적 검사를 표현한 IR

다음 전체 형식 3 모듈을 `integer-operations.wir`로 저장하세요. 텍스트 reader가 받는 입력이며 스칼라 인터프리터에서 실행할 수 있습니다.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "add_u8": whale () -> u8, linkage internal
  declare @f1 "require_no_overflow": whale () -> u8, linkage internal

  declare @f2 "minimum_division": whale () -> i8, linkage internal
  declare @f3 "minimum_remainder": whale () -> i8, linkage internal
  declare @f4 "negative_shift": whale () -> i8, linkage internal

  fn @f0 "add_u8"() -> u8, entry %b0 {
  %b0 "entry":
    %v0: u8 = const u8 255
    %v1: u8 = const u8 1
    %v2: u8 = add u8 %v0, %v1
    ret u8 %v2
  }

  fn @f1 "require_no_overflow"() -> u8, entry %b1 {
  %b1 "entry":
    %v3: u8 = const u8 255
    %v4: u8 = const u8 1
    %v5: tuple<u8, bool> = uadd_chk u8 %v3, %v4
    %v6: u8 = extract %v5, 0
    %v7: bool = extract %v5, 1
    trap_if bool %v7, reason="integer overflow"
    ret u8 %v6
  }

  fn @f2 "minimum_division"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -128
    %v1: i8 = const i8 -1
    %v2: i8 = sdiv i8 %v0, %v1
    ret i8 %v2
  }

  fn @f3 "minimum_remainder"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -128
    %v1: i8 = const i8 -1
    %v2: i8 = srem i8 %v0, %v1
    ret i8 %v2
  }

  fn @f4 "negative_shift"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -1
    %v1: i8 = const i8 -1
    %v2: i8 = shl i8 %v0, %v1
    ret i8 %v2
  }

}
```

`add_u8`은 256의 낮은 8비트인 0을 반환합니다. `minimum_division`은 −128, `minimum_remainder`는 0을 반환합니다. `negative_shift`는 −1을 unsigned count 255로 해석한 뒤 255 mod 8 = 7을 적용합니다. 명시적인 함수 ID로 실행하세요.

```shell
whale ir run integer-operations.wir --function @f0
whale ir run integer-operations.wir --function @f2
whale ir run integer-operations.wir --function @f3
whale ir run integer-operations.wir --function @f4
```

```text
u8 0
i8 -128
i8 0
i8 -128
```

`require_no_overflow`는 인덱스 0에서 감긴 결과, 인덱스 1에서 Bool overflow 플래그를 추출합니다. 명시적인 trap_if가 반환 전에 중단합니다. 다음 명령은 상태 1로 종료하고 stderr에 진단을 출력합니다.

```shell
whale ir run integer-operations.wir --function @f1
```

```text
Error: integer-operations.wir: trap at @f1 %b1 instruction 5: "integer overflow" (after 6 steps)
```

## 나눗셈과 나머지

정수의 0 나눗셈과 0에 대한 나머지는 trap입니다. signed 최솟값을 −1로 나누면 최솟값으로 wrap합니다. 이 경우의 나머지는 0입니다.

| 연산 | 결과 |
| --- | --- |
| i8: −128 / −1 | −128 |
| i8: −128 % −1 | 0 |
| 정수 0 나눗셈 | trap |
| 정수 0에 대한 나머지 | trap |

## 시프트

N비트 값의 시프트에서는 count의 비트열을 unsigned로 해석한 뒤 N으로 나눈 나머지를 사용합니다. count가 0부터 N−1까지의 범위 밖이라는 이유만으로 trap을 발생시키지 않습니다.

8비트 값에서 count 0·8·16은 모두 0비트 시프트입니다. 8비트 count의 비트열 `11111111`은 7비트 시프트가 됩니다. 이 비트열이 signed −1을 나타내더라도 같습니다. unsigned 해석은 나머지 계산보다 먼저 이루어집니다.

음수나 과도한 count를 거부하는 소스 언어는 시프트 앞에 명시적인 검사를 넣어야 합니다.

## 형변환

| 변환 | 의미 |
| --- | --- |
| Zero extension | 상위 비트를 0으로 채워 폭을 늘림 |
| Sign extension | 부호 비트를 복제해 폭을 늘림 |
| 비트 절단 | 목적지 폭에 해당하는 하위 비트만 유지 |
| 비트 재해석 | 같은 비트열을 다른 타입으로 해석 |
| 손실 없는 수치 변환 | 수치를 보존하며 표현할 수 없으면 trap |

예를 들어 8비트 `11111111`을 16비트로 zero extension하면 `0000000011111111`, sign extension하면 `1111111111111111`입니다. 입력 비트가 같아도 서로 다른 연산입니다.

float→int는 0 방향으로 절삭한 뒤 정수 범위를 검사합니다. NaN과 무한대는 trap입니다. i8로 변환할 때 127.9는 127이 되고, 128.0은 trap입니다. Bool은 정수 0 또는 1로 변환합니다. signed i1은 1을 표현할 수 없으므로 이 변환의 목적지로 사용할 수 없습니다.

주소를 정수로 변환해도 그 정수로 포인터 접근 권한을 복구할 수 있는 것은 아닙니다. [포인터 유효성](memory-model)을 참고하세요.

### cast와 checked 연산의 검증 형식

검증은 실제 피연산자 타입과 `src_ty`가 같은지 확인한 뒤, opcode가 허용하는 원본·목적지 종류와 폭을 검사합니다. 결과 타입 표기도 정의와 일치해야 합니다. 암묵적 변환은 없습니다.

| Opcode | 허용 타입 |
| --- | --- |
| `zext`, `sext` | Bool이 아닌 정수; 목적지 폭이 반드시 더 큼 |
| Bool에서 `zext` | 정수 0/1 변환; signed `i1`을 제외한 모든 정수 목적지, `u1` 포함 |
| `trunc` | Bool이 아닌 정수; 목적지 폭이 반드시 더 작음 |
| `fext`, `ftrunc` | float; 목적지 폭이 반드시 더 큼 / 더 작음 |
| `itof_s`, `itof_u` | signed / unsigned 정수에서 float |
| `ftoi_s`, `ftoi_u` | float에서 signed / unsigned 정수 |
| `bitcast` | 폭이 같은 정수·float 스칼라, 또는 데이터 포인터에서 데이터 포인터 |
| `ptrtoint`, `inttoptr` | 데이터 포인터에서 정수 / 정수에서 데이터 포인터 |

정수 확장·절단은 비트 연산이므로 정수 피연산자의 signedness가 달라도 됩니다. Bool은 별도 논리 타입이며 sign extension, 절단, `i1`/`u1`로의 bitcast를 허용하지 않습니다. aggregate와 함수 포인터 cast는 거부합니다. 포인터·정수의 종류 검증은 유효한 할당이나 접근 권한을 복구하지 않습니다. 실행 시 변환 검사와 기계어 lowering은 별도 미완료 범위입니다.

다음 전체 Rust 프로그램은 유효한 변환 두 개를 출력하고, `zext`를 `fext`로 바꾸면 거부되는지 확인합니다.

```rust
use ir::*;

fn main() {
    let mut builder = ModuleBuilder::new("x86_64-whale-linux", DataLayout::default_64bit_le());
    let mut function = builder.begin_function("widen", vec![("byte".into(), Type::U8)], Type::Void);
    function.ret(None);
    function.finish();
    let mut module = builder.finish();
    let function = &mut module.functions[0];
    function.value_types.extend([(ValueId(1), Type::U32), (ValueId(2), Type::I32)]);
    function.blocks[0].instructions.extend([
        Instruction::Cast {
            dst: ValueId(1), op: CastOp::ZExt,
            src_ty: Type::U8, src: ValueId(0), dst_ty: Type::U32,
        },
        Instruction::Cast {
            dst: ValueId(2), op: CastOp::Bitcast,
            src_ty: Type::U32, src: ValueId(1), dst_ty: Type::I32,
        },
    ]);
    verify_module(&module).unwrap();
    print!("{}", print_module(&module));
    // An integer source cannot be annotated as a floating widening operation.
    if let Instruction::Cast { op, .. } = &mut module.functions[0].blocks[0].instructions[0] {
        *op = CastOp::FExt;
    }
    assert!(matches!(verify_module(&module), Err(VerifyError::InvalidCast { .. })));
}
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "widen": whale (u8) -> void, linkage internal

  fn @f0 "widen"(%v0 "byte": u8) -> void, entry %b0 {
  %b0 "entry":
    %v1: u32 = zext u8 %v0 to u32
    %v2: i32 = bitcast u32 %v1 to i32
    ret void
  }

}
```

다음 명령 조각은 의도적으로 잘못된 IR입니다.

```text
%v1: i32 = ftoi_s f64 %v0 to i32   // Invalid when %v0 is actually Bool.
%v2: i16 = zext i32 %v3 to i16     // zext cannot narrow.
%v4: i1 = zext bool %v5 to i1      // Signed i1 cannot represent true as 1.
```

실제 원본 타입의 불일치는 `OperandTypeMismatch`, 허용되지 않는 타입 쌍은 opcode와 두 타입을 담은 `InvalidCast`입니다. 검증 오류는 변환을 삽입하거나 모듈을 변경하지 않습니다.

`sadd_chk`, `ssub_chk`, `smul_chk`는 signed 정수, `uadd_chk`, `usub_chk`, `umul_chk`는 unsigned 정수 피연산자를 요구합니다. 양쪽 피연산자는 연산 타입 `T`와 정확히 같아야 하며 결과는 `tuple<T, bool>`이어야 합니다. 추출은 tuple, 존재하는 필드 인덱스, 해당 필드의 정확한 타입을 요구합니다. 특히 overflow 필드를 `i1`로 추출하면 거부합니다. 위의 wrap·overflow 결과는 실행 계약이며, 이 테스트는 구조 검증을 확인합니다. 구현된 실행 백엔드의 결과를 뜻하지 않습니다.

## 부동소수점 연산

부동소수점 값은 정확한 f16·f32·f64 비트열을 갖습니다. 연산은 선언된 폭에서 가장 가까운 값으로 반올림하고, 정확히 중간인 경우 유효숫자의 최하위 비트가 짝수인 값을 선택하는 ties-to-even을 사용합니다.

기본 연산은 fast-math, 암묵적 FMA, 작은 값의 강제 0 처리를 허용하지 않습니다. 곱셈 다음 덧셈은 각각의 반올림 단계를 유지하며 백엔드가 암묵적으로 하나의 연산으로 합쳐서는 안 됩니다.

수치 연산의 결과는 NaN이나 무한대일 수 있습니다. 수치 연산과 폭 변경에서 나오는 NaN은 폭별로 하나의 고정된 양의 quiet NaN으로 정규화합니다. 저장과 복사는 원래 NaN 비트를 보존합니다. 따라서 NaN payload를 산술 연산 없이 메모리로 전달하는 경우와 계산하는 경우의 동작이 다릅니다.

부동소수점 상태 플래그는 노출하지 않습니다. 기본 부동소수점 산술이 NaN·무한대 결과를 허용하더라도 float→int 변환에는 위의 trap 규칙을 적용합니다.


### 정확한 상수 저장

`FloatBits` variant 또는 정확한 폭의 16진 비트 문자열을 사용합니다. 저장 값의 동등성은 음수 0과 NaN payload를 포함한 비트열로 비교합니다. 검증기는 IR 타입과 payload의 폭이 다르면 거부합니다.

```rust
use ir::{FloatBits, ModuleBuilder, Target, Type};
fn main() {
    let bits = FloatBits::parse(32, "0xffc01234").unwrap();
    assert_eq!(bits, FloatBits::F32(0xffc01234));
    let target = Target::X86_64WhaleLinux;
    let mut module = ModuleBuilder::new(target.name(), target.data_layout());
    let mut function = module.begin_function("payload", vec![], Type::F32);
    let value = function.const_float_bits(Type::F32, bits);
    function.ret(Some(value));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    assert!(ir::print_module(&module).contains("const f32 0xffc01234"));
    println!("{}", bits);
}
```

```text
0xffc01234
```

f16/f32/f64 문자열은 `0x` 뒤에 정확히 4/8/16개의 16진 숫자를 사용합니다. `0x80000000`은 f32 음수 0이고 `0x7f800000`은 양의 무한대입니다. `const_float`은 host f64에서 수치 변환하는 편의 API이며 원본 비트 보존에는 `const_float_bits`를 사용합니다. 정확한 저장 표현이 float 연산 백엔드의 완성을 의미하지는 않습니다. 컴파일 시점 산술은 여전히 host f64 중간값을 사용하므로 선언된 폭의 전체 반올림 계약은 미완성입니다.
