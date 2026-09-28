---
translation_set_id: storage-duration
path: language/storage-duration
locale: de
group: language
group_order: 2
order: 18
title: Speicherdauer und Veränderlichkeit
summary: Unterscheiden Sie zwischen dem Umfang und der Beschreibbarkeit von var, const und static.
---

## Bedeutung jeder Erklärung

|formatieren|Zulässiger Standort|erneut einreichen|Benutzen|
| --- | --- | --- | --- |
| `var` |Funktion/Block|möglich|Gemeinsame veränderbare lokale Variablen|
| `const` |oben|Unmöglich|Globale Konstantendeklaration|
| `static` |oben|möglich|Statisch gespeicherte Deklarationen, die für die gesamte Lebensdauer des Programms bestehen|

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

## Regeln für lokale Deklarationen

```wave
var value: i32 = 1;
value = 2;
```

Einer lokalen Variablen, die als `var` deklariert ist, kann ein neuer Wert zugewiesen werden. Konstanten, die im gesamten Programm verwendet werden sollen, werden auf der obersten Ebene als `const` deklariert.

## Lokale Verwendung von const und static

`const` und `static` sind Deklarationen der obersten Ebene. Der Funktionskörper und die `for`-Initialisierung verwenden die lokale Deklaration `var`.

## Lebensdauer und Zeiger

Sie können die Adresse einer lokalen Variablen als `&` erhalten, aber der Typ `ptr<T>` verfolgt nicht die tatsächliche Lebensdauer des Speichers, auf den der Zeiger zeigt. Bei der Übergabe einer lokalen Speicheradresse aus einer Funktion muss die Programmstruktur direkt sicherstellen, dass die Adresse gültig bleibt.

## Lern- und Beispielbereich

[Üben Sie mit dem vollständigen Programm](/docs/de/getting-started/overview) · [Standardbibliothek](/docs/de/stdlib)
