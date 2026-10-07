---
translation_set_id: whale-overview
path: whale/overview
locale: es
group: whale
group_order: 1
order: 1
title: Documentación de Whale
summary: Guía de componentes de la cadena de herramientas, comandos disponibles y documentación de referencia.
---

## Introducción

Whale es una cadena de herramientas de compilación de propósito general para implementaciones de lenguajes de programación y herramientas de compilación. Proporciona expresiones intermedias escritas (IR), ensamblador AMD64, biblioteca de archivos de objetos y funciones basadas en enlazadores. Cada componente se utiliza a través de la biblioteca Rust y el comando `whale`.

IR expresa la semántica de un cálculo independientemente de la codificación del lenguaje de máquina. El ensamblador codifica instrucciones de máquina para crear objetos reubicables. Las bibliotecas de objetos representan secciones, símbolos y reordenamientos. El vinculador resuelve referencias entre objetos y coloca el archivo ejecutable.

## Construir y usar herramientas

|trabajo|documento|
| --- | --- |
|Qué hace cada herramienta| [Componentes de la cadena de herramientas](/docs/es/whale/ecosystem) |
|Wave Creación de programa/enlace/selección de destino| [Construir y vincular](/docs/es/whale/build-link-targets) |
|Gestión de paquetes/dependencias| [Vex](/docs/es/whale/vex-package-manager) |
|Whale Ejecutar comando| [Whale CLI](/docs/es/whale/whale-cli) |

## Guía de documentos

|documento de referencia|contenido|
| --- | --- |
| [Ver IR](ir-reference) |Tipos, valores, funciones, flujo de control, validación, formatos de intercambio.|
| [Operaciones numéricas](numeric-operations) |Aritmética de enteros, desplazamiento, conversión de tipos, punto flotante|
| [Modelo de memoria](memory-model) |Inicialización, validación de puntero, cálculo de direcciones, diseño, cadenas.|
| [Depurando con O0](o0-debugging) |Retención de cálculo, almacenamiento de variables, información de depuración.|
| [AMD64 Objetivo](amd64-target) |Identificador de destino, convención de llamada, alcance de función native|
| [Ensamblador y enlazador](assembler-linker) |Operandos de ensamblaje, secciones, símbolos, enlaces estáticos.|

El documento de referencia define las reglas semánticas para Whale. Las funciones disponibles siguen la siguiente tabla. Es posible que las operaciones descritas en la documentación de referencia no estén disponibles en todas las versiones.

## Estado de soporte

|componente|Interfaz proporcionada|límite|
| --- | --- | --- |
|ensamblador|Cree ELF64 objeto reubicable a partir del ensamblaje AMD64|Cobertura incompleta de comandos y directivas.|
|biblioteca de objetos|Composición de objetos, payload sin BSS, validación de objetivos, AMD64 ELF64 serialización e implementación de registro Wave seleccionable para verificar el tamaño.|Los objetos reubicables no son ejecutables.|
| IR | Construcción, impresión, llamadas directas/indirectas con verificación de firma, reducción de versiones AST, validación de objetivos y diseños tipográficos verificados | AST format 2 / typed IR format 4, constantes de bits exactos y lectura, verificación e ida y vuelta de texto disponibles; la emisión de llamadas máquina sigue pendiente |
|enlazador|Interpretación de símbolos, verificación de objetivos de entrada, inspección de ubicación de sección de archivo/memoria|No se admite la aplicación de reubicación completa ni la generación de archivos ejecutables.|
| Ejecución y depuración | Enteros/Bool, control de flujo, pila rastreada y comprobación de inicialización | Float, llamadas, direcciones globales, generación native y DWARF no soportados |

Las reglas que prohíben el comportamiento indefinido se aplican a las memorias de rastreo y IR verificadas. La implementación experimental aún no implementa todas las comprobaciones de tiempo de ejecución en la documentación de referencia de memoria/ejecución.

## Construir y ensamblar

Compile con Rust 1.86.0 o superior.

```sh
git clone https://github.com/wavefnd/Whale.git
cd Whale
cargo build --release --locked
```

Guarde el siguiente ensamblaje como `answer.asm`.

```asm
section .text
global answer

answer:
    mov eax, 42
    ret
```

Crea un objeto ELF64.

```sh
./target/release/whale asm --amd64 answer.asm -o answer.o
```

Utiliza el propio ensamblador de Whale, por lo que no se requiere ningún ensamblador externo. La salida contiene una función invocable y no contiene ningún código de inicio de proceso.

Los comandos experimentales AST→IR están disponibles compilando con `--features socket-cli`. Para conocer el comando de CLI, consulte [Referencia de comando](/docs/es/whale/whale-cli).

## Uso y diagnóstico de la biblioteca.

IR debe verificarse antes de la ejecución o generación de código. Los errores de entrada y el mal uso de builder devolverán errores estructurales. Los errores de entrada en la biblioteca no deben finalizar el proceso del host ni sobrescribir el contenido existente. Las herramientas que aceptan entradas que no son de confianza deberían poder ajustar los límites de recursos.

La misma versión, entrada, destino y configuración de la cadena de herramientas deberían producir resultados deterministas. Los metadatos de la distribución identifican la versión·commit·y las funciones disponibles. CI verifica si hay buena entrada, entrada para rechazar, trap, O0 retención, round-trip, native semántica de ejecución para cada interfaz. Para conocer los comandos de desarrollo/verificación, consulte [Whale Información de contribución](https://github.com/wavefnd/Whale/blob/master/CONTRIBUTING.md).
