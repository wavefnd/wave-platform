---
translation_set_id: language-variants
path: language/variants
locale: de
group: language
group_order: 2
order: 17
title: Varianten und Mustervergleich
summary: Speichern Sie fallspezifische Daten zusammen und sicher getrennt von match.
---

## Unterscheidet sich von enum

enum stellt einen benannten ganzzahligen Wert dar und variant enthält payload, was jeweils unterschiedlich ist. Result unten speichert den Wert i32 im Erfolgsfall und die Fehlernummer i32 im Fehlerfall. Auch wenn Fehler und Ergebnis vom gleichen Ganzzahltyp sind, kann ihre Bedeutung anhand des Namens unterschieden werden.

## Deklaration, Erstellung, Inspektion

Speichern Sie es unter `main.wave` und führen Sie es aus.

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

Ausführungsergebnis:

```text
value=42
```

Macht den Fall, dass `Result::Value(42)` payload hat. Verwenden Sie payload unter dem Namen value nur innerhalb des entsprechenden Musters von `match`. In anderen Fällen wird das Lesen von payload nicht erzwungen. arm Der Text ist in Blöcken geschrieben. Behandeln Sie alle Fälle oder den Rest mit `_`. Es ist besser, jeden Fall explizit aufzuschlüsseln, um zu erkennen, ob eine Bearbeitung erforderlich ist, wenn ein neuer Fall hinzugefügt wird.

## Generika und Langlebigkeit

Typparameter können wie in `variant Optional<T> { Some(T), None }` verwendet werden. Geben Sie einen konkreten Typ für lokale Variablen an, z. B. `Optional<i32>`. Eine Variante, die einen Zeiger enthält, verwaltet den Speicherbesitz nicht automatisch. Durch das Kopieren des Werts wird die Zuordnung, auf die verwiesen wird, nicht dupliziert.

Gehen Sie nicht davon aus, dass die Speicherdarstellung von variant mit jedem beliebigen C union identisch ist. Für extern zu versendende Daten ABI wird ein eigener Ausdruck ermittelt und als zulässiger Typ FFI ausgeliefert.

## üben

Wenn Sie es in calculate (false) ändern, sollten Sie `error=1` erhalten. Versuchen Sie, die Groß-/Kleinschreibung Empty ohne payload hinzuzufügen, und behandeln Sie diese Groß-/Kleinschreibung dann auch in match.

[Struktur und enum](/docs/de/language/structures-enums-and-aliases) · [Fehlerbehandlungsklasse](/docs/de/language/errors)
