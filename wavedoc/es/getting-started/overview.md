---
translation_set_id: overview
path: getting-started/overview
locale: es
group: getting-started
group_order: 1
order: 1
title: Introducción al lenguaje Wave
summary: Una introducción breve a la estructura, las variables y el compilador de Wave.
---

## Acerca de Wave

Wave es un lenguaje de programación de sistemas con tipado estático, generación de código nativo y control explícito de bajo nivel.

## Primer programa

```wave
fun main() {
    println("Hello, Wave!");
}
```

Guarda el archivo como `main.wave` y ejecútalo con:

```shell
wavec run main.wave
```

## Declaración de variables

Las variables locales se declaran con `var`.

```wave
var count: i32 = 1;
count += 1;
```

Las formas anteriores `let` y `let mut` fueron eliminadas y ahora producen un error de sintaxis. Usa `const` y `static` para declaraciones de nivel superior.

## Estado de la traducción

Las páginas todavía no traducidas al español se muestran claramente como contenido original en inglés.
