---
translation_set_id: whale-ir-reference
path: whale/ir-reference
locale: de
group: whale
group_order: 1
order: 6
title: Whale-IR-Referenz
summary: Beschreibt Typen, Bezeichner, Funktionsgültigkeit, Auswertungsreihenfolge und Austauschformat.
---

## Module und Bezeichner

Ein Modul besteht aus Zielinformationen, globalen Definitionen und Funktionen. Werte haben explizite Typen. Das Frontend löst den Namen, den Typ, die Überladung und die Generika der Quellsprache auf und generiert typed IR.

Funktionen und globale Variablen verwenden unterschiedliche interne Namensräume. Daher können Funktionen und Variablen denselben Namen haben. Der interne Bezeichner unterscheidet sich vom externen Verbindungsnamen `link_name` und der externe Name wird vom Frontend angegeben. Whale löst externe Konflikte nicht durch die automatische Generierung neuer Namen. Bitte beachten Sie [Symbole und Links](assembler-linker).

Jede Wertdefinition verfügt über einen Bezeichner. Definitionen können nicht dupliziert werden und Typmetadaten müssen mit dem in der Definition angegebenen Typ übereinstimmen. Der Name allein identifiziert keine Definitionen, selbst wenn Deklarationen mit demselben Namen einander verdecken.

## IR Konfiguration und Auslesen

Unten finden Sie ein vollständiges Rust-Beispiel, das die Funktion mit der Kiste `ir` erstellt und überprüft und dann typed IR ausgibt.

```rust
use ir::{ModuleBuilder, Target, Type};

fn main() {
    let target = Target::lookup("x86_64-whale-linux").unwrap();
    let mut module = ModuleBuilder::new(target.name(), target.data_layout());
    let mut function = module.begin_function("answer", vec![], Type::I32);
    let left = function.const_i32(40);
    let right = function.const_i32(2);
    let answer = function.add(Type::I32, left, right);
    function.ret(Some(answer));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    print!("{}", ir::print_module(&module));
}
```

Der Drucker gibt IR aus:

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "answer": whale () -> i32, linkage internal

  fn @f0 "answer"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 40
    %v1: i32 = const i32 2
    %v2: i32 = add i32 %v0, %v1
    ret i32 %v2
  }

}
```

Speichern Sie die Ausgabe als `answer.wir`, um sie mit Textparser oder CLI einzulesen und zu prüfen. Skalare Ganzzahlausführung wird unten erklärt.

Speichern Sie die Ausgabe als `answer.wir`, um sie mit Textleser und CLI einzulesen und zu prüfen. IR-Ausführung ist noch nicht verfügbar.

## Typ

|Typ|Bedeutung|
| --- | --- |
| `bool` |Logischer Wert false oder true|
| `i1`, `i8`, `i16`, `i32`, `i64`, `i128` |Ganzzahl mit Vorzeichen und angegebener Bitbreite|
| `u1`, `u8`, `u16`, `u32`, `u64`, `u128` |Ganzzahl ohne Vorzeichen der angegebenen Bitbreite|
| `f16`, `f32`, `f64` |Gleitkommawert mit angegebener Bitbreite|
| `ptr<T>` |T Zeiger auf Typwert|
| `fnptr<signature>` | Aufrufbarer Zeiger mit genauen Parameter-/Ergebnistypen und Aufrufkonvention |
| `array<T, N>` |N Elemente des gleichen Typs|
| `struct{T, ...}` |Geordnetes Strukturfeld|
| `tuple<T, ...>` |geordnete Tupelelemente|
| `void` |Keine Ergebnisse|

`bool`, `i1` und `u1` sind unterschiedliche Typen. signed `i1` steht für −1 und 0 und unsigned `u1` steht für 0 und 1. Die Ganzzahl 1 ist keine implizite logische Bedingung. Bedingter Zweig, die Bedingung von Select, `trap_if` erfordert den Operanden Bool.

Die Speichergröße wird nicht allein durch die Anzahl der Bits im Wert bestimmt und folgt [Ziellayout](memory-model). Der Wert von `i1` beträgt beispielsweise 1 Bit, belegt aber mindestens 1 Byte im Speicher.

## Funktionen und Aufrufe

Die Funktion gibt alle Parameter, Ergebnistypen, Aufrufkonventionen und linkage an. Direkte und indirekte Anrufe müssen mit der Signatur des Anrufers übereinstimmen. Der Anruf void hat kein Ergebnis ID. Ein Aufruf von nonvoid behält die Ergebnisdefinition bei, auch wenn O0 das Ergebnis nicht verwendet.

Die Rückgabe muss mit dem Ergebnistyp der Funktion übereinstimmen. Die Rückgabe void enthält keinen Wert, und die Rückgabe nonvoid enthält einen Wert des deklarierten Ergebnistyps.

### Deklarationen, Identitäten und Anrufe

`Module.declarations` zeichnet den `FunctionId` jeder Funktion, den Namen, die vollständige Signatur, die Verknüpfung und den Namen des externen Links auf. Eine Definition bezieht sich auf diese Identität; Parameter- und Rückgabetypen müssen mit seiner Deklaration übereinstimmen. Identische wiederholte Deklarationen werden über `declare_function` in dieselbe ID aufgelöst; Konflikte und doppelte Definitionen sind Fehler. Eine interne Deklaration benötigt einen Hauptteil im Modul. Eine externe Deklaration kann bis zur Verknüpfung ungelöst sein oder einen exportierten Textkörper haben. Interne Funktionen haben kein `link_name`; Externe Funktionen erfordern einen expliziten, nicht leeren Namen ohne NUL. Zwei unterschiedliche Funktionsdeklarationen können nicht denselben externen Namen beanspruchen. Globale und Funktionen verwenden weiterhin separate interne Namespaces.

Registrieren Sie Deklarationen vor dem Erstellen von Körpern mit `begin_declared_function`, um Weiterleitungsaufrufe und Rekursion zu unterstützen. `begin_function` bleibt eine Annehmlichkeit für eine neue interne Whale-Funktion. Die überprüften APIs `declare_function`, `begin_declared_function`, `function_addr`, `null_function` und `call` geben `Result` zurück; Bei einem abgelehnten Aufruf wird weder eine Anweisung angehängt noch ihre Ergebnis-ID zugewiesen.

Das folgende vollständige Rust-Programm deklariert eine externe Funktion, übernimmt deren typisierte Adresse und gibt sowohl direkte als auch indirekte Aufrufe aus:

```rust
use ir::{Callee, CallingConvention, DataLayout, FunctionSignature, Linkage, ModuleBuilder, Type};

fn main() {
    let mut module = ModuleBuilder::new("x86_64-whale-linux", DataLayout::default_64bit_le());
    let signature = FunctionSignature {
        params: vec![Type::I32], ret: Type::I32,
        convention: CallingConvention::SysV64, variadic: false,
    };
    let identity = module.declare_function(
        "identity", signature, Linkage::External, Some("identity_i32".into()),
    ).unwrap();
    let mut function = module.begin_function("answer", vec![], Type::I32);
    let input = function.const_i32(42);
    let callback = function.function_addr(identity).unwrap();
    // The direct call's result remains defined even though it is unused.
    function.call(Callee::Direct(identity), vec![input]).unwrap();
    let result = function.call(Callee::Indirect(callback), vec![input]).unwrap().unwrap();
    function.ret(Some(result));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    print!("{}", ir::print_module(&module));
}
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "identity": sysv64 (i32) -> i32, linkage external, link_name "identity_i32"
  declare @f1 "answer": whale () -> i32, linkage internal

  fn @f1 "answer"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 42
    %v1: fnptr<sysv64 (i32) -> i32> = function_addr @f0
    %v2: i32 = call sysv64 i32 @f0(%v0)
    %v3: i32 = call sysv64 i32 indirect %v1(%v0)
    ret i32 %v3
  }

}
```

`Callee::Direct(FunctionId)` wird über die Deklarationstabelle aufgelöst; `Callee::Indirect(ValueId)` erfordert einen `Type::FnPtr(FunctionSignature)`-Wert. Die Signatur umfasst alle Parametertypen, den Ergebnistyp und `CallingConvention::{Whale, SysV64}`. Es bleibt durch Kopien, Speicherung, Parameter, Rückgaben, Phi und Auswahl erhalten. Datenzeiger und Ganzzahlwerte sind nicht aufrufbar. Umwandlungen mit Funktionszeigertypen werden abgelehnt. Durch das Ändern einer Typanmerkung kann eine aufrufbare Signatur nicht geändert werden. Ein Funktionszeiger verfügt über einen 64-Bit-Adressspeicher auf diesem Ziel; Dadurch werden selbst keine Runtime-Shadow-Metadaten implementiert.

Arität, genaue Argument-/Ergebnistypen, Vorhandensein der Ergebnis-ID und Aufrufkonvention müssen übereinstimmen. Es gibt keine impliziten Konvertierungen. Der indirekte Angerufene muss den Anruf genauso dominieren wie seine Argumente. `variadic: true`, Parameter vom Typ void und SysV64 Aggregatparameter/Ergebnissignaturen werden abgelehnt. Whale aggregierte Signaturen können in IR dargestellt werden; Für keine der beiden Konventionen sind die native ABI-Klassifizierung und die Ausgabe von Maschinenaufrufen noch verfügbar.

`null_function(signature)` stellt einen typisierten Nullfunktionszeiger dar. Der Aufruf erfolgt durch ein gut eingegebenes IR mit einem erforderlichen Laufzeit-Trap vor der Eingabe eines Angerufenen. Ein Ziel ungleich Null, das ungültig, abgelaufen oder mit der überprüften Signatur nicht kompatibel ist, muss ebenfalls abgefangen werden. Diese Laufzeitprüfungen und die Verwaltung der Lebensdauer ausländischer Rückrufe warten auf die Interpreter-/native Ausführungsschicht. Ein erfolgreicher Verifizierer bedeutet nicht, dass beliebige externe Adressen sicher sind.

### AST-Aufrufformen

Dies sind Ausdrucksfragmente innerhalb eines AST-Format-2-Programms:

```json
{"Call":{"callee":{"Direct":"increment"},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

```json
{"Call":{"callee":{"Indirect":{"FunctionRef":"increment"}},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

`Direct` und `FunctionRef` verwenden den Funktionsnamensraum, auch wenn eine Variable denselben Namen hat. `Indirect` wertet zuerst seinen Ausdruck und dann die Argumente von links nach rechts aus. Ein void-Aufruf ist als `ExprStmt` gültig, jedoch nicht als Variableninitialisierer, Argument, Operand oder Rückgabewert. Aufrufe und Funktionsverweise sind keine numerischen Konstantenausdrücke zur Kompilierungszeit. `NullFunction` nimmt ein Signaturobjekt mit den Feldern `params`, `ret`, `convention` und `variadic`.

Das [vollständige JSON-Beispiel](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/ast-v2-calls.json) speichert einen Rückruf und ruft ihn vor einem externen Anruf auf. Senken Sie es ab mit:

```sh
cargo run --locked --features socket-cli -- ir lower ir/tests/fixtures/ast-v2-calls.json
```

Sein [erwarteter IR](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/calls-v3.wir) wird in den Tieferlegungstests überprüft. Funktionsidentität und Linknamen werden an der Grenze IR dargestellt; Ihre Erhaltung durch native Objektgenerierung und -verknüpfung ist immer noch eine separate Arbeit.

## Verfügbarkeit von Blöcken und Werten

Jeder Block hat eine eindeutige Kennung und genau einen terminator. Verzweigungsziele müssen zur gleichen Funktion gehören. Der Eintrittsblock muss vorhanden sein und darf keine Vorderkante und phi haben. Wenn Sie eine Schleife erstellen, verzweigen Sie vom Eingangsblock zu einem separaten Schleifenkopf.

Die Definition des Werts im ausführbaren Pfad sollte seinen typischen Verwendungspunkt bestimmen. Das bedeutet, dass alle Pfade vom Einstiegspunkt bis zum Verwendungspunkt diese Definition durchlaufen müssen. Im selben Block müssen Definitionen vor Verwendungen stehen. Die Speicherreihenfolge der Blöcke bestimmt nicht die Dominanz.

Gehen Sie davon aus, dass der Einstiegspunkt zu left oder right verzweigt und dann bei join zusammentrifft. Werte, die nur in left definiert sind, können nicht als allgemeine Werte in join verwendet werden. Dies liegt daran, dass der durch right verlaufende Pfad undefiniert ist. Die Werte jedes vorangehenden Blocks müssen zu phi kombiniert werden, das als Eingabe empfangen wird.

Auch nicht erreichbare Blöcke bleiben im Modul erhalten. Der Verifizierer überprüft kontinuierlich die Kennung, den Typ, den Operanden und die Verzweigungsstruktur des Blocks. Die Definition eines nicht erreichbaren Blocks kann keinen Mehrwert für die normale Nutzung eines erreichbaren Pfads liefern.

## Phi Befehl

phi wird vor allen regulären Befehlen im Block platziert. Für jeden unterschiedlichen Vorgängerblock ist genau eine Eingabe erforderlich. Der Eingabewert muss vom Typ phi sein und am Ende des entsprechenden Vorgängersatzes verfügbar sein.

Auch wenn in einem vorhergehenden Block mehrere Kanten vorhanden sind, gibt es nur eine Eingabe. Schleife phi kann sich auf den Wert eines Blocks beziehen, der später in der Modulspeicherreihenfolge erscheint, solange es sich um einen Wert handelt, der an der Wiederholungskante berechnet wird. Fehlende, doppelte oder irrelevante vorangehende Blöcke und falsch eingegebene Eingaben sind Validierungsfehler.

## Bewertung und Auswahl

Whale AST wertet Aufrufziele und Unterausdrücke von links aus, wo sie angegeben sind. Das Frontend drückt die Kurzschlussauswertung als Kontrollflusszweig aus.

Select wählt einen der bereits berechneten Werte aus. Die Berechnung beider Eingaben wird nicht ausgelassen. Selbst wenn Sie beispielsweise einen sicheren Wert wählen, können Sie das Auftreten von trap bei der Berechnung anderer Eingaben nicht vermeiden. Berechnungen, die nur auf einem bestimmten Pfad ausgeführt werden müssen, sollten in einen bedingten Block eingefügt werden.

## Verifizierungsabteilung trap

Ungültige IR ist ein Prüfungsfehler. Der Prüfer lehnt legacy `undef` ab; erzeugen Sie die IR erneut aus AST. Deklarationen ohne Initialisierung verwenden format 4 `uninit` und geprüfte tatsächliche Lesezugriffe, ohne Null- oder beliebige Ersatzwerte. Laufzeitverletzungen erzeugen definierte Traps mit IR-Position.

`InterpreterTrap` meldet Grund, ausgeführte Schritte und `ExecutionSite`: Funktions-ID, Block-ID, Anweisungsindex ab null und optionale Ergebniswert-ID. Der terminator folgt den Anweisungen. Die CLI nennt auch die Eingabedatei. Typed IR trägt noch keine Quell-spans; dies sind IR-Positionen, keine Quellzeilennummern. Ein trap liefert einen Fehler und stoppt nachfolgende Ausführung; die Bibliothek beendet nicht den Hostprozess.

Diese Garantie gilt für verifizierte IR- und Trace-Speicher. Die externe C·Rohadresse·Inline-Assembly verfügt über einen separaten Vertrag und erkennt nicht immer Verstöße außerhalb seiner Grenzen. Bitte beachten Sie [Speichermodell](memory-model).

## Austauschformat und Textdarstellung

AST und typed IR verwenden das entsprechende format version und das gemeinsame semantics version. Der Leser sollte den unversionierten/unbekannten Versions-/Feld-/Funktions-/Duplikatschlüssel JSON ablehnen. Konstrukteure sollten nicht davon ausgehen, dass nicht unterstützte Eigenschaften stillschweigend ignoriert werden.

Ganzzahlen werden als Bitbreite·signedness·String-Zahlen übergeben. Gleitkommakonstanten werden als Breite und genaue Bitfolge übergeben. Der Text round-trip in IR muss den Namen·ID·Typ·Konstante·Sequenz·Eigenschaft·Metadaten beibehalten. Leerzeichen und Kommentarplatzierung unterliegen nicht der Aufbewahrung.

Der folgende AST-JSON-Vertrag sowie Lesen, Prüfen und Round-trip-Ausgabe von typed IR format 4 sind verfügbar.

### Ausgegebene Identitäten und Namen in Anführungszeichen

Typed IR format 4 gibt Funktionen als `@fN`, globale Werte als `@gN`, Werte als `%vN` und Blöcke als `%bN` aus. Funktions- und globale IDs gehören zum Modul; Wert- und Block-IDs zur jeweiligen Funktion. Vorgegebene IDs bleiben einschließlich Nummernlücken erhalten. Namen in Anführungszeichen dienen der Beschreibung, nicht der Referenzauflösung. Jede Funktion nennt ausdrücklich `entry %bN`, unabhängig von der Speicherreihenfolge ihrer Blöcke.

Dieses vollständige Modul wurde über die Rust-IR-API geprüft und ausgegeben. Beide Verzweigungsblöcke heißen `"branch"`; ihre IDs unterscheiden Definitionen und phi-Eingaben.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "choose": whale (bool) -> i32, linkage internal

  fn @f0 "choose"(%v0 "condition": bool) -> i32, entry %b0 {
  %b0 "entry":
    cbr bool %v0, label %b1, label %b2
  %b1 "branch":
    %v1: i32 = const i32 1
    br label %b3
  %b2 "branch":
    %v2: i32 = const i32 2
    br label %b3
  %b3 "join":
    %v99: i32 = phi i32 [ %v1, %b1 ], [ %v2, %b2 ]
    ret i32 %v99
  }

}
```

`%v0` wird in der Parameterliste definiert. `%b1` und `%b2` bleiben trotz gleicher Namen verschiedene Blöcke; phi bezeichnet die Vorgänger mit IDs. Verzweigungen und switch-Ziele verwenden dieselbe Block-ID-Syntax. Der Drucker nummeriert das ausdrücklich vorgegebene `%v99` nicht neu.

Alle Namen und Zeichenketten stehen in doppelten Anführungszeichen: Ziel, Funktions-, globale, Parameter- und Blocknamen, externe Linknamen, Konstantendeklarationen und trap-Gründe. Druckbares Unicode bleibt erhalten. Die Escape-Sequenzen sind `\"`, `\\`, `\n`, `\r`, `\t`, `\0` und `\u{hex}` mit kleinen Hexadezimalziffern für weitere Steuerzeichen und U+2028/U+2029. Auch Namen mit Zeilenumbruch, Tabulator, Anführungszeichen, Backslash und koreanischem Text bleiben in einem Datensatz.

```text
"line\ncolumn\tquote\"slash\\한글"
```

Der Leser akzeptiert typed IR formats 3 und 4 mit semantics version 1 und schreibt immer format 4. `uninit` erfordert format 4. Gültige bisherige format-3-Anweisungen bleiben lesbar; legacy `undef` ist in beiden Formaten ein Prüfungsfehler und muss aus AST neu erzeugt werden. Format-2-Text benötigt manuelle Migration zu expliziten IDs, zitierten Namen und Eintrittsreferenzen. AST JSON hat sein separates format 2.

### Text-IR lesen und prüfen

`ir::parse_module` liest typed IR, prüft es und liefert ein `Module`. Es akzeptiert die aktuelle Druckersyntax für Skalare, Kontrollfluss, Speicher, direkte/indirekte Aufrufe und konstante Ausdrücke. IDs, Namen, Typen, genaue Ganzzahl- und float-Bits, Ausdrucksbäume und berechnete Ergebnisse, Blockreihenfolge, Eintritt, Ausrichtung, Signaturen und link_name bleiben erhalten. Leerraum und `//`-Zeilenkommentare werden kanonisiert. Speichern Sie das vollständige IR oben als `answer.wir` und führen Sie diese Befehle im Standardbuild aus.

```sh
cargo run --locked -- ir verify answer.wir
cargo run --locked -- ir print answer.wir -o canonical.wir
```

Unbekannte Versionen, Felder, Befehle und Escapes, Literale außerhalb des Wertebereichs, doppelte IDs, widersprüchliche Typangaben und nachfolgende Eingabe werden abgelehnt. `ParseError` enthält Byteposition und Zeile/Unicode-Skalarspalte ab 1. Prüfungsfehler werden nach Möglichkeit der betreffenden Funktion oder globalen Deklaration zugeordnet. Auch `print` prüft und bewahrt bei Fehlern eine bestehende Ausgabe. Neue Syntax oder Semantik erfordert die entsprechende Änderung der format/semantics version; unbekannte Versionen sind Fehler.

### Ressourcenlimits der Prüfung

`IrLimits` begrenzt standardmäßig Eingaben auf 8 MiB, Tokens und besuchte Knoten jeweils auf 1,000,000 und Typ-/Ausdruckstiefe jeweils auf 128. Die Wurzel hat Tiefe 0, jedes Kind erhöht sie um 1. Tiefen dürfen gesenkt oder bis `MAX_IR_NESTING`, 256, erhöht werden. Überschreitungen liefern `LimitError`, `VerifyError::ResourceLimit`, `ConstEvalError::ResourceLimit` oder `CallError::ResourceLimit`. Knoten zählen die Traversierung an jeder Eingabegrenze einschließlich Typangaben und Ausdrücken, nicht die verstrichene Zeit.

```rust
use ir::{parse_module_with_limits, print_module, IrLimits};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let source = std::fs::read_to_string("answer.wir")?;
    let limits = IrLimits {
        max_type_depth: 32,
        max_const_depth: 32,
        max_nodes: 10_000,
        ..IrLimits::default()
    };
    let module = parse_module_with_limits(&source, limits)?;
    let canonical = print_module(&module);
    let reread = parse_module_with_limits(&canonical, limits)?;
    assert_eq!(print_module(&reread), canonical);
    let invalid = source.replacen("format_version 4", "format_version 99", 1);
    assert!(parse_module_with_limits(&invalid, limits).is_err());
    Ok(())
}
```

Limits gelten auch für `verify_module_with_limits`, `ConstExpr::evaluate_with_limits`, `validate_signature_with_limits` und `ModuleBuilder::declare_function_with_limits`. Iterative Typprüfung erfolgt vor rekursivem clone, Vergleich und Diagnose; Konstanten werden mit einem Arbeitsstack ausgewertet. Geliehene Rust-Bäume samt Drop bleiben beim Aufrufer. Beliebige ungeprüfte Bäume haben weiterhin rekursives clone/Drop; die von der checked-Deklarations-API besessene abgelehnte Signatur wird iterativ freigegeben. Ungültige IR ist ein Prüfungsfehler. Der Prüfer lehnt legacy `undef` ab; erzeugen Sie die IR erneut aus AST. Deklarationen ohne Initialisierung verwenden format 4 `uninit` und geprüfte tatsächliche Lesezugriffe, ohne Null- oder beliebige Ersatzwerte. Laufzeitverletzungen erzeugen definierte Traps mit IR-Position.

### Angegebene Version AST JSON

Speichern Sie Folgendes als `program.json`. Alle vier Umschlagfelder sind Pflichtfelder. `program` enthält die erforderlichen Arrays `declarations`, `globals` und `functions`, die möglicherweise leer sind. Funktionsname, Parameter, Rückgabetyp, Hauptteil, `convention` und `linkage` sind erforderlich. `link_name` kann für interne Funktionen fehlen/null sein und muss eine nicht leere Zeichenfolge ohne NUL für externe Funktionen sein. Jede Aufzählung verwendet entweder ihren Einheitennamen oder ein einzelnes Variantenschlüsselobjekt. Einheitenvarianten akzeptieren auch ein Nullwertobjekt, z. B. `{"Void":null}`; Der Encoder gibt den Gerätenamen `"Void"` aus. `VarDecl.init` kann fehlen oder null sein; weitere erforderliche Felder müssen vorhanden sein.

```json
{
  "format_version": 2,
  "semantics_version": 1,
  "features": [],
  "program": {
    "globals": [],
    "functions": [
      {
        "name": "answer",
        "parameters": [],
        "return_type": {
          "Int": {
            "bits": 128,
            "signed": false
          }
        },
        "body": [
          {
            "Return": {
              "Lit": {
                "Int": {
                  "bits": 128,
                  "signed": false,
                  "value": "340282366920938463463374607431768211455"
                }
              }
            }
          }
        ],
        "convention": "Whale",
        "linkage": "Internal",
        "link_name": null
      }
    ],
    "declarations": []
  }
}
```

```sh
cargo run --locked --features socket-cli -- ir lower program.json
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "answer": whale () -> u128, linkage internal

  fn @f0 "answer"() -> u128, entry %b0 {
  %b0 "entry":
    %v0: u128 = const u128 340282366920938463463374607431768211455
    ret u128 %v0
  }

}
```

Die Ganzzahl `value` ist eine Dezimalzeichenfolge. Signed Eine Zahl wird nach dem optionalen Minus einer Ganzzahl verwendet und Leerzeichen, Pluszeichen, Exponenten und Trennzeichen sind nicht zulässig. Der zulässige Bereich wird durch die angegebene Breite und signedness bestimmt. Das obige `u128::MAX` bleibt unverändert bis JSON und lowering erhalten. Negative unsigned-Werte oder Werte außerhalb des Bereichs sind keine wrap und sind Fehler. Der Wert Float verwendet eine hexadezimale Bitfolge mit der genauen Breite, die in [Numerische Operationen](numeric-operations) beschrieben ist.

`format_version` ist 2 für dieses AST-Format; `semantics_version` ist 1. `features` muss ein leeres Array sein. Unbekannte Felder, Versionen, Funktionen, doppelte rohe JSON-Schlüssel (einschließlich maskierter äquivalenter Schlüssel) und nachgestellte Werte sind Fehler, auch mit `--no-verify`. Der Bibliothekseinstiegspunkt ist `ir::lower_ast::interchange::decode`; `encode` gibt den Umschlag aus. `decode` ist standardmäßig auf ein Quellbyte-Limit von 8 MiB eingestellt; `decode_with_limit` akzeptiert ein Anruflimit. JSON Die Verschachtelung ist begrenzt. Verwenden Sie diesen Rohdecoder, anstatt ihn in eine generische Karte zu analysieren, die bereits doppelte Schlüssel verwerfen könnte.

[Das vollständige JSON Schema](https://github.com/wavefnd/Whale/blob/master/ir/schema/ast-v2.schema.json) spezifiziert Formen, erforderliche Felder und Varianten. Zusätzlich kommen Bereichs-/Typprüfungen und die Erkennung doppelter Schlüssel zum Einsatz. Die Teilmenge der Skalarreduzierung umfasst Literale, Variablen/Konstanten, Add/Sub/Mul, Vergleiche, Zuweisungen, If/While, Return und Break/Continue. Funktionsreferenzen, direkte Aufrufe und indirekte Aufrufe werden unterstützt; Aggregatausdrücke werden nicht unterstützt. `Opaque` ist im Schema darstellbar, wird jedoch durch Absenken nicht unterstützt.

Für die Migration wird das alte nackte Program in einen Envelope eingebettet; JSON-Zahlen werden durch dezimale Ganzzahlzeichenketten oder Gleitkommabitzeichenketten ersetzt. Eingaben ohne Version werden abgelehnt. Format 1 muss auf Format 2 umgestellt werden: `program.declarations` (bei Nichtverwendung ein leeres Array) und ausdrückliche `convention`/`linkage` an Definitionen ergänzen. Die Versionen sind unabhängig: AST format 2, typed IR format 4 und semantics version 1.

### Abgelehnte Eingabe und CLI Wiederherstellung

Speichern Sie die folgende vollständige Eingabe als `invalid.json`.

```json
{
  "format_version": 99,
  "semantics_version": 1,
  "features": [],
  "program": {
    "globals": [],
    "functions": [],
    "declarations": []
  }
}
```

```sh
cargo run --locked --features socket-cli -- ir lower invalid.json -o rejected.wir
```

```text
Failed to parse socket JSON: unsupported AST format_version 99; expected 2
```

Der Befehl wird mit einem Status ungleich Null beendet und erstellt keine neue Ausgabe und überschreibt keine vorhandenen Dateien. Typkonflikte schlagen ebenfalls fehl, bevor die Ausgabe veröffentlicht wird. Ohne `socket-cli` erstellte Binärdateien werden mit Status 2 beendet und geben einen Wiederherstellungsbefehl aus, der `--features socket-cli` enthält.


## Interpreter für skalare Ganzzahlen

Der Standardinterpreter führt Ganzzahlen/Bool, Kontrollfluss, Stapelzuweisungen, Datenzeigerspeicherung und -lesen, typed GEP, memcpy und memset aus. Beim Lesen von checked-Paaren wird Padding ausgeschlossen. Adressen sind synthetische 64-Bit-Werte ohne Zugriff auf Hostspeicher. Argumente und Rückgaben bleiben Ganzzahlen/Bool oder void; float, Aufrufe, Funktionszeiger, allgemeine Aggregatwerte, globale Adressen und native Ausführung fehlen. Stapelzuweisungen leben bis zur Rückgabe. Lexikalisches Lebensdauerende, Zeigerübergabe bei Aufrufen/Rückgaben, Fremdspeicheradapter und native shadow metadata müssen noch implementiert werden.

`InterpreterOptions::max_steps` ist standardmäßig 1,000,000. Jede ausgeführte Anweisung einschließlich phi und jeder terminator verbraucht einen Schritt. Null stoppt vor der ersten Operation; eine Endlosschleife gibt `InterpreterError::StepLimit` zurück. `ir_limits` begrenzt die Prüfung separat. Auch ungenutzte Arithmetik wird ausgeführt und kann trap auslösen. checked overflow liefert Bool; erst ein explizites trap_if löst einen trap aus.

Speichern Sie dieses vollständige Modul als `swap-loop.wir`. Der Einstieg ist explizit, auch wenn der Ausgang zuerst gespeichert ist. Beim Blockeintritt werden alle phi-Eingaben aus dem Vorgänger gelesen, bevor Ergebnisse gemeinsam geschrieben werden. Drei Iterationen tauschen 11 und 22 dreimal und liefern 22.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f7 "swap_loop": whale (u32) -> i32, linkage internal

  fn @f7 "swap_loop"(%v0 "iterations": u32) -> i32, entry %b11 {
  %b90 "exit":
    ret i32 %v5
  %b11 "entry":
    %v1: i32 = const i32 11
    %v2: i32 = const i32 22
    %v3: u32 = const u32 0
    %v4: u32 = const u32 1
    br label %b20
  %b20 "loop":
    %v5: i32 = phi i32 [ %v1, %b11 ], [ %v6, %b30 ]
    %v6: i32 = phi i32 [ %v2, %b11 ], [ %v5, %b30 ]
    %v7: u32 = phi u32 [ %v3, %b11 ], [ %v9, %b30 ]
    %v8: bool = icmp ult u32 %v7, %v0
    cbr bool %v8, label %b30, label %b90
  %b30 "next":
    %v9: u32 = add u32 %v7, %v4
    br label %b20
  }

}
```

```shell
whale ir run swap-loop.wir --function @f7 --arg 3 --max-steps 100
```

```text
i32 22
```

Die Rust-API liefert Wert und Schrittzahl oder einen strukturierten Prüfungs-, Unterstützungs-, Argument-, Limit- oder trap-Fehler. Dieses vollständige Programm liest dieselbe Datei `swap-loop.wir`:

```rust
use ir::{interpret_with_options, parse_module, ConstValue, FunctionId,
         InterpreterError, InterpreterOptions};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let source = std::fs::read_to_string("swap-loop.wir")?;
    let module = parse_module(&source)?;
    let options = InterpreterOptions {
        max_steps: 100,
        ..InterpreterOptions::default()
    };
    let result = interpret_with_options(
        &module, FunctionId(7), &[ConstValue::U(3)], options,
    )?;
    assert_eq!(result.value, Some(ConstValue::I(22)));
    assert_eq!(result.steps, 32);
    let stopped = interpret_with_options(
        &module, FunctionId(7), &[ConstValue::U(3)],
        InterpreterOptions { max_steps: 0, ..options },
    );
    assert!(matches!(stopped, Err(InterpreterError::StepLimit { .. })));
    println!("i32 22");
    Ok(())
}
```
