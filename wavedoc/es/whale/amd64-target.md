---
translation_set_id: whale-amd64-target
path: whale/amd64-target
locale: es
group: whale
group_order: 1
order: 10
title: Objetivos AMD64 y ABI
summary: Linux AMD64 Describe las propiedades de destino, los límites de llamada y el alcance de la función native.
---

## identificador de destino

El identificador de destino para el perfil native es `x86_64-whale-linux`.

|atributo|valor|
| --- | --- |
|sistema operativo| Linux |
|conjunto de instrucciones| AMD64 |
|orden de bytes| Little endian |
|Native Ancho de dirección|64 bits|
|formato de objeto| ELF64 |
|C Protocolo de llamada| SysV AMD64 ABI |
|formato de archivo ejecutable estático| ELF ET_EXEC |

El host de compilación y el destino de salida son conceptos diferentes. El hecho de que pueda ejecutar Whale en otro host no significa que pueda generar el conjunto de comandos o el formato de objeto de ese host. Consulte [Estado de soporte](overview) para conocer la ruta de compilación implementada.

## Selección y verificación de objetivos.

El comando experimental `ir lower` utiliza `x86_64-whale-linux` como destino predeterminado y único compatible. Para usar este comando, compílelo como `--features socket-cli`.

```sh
whale ir lower program.json --target x86_64-whale-linux
```

El destino de salida proporciona un diseño de datos little-endian de 64 bits independientemente del host de compilación. Los identificadores desconocidos o combinaciones no admitidas como `aarch64-whale-linux`, `x86_64-whale-windows` fallarán y generarán un error al guiar el destino admitido antes de leer la entrada o reemplazar el archivo de salida. `--no-verify` no deshabilita la verificación de selección de objetivos.

Al ingresar el espacio en blanco AST(`{"format_version":2,"semantics_version":1,"features":[],"program":{"declarations":[],"globals":[],"functions":[]}}`) se generarán los siguientes encabezados y módulos:

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

}
```

Si especifica un destino no compatible, terminará en un estado fallido.

```sh
whale ir lower program.json --target banana
```

```text
Error: unsupported target "banana"; supported targets: x86_64-whale-linux
```

En Rust, puede seleccionar el objetivo con `ir::Target::lookup("x86_64-whale-linux")` y transferir `name()` y `data_layout()` a `lower_o0`. Lowering rechaza si el diseño entregado es diferente del objetivo seleccionado. Compruebe si hay nombres de objetivos no admitidos e inconsistencias en el diseño en el IR y en el `verify_module` autoconfigurado.

El modelo de objetos almacena `format`, `machine`, `endian` y `address_bits` en `ObjectTarget`. `ObjectFile::with_target` conserva la información de identificación especificada y la serialización solo permite la combinación de AMD64·little-endian·64 bits ELF64. La presencia del identificador machine para una arquitectura diferente no implica compatibilidad con ese codificador. Ambos puntos de entrada ELF writer inspeccionan los metadatos y la entrada del vinculador se inspecciona antes de la resolución o vinculación del símbolo. El encabezado ELF para los objetos admitidos conserva `EM_X86_64`.

`ObjectFile::new(ObjectFormat::ELF64)` es un constructor conveniente que utiliza este identificador AMD64 como antes. Los códigos que accedieron anteriormente a `object.format` deben usar `object.target.format`.

Esta inspección proporciona selección de objetivos e identificación de objetos. El tamaño de los campos de estructuras, matrices, etc. offset, stride se puede buscar con el diseño IR API. La lectura de archivos de objetos, native ABI lowering, aún no se admite la vinculación a archivos ejecutables. El escalar lowering continúa especificando la alineación explícita y las reglas de diseño de tipos complejos completos se definen en [Documento de referencia de la memoria](memory-model).

## Llamar y firmar

El verificador de declaración y llamada IR acepta convenciones explícitas `Whale` y `SysV64`. La convención forma parte de un tipo `fnptr` y debe coincidir en el sitio de la llamada. Se rechazan las firmas variables y las firmas agregadas SysV64. Consulte el [ejemplo de construcción de llamada ejecutable](ir-reference). Esto valida los contratos IR; La clasificación ABI de la máquina, los registros de argumentos/ubicación de la pila y la emisión de llamadas nativas aún no están implementados.

Las llamadas admitidas requieren una firma explícita y una convención de llamada. Una firma no admitida es un error. El backend no debe realizar llamadas similares faltando argumentos o reemplazándolos con expresiones diferentes.

El soporte de firmas se divide en enteros/punteros básicos, f32/f64, estructuras/valores amplios y argumentos variables. El hecho de que haya un tipo en el sistema de tipos IR no significa que ABI admita pasar o devolver argumentos de ese tipo. Al elegir un backend, debes verificar si cada uno admite las categorías que necesitas.

Por ejemplo, las firmas de devolución de estructura i32 que tampoco sean compatibles con el backend que maneja la devolución deben rechazarse. La convención de retorno escalar no se puede aplicar simplemente porque parte de la estructura entra en un registro escalar.

## Llamadas internas y límites C

native La dirección del puntero es de 64 bits. shadow metadata del puntero de seguimiento deben pasarse juntos incluso después de copiar, guardar, argumentar y devolver.

C ABI y el protocolo de entrega de metadatos internos son distintos. C Las llamadas transfronterizas requieren un adaptador explícito. No se debe considerar que el hecho de pasar una dirección numérica a C ABI preserva la identidad de la asignación, la duración, el alcance o los derechos de acceso.

## marco de pila

Conserva el puntero del marco y no utiliza red zone. En O0, el espacio de almacenamiento de diferentes variables locales no se reutiliza. La información de depuración del marco de llamada describe la correspondencia del marco native con el programa original.

Esta regla es solo para fines de depuración y no pretende admitir la excepción unwinding. Consulte [Depurando con O0](o0-debugging).

## Restricciones de perfil

Los primeros perfiles O0 native no incluyen las siguientes características:

- O1 Optimización, vectorización, LTO.
- Memoria compartida y operaciones atomic.
- Excepción unwinding, función async, coroutine.
- GC y verificación de propiedad específica del idioma.
- Transferencia de propiedad de memoria externa arbitraria.
- Soporte completo de restricciones de ensamblaje en línea.
- Enlace dinámico, TLS, conjunto de comandos adicional, tipos de objetos adicionales.

Las solicitudes que estén fuera del alcance del soporte deben rechazarse en lugar de reemplazarse silenciosamente con otra característica. Esta restricción es para la ruta de compilación native y no significa eliminar otros componentes de la cadena de herramientas proporcionados de forma independiente.
