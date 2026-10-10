---
translation_set_id: language-variants
path: language/variants
locale: es
group: language
group_order: 2
order: 17
title: Variantes y combinación de patrones.
summary: Almacene los datos específicos del caso juntos y sepárelos de forma segura de match.
---

## Distinguido de enum

enum representa un valor entero con nombre y variant contiene payload, que es diferente en cada caso. Result a continuación almacena el valor i32 en caso de éxito y el número de error i32 en caso de error. Incluso si el error y el resultado son del mismo tipo entero, su significado se puede diferenciar por su nombre.

## Declaración, creación, inspección.

Guárdelo en `main.wave` y ejecútelo.

<!-- wave-example: variant-api -->
```wave playground
variant Result {
    Value(i32), Error(i32)
}

fun calculate(valid: bool) -> Result {
    if (!valid) {
        return Result::Error(1);
    }

    return Result::Value(42);
}

fun main() {
    var result: Result = calculate(true);
    match (result) {
        Result::Value(value) => {
            println("value={}", value);
        }
        Result::Error(code) => {
            println("error={}", code);
        }
    }
}
```

Resultado de la ejecución:

```text
value=42
```

Se argumenta que `Result::Value(42)` tiene payload. Utilice payload bajo el nombre value solo dentro del patrón correspondiente de `match`. No fuerza la lectura de payload en otros casos. arm El texto está escrito en bloques. Maneje todos los casos o el resto con `_`. Es mejor desglosar explícitamente cada caso para revelar si es necesario procesarlo cuando se agrega un nuevo caso.

## Genéricos y longevidad

Los parámetros de tipo se pueden utilizar como en `variant Optional<T> { Some(T), None }`. Especifique un tipo concreto para variables locales, como `Optional<i32>`. Una variante que contiene un puntero no gestiona automáticamente la propiedad de la memoria. Copiar el valor no duplica la asignación apuntada.

No asuma que la representación en memoria de variant es idéntica a cualquier C union arbitraria. Para que los datos se envíen externamente ABI, se determina una expresión separada y se entrega como el tipo FFI permitido.

## practica

Si lo cambia a calculate(false), debería obtener `error=1`. Intente agregar el caso Empty sin payload y luego maneje ese caso en match también.

[Estructura y enum](/docs/es/language/structures-enums-and-aliases) · [Clase de manejo de errores](/docs/es/language/errors)
