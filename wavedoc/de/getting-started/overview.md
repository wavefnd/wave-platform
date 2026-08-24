---
translation_set_id: overview
path: getting-started/overview
locale: de
group: getting-started
group_order: 1
order: 1
title: Überblick über Wave
summary: Eine kurze Einführung in Programmstruktur, Variablen und den Wave-Compiler.
---

## Über Wave

Wave ist eine statisch typisierte Systemprogrammiersprache für native Codeerzeugung und explizite Kontrolle auf niedriger Ebene.

## Erstes Programm

```wave
fun main() {
    println("Hello, Wave!");
}
```

Speichere die Datei als `main.wave` und führe sie so aus:

```shell
wavec run main.wave
```

## Variablendeklarationen

Lokale Variablen werden mit `var` deklariert.

```wave
var count: i32 = 1;
count += 1;
```

Die früheren Formen `let` und `let mut` wurden entfernt und sind Syntaxfehler. Für Deklarationen auf oberster Ebene werden `const` und `static` verwendet.

## Übersetzungsstatus

Noch nicht übersetzte Seiten werden eindeutig als englischer Originaltext angezeigt.
