---
translation_set_id: whale-o0-debugging
path: whale/o0-debugging
locale: de
group: whale
group_order: 1
order: 9
title: O0 und Debugging
summary: Regeln zum Beibehalten von Berechnungen, konstanten Ausdrücken, Speicherplatz und Debug-Antwortinformationen.
---

## Naturschutzmodell

O0 behält die Berechnungen, Variablen und den Kontrollfluss des Originals typed IR für Debugging-Zwecke bei. Außerdem bleiben ungenutzte Ergebnisse und strukturell nicht erreichbare Blöcke erhalten. Die Überprüfung diagnostiziert fehlerhafte IR und löscht keine Blöcke oder vereinfacht Vorgänge.

Die zur Generierung der Maschinensprache erforderlichen Konvertierungen werden in einem separaten Sub-Sub IR durchgeführt. Auch nach der Umstellung muss die Übereinstimmung mit dem Original ID gewahrt bleiben. Die Konservierung von typed IR bedeutet nicht, dass alle IR Vorgänge und Maschinenanweisungen eins zu eins übereinstimmen.

## Konvertierung nicht durchgeführt von O0

|Konvertierung|O0 Betrieb|
| --- | --- |
|Inlining|Behalten Sie Aufruf- und Funktionsgrenzen bei|
|Konvertieren Tail-call|Behalten Sie die allgemeine Anruf-/Rückgabestruktur bei|
|Entfernen Sie toten Code|Behalten Sie ungenutzte Berechnungen und nicht erreichbare Blöcke bei|
|Laufzeitkonstanten ausblenden|Behalten Sie den ursprünglichen Betrieb bei|
|Kombinieren Sie gängige Berechnungen|Führen Sie getrennte Berechnungen durch|
|Wiederverwendung des lokalen Variablenspeicherplatzes|Behalten Sie jeden Speicherplatz bei|
|Rahmenzeiger weglassen|Rahmenzeiger beibehalten|
|Zeichenfolgen automatisch zusammenführen|Pflegen Sie separate String-Objekte|

Beispielsweise bleibt eine Laufzeitoperation, die zwei Konstanten hinzufügt, auch dann addiert, wenn das Ergebnis nicht verwendet wird. Zweige mit konstanten Bedingungen behalten außerdem die ursprüngliche Kontrollflussstruktur bei.

## Konstanten zur Kompilierungszeit

Deklarationen von Konstanten zur Kompilierungszeit bewahren sowohl den typed-Initialisierungsausdruck als auch das Auswertungsergebnis. Dies unterscheidet sich vom ständigen Falten in regulären Laufzeitanweisungen.

Eine Deklaration mit einem Initialisierungsausdruck von `1 + 2` bleibt mit einem Additionsausdruck und einem Ergebnis von `3` zurück. Dieses Beispiel veranschaulicht die Bedeutung des Ausdrucks und nicht die deklarative Grammatik der spezifischen Quellsprache. Das Ergebnis kann beim Aufbau statischer Daten verwendet werden und ersetzt den Initialisierungsausdruck nicht durch einen Laufzeitzusatz.

Auch ungenutzte und nicht erreichbare Deklarationen müssen identifiziert werden. Auch wenn Deklarationen mit demselben Namen einander verdecken, müssen sie durch Namen·Deklaration ID·Referenz getrennt werden. Die Validierung lehnt gespeicherte Ergebnisse ab, die ungültige Referenzen, Abhängigkeitszyklen oder ungültige Typen aufweisen oder nicht mit dem ursprünglichen Ausdruck übereinstimmen.

## nicht erreichbare Quellanweisung

Auch die Sätze nach return·break·continue bleiben in einem unzusammenhängenden Block. Aufgrund dieser Anweisung sollten Sie den vorangehenden terminator nicht ändern oder einen neuen Ausführungspfad erstellen. Außerdem werden ungültige Ausdrücke in nicht erreichbaren Sätzen diagnostiziert.

Diese Erhaltungsregel ermöglicht die Überprüfung der ursprünglichen Programmstruktur. Dies bedeutet nicht, dass die Anweisung nach terminator tatsächlich ausgeführt wird.

## Erhaltenes Beispiel IR

Nachfolgend finden Sie die aktuelle Druckerausgabe für das Modul, das die Prüfung bestanden hat. Zeigt nicht verwendete Berechnungen, Deklarationen zur Kompilierungszeit und nicht verknüpfte Blöcke zusammen.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "preserved": whale () -> void, linkage internal

  fn @f0 "preserved"() -> void, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 1
    %v1: i32 = const i32 2
    %v2: i32 = add i32 %v0, %v1
    %v3: i32 = const_decl "count" add(i32 1, i32 2) => const i32 3
    ret void
  %b1 "unreachable.cont":
    %v4: i32 = const i32 4
    %v5: i32 = const i32 5
    %v6: i32 = add i32 %v4, %v5
    ret void
  }

}
```

`%v2` bleibt als `add`, auch wenn es keinen Nutzen hat. `%v3` ist ein separates `const_decl`, das den Ausdruck `add(i32 1, i32 2)` und das Bewertungsergebnis 3 zusammen behält. Da `unreachable.cont` keine eingehende Kante hat, können wir das Hinzufügen von `%v6` überprüfen, ohne nach `ret void` einen Ausführungspfad hinzuzufügen.

Bei der Überprüfung bleiben diese Anweisungen und Blöcke erhalten. Dieses Beispiel zeigt die Konfiguration/Verifizierung/Ausgabe von IR und impliziert nicht, dass eine Ausführung von native oder eine Ausgabe von DWARF bereitgestellt wird.

## Quell- und Stack-Informationen

Die Debug-Schnittstelle verwendet Funktionsinformationen, Quellzeilen, lokale Standardvariablen und Aufrufrahmeninformationen von DWARF 5. Auch bei der Konvertierung in einen Unterausdruck muss die Entsprechung zwischen Funktions-/lokalen Variableninformationen und der ursprünglichen IR-Kennung erhalten bleiben.

Das Profil AMD64 behält den Frame-Zeiger bei und verwendet nicht red zone. Call-Frame-Informationen werden zur Stack-Inspektion verwendet und bedeuten keine Unterstützung für die Ausnahme unwinding. trap beendet die Ausführung, ohne einen Destruktor zu garantieren oder unwinding.

Informationen zur Verfügbarkeit der Ausgabe DWARF und der Ausführung native finden Sie unter [Übersicht über die Toolchain](overview). Das Verhalten von O1 und höheren Optimierungen liegt außerhalb des Rahmens dieser O0-Referenz.

## Wiederholte Deklarationen in Schleifen

Speichern Sie dieses vollständige Modul als `initialization-loop.wir`. Trotz Speicherung von 42 im ersten Durchlauf setzt uninit der zweiten Deklaration die Initialisierung zurück; der Lesezugriff löst einen Trap aus. O0 belässt uninit an der ausgeführten Deklaration, auch wenn alloca im Eintrittsblock steht. Ungenutzte ausgeführte Lesezugriffe werden geprüft; erhaltene unerreichbare Blöcke werden nicht ausgeführt.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "redeclaration": whale () -> u32, linkage internal

  fn @f0 "redeclaration"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    %v1: u32 = const u32 0
    %v2: u32 = const u32 1
    %v3: u32 = const u32 42
    br label %b1
  %b1 "declaration":
    %v4: u32 = phi u32 [ %v1, %b0 ], [ %v6, %b2 ]
    uninit u32, ptr<u32> %v0, align 4
    %v5: bool = icmp eq u32 %v4, %v1
    cbr bool %v5, label %b2, label %b3
  %b2 "first_iteration":
    store u32 %v3, ptr<u32> %v0, align 4
    %v6: u32 = add u32 %v4, %v2
    br label %b1
  %b3 "second_iteration":
    %v7: u32 = load u32, ptr<u32> %v0, align 4
    ret u32 %v7
  }

}
```

```shell
whale ir run initialization-loop.wir --function @f0
```

```text
Error: initialization-loop.wir: trap at @f0 %b3 instruction 0 (%v7): uninitialized byte at allocation offset 0 (after 17 steps)
```
