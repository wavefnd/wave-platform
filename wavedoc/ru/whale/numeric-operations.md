---
translation_set_id: whale-numeric-operations
path: whale/numeric-operations
locale: ru
group: whale
group_order: 1
order: 7
title: Числовые операции
summary: Описывает целочисленные операции wrap, checked, сдвиг, ошибки преобразования типов и результаты с плавающей запятой.
---

## целочисленное представление

Битовое целое число N имеет биты значения N. Целые числа без знака находятся в диапазоне от 0 до 2^N - 1, а целые числа со знаком - от -2^(N-1) до 2^(N-1) - 1. signedness операции определяет интерпретацию битовой строки.

В таблице ниже поясняются результаты расчетов. IR В примере кода используется текущее представление принтера.

## Сложение, вычитание, умножение

Базовые целые числа add·sub·mul содержат младшие биты N результата. overflow не вызывает trap. Операция checked возвращает тот же результат wrap, а также Bool, который указывает, находится ли математический результат за пределами диапазона соответствующего signed или unsigned.

|операция|Wrap Результат| Checked overflow |
| --- | --- | --- |
| u8: 255 + 1 | 0 | true |
| i8: 127 + 1 | −128 | true |
| u8: 0 − 1 | 255 | true |
| i8: 12 × 3 | 36 | false |

Интерфейсы для языков, которые прерывают выполнение на overflow, должны использовать явный `trap_if` для результата overflow операции checked. Операция по умолчанию не применяет неявно политику overflow исходного языка.

### Wrap и IR выражают явную проверку

Сохраните полный модуль формата 3 как `integer-operations.wir`: текстовый читатель принимает его, скалярный интерпретатор исполняет.

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

`add_u8` возвращает 0, младшие восемь бит числа 256. `minimum_division` возвращает −128, `minimum_remainder` — 0. `negative_shift` трактует −1 как unsigned count 255, затем использует 255 mod 8 = 7. Запускайте по явному ID:

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

`require_no_overflow` извлекает обёрнутый результат по индексу 0 и флаг overflow Bool по индексу 1. Явный trap_if останавливает перед возвратом. Команда завершится со статусом 1 и выведет диагностику в stderr:

```shell
whale ir run integer-operations.wir --function @f1
```

```text
Error: integer-operations.wir: trap at @f1 %b1 instruction 5: "integer overflow" (after 6 steps)
```

## Деление и остаток

Целочисленное деление или остаток на ноль вызывает ловушку. Разделив минимальное значение со знаком на -1, мы получим минимальное значение. Остаток в этом случае равен нулю.

|операция|результат|
| --- | --- |
| i8: −128 / −1 | −128 |
| i8: −128 % −1 | 0 |
|Разделить на целое число 0| trap |
|остаток для целого числа 0| trap |

## сдвиг

При сдвиге битового значения N битовая строка count интерпретируется как unsigned, и используется остаток, разделенный на N. Он не генерирует trap только потому, что count находится вне диапазона от 0 до N-1.

В 8-битных значениях count 0·8·16 все представляют собой 0-битные сдвиги. Битовая строка `11111111` из 8 бит count сдвигается на 7 бит. Это то же самое, даже если эта битовая строка представляет собой signed -1. unsigned Анализ выполняется перед остальными расчетами.

Исходные языки, которые отвергают отрицательные или чрезмерные count, должны поставить явную проверку перед сдвигом.

## преобразование типов

|преобразование|смысл|
| --- | --- |
| Zero extension |Увеличьте ширину, заполнив старшие биты 0.|
| Sign extension |Увеличьте ширину, дублируя знаковый бит.|
|резка бит|Сохраняйте только младшие биты, соответствующие ширине назначения.|
|Ударная реинтерпретация|Интерпретация одной и той же битовой строки как разных типов|
|Числовое преобразование без потерь|Если его невозможно выразить, сохранив числовое значение, trap|

Например, преобразование 8-битного `11111111` в 16-битное zero extension становится `0000000011111111`, а sign extension становится `1111111111111111`. Даже если входные биты одинаковы, это разные операции.

Преобразование float-to-int отбрасывает дробную часть в направлении нуля, затем проверяется диапазон целых чисел. NaN и бесконечность вызывают ловушку. При преобразовании в i8 127,9 становится 127, а 128,0 — ловушкой. Bool преобразуется в целое число 0 или 1. Знак i1 не может представлять 1, поэтому он не может быть местом назначения этого преобразования.

Преобразование адреса в целое число не восстанавливает доступ указателя к этому целому числу. Пожалуйста, обратитесь к [достоверность указателя](memory-model).

### Проверяемые формы cast и операций checked

Проверка сопоставляет фактический тип операнда с `src_ty`, затем проверяет разрешённые opcode категории и разрядности. Аннотация типа результата также должна совпадать с определением. Неявных преобразований нет.

| Opcode | Допустимые типы |
| --- | --- |
| `zext`, `sext` | Целые кроме Bool; разрядность результата строго больше |
| `zext` (Bool) | Bool в целое 0/1; все целые типы назначения кроме знакового `i1`, включая `u1` |
| `trunc` | Целые кроме Bool; разрядность результата строго меньше |
| `fext`, `ftrunc` | Числа с плавающей точкой; разрядность результата строго больше / меньше |
| `itof_s`, `itof_u` | Знаковое / беззнаковое целое в число с плавающей точкой |
| `ftoi_s`, `ftoi_u` | Число с плавающей точкой в знаковое / беззнаковое целое |
| `bitcast` | Целые/вещественные скаляры одной разрядности либо указатель данных в указатель данных |
| `ptrtoint`, `inttoptr` | Указатель данных в целое / целое в указатель данных |

Расширение и усечение целых — битовые операции; знаковость целых операндов может различаться. Bool — отдельный логический тип: знаковое расширение, усечение и bitcast в `i1`/`u1` запрещены. Преобразования агрегатов и указателей функций отклоняются. Проверка категорий указателя и целого не создаёт корректного выделения памяти и не восстанавливает права доступа. Проверки преобразований во время исполнения и lowering в машинный код остаются отдельными незавершёнными задачами.

Следующая полная программа Rust печатает два допустимых преобразования и подтверждает, что замена `zext` на `fext` отклоняется.

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

Следующие фрагменты инструкций намеренно содержат некорректный IR.

```text
%v1: i32 = ftoi_s f64 %v0 to i32   // Invalid when %v0 is actually Bool.
%v2: i16 = zext i32 %v3 to i16     // zext cannot narrow.
%v4: i1 = zext bool %v5 to i1      // Signed i1 cannot represent true as 1.
```

Несовпадение фактического исходного типа даёт `OperandTypeMismatch`; недопустимая пара — `InvalidCast` с opcode и обоими типами. Ошибка проверки не вставляет преобразований и не изменяет модуль.

`sadd_chk`, `ssub_chk`, `smul_chk` требуют знаковые целые, а `uadd_chk`, `usub_chk`, `umul_chk` — беззнаковые. Оба операнда должны точно соответствовать типу операции `T`, результат — `tuple<T, bool>`. Извлечение требует кортеж, существующий индекс поля и точный тип этого поля. В частности, извлечение поля переполнения как `i1` отклоняется. Описанные выше результаты wrap/переполнения остаются контрактом исполнения; эти тесты проверяют структуру, а не выполнение реализованным backend.

## арифметика с плавающей запятой

Значения с плавающей запятой имеют точную битовую строку f16·f32·f64. Операция округляет до ближайшего значения в объявленной ширине и, если ровно посередине, использует ties-to-even, которое выбирает значение с четными младшими битами значащих цифр.

Операция по умолчанию — fast-math, неявная FMA, которая не позволяет принудительно обнулять небольшие значения. Умножение, за которым следует сложение, сохраняет соответствующие шаги округления, и серверная часть не должна неявно объединять их в одну операцию.

Результатом числовой операции может быть NaN или бесконечность. NaN от математических операций и изменений ширины нормализуются до одного фиксированного количества тихих NaN на ширину. Сохраните и скопируйте исходные биты NaN. Поэтому поведение различается при передаче полезных данных NaN в память без арифметических операций и при их вычислении.

Он не предоставляет флаги состояния с плавающей запятой. Несмотря на то, что базовая арифметика с плавающей запятой позволяет получить результат NaN·бесконечность, преобразование float→int применяет приведенные выше правила trap.


### Храните точные константы

Используйте `FloatBits` variant или любую строку шестнадцатеричных бит точной ширины. Равенство сохраненных значений сравнивается с битовой строкой, содержащей отрицательные 0 и NaN payload. Верификатор отклоняет, если ширина типов IR и payload различна.

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

Битовая строка f16, f32 или f64 начинается с `0x`, за которой следуют ровно 4, 8 или 16 шестнадцатеричных цифр соответственно. `0x80000000` представляет собой f32 отрицательный ноль; `0x7f800000` представляет положительную бесконечность. `const_float` преобразует значение хоста f64 в числовую форму; используйте `const_float_bits`, чтобы сохранить исходные биты. Точное представление хранилища не подразумевает полную обработку операций с плавающей запятой. Арифметика времени компиляции по-прежнему использует промежуточные звенья f64 хоста, поэтому контракт полного округления для каждой объявленной ширины еще не реализован.
