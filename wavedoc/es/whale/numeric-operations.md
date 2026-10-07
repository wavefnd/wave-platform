---
translation_set_id: whale-numeric-operations
path: whale/numeric-operations
locale: es
group: whale
group_order: 1
order: 7
title: Operaciones numéricas
summary: Describe operaciones de números enteros wrap, checked, cambios, errores de conversión de tipos y resultados de punto flotante.
---

## representación entera

Un entero de bits N tiene bits de valor N. Los enteros sin signo varían de 0 a 2^N − 1, y los enteros con signo varían de −2^(N−1) a 2^(N−1) − 1. El signedness de la operación determina la interpretación de la cadena de bits.

La siguiente tabla explica los resultados del cálculo. IR El ejemplo de código utiliza la representación de la impresora actual.

## Suma, resta, multiplicación.

Los enteros básicos add·sub·mul contienen los bits N inferiores del resultado. overflow no causa trap. La operación checked devuelve el mismo resultado wrap, así como Bool, que indica si el resultado matemático está fuera del rango del signed o unsigned correspondiente.

|operación|Wrap Resultado| Checked overflow |
| --- | --- | --- |
| u8: 255 + 1 | 0 | true |
| i8: 127 + 1 | −128 | true |
| u8: 0 − 1 | 255 | true |
| i8: 12 × 3 | 36 | false |

Las interfaces de los lenguajes que interrumpen la ejecución en overflow deben usar un `trap_if` explícito para el overflow resultado de la operación checked. La operación predeterminada no aplica implícitamente la política overflow del idioma de origen.

### Wrap y IR expresando un cheque explícito

Guarde este módulo completo de formato 3 como `integer-operations.wir`; el lector de texto lo acepta y el intérprete escalar lo ejecuta.

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

`add_u8` devuelve 0, los ocho bits bajos de 256. `minimum_division` devuelve −128 y `minimum_remainder`, 0. `negative_shift` interpreta −1 como unsigned count 255 y aplica 255 mod 8 = 7. Ejecute por ID explícito:

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

`require_no_overflow` extrae el resultado envuelto en el índice 0 y el indicador Bool de overflow en el 1. El trap_if explícito detiene antes del retorno. El comando siguiente termina con estado 1 y escribe el diagnóstico en stderr:

```shell
whale ir run integer-operations.wir --function @f1
```

```text
Error: integer-operations.wir: trap at @f1 %b1 instruction 5: "integer overflow" (after 6 steps)
```

## División y resto

La división de números enteros o el resto por cero provoca una trampa. Dividiendo el valor mínimo con signo por −1 se ajusta al valor mínimo. El resto en este caso es cero.

|operación|resultado|
| --- | --- |
| i8: −128 / −1 | −128 |
| i8: −128 % −1 | 0 |
|Dividir por número entero 0| trap |
|resto para el entero 0| trap |

## turno

Al cambiar el valor de bit N, la cadena de bits de count se interpreta como unsigned y se utiliza el resto dividido por N. No genera trap solo porque count está fuera del rango de 0 a N−1.

En valores de 8 bits, count 0·8·16 son todos desplazamientos de 0 bits. La cadena de bits `11111111` de 8 bits count se desplaza 7 bits. Esto es lo mismo incluso si esta cadena de bits representa signed −1. unsigned El análisis se realiza antes que el resto de los cálculos.

Los idiomas de origen que rechazan count negativo o excesivo deben realizar una verificación explícita antes del cambio.

## conversión de tipo

|conversión|significado|
| --- | --- |
| Zero extension |Aumente el ancho llenando bits de orden superior con 0|
| Sign extension |Aumente el ancho duplicando el bit del signo.|
|corte de bits|Mantenga solo los bits de orden inferior correspondientes al ancho de destino|
|Beat reinterpretación|Interpretar la misma cadena de bits como diferentes tipos|
|Conversión numérica sin pérdidas|Si no se puede expresar conservando el valor numérico, trap|

Por ejemplo, convertir `11111111` de 8 bits en zero extension de 16 bits se convierte en `0000000011111111` y sign extension se convierte en `1111111111111111`. Incluso si los bits de entrada son los mismos, son operaciones diferentes.

La conversión de flotante a int se trunca hacia cero y luego verifica el rango de enteros. NaN y el infinito provocan una trampa. Al convertir a i8, 127,9 se convierte en 127, mientras que 128,0 queda atrapado. Bool se convierte a un número entero 0 o 1. i1 con signo no puede representar 1, por lo que no puede ser el destino de esta conversión.

Convertir una dirección a un número entero no restaura el acceso del puntero a ese número entero. Consulte [validez del puntero](memory-model).

### Formas verificadas de cast y operaciones checked

La verificación compara el tipo real del operando con `src_ty` y después comprueba las categorías y anchuras permitidas por el opcode. La anotación del resultado también debe coincidir con su definición. No hay conversiones implícitas.

| Opcode | Tipos aceptados |
| --- | --- |
| `zext`, `sext` | Enteros distintos de Bool; anchura de destino estrictamente mayor |
| `zext` (Bool) | Conversión de Bool a entero 0/1; todos los destinos enteros excepto `i1` con signo, incluido `u1` |
| `trunc` | Enteros distintos de Bool; anchura de destino estrictamente menor |
| `fext`, `ftrunc` | Flotantes; destino estrictamente más ancho / más estrecho |
| `itof_s`, `itof_u` | Entero con signo / sin signo a flotante |
| `ftoi_s`, `ftoi_u` | Flotante a entero con signo / sin signo |
| `bitcast` | Escalares enteros/flotantes de igual anchura, o puntero de datos a puntero de datos |
| `ptrtoint`, `inttoptr` | Puntero de datos a entero / entero a puntero de datos |

La extensión y el truncamiento enteros son operaciones de bits; los operandos enteros pueden diferir en signo. Bool es un tipo lógico separado: no admite extensión de signo, truncamiento ni bitcast a `i1`/`u1`. Se rechazan casts de agregados y punteros a funciones. Validar las categorías puntero/entero no establece una asignación válida ni recupera permisos. Las comprobaciones de conversión en ejecución y el lowering a código máquina siguen pendientes por separado.

Este programa Rust completo imprime dos conversiones válidas y confirma que sustituir `zext` por `fext` se rechaza.

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

Estos fragmentos de instrucciones son IR deliberadamente inválido.

```text
%v1: i32 = ftoi_s f64 %v0 to i32   // Invalid when %v0 is actually Bool.
%v2: i16 = zext i32 %v3 to i16     // zext cannot narrow.
%v4: i1 = zext bool %v5 to i1      // Signed i1 cannot represent true as 1.
```

Un tipo real de origen distinto produce `OperandTypeMismatch`; un par ilegal produce `InvalidCast` con el opcode y ambos tipos. Ningún error inserta conversiones ni modifica el módulo.

`sadd_chk`, `ssub_chk` y `smul_chk` requieren enteros con signo; `uadd_chk`, `usub_chk` y `umul_chk`, enteros sin signo. Ambos operandos deben coincidir exactamente con el tipo `T`, y el resultado debe ser `tuple<T, bool>`. La extracción exige una tupla, un índice existente y el tipo exacto de ese campo. En particular, se rechaza extraer el desbordamiento como `i1`. Los resultados de wrap/desbordamiento anteriores siguen siendo el contrato de ejecución; las pruebas comprueban la estructura, no la ejecución mediante un backend implementado.

## aritmética de coma flotante

Los valores de punto flotante tienen la cadena de bits exacta f16·f32·f64. La operación redondea al valor más cercano en el ancho declarado y, si está exactamente en el medio, utiliza ties-to-even, que selecciona un valor incluso con los bits menos significativos de los dígitos significativos.

La operación predeterminada es fast-math, implícita FMA, que no permite el manejo de cero forzado de valores pequeños. La multiplicación seguida de la suma conserva sus respectivos pasos de redondeo y el backend no debe combinarlos implícitamente en una sola operación.

El resultado de una operación numérica puede ser NaN o infinito. Los NaN de operaciones matemáticas y cambios de ancho se normalizan a una cantidad fija de NaN silencioso por ancho. Guarde y copie para conservar los bits NaN originales. Por lo tanto, el comportamiento es diferente al pasar la carga útil de NaN a la memoria sin operaciones aritméticas y al calcularla.

No expone indicadores de estado de punto flotante. Aunque la aritmética básica de punto flotante permite NaN·resultados infinitos, la conversión float→int aplica las reglas trap anteriores.


### Almacenar constantes exactas

Utilice `FloatBits` variant o cualquier cadena de bits hexadecimales del ancho exacto. La igualdad de los valores almacenados se compara con una cadena de bits que contiene 0 negativo y NaN payload. El verificador rechaza si los anchos de tipo IR y payload son diferentes.

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

Una cadena de bits f16, f32 o f64 comienza con `0x`, seguida exactamente de 4, 8 o 16 dígitos hexadecimales, respectivamente. `0x80000000` representa f32 cero negativo; `0x7f800000` representa infinito positivo. `const_float` convierte un valor de host f64 numéricamente; utilice `const_float_bits` para conservar los bits originales. La representación exacta del almacenamiento no implica un backend de ejecución de punto flotante completo. La aritmética en tiempo de compilación todavía utiliza intermediarios del host f64, por lo que aún no se ha implementado el contrato de redondeo completo para cada ancho declarado.
