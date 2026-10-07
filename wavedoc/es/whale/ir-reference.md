---
translation_set_id: whale-ir-reference
path: whale/ir-reference
locale: es
group: whale
group_order: 1
order: 6
title: Referencia de Whale IR
summary: Describe tipos, identificadores, validez de función, orden de evaluación y formato de intercambio.
---

## Módulos e identificadores

Un módulo consta de información de destino, definiciones globales y funciones. Los valores tienen tipos explícitos. La interfaz resuelve el nombre, el tipo, la sobrecarga y los genéricos del idioma de origen y genera typed IR.

Las funciones y las variables globales utilizan diferentes espacios de nombres internos. Por tanto, funciones y variables pueden tener el mismo nombre. El identificador interno es distinto del nombre de la conexión externa `link_name`, y el nombre externo lo especifica la interfaz. Whale no resuelve conflictos externos generando automáticamente nuevos nombres. Consulte [Símbolos y enlaces](assembler-linker).

Cada definición de valor tiene un identificador. Las definiciones no se pueden duplicar y los metadatos de tipo deben coincidir con el tipo especificado en la definición. El nombre por sí solo no identifica definiciones, incluso si las declaraciones con el mismo nombre se oscurecen entre sí.

## IR Configuración y lectura

A continuación se muestra un ejemplo completo de Rust que construye y verifica la función con la caja `ir` y luego genera typed IR.

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

La impresora genera IR:

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

Guarde la salida como `answer.wir` para leerla y verificarla con el parser o CLI. La ejecución entera escalar se explica abajo.

Guarde la salida como `answer.wir` para leerla y verificarla con el analizador de texto y la CLI. La ejecución de IR aún no está disponible.

## tipo

|tipo|significado|
| --- | --- |
| `bool` |Valor lógico false o true|
| `i1`, `i8`, `i16`, `i32`, `i64`, `i128` |entero con signo de ancho de bits especificado|
| `u1`, `u8`, `u16`, `u32`, `u64`, `u128` |entero sin signo de ancho de bits especificado|
| `f16`, `f32`, `f64` |Valor de coma flotante con ancho de bits especificado|
| `ptr<T>` |T Puntero para escribir valor|
| `fnptr<signature>` | Puntero invocable con tipos exactos de parámetros/resultados y convención de llamada |
| `array<T, N>` |N elementos del mismo tipo|
| `struct{T, ...}` |Campo de estructura ordenada|
| `tuple<T, ...>` |elementos de tupla ordenados|
| `void` |Sin resultados|

`bool`, `i1` y `u1` son tipos diferentes. signed `i1` representa −1 y 0, y unsigned `u1` representa 0 y 1. El número entero 1 no es una condición lógica implícita. Rama condicional, la condición de Select, `trap_if` requiere el operando Bool.

El tamaño de almacenamiento no está determinado únicamente por la cantidad de bits en el valor y sigue a [diseño de destino](memory-model). Por ejemplo, el valor de `i1` es de 1 bit, pero ocupa al menos 1 byte en la memoria.

## Funciones y llamadas

La función especifica todos los parámetros, tipos de resultados, convención de llamada y linkage. Las llamadas directas e indirectas deben coincidir con la firma de quien llama. La llamada void no tiene resultado ID. Una llamada a nonvoid conserva la definición del resultado incluso si O0 no utiliza el resultado.

La devolución debe coincidir con el tipo de resultado de la función. El retorno void no lleva un valor y el retorno nonvoid lleva un valor del tipo de resultado declarado.

### Declaraciones, identidades y convocatorias

`Module.declarations` registra el `FunctionId`, el nombre, la firma completa, el enlace y el nombre del enlace externo de cada función. Una definición se refiere a esta identidad; Los tipos de parámetros y de retorno deben coincidir con su declaración. Las declaraciones repetidas idénticas se resuelven en el mismo ID hasta `declare_function`; Los conflictos y las definiciones duplicadas son errores. Una declaración interna necesita un cuerpo en el módulo. Una declaración externa puede quedar sin resolver hasta vincularse, o tener cuerpo exportado. Las funciones internas no tienen `link_name`; Las funciones externas requieren un nombre explícito no vacío sin NUL. Dos declaraciones de funciones distintas no pueden reclamar el mismo nombre externo. Los globales y las funciones todavía usan espacios de nombres internos separados.

Registre declaraciones antes de construir cuerpos con `begin_declared_function` para admitir llamadas directas y recursividad. `begin_function` sigue siendo una conveniencia para una nueva función interna Whale. Las API `declare_function`, `begin_declared_function`, `function_addr`, `null_function` y `call` marcadas devuelven `Result`; una llamada rechazada no agrega una instrucción ni asigna su ID de resultado.

El siguiente programa completo de Rust declara una función externa, toma su dirección escrita y emite llamadas directas e indirectas:

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

`Callee::Direct(FunctionId)` resuelve mediante la tabla de declaración; `Callee::Indirect(ValueId)` requiere un valor `Type::FnPtr(FunctionSignature)`. La firma incluye todos los tipos de parámetros, el tipo de resultado y `CallingConvention::{Whale, SysV64}`. Se conserva a través de copias, almacenamiento, parámetros, devoluciones, phi y selección. Los punteros de datos y los valores enteros no se pueden llamar. Se rechazan las conversiones que involucran tipos de puntero de función; cambiar una anotación de tipo no puede cambiar una firma invocable. Un puntero de función tiene almacenamiento de direcciones de 64 bits en este destino; esto por sí solo no implementa metadatos ocultos en tiempo de ejecución.

La aridad, los tipos exactos de argumento/resultado, la presencia de ID de resultado y la convención de llamada deben coincidir. No hay conversiones implícitas. El destinatario indirecto debe dominar la llamada al igual que sus argumentos. `variadic: true`, los parámetros de tipo void y las firmas de parámetros/resultados agregados SysV64 se rechazan. Whale las firmas agregadas se pueden representar en IR; La clasificación ABI nativa y la emisión de llamadas automáticas aún no están disponibles para ninguna de las convenciones.

`null_function(signature)` representa un puntero de función nula escrito. Llamarlo está bien escrito IR con una captura de tiempo de ejecución requerida antes de ingresar a un destinatario. También se debe capturar un destino no nulo que no sea válido, haya caducado o sea incompatible con la firma comprobada. Estas comprobaciones de tiempo de ejecución y la gestión de la vida útil de las devoluciones de llamadas externas esperan a la capa de ejecución nativa/intérprete; El éxito del verificador no significa que las direcciones externas arbitrarias sean seguras.

### Formas de llamada en el AST

Estos son fragmentos de expresión dentro de un programa de formato 2 AST:

```json
{"Call":{"callee":{"Direct":"increment"},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

```json
{"Call":{"callee":{"Indirect":{"FunctionRef":"increment"}},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

`Direct` y `FunctionRef` usan el espacio de nombres de la función incluso cuando una variable tiene el mismo nombre. `Indirect` primero evalúa su expresión y luego evalúa los argumentos de izquierda a derecha. Una llamada nula es válida como `ExprStmt`, pero no como inicializador de variable, argumento, operando o valor devuelto. Las llamadas y referencias de funciones no son expresiones constantes numéricas en tiempo de compilación. `NullFunction` toma un objeto de firma con los campos `params`, `ret`, `convention` y `variadic`.

El [ejemplo JSON completo](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/ast-v2-calls.json) almacena una devolución de llamada y la invoca antes de una llamada externa. Bájalo con:

```sh
cargo run --locked --features socket-cli -- ir lower ir/tests/fixtures/ast-v2-calls.json
```

Su [esperado IR](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/calls-v3.wir) se verifica en las pruebas de reducción. La identidad de la función y los nombres de los enlaces se representan en el límite IR; preservarlos mediante la generación y vinculación de objetos nativos sigue siendo un trabajo separado.

## Disponibilidad de bloques y valores.

Cada bloque tiene un identificador único y exactamente un terminator. Los objetivos de sucursal deben pertenecer a la misma función. El bloque de entrada debe existir y no puede tener un borde anterior y phi. Al crear un bucle, pase del bloque de entrada a un encabezado de bucle separado.

La definición del valor en la ruta ejecutable debe regir su punto de uso típico. Esto significa que todos los caminos desde el punto de entrada hasta el punto de uso deben pasar por esa definición. En el mismo bloque, las definiciones deben ir antes que los usos. El orden de almacenamiento de los bloques no determina la dominancia.

Supongamos que el punto de entrada se bifurca a left o right y luego se une en join. Los valores definidos únicamente en left no se pueden utilizar como valores generales en join. Esto se debe a que la ruta que pasa por right no está definida. Los valores de cada bloque anterior deben combinarse en phi, que se recibe como entrada.

Los bloques inalcanzables también se conservan en el módulo. El verificador verifica continuamente el identificador, tipo, operando y estructura de rama del bloque. La definición de un bloque inalcanzable no puede aportar valor al uso normal de una ruta accesible.

## comando Phi

phi se coloca antes de todos los comandos normales del bloque. Se requiere exactamente una entrada para cada bloque anterior diferente. El valor de entrada debe ser de tipo phi y debe estar disponible al final del bloque anterior correspondiente.

Incluso si hay múltiples aristas en un bloque anterior, solo hay una entrada. El bucle phi puede referirse al valor de un bloque que aparece más adelante en el orden de almacenamiento del módulo siempre que sea un valor calculado en el flanco de repetición. Los bloques anteriores faltantes, duplicados o irrelevantes y las entradas escritas incorrectamente son errores de validación.

## Evaluación y Selección

Whale AST evalúa los objetivos de llamada y las subexpresiones desde la izquierda donde se especifican. El frontend expresa la evaluación de cortocircuito como una rama del flujo de control.

Select selecciona uno de los valores ya calculados. No omite el cálculo de ninguna de las entradas. Por ejemplo, incluso si elige un valor seguro, no puede evitar que se encuentre trap al calcular otras entradas. Los cálculos que deben ejecutarse sólo en una ruta específica deben colocarse dentro de un bloque condicional.

## Departamento de Verificación trap

IR mal formado es un error de verificación. El verificador rechaza legacy `undef`; regenere desde AST. Una declaración sin inicializador usa `uninit` de format 4 y lecturas comprobadas, sin sustituir por cero ni valores arbitrarios. Las condiciones de ejecución violadas producen traps definidos con ubicación IR.

`InterpreterTrap` informa del motivo, pasos ejecutados y `ExecutionSite`: ID de función, ID de bloque, índice de instrucción desde cero e ID opcional del resultado. El terminator sigue a las instrucciones. El CLI indica también el archivo. El IR aún no lleva spans de fuente: son posiciones IR, no números de línea. El trap devuelve un error y detiene la ejecución posterior; la biblioteca no aborta el proceso anfitrión.

Esta garantía se aplica a la IR verificada y a la memoria de seguimiento. Externo C·dirección sin formato·el ensamblaje en línea tiene un contrato separado y no siempre detecta violaciones fuera de sus límites. Consulte [Modelo de memoria](memory-model).

## Formato de intercambio y representación de texto

AST y typed IR utilizan el format version respectivo y el semantics version común. El lector debe rechazar la clave JSON no versionada/desconocida/campo/función/duplicada. Los constructores no deben asumir que las propiedades no admitidas se ignorarán silenciosamente.

Los números enteros se pasan como ancho de bits·signedness·números de cadena. Las constantes de punto flotante se pasan como una cadena de bits exacta y de ancho. El texto round-trip en IR debe conservar el nombre·ID·tipo·constante·secuencia·propiedad·metadatos. Los espacios y la ubicación de comentarios no están sujetos a conservación.

Están disponibles el contrato AST JSON siguiente y la lectura, verificación e impresión de ida y vuelta de typed IR format 4.

### Identidades impresas y nombres entre comillas

El formato 3 de typed IR imprime funciones como `@fN`, globales como `@gN`, valores como `%vN` y bloques como `%bN`. Los ID de funciones y globales pertenecen al módulo; los de valores y bloques, a la función que los contiene. Se conservan los ID suministrados, incluso si hay huecos. Los nombres entre comillas son anotaciones descriptivas y no resuelven referencias. Cada función declara `entry %bN`, independientemente del orden de almacenamiento de los bloques.

Este módulo completo se verificó e imprimió con la API Rust de IR. Ambos bloques de rama se llaman `"branch"`; sus ID distinguen las definiciones y las entradas de phi.

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

`%v0` se define en la lista de parámetros. `%b1` y `%b2` son distintos aunque tengan el mismo nombre; phi identifica cada predecesor por ID. Las ramas y los destinos de switch usan la misma sintaxis de ID de bloque. El impresor no renumera el `%v99` explícito.

Todos los nombres y cadenas usan comillas dobles: destino, nombres de funciones, globales, parámetros y bloques, nombres de enlace externo, declaraciones de constantes y motivos de trap. Unicode imprimible se conserva. Los escapes son `\"`, `\\`, `\n`, `\r`, `\t`, `\0` y `\u{hex}`, con hexadecimal en minúsculas para los demás caracteres de control y U+2028/U+2029. Un nombre con salto de línea, tabulación, comillas, barra inversa y texto coreano se imprime en un solo registro.

```text
"line\ncolumn\tquote\"slash\\한글"
```

El lector acepta typed IR formats 3 y 4 con semantics version 1 y siempre imprime format 4. `uninit` exige format 4. Las instrucciones válidas de format 3 siguen siendo legibles, pero legacy `undef` es error de verificación en ambos formatos y debe regenerarse desde AST. El texto format 2 exige migración manual a IDs explícitos, nombres entre comillas y referencias de entrada. AST JSON mantiene su format 2 independiente.

### Lectura y verificación de IR textual

`ir::parse_module` lee typed IR, lo verifica y devuelve un `Module`. Acepta la sintaxis actual del impresor para escalares, control de flujo, memoria, llamadas directas/indirectas y expresiones constantes. Conserva IDs, nombres, tipos, enteros y bits float exactos, árboles de expresiones y resultados evaluados, orden de bloques, entrada, alineación, firmas y link_name. Los espacios y comentarios de línea `//` se normalizan. Guarde el IR completo anterior como `answer.wir` y ejecute estos comandos con la compilación predeterminada.

```sh
cargo run --locked -- ir verify answer.wir
cargo run --locked -- ir print answer.wir -o canonical.wir
```

Se rechazan versiones, campos, instrucciones y escapes desconocidos, literales fuera de rango, IDs duplicados, anotaciones de tipo contradictorias y entrada sobrante. `ParseError` indica el desplazamiento en bytes y la línea/columna de escalares Unicode desde 1. Los errores de verificación se ubican en la función o declaración global correspondiente cuando es posible. `print` también verifica y conserva la salida existente si falla. La nueva sintaxis o semántica requiere cambiar la format/semantics version correspondiente; las versiones desconocidas son errores.

### Límites de recursos de verificación

Los valores predeterminados de `IrLimits` son 8 MiB de entrada, 1,000,000 tokens, 1,000,000 nodos de recorrido y profundidades de tipos y expresiones de 128. La raíz tiene profundidad 0 y cada hijo suma 1. Pueden reducirse o elevarse hasta `MAX_IR_NESTING`, 256. Los excesos devuelven `LimitError`, `VerifyError::ResourceLimit`, `ConstEvalError::ResourceLimit` o `CallError::ResourceLimit`. Los nodos cuentan el recorrido en cada límite de entrada, incluidas las anotaciones y expresiones, no el tiempo transcurrido.

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

También puede pasar límites a `verify_module_with_limits`, `ConstExpr::evaluate_with_limits`, `validate_signature_with_limits` y `ModuleBuilder::declare_function_with_limits`. El recorrido iterativo precede al clone, comparación y diagnóstico recursivos; las constantes se evalúan con una pila de trabajo. Los árboles Rust prestados y su Drop siguen siendo responsabilidad del llamador. Los árboles arbitrarios no verificados conservan clone/Drop recursivos; la firma propia rechazada por la API de declaración checked se libera iterativamente. IR mal formado es un error de verificación. El verificador rechaza legacy `undef`; regenere desde AST. Una declaración sin inicializador usa `uninit` de format 4 y lecturas comprobadas, sin sustituir por cero ni valores arbitrarios. Las condiciones de ejecución violadas producen traps definidos con ubicación IR.

### Versión especificada AST JSON

Guarde lo siguiente como `program.json`. Los cuatro campos del sobre son obligatorios. `program` contiene las matrices `declarations`, `globals` y `functions` requeridas, que pueden estar vacías. Se requieren el nombre de la función, los parámetros, el tipo de retorno, el cuerpo, `convention` y `linkage`. `link_name` puede estar ausente/nulo para funciones internas y debe ser una cadena no vacía sin NUL para funciones externas. Cada enumeración utiliza el nombre de su unidad o un único objeto de clave variante. Las variantes de unidades también aceptan un objeto con valor nulo, como `{"Void":null}`; el codificador emite el nombre de la unidad `"Void"`. `VarDecl.init` puede estar ausente o ser nulo; otros campos obligatorios deben estar presentes.

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

El número entero `value` es una cadena decimal. Signed Se utiliza un número después del menos opcional de un número entero y no se permiten espacios, más, exponentes ni separadores. El rango permitido está determinado por el ancho declarado y signedness. El `u128::MAX` anterior se conserva tal cual hasta JSON y lowering. Los valores unsigned negativos o los valores fuera de rango no son wrap y son errores. El valor Float utiliza una cadena de bits hexadecimal del ancho exacto descrito en [Operaciones numéricas](numeric-operations).

`format_version` es 2 para este formato AST; `semantics_version` es 1. `features` debe ser una matriz vacía. Los campos desconocidos, las versiones, las características, las claves JSON sin procesar duplicadas (incluidas las claves equivalentes con escape) y los valores finales son errores, incluso con `--no-verify`. El punto de entrada a la biblioteca es `ir::lower_ast::interchange::decode`; `encode` emite la envolvente. `decode` tiene por defecto un límite de bytes de origen de 8 MiB; `decode_with_limit` acepta un límite de bytes elegido por quien llama. JSON el anidamiento está limitado. Utilice este decodificador sin formato en lugar de analizarlo en un mapa genérico que ya podría descartar claves duplicadas.

[El esquema JSON completo](https://github.com/wavefnd/Whale/blob/master/ir/schema/ast-v2.schema.json) especifica formas, campos obligatorios y variantes. También se aplican comprobaciones de rango/tipo y detección de claves duplicadas. El subconjunto de reducción escalar incluye literales, variables/constantes, agregar/sub/mul, comparaciones, asignación, si/mientras, regresar y romper/continuar. Se admiten referencias de funciones, llamadas directas y llamadas indirectas; Las expresiones agregadas no son compatibles. `Opaque` se puede representar en el esquema, pero no se puede reducir.

La migración exige envolver el antiguo Program sin envelope y sustituir los números JSON por cadenas decimales de enteros o cadenas de bits flotantes. Se rechazan entradas sin versión. El formato 1 debe migrarse al 2 añadiendo `program.declarations` (un array vacío si no se usa) y `convention`/`linkage` explícitos en las definiciones. Las versiones son independientes: AST format 2, typed IR format 4 y semantics version 1.

### Entrada rechazada y recuperación CLI

Guarde la siguiente entrada completa como `invalid.json`.

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

El comando sale con un estado distinto de cero y no crea nuevos resultados ni sobrescribe archivos existentes. Las discrepancias de tipos también fallarán antes de publicar el resultado. Los binarios creados sin `socket-cli` salen con estado 2 y generan un comando de recuperación que contiene `--features socket-cli`.


## Intérprete de enteros escalares

El intérprete predeterminado ejecuta enteros/Bool, control de flujo, asignaciones de pila, almacenamiento y lectura de punteros de datos, typed GEP, memcpy y memset. Las lecturas de pares checked excluyen el relleno. Las direcciones son valores sintéticos de 64 bits, sin desreferenciar memoria del anfitrión. Los argumentos y retornos siguen limitados a enteros/Bool o void; float, llamadas, punteros de función, valores agregados generales, direcciones globales y ejecución native no están soportados. La pila vive hasta el retorno. El fin de vida léxico, la transferencia de punteros en llamadas/retornos, adaptadores de memoria externa y native shadow metadata requieren implementación posterior.

`InterpreterOptions::max_steps` vale 1,000,000 por defecto. Cada instrucción ejecutada, incluida phi, y cada terminator consumen un paso. Cero detiene antes de la primera operación; un bucle infinito devuelve `InterpreterError::StepLimit`. `ir_limits` limita la verificación por separado. La aritmética sin usos también se ejecuta y puede producir trap. El overflow checked es un resultado Bool; solo un trap_if explícito lo convierte en trap.

Guarde este módulo completo como `swap-loop.wir`. La entrada es explícita aunque la salida se almacene primero. Al entrar en un bloque, se leen todos los operandos phi del bloque anterior antes de escribir ningún resultado phi. Tres iteraciones intercambian 11 y 22 tres veces y devuelven 22.

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

La API Rust devuelve valor y pasos, o un error estructurado de verificación, operación no admitida, argumentos, límite o trap. Este programa completo lee el mismo archivo `swap-loop.wir`:

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
