---
translation_set_id: whale-memory-model
path: whale/memory-model
locale: es
group: whale
group_order: 1
order: 8
title: Modelo de memoria
summary: Seguimiento de asignaciones, lectura de valores inicializados, aritmética de punteros, diseño y reglas de almacenamiento de cadenas.
---

## Seguimiento de asignaciones

Un puntero rastreado asocia una dirección con un ID de asignación, generación, límites, desplazamiento y permisos de acceso. El modelo de memoria también rastrea la vida útil y el estado de inicialización de la asignación. El acceso debe cumplir estas condiciones; una infracción provoca una trampa.

Aunque se reutiliza la misma dirección física, las generaciones se separan en cada vida. La mera presencia de una dirección no determina que el puntero sea válido o que la persona que llama tenga acceso a ese espacio de almacenamiento.

La dirección native mantiene 64 bits. Los shadow metadata separados se pasan junto con el puntero a través de copiar/guardar/llamar/devolver. El alcance de la memoria inicial native es la pila rastreable y las asignaciones globales. C Los límites requieren un adaptador explícito y la transferencia de propiedad de una memoria externa arbitraria no está incluida en este alcance.

## Inicialización y lectura

Declarar un espacio de almacenamiento no inicializa el valor. Comprueba si se ha inicializado el rango de bytes que realmente se están leyendo. Si incluso parte del valor se lee en un estado no inicializado, es trap. No reemplaza el resultado de la lectura con 0 o un valor no especificado.

La verificación de inicialización del valor leído excluye el byte padding. Por ejemplo, si se inicializan todos los campos de una estructura, la lectura de los valores no será incorrecta simplemente porque los bytes vacíos resultantes de la alineación de los campos no estén inicializados.

La copia de memoria lleva el estado de inicialización junto con los bytes. Copiar espacio de almacenamiento no inicializado no lo cambia a espacio de almacenamiento inicializado. Al leer valores del destino más tarde, se aplican las mismas comprobaciones que al leer el origen.

El mero hecho de que el byte físico de BSS sea 0 no permite la inicialización de la variable IR.

### IR en espacio de almacenamiento escalar inicializado

El siguiente módulo se configuró como builder y pasó el verificador. `store` antes de leer el valor, y las tres instrucciones de memoria especifican una clasificación de potencia de 2 distinta de cero.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "initialized_local": whale () -> i32, linkage internal

  fn @f0 "initialized_local"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: ptr<i32> = alloca i32, align 4
    %v1: i32 = const i32 42
    store i32 %v1, ptr<i32> %v0, align 4
    %v2: i32 = load i32, ptr<i32> %v0, align 4
    ret i32 %v2
  }

}
```

Guarde este módulo completo como `initialized.wir`; al ejecutarlo imprime `i32 42`. Quitar store provoca un trap por lectura no inicializada en load. Alineación 3 es un error de verificación. Los ceros físicos de alloca no establecen inicialización.

## Comparación con la aritmética de punteros

El cálculo de la dirección overflow es trap. Puede crear un puntero one-past que apunte inmediatamente después del rango de asignación, pero no puede acceder a la memoria a través de él.

La igualdad de punteros utiliza una identidad de asignación, no solo una dirección numérica. Ordenar o restar punteros de diferentes asignaciones provoca una trampa. Reconstruir una dirección a partir de un número entero no restaura los permisos de acceso.

### GEP

GEP calcula direcciones por elementos y campos. Este no es un comando para leer un valor.

El primer índice es la unidad de elemento offset del tipo al que apunta el puntero base. El índice posterior selecciona un elemento de matriz o campo de estructura/tupla. El índice de campo de una estructura/tupla es el número de orden del campo en el momento de la compilación, no el byte offset. El tipo seleccionado determina el tipo de puntero resultante.

Cuando el tipo base es `ptr<array<i32, 4>>`, el índice `[0, 2]` selecciona el tercer elemento i32 de la matriz, lo que da como resultado `ptr<i32>`. Al especificar el primer índice como 1, se mueve i32 un elemento, en lugar de un elemento en la matriz. Este ejemplo ilustra la semántica del índice y no la sintaxis del comando de texto.

El objetivo inicial native rechaza la aritmética de punteros en elementos de tamaño 0. Calcular la dirección no elimina las comprobaciones de duración, alcance, inicialización y permisos necesarias para el acceso posterior.

## diseño de datos

El objetivo de salida determina el campo de alineación de tamaño offset·matriz stride. El orden de los campos en estructuras y tuplas conserva su orden de declaración. No se debe asumir que el diseño del host que ejecuta el compilador es el diseño del destino de salida.

|valor|Guardar reglas|
| --- | --- |
| Bool, signed i1, unsigned u1 |Al menos 1 byte|
|Estructura/tupla vacía|Tamaño 0, alineación 1|
|matriz|El diseño del objetivo determina el elemento stride|
|Estructura/tupla|Preservar el orden de declaración y los requisitos de alineación del objetivo.|

La alineación completa de IR es una potencia de 2, no de 0. La alineación automática debe determinarse antes de generar este IR. Los diseños packed·union·bitfield no son compatibles con este perfil y deben rechazarse.

### Consulta de diseño de salida

Rust API calcula el diseño de almacenamiento independientemente del host de compilación. En el siguiente ejemplo, hay 7 bytes padding antes del campo u64 y 6 bytes al final padding.

```rust
use ir::{allocation_align, layout_of, Target, Type};

fn main() {
    let target = Target::X86_64WhaleLinux;
    let record = Type::Struct(vec![Type::U8, Type::U64, Type::U16]);
    let layout = layout_of(&record, target).unwrap();
    assert_eq!((layout.size, layout.align), (24, 8));
    assert_eq!(layout.field_offsets, [0, 8, 16]);

    let array = Type::Array(Box::new(record), 3);
    let layout = layout_of(&array, target).unwrap();
    assert_eq!((layout.size, layout.align), (72, 8));
    assert_eq!(layout.element_stride, Some(24));
    assert_eq!(allocation_align(&array, target).unwrap(), 16);
}
```

```text
struct{u8, u64, u16}: size 24, natural alignment 8
field 0: byte 0
field 1: byte 8
field 2: byte 16
array of 3: size 72, element stride 24
standalone array placement alignment: 16
```

El tamaño devuelto por `layout_of` y el paso de una matriz incluyen el relleno final. Las estructuras y las tuplas utilizan las mismas reglas de orden de campos. Bool, i1 y u1 ocupan cada uno un byte. Las estructuras y tuplas vacías tienen tamaño 0 y alineación 1; una matriz de longitud cero conserva la alineación natural de su elemento. `void` no tiene diseño de almacenamiento, mientras que `ptr<void>` ocupa 8 bytes.

Utilice clasificación natural para campos y elementos de matriz. `allocation_align` aplica la regla SysV AMD64, que requiere una alineación de al menos 16 bytes, a matrices locales/globales independientes de 16 bytes o más. No aumenta la alineación de los campos de la matriz ni del elemento stride. AST lowering utiliza esta búsqueda de clasificación por lotes para el almacenamiento local.

Multiplicación de magnitud, campo offset suma, overflow en el cálculo padding devuelve `LayoutError::Overflow`. `layout_of` tiene un límite de anidamiento de tipo complejo de 128 niveles y `layout_of_with_limit` permite a la persona que llama especificar el límite. No asigna un espacio de almacenamiento igual al número de elementos de la matriz. `pointer_stride` rechaza pointee con tamaño 0 en native aritmética de punteros, pero el diseño de almacenamiento de ese tipo en sí es válido. Packed·union·bitfield no tiene una expresión de tipo compatible. Esta búsqueda de diseño almacenado no implementa convenciones de llamada de tipos complejos ni verificación de límites de tiempo de ejecución.

## Cadena y límite C

Una cadena es una cadena inmutable de bytes con una longitud especificada. La codificación predeterminada es UTF-8. Permite un NUL interno y no agrega automáticamente la terminación NUL. O0 no combina automáticamente objetos de cadena con el mismo contenido.

Por lo tanto, la longitud de la cadena de bytes que consta de `A`, NUL y `B` es 3. No se puede utilizar para conversiones de cadenas C explícitas, que rechazan el NUL interno. La interfaz no debería truncar esto silenciosamente a `A`.

Externo C·Dirección sin formato·El ensamblaje en línea es un límite de contrato separado. No se garantiza que la inspección en tiempo de ejecución de la memoria de seguimiento detecte todos los comportamientos incorrectos en el código externo.

## Ejecución de memoria de pila rastreada

El intérprete predeterminado ejecuta enteros/Bool, control de flujo, asignaciones de pila, almacenamiento y lectura de punteros de datos, typed GEP, memcpy y memset. Las lecturas de pares checked excluyen el relleno. Las direcciones son valores sintéticos de 64 bits, sin desreferenciar memoria del anfitrión. Los argumentos y retornos siguen limitados a enteros/Bool o void; float, llamadas, punteros de función, valores agregados generales, direcciones globales y ejecución native no están soportados. La pila vive hasta el retorno. El fin de vida léxico, la transferencia de punteros en llamadas/retornos, adaptadores de memoria externa y native shadow metadata requieren implementación posterior.

`--max-memory` fija el presupuesto de bytes lógicos asignados, por defecto 64 MiB. `InterpreterOptions::memory_limits` también limita a 16384 asignaciones, 262144 fragmentos de metadatos de bytes de puntero y 256 Mi unidades de trabajo de bytes/metadatos. Superar un límite devuelve `MemoryLimit` con ubicación IR, separado de un trap del programa. La verificación rechaza además el overflow de tamaños de almacenamiento conocidos del destino antes de ejecutar.

```shell
whale ir run initialized.wir --function @f0
```

```text
i32 42
```

## Copias de bytes y reinicio de inicialización

Guarde el siguiente módulo completo como `tracked-memory.wir`. @f0 copia bytes del puntero y sus metadatos separados, y lee 42. En @f1, uninit marca el rango de almacenamiento del tipo como no inicializado y borra metadatos de punteros sin cambiar los bytes existentes. memcpy puede copiar bytes no inicializados; la lectura posterior del destino comprueba su estado. Copias parciales conservan autoridad solo al reunir los ocho fragmentos coherentes. Escribir bits idénticos con enteros o memset no recupera autoridad.

memcpy/memset de longitud cero tienen éxito sin acceso ni comprobación de alineación, incluso con null y one-past. memcpy no vacío con solapamiento provoca trap. memset inicializa los bytes escritos y borra sus metadatos de punteros. Bool debe almacenarse como 0 o 1; leer otra representación provoca trap.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "copied_pointer": whale () -> u32, linkage internal
  declare @f1 "uninitialized": whale () -> u32, linkage internal

  fn @f0 "copied_pointer"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    %v1: u32 = const u32 42
    store u32 %v1, ptr<u32> %v0, align 4
    %v2: ptr<ptr<u32>> = alloca ptr<u32>, align 8
    %v3: ptr<ptr<u32>> = alloca ptr<u32>, align 8
    store ptr<u32> %v0, ptr<ptr<u32>> %v2, align 8
    %v4: ptr<u8> = bitcast ptr<ptr<u32>> %v2 to ptr<u8>
    %v5: ptr<u8> = bitcast ptr<ptr<u32>> %v3 to ptr<u8>
    %v6: u64 = const u64 8
    memcpy ptr<u8> %v5, ptr<u8> %v4, u64 %v6, align 8
    %v7: ptr<u32> = load ptr<u32>, ptr<ptr<u32>> %v3, align 8
    %v8: u32 = load u32, ptr<u32> %v7, align 4
    ret u32 %v8
  }

  fn @f1 "uninitialized"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    uninit u32, ptr<u32> %v0, align 4
    %v1: u32 = load u32, ptr<u32> %v0, align 4
    ret u32 %v1
  }

}
```

```shell
whale ir run tracked-memory.wir --function @f0
whale ir run tracked-memory.wir --function @f1
```

```text
u32 42
Error: tracked-memory.wir: trap at @f1 %b0 instruction 2 (%v1): uninitialized byte at allocation offset 0 (after 3 steps)
```

## Obligaciones de verificación, lowering y ejecución

| Etapa | Obligación |
| --- | --- |
| Verifier | Comprobar operandos/tipos de puntero, forma de alineación y tamaños del destino; rechazar undef. |
| O0 lowering | Mantener uninit en la declaración ejecutada, almacenar inicializadores y conservar lecturas. |
| Runtime | Comprobar identidad, generación/vida, rango, permisos, alineación, overflow de dirección e inicialización; transferir el estado copiado. |
