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

Las constantes y el almacenamiento estático de nivel superior se declaran con `const` y `static`.

## Estado de la traducción

Las páginas todavía no traducidas al español se muestran claramente como contenido original en inglés.
