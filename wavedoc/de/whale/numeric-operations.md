---
translation_set_id: whale-numeric-operations
path: whale/numeric-operations
locale: de
group: whale
group_order: 1
order: 7
title: Numerische Operationen
summary: Beschreibt Ganzzahloperationen wrap, checked, Verschiebung, Typkonvertierungsfehler und Gleitkommaergebnisse.
---

## ganzzahlige Darstellung

Eine N-Bit-Ganzzahl hat N-Wertbits. Vorzeichenlose Ganzzahlen reichen von 0 bis 2^N − 1 und vorzeichenbehaftete Ganzzahlen reichen von −2^(N−1) bis 2^(N−1) − 1. Das signedness der Operation bestimmt die Interpretation der Bitfolge.

Die folgende Tabelle erläutert die Berechnungsergebnisse. IR Das Codebeispiel verwendet die Darstellung des aktuellen Druckers.

## Addition, Subtraktion, Multiplikation

Die grundlegenden Ganzzahlen add·sub·mul enthalten die unteren N Bits des Ergebnisses. overflow verursacht nicht trap. Die Operation checked gibt das gleiche Ergebnis wrap sowie Bool zurück, das angibt, ob das mathematische Ergebnis außerhalb des Bereichs des entsprechenden signed oder unsigned liegt.

|Betrieb|Wrap Ergebnis| Checked overflow |
| --- | --- | --- |
| u8: 255 + 1 | 0 | true |
| i8: 127 + 1 | −128 | true |
| u8: 0 − 1 | 255 | true |
| i8: 12 × 3 | 36 | false |

Frontends für Sprachen, die die Ausführung bei overflow unterbrechen, müssen ein explizites `trap_if` für das overflow-Ergebnis der checked-Operation verwenden. Der Standardvorgang wendet nicht implizit die overflow-Richtlinie der Quellsprache an.

### Wrap und IR drücken eine explizite Prüfung aus

Speichern Sie dieses vollständige Format-3-Modul als `integer-operations.wir`; Textreader und skalarer Interpreter akzeptieren es.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "add_u8": whale () -> u8, linkage internal
  declare @f1 "require_no_overflow": whale () -> u8, linkage internal

  declare @f2 "minimum_division": whale () -> i8, linkage internal
  declare @f3 "minimum_remainder": whale () -> i8, linkage internal
  declare @f4 "negative_shift": whale () -> i8, linkage internal

  fn @f0 "add_u8"() -> u8, entry %b0 {
  %b0 "entry":
    %v0: u8 = const u8 255
    %v1: u8 = const u8 1
    %v2: u8 = add u8 %v0, %v1
    ret u8 %v2
  }

  fn @f1 "require_no_overflow"() -> u8, entry %b1 {
  %b1 "entry":
    %v3: u8 = const u8 255
    %v4: u8 = const u8 1
    %v5: tuple<u8, bool> = uadd_chk u8 %v3, %v4
    %v6: u8 = extract %v5, 0
    %v7: bool = extract %v5, 1
    trap_if bool %v7, reason="integer overflow"
    ret u8 %v6
  }

  fn @f2 "minimum_division"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -128
    %v1: i8 = const i8 -1
    %v2: i8 = sdiv i8 %v0, %v1
    ret i8 %v2
  }

  fn @f3 "minimum_remainder"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -128
    %v1: i8 = const i8 -1
    %v2: i8 = srem i8 %v0, %v1
    ret i8 %v2
  }

  fn @f4 "negative_shift"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -1
    %v1: i8 = const i8 -1
    %v2: i8 = shl i8 %v0, %v1
    ret i8 %v2
  }

}
```

`add_u8` liefert 0, die unteren acht Bits von 256. `minimum_division` liefert −128 und `minimum_remainder` 0. `negative_shift` interpretiert −1 als unsigned count 255 und verwendet 255 mod 8 = 7. Führen Sie die Funktionen über explizite IDs aus:

```shell
whale ir run integer-operations.wir --function @f0
whale ir run integer-operations.wir --function @f2
whale ir run integer-operations.wir --function @f3
whale ir run integer-operations.wir --function @f4
```

```text
u8 0
i8 -128
i8 0
i8 -128
```

`require_no_overflow` extrahiert das umgebrochene Ergebnis an Index 0 und das Bool-overflow-Flag an Index 1. Das explizite trap_if stoppt vor der Rückgabe. Der folgende Befehl endet mit Status 1 und schreibt die Diagnose nach stderr:

```shell
whale ir run integer-operations.wir --function @f1
```

```text
Error: integer-operations.wir: trap at @f1 %b1 instruction 5: "integer overflow" (after 6 steps)
```

## Division und Rest

Eine ganzzahlige Division oder ein Rest durch Null führt zu einer Falle. Durch Teilen des minimalen vorzeichenbehafteten Werts durch −1 wird auf den minimalen Wert umgebrochen. Der Rest ist in diesem Fall Null.

|Betrieb|Ergebnis|
| --- | --- |
| i8: −128 / −1 | −128 |
| i8: −128 % −1 | 0 |
|Durch die ganze Zahl 0 dividieren| trap |
|Rest für Ganzzahl 0| trap |

## Verschiebung

Beim Verschieben des Bitwerts N wird die Bitfolge von count als unsigned interpretiert und der durch N geteilte Rest verwendet. Es generiert kein trap, nur weil count außerhalb des Bereichs von 0 bis N−1 liegt.

Bei 8-Bit-Werten sind count 0·8·16 alle 0-Bit-Verschiebungen. Die Bitfolge `11111111` von 8 Bit count wird um 7 Bit verschoben. Dies gilt auch dann, wenn diese Bitfolge signed −1 darstellt. unsigned Die Analyse erfolgt vor den restlichen Berechnungen.

Ausgangssprachen, die negative oder übermäßige count ablehnen, müssen vor der Verschiebung ein explizites Häkchen setzen.

## Typkonvertierung

|Konvertierung|Bedeutung|
| --- | --- |
| Zero extension |Erhöhen Sie die Breite, indem Sie höherwertige Bits mit 0 füllen|
| Sign extension |Erhöhen Sie die Breite, indem Sie das Vorzeichenbit duplizieren|
|Bitschneiden|Behalten Sie nur die niederwertigen Bits bei, die der Zielbreite entsprechen|
|Beat-Neuinterpretation|Interpretieren derselben Bitfolge als unterschiedliche Typen|
|Verlustfreie numerische Konvertierung|Wenn es nicht unter Beibehaltung des numerischen Werts ausgedrückt werden kann, trap|

Wenn Sie beispielsweise 8-Bit `11111111` in 16-Bit zero umwandeln, wird extension zu `0000000011111111` und sign extension wird zu `1111111111111111`. Auch wenn die Eingabebits gleich sind, handelt es sich um unterschiedliche Operationen.

Die Konvertierung von Float zu Int schneidet in Richtung Null ab und überprüft dann den Ganzzahlbereich. NaN und Unendlichkeit verursachen eine Falle. Bei der Konvertierung in i8 wird 127,9 zu 127, während 128,0 Traps sind. Bool wird in die Ganzzahl 0 oder 1 konvertiert. i1 mit Vorzeichen kann nicht 1 darstellen und kann daher nicht das Ziel dieser Konvertierung sein.

Durch das Konvertieren einer Adresse in eine Ganzzahl wird der Zeigerzugriff auf diese Ganzzahl nicht wiederhergestellt. Bitte beachten Sie [Zeigergültigkeit](memory-model).

### Geprüfte Formen von Casts und checked-Operationen

Die Prüfung vergleicht den tatsächlichen Operandentyp mit `src_ty` und kontrolliert anschließend die erlaubten Kategorien und Breiten des Opcodes. Auch die Ergebnistypangabe muss zur Definition passen. Es gibt keine impliziten Konvertierungen.

| Opcode | Erlaubte Typen |
| --- | --- |
| `zext`, `sext` | Ganzzahlen außer Bool; Zielbreite strikt größer |
| `zext` (Bool) | Bool zu Ganzzahl 0/1; alle Ganzzahlziele außer vorzeichenbehaftetem `i1`, einschließlich `u1` |
| `trunc` | Ganzzahlen außer Bool; Zielbreite strikt kleiner |
| `fext`, `ftrunc` | Gleitkommawerte; Zielbreite strikt größer / kleiner |
| `itof_s`, `itof_u` | Vorzeichenbehaftete / vorzeichenlose Ganzzahl zu Gleitkomma |
| `ftoi_s`, `ftoi_u` | Gleitkomma zu vorzeichenbehafteter / vorzeichenloser Ganzzahl |
| `bitcast` | Gleich breite Ganzzahl-/Gleitkommaskalare oder Datenzeiger zu Datenzeiger |
| `ptrtoint`, `inttoptr` | Datenzeiger zu Ganzzahl / Ganzzahl zu Datenzeiger |

Ganzzahlerweiterung und -kürzung sind Bitoperationen; die Ganzzahloperanden dürfen unterschiedliche Vorzeicheninterpretationen haben. Bool ist ein eigener logischer Typ: Vorzeichenerweiterung, Kürzung und bitcast zu `i1`/`u1` sind verboten. Casts von Aggregaten und Funktionszeigern werden abgelehnt. Die Prüfung der Zeiger-/Ganzzahlkategorien begründet weder eine gültige Allokation noch Zugriffsrechte. Laufzeitprüfungen für Konvertierungen und das Maschinen-Lowering bleiben eigenständige offene Arbeiten.

Dieses vollständige Rust-Programm gibt zwei gültige Konvertierungen aus und bestätigt, dass das Ersetzen von `zext` durch `fext` abgelehnt wird.

```rust
use ir::*;

fn main() {
    let mut builder = ModuleBuilder::new("x86_64-whale-linux", DataLayout::default_64bit_le());
    let mut function = builder.begin_function("widen", vec![("byte".into(), Type::U8)], Type::Void);
    function.ret(None);
    function.finish();
    let mut module = builder.finish();
    let function = &mut module.functions[0];
    function.value_types.extend([(ValueId(1), Type::U32), (ValueId(2), Type::I32)]);
    function.blocks[0].instructions.extend([
        Instruction::Cast {
            dst: ValueId(1), op: CastOp::ZExt,
            src_ty: Type::U8, src: ValueId(0), dst_ty: Type::U32,
        },
        Instruction::Cast {
            dst: ValueId(2), op: CastOp::Bitcast,
            src_ty: Type::U32, src: ValueId(1), dst_ty: Type::I32,
        },
    ]);
    verify_module(&module).unwrap();
    print!("{}", print_module(&module));
    // An integer source cannot be annotated as a floating widening operation.
    if let Instruction::Cast { op, .. } = &mut module.functions[0].blocks[0].instructions[0] {
        *op = CastOp::FExt;
    }
    assert!(matches!(verify_module(&module), Err(VerifyError::InvalidCast { .. })));
}
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "widen": whale (u8) -> void, linkage internal

  fn @f0 "widen"(%v0 "byte": u8) -> void, entry %b0 {
  %b0 "entry":
    %v1: u32 = zext u8 %v0 to u32
    %v2: i32 = bitcast u32 %v1 to i32
    ret void
  }

}
```

Die folgenden Anweisungsfragmente sind absichtlich ungültige IR.

```text
%v1: i32 = ftoi_s f64 %v0 to i32   // Invalid when %v0 is actually Bool.
%v2: i16 = zext i32 %v3 to i16     // zext cannot narrow.
%v4: i1 = zext bool %v5 to i1      // Signed i1 cannot represent true as 1.
```

Ein abweichender tatsächlicher Quelltyp führt zu `OperandTypeMismatch`; ein unerlaubtes Typenpaar zu `InvalidCast` mit Opcode und beiden Typen. Kein Prüfungsfehler fügt Konvertierungen ein oder verändert das Modul.

`sadd_chk`, `ssub_chk` und `smul_chk` verlangen vorzeichenbehaftete Ganzzahlen; `uadd_chk`, `usub_chk` und `umul_chk` vorzeichenlose. Beide Operanden müssen exakt zum Operationstyp `T` passen; das Ergebnis muss `tuple<T, bool>` sein. Extraktion erfordert ein Tupel, einen vorhandenen Feldindex und dessen exakten Typ. Insbesondere wird die Extraktion des Überlauffelds als `i1` abgelehnt. Die obigen Wrap-/Überlaufergebnisse bleiben der Ausführungsvertrag; diese Tests prüfen die Struktur, nicht die Ausführung durch ein implementiertes Backend.

## Gleitkomma-Arithmetik

Gleitkommawerte haben die exakte Bitfolge f16·f32·f64. Die Operation rundet auf den nächsten Wert in der angegebenen Breite und verwendet, wenn er genau in der Mitte liegt, ties-to-even, das einen Wert mit geraden niedrigstwertigen Bits der signifikanten Ziffern auswählt.

Die Standardoperation ist fast-math, implizit FMA, was keine erzwungene Nullbehandlung kleiner Werte zulässt. Bei der Multiplikation mit anschließender Addition bleiben die jeweiligen Rundungsschritte erhalten und das Backend sollte sie nicht implizit in einer Operation kombinieren.

Das Ergebnis einer numerischen Operation kann NaN oder unendlich sein. NaNs aus mathematischen Operationen und Breitenänderungen werden auf eine feste Menge stiller NaN pro Breite normalisiert. Durch Speichern und Kopieren bleiben die ursprünglichen NaN-Bits erhalten. Daher ist das Verhalten unterschiedlich, wenn NaN-Nutzdaten ohne Rechenoperationen an den Speicher übergeben werden und wenn sie berechnet werden.

Es werden keine Gleitkomma-Statusflags verfügbar gemacht. Obwohl die grundlegende Gleitkomma-Arithmetik NaN·unendliche Ergebnisse ermöglicht, wendet die float→int-Konvertierung die oben genannten trap-Regeln an.


### Speichern Sie genaue Konstanten

Verwenden Sie `FloatBits` variant oder eine beliebige Folge hexadezimaler Bits mit der exakten Breite. Die Gleichheit der gespeicherten Werte wird mit einer Bitfolge verglichen, die negative 0 und NaN payload enthält. Der Prüfer lehnt ab, wenn die Breiten der Typen IR und payload unterschiedlich sind.

```rust
use ir::{FloatBits, ModuleBuilder, Target, Type};
fn main() {
    let bits = FloatBits::parse(32, "0xffc01234").unwrap();
    assert_eq!(bits, FloatBits::F32(0xffc01234));
    let target = Target::X86_64WhaleLinux;
    let mut module = ModuleBuilder::new(target.name(), target.data_layout());
    let mut function = module.begin_function("payload", vec![], Type::F32);
    let value = function.const_float_bits(Type::F32, bits);
    function.ret(Some(value));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    assert!(ir::print_module(&module).contains("const f32 0xffc01234"));
    println!("{}", bits);
}
```

```text
0xffc01234
```

Eine f16-, f32- oder f64-Bitfolge beginnt mit `0x`, gefolgt von genau 4, 8 bzw. 16 Hexadezimalziffern. `0x80000000` steht für f32 negative Null; `0x7f800000` steht für positive Unendlichkeit. `const_float` konvertiert einen Host-f64-Wert numerisch; Verwenden Sie `const_float_bits`, um die ursprünglichen Bits beizubehalten. Eine exakte Speicherdarstellung impliziert kein vollständiges Gleitkomma-Ausführungs-Backend. Die Arithmetik zur Kompilierungszeit verwendet immer noch Zwischenprodukte des Hosts f64, sodass der vollständige Rundungsvertrag für jede deklarierte Breite noch nicht implementiert ist.
