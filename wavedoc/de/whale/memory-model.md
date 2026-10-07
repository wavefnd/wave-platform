---
translation_set_id: whale-memory-model
path: whale/memory-model
locale: de
group: whale
group_order: 1
order: 8
title: Speichermodell
summary: Zuordnungsverfolgung, Lesen initialisierter Werte, Zeigerarithmetik, Layout und String-Speicherregeln.
---

## Zuordnungsverfolgung

Ein verfolgter Zeiger verknüpft eine Adresse mit einer Zuordnungs-ID, Generation, Grenzen, Offset und Zugriffsberechtigungen. Das Speichermodell verfolgt auch die Lebensdauer und den Initialisierungsstatus der Zuweisung. Der Zugang muss diese Bedingungen erfüllen; Ein Verstoß führt zu einer Falle.

Auch wenn dieselbe physische Adresse wiederverwendet wird, trennen die einzelnen Generationen jede Lebensdauer. Das bloße Vorhandensein einer Adresse bestimmt nicht, dass der Zeiger gültig ist oder dass der Aufrufer Zugriff auf diesen Speicherplatz hat.

Die Adresse native umfasst 64 Bit. Separate shadow metadata werden zusammen mit dem Zeiger durch Kopieren/Speichern/Aufrufen/Zurück übergeben. Der anfängliche Speicherbereich native besteht aus nachverfolgbaren Stapel- und globalen Zuordnungen. C Grenzen erfordern einen expliziten Adapter und die Übertragung des Eigentums an beliebigem externen Speicher ist in diesem Bereich nicht enthalten.

## Initialisierung und Lesen

Durch die Deklaration eines Speicherplatzes wird der Wert nicht initialisiert. Überprüft, ob der tatsächlich gelesene Bytebereich initialisiert wurde. Wenn auch nur ein Teil des Werts im nicht initialisierten Zustand gelesen wird, ist er trap. Das Leseergebnis wird nicht durch 0 oder einen nicht spezifizierten Wert ersetzt.

Die Initialisierungsprüfung des gelesenen Werts schließt das Byte padding aus. Wenn beispielsweise alle Felder einer Struktur initialisiert sind, sind die Lesewerte nicht einfach deshalb falsch, weil leere Bytes, die aus der Feldausrichtung resultieren, nicht initialisiert sind.

Die Speicherkopie trägt den Initialisierungsstatus zusammen mit den Bytes. Durch das Kopieren von nicht initialisiertem Speicherplatz wird dieser nicht in initialisierten Speicherplatz umgewandelt. Beim späteren Lesen von Werten vom Ziel werden die gleichen Prüfungen durchgeführt wie beim Lesen der Quelle.

Die bloße Tatsache, dass das physikalische Byte von BSS 0 ist, erlaubt keine Initialisierung der Variable IR.

### IR im initialisierten Skalarspeicherplatz

Das folgende Modul wurde als builder konfiguriert und hat den Prüfer bestanden. `store` vor dem Lesen des Werts, und alle drei Speicheranweisungen geben eine Zweierpotenzsortierung ungleich Null an.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "initialized_local": whale () -> i32, linkage internal

  fn @f0 "initialized_local"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: ptr<i32> = alloca i32, align 4
    %v1: i32 = const i32 42
    store i32 %v1, ptr<i32> %v0, align 4
    %v2: i32 = load i32, ptr<i32> %v0, align 4
    ret i32 %v2
  }

}
```

Speichern Sie dieses vollständige Modul als `initialized.wir`; es gibt `i32 42` aus. Ohne store löst load einen Trap wegen nicht initialisierten Speichers aus. Ausrichtung 3 ist ein Prüfungsfehler. Physische Nullbytes von alloca bedeuten keine Initialisierung.

## Vergleich mit Zeigerarithmetik

Adressberechnung overflow ist trap. Sie können einen one-past-Zeiger erstellen, der direkt auf den Zuordnungsbereich zeigt, aber Sie können über ihn nicht auf den Speicher zugreifen.

Zeigergleichheit verwendet Zuordnungsidentität, nicht nur eine numerische Adresse. Das Ordnen oder Subtrahieren von Zeigern aus verschiedenen Zuordnungen führt zu einer Falle. Durch die Rekonstruktion einer Adresse aus einer Ganzzahl werden die Zugriffsberechtigungen nicht wiederhergestellt.

### GEP

GEP berechnet Adressen auf Element- und Feldbasis. Dies ist kein Befehl zum Lesen eines Werts.

Der erste Index ist die Elementeinheit offset des Typs, auf den der Basiszeiger zeigt. Der nachfolgende Index wählt ein Array-Element oder ein Struktur-/Tupelfeld aus. Der Feldindex einer Struktur/eines Tupels ist die Feldreihenfolgenummer zum Zeitpunkt der Kompilierung, nicht das Byte offset. Der ausgewählte Typ bestimmt den resultierenden Zeigertyp.

Wenn der Basistyp `ptr<array<i32, 4>>` ist, wählt der Index `[0, 2]` das dritte i32-Element des Arrays aus, was zu `ptr<i32>` führt. Wenn Sie den ersten Index als 1 angeben, wird i32 um ein Element und nicht um ein Element im Array verschoben. Dieses Beispiel veranschaulicht die Indexsemantik und nicht die Textbefehlssyntax.

Das anfängliche native-Ziel lehnt Zeigerarithmetik für Elemente der Größe 0 ab. Durch die Berechnung der Adresse entfallen nicht die Lebensdauer-, Bereichs-, Initialisierungs- und Berechtigungsprüfungen, die für den nachfolgenden Zugriff erforderlich sind.

## Datenlayout

Das Ausgabeziel bestimmt die Größe·Ausrichtung·Feld offset·Array stride. Die Reihenfolge der Felder in Strukturen und Tupeln behält ihre Deklarationsreihenfolge bei. Es sollte nicht davon ausgegangen werden, dass das Layout des Hosts, auf dem der Compiler ausgeführt wird, das Layout des Ausgabeziels ist.

|Wert|Regeln speichern|
| --- | --- |
| Bool, signed i1, unsigned u1 |Mindestens 1 Byte|
|Leere Struktur/Tupel|Größe 0, Ausrichtung 1|
|Array|Ziellayout bestimmt Element stride|
|Struktur/Tupel|Behalten Sie die Deklarationsreihenfolge und die Ausrichtungsanforderungen des Ziels bei|

Die abgeschlossene Ausrichtung von IR ist eine Potenz von 2, nicht 0. Die automatische Ausrichtung muss vor der Generierung dieses IR bestimmt werden. Die Layouts packed·union·bitfield werden in diesem Profil nicht unterstützt und sollten abgelehnt werden.

### Anfrage zum Ausgabelayout

Rust API berechnet das Speicherlayout unabhängig vom Build-Host. Im folgenden Beispiel gibt es 7 Bytes padding vor dem Feld u64 und 6 Bytes am Ende padding.

```rust
use ir::{allocation_align, layout_of, Target, Type};

fn main() {
    let target = Target::X86_64WhaleLinux;
    let record = Type::Struct(vec![Type::U8, Type::U64, Type::U16]);
    let layout = layout_of(&record, target).unwrap();
    assert_eq!((layout.size, layout.align), (24, 8));
    assert_eq!(layout.field_offsets, [0, 8, 16]);

    let array = Type::Array(Box::new(record), 3);
    let layout = layout_of(&array, target).unwrap();
    assert_eq!((layout.size, layout.align), (72, 8));
    assert_eq!(layout.element_stride, Some(24));
    assert_eq!(allocation_align(&array, target).unwrap(), 16);
}
```

```text
struct{u8, u64, u16}: size 24, natural alignment 8
field 0: byte 0
field 1: byte 8
field 2: byte 16
array of 3: size 72, element stride 24
standalone array placement alignment: 16
```

Die von `layout_of` zurückgegebene Größe und der Schritt eines Arrays umfassen nachfolgendes Auffüllen. Strukturen und Tupel verwenden dieselben Regeln für die Feldreihenfolge. Bool, i1 und u1 belegen jeweils ein Byte. Leere Strukturen und Tupel haben Größe 0 und Ausrichtung 1; Ein Array mit der Länge Null behält die natürliche Ausrichtung seines Elements bei. `void` hat kein Speicherlayout, während `ptr<void>` 8 Bytes belegt.

Verwenden Sie eine natürliche Sortierung für Felder und Array-Elemente. `allocation_align` wendet die Regel SysV AMD64 an, die eine mindestens 16-Byte-Ausrichtung erfordert, auf unabhängige lokale/globale Arrays mit 16 Byte oder mehr. Die Ausrichtung von Array-Feldern oder dem Element stride wird dadurch nicht erhöht. AST lowering verwendet diese Stapelsortierungssuche für die lokale Speicherung.

Betragsmultiplikation, Feldaddition offset, overflow in der padding-Berechnung ergibt `LayoutError::Overflow`. `layout_of` hat ein Verschachtelungslimit für komplexe Typen von 128 Ebenen und `layout_of_with_limit` ermöglicht dem Aufrufer, das Limit anzugeben. Es wird kein Speicherplatz zugewiesen, der der Anzahl der Array-Elemente entspricht. `pointer_stride` lehnt pointee mit der Größe 0 in der native-Zeigerarithmetik ab, aber das Speicherlayout dieses Typs selbst ist gültig. Packed·union·bitfield verfügt über keinen unterstützten Typausdruck. Diese gespeicherte Layoutsuche implementiert keine komplexen Typaufrufkonventionen oder die Überprüfung von Laufzeitgrenzen.

## Zeichenfolge und C-Grenze

Eine Zeichenfolge ist eine unveränderliche Folge von Bytes mit einer bestimmten Länge. Die Standardkodierung ist UTF-8. Ermöglicht ein internes NUL und hängt die Endung NUL nicht automatisch an. O0 kombiniert String-Objekte mit demselben Inhalt nicht automatisch.

Daher beträgt die Länge der Bytefolge bestehend aus `A`, NUL und `B` 3. Kann nicht für explizite C-Stringkonvertierungen verwendet werden, die das interne NUL ablehnen. Das Frontend sollte dies nicht stillschweigend auf `A` kürzen.

Externe C·Rohadresse·Inline Assembly ist eine separate Vertragsgrenze. Bei der Laufzeitüberprüfung des Trace-Speichers kann nicht garantiert werden, dass alle fehlerhaften Verhaltensweisen im externen Code erkannt werden.

## Ausführung mit verfolgtem Stapelspeicher

Der Standardinterpreter führt Ganzzahlen/Bool, Kontrollfluss, Stapelzuweisungen, Datenzeigerspeicherung und -lesen, typed GEP, memcpy und memset aus. Beim Lesen von checked-Paaren wird Padding ausgeschlossen. Adressen sind synthetische 64-Bit-Werte ohne Zugriff auf Hostspeicher. Argumente und Rückgaben bleiben Ganzzahlen/Bool oder void; float, Aufrufe, Funktionszeiger, allgemeine Aggregatwerte, globale Adressen und native Ausführung fehlen. Stapelzuweisungen leben bis zur Rückgabe. Lexikalisches Lebensdauerende, Zeigerübergabe bei Aufrufen/Rückgaben, Fremdspeicheradapter und native shadow metadata müssen noch implementiert werden.

`--max-memory` begrenzt logische Zuweisungsbytes, standardmäßig 64 MiB. `InterpreterOptions::memory_limits` begrenzt außerdem Zuweisungen auf 16384, Zeigerbyte-Metadaten auf 262144 Fragmente und Byte-/Metadatenarbeit auf 256 Mi Einheiten. Überschreitungen liefern `MemoryLimit` mit IR-Position, getrennt von Programm-Traps. Die Prüfung lehnt auch bekannten Speichergrößen-overflow des Ausgabeziels vor Ausführung ab.

```shell
whale ir run initialized.wir --function @f0
```

```text
i32 42
```

## Bytekopien und Zurücksetzen der Initialisierung

Speichern Sie das vollständige folgende Modul als `tracked-memory.wir`. @f0 kopiert Zeigerbytes und getrennte Metadaten und liest 42. Bei @f1 markiert uninit den Speicherbereich des Typs als nicht initialisiert und löscht Zeigermetadaten, ohne bestehende Bytes zu ändern. memcpy darf nicht initialisierte Bytes kopieren; spätere Wertzugriffe prüfen den Zielzustand. Teilkopien erhalten Rechte erst, wenn alle acht konsistenten Metadatenfragmente vorliegen. Gleiche Bits durch Ganzzahlen oder memset stellen keine Rechte wieder her.

memcpy/memset mit Länge null gelingen ohne Zugriffs- oder Ausrichtungsprüfung, auch für null und one-past. Überlappende nichtleere memcpy löst einen Trap aus. memset initialisiert geschriebene Bytes und löscht ihre Zeigermetadaten. Bool muss als 0 oder 1 gespeichert sein; andere Darstellungen lösen beim Lesen einen Trap aus.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "copied_pointer": whale () -> u32, linkage internal
  declare @f1 "uninitialized": whale () -> u32, linkage internal

  fn @f0 "copied_pointer"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    %v1: u32 = const u32 42
    store u32 %v1, ptr<u32> %v0, align 4
    %v2: ptr<ptr<u32>> = alloca ptr<u32>, align 8
    %v3: ptr<ptr<u32>> = alloca ptr<u32>, align 8
    store ptr<u32> %v0, ptr<ptr<u32>> %v2, align 8
    %v4: ptr<u8> = bitcast ptr<ptr<u32>> %v2 to ptr<u8>
    %v5: ptr<u8> = bitcast ptr<ptr<u32>> %v3 to ptr<u8>
    %v6: u64 = const u64 8
    memcpy ptr<u8> %v5, ptr<u8> %v4, u64 %v6, align 8
    %v7: ptr<u32> = load ptr<u32>, ptr<ptr<u32>> %v3, align 8
    %v8: u32 = load u32, ptr<u32> %v7, align 4
    ret u32 %v8
  }

  fn @f1 "uninitialized"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    uninit u32, ptr<u32> %v0, align 4
    %v1: u32 = load u32, ptr<u32> %v0, align 4
    ret u32 %v1
  }

}
```

```shell
whale ir run tracked-memory.wir --function @f0
whale ir run tracked-memory.wir --function @f1
```

```text
u32 42
Error: tracked-memory.wir: trap at @f1 %b0 instruction 2 (%v1): uninitialized byte at allocation offset 0 (after 3 steps)
```

## Pflichten von Prüfung, lowering und Ausführung

| Stufe | Pflicht |
| --- | --- |
| Verifier | Operanden/Zeigertypen, Ausrichtungsform und Speichergrößen des Ausgabeziels prüfen; undef ablehnen. |
| O0 lowering | uninit an der ausgeführten Deklaration belassen, Anfangswerte speichern und Lesezugriffe erhalten. |
| Runtime | Zuweisungsidentität, Generation/Lebensdauer, Bereich, Rechte, Ausrichtung, Adress-overflow und Initialisierung prüfen; Kopierzustand übertragen. |
