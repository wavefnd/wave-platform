---
translation_set_id: storage-duration
path: language/storage-duration
locale: es
group: language
group_order: 2
order: 18
title: Duración del almacenamiento y mutabilidad.
summary: Distinga entre el alcance y la capacidad de escritura de var, const y static.
---

## Significado de cada declaración

|formato|Ubicación permitida|reasignación|uso|
| --- | --- | --- | --- |
| `var` |Función/bloque|posible|Variables locales mutables comunes|
| `const` |arriba|Imposible|Declaración constante global|
| `static` |arriba|posible|Declaraciones almacenadas estáticas que existen durante la vida del programa.|

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;

fun main() {
    var limit: i32 = 4;
    var current: i32 = 0;
    var retries: i32 = 0;

    current += 1;
    retries += 1;
    println("{} {} {}", limit, current, retries);
}
```

## Reglas de declaración local

```wave
var value: i32 = 1;
value = 2;
```

Se puede asignar un nuevo valor a una variable local declarada como `var`. Las constantes que se utilizarán en todo el programa se declaran en el nivel superior como `const`.

## Uso local de const y static

`const` y `static` son declaraciones de nivel superior. El cuerpo de la función y la inicialización `for` utilizan la declaración local `var`.

## Tiempo de vida y punteros

Puede obtener la dirección de una variable local como `&`, pero el tipo `ptr<T>` no rastrea la vida útil real del almacenamiento al que apunta el puntero. Al pasar una dirección de almacenamiento local fuera de una función, la estructura del programa debe garantizar directamente que la dirección siga siendo válida.

## Rango de aprendizaje y ejemplo

[Practica con el programa completo](/docs/es/getting-started/overview) · [Biblioteca estándar](/docs/es/stdlib)
