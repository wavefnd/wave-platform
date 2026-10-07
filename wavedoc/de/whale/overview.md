---
translation_set_id: whale-overview
path: whale/overview
locale: de
group: whale
group_order: 1
order: 1
title: Whale-Dokumentation
summary: Leitfaden zu Toolchain-Komponenten, verfügbaren Befehlen und Referenzdokumentation.
---

## Einführung

Whale ist eine universelle Compiler-Toolchain für Programmiersprachenimplementierungen und Compiler-Tools. Bietet typisierte Zwischenausdrücke (IR), AMD64-Assembler, Objektdateibibliothek und Linker-basierte Funktionen. Jede Komponente wird über die Bibliothek Rust und den Befehl `whale` verwendet.

IR drückt die Semantik einer Berechnung unabhängig von der Maschinensprachenkodierung aus. Der Assembler kodiert Maschinenanweisungen, um verschiebbare Objekte zu erstellen. Objektbibliotheken repräsentieren Abschnitte, Symbole und Neuanordnungen. Der Linker löst Referenzen zwischen Objekten auf und platziert die ausführbare Datei.

## Erstellen und verwenden Sie Tools

|Arbeit|Dokument|
| --- | --- |
|Was jedes Tool macht| [Toolchain-Komponenten](/docs/de/whale/ecosystem) |
|Wave Auswahl von Programmaufbau/Link/Ziel| [Erstellen und verknüpfen](/docs/de/whale/build-link-targets) |
|Paket-/Abhängigkeitsverwaltung| [Vex](/docs/de/whale/vex-package-manager) |
|Whale Befehl ausführen| [Whale CLI](/docs/de/whale/whale-cli) |

## Dokumenthandbuch

|Referenzdokument|Inhalt|
| --- | --- |
| [Siehe IR](ir-reference) |Typen, Werte, Funktionen, Kontrollfluss, Validierung, Austauschformate|
| [Numerische Operationen](numeric-operations) |Ganzzahlarithmetik, Verschiebung, Typkonvertierung, Gleitkomma|
| [Speichermodell](memory-model) |Initialisierung, Zeigervalidierung, Adressberechnung, Layout, Strings|
| [Debuggen mit O0](o0-debugging) |Berechnungsaufbewahrung, Variablenspeicher, Debug-Informationen|
| [AMD64 Ziel](amd64-target) |Zielkennung, Aufrufkonvention, native Funktionsumfang|
| [Assembler und Linker](assembler-linker) |Assembly-Operanden, Abschnitte, Symbole, statische Links|

Das Referenzdokument definiert die semantischen Regeln für Whale. Verfügbare Funktionen folgen der Tabelle unten. Die in der Referenzdokumentation beschriebenen Vorgänge sind möglicherweise nicht in allen Builds verfügbar.

## Supportstatus

|Komponente|Schnittstelle vorhanden|Grenze|
| --- | --- | --- |
|Assembler|Erstellen Sie ein verschiebbares Objekt ELF64 aus der Baugruppe AMD64|Unvollständige Abdeckung von Befehlen und Anweisungen|
|Objektbibliothek|Objektzusammensetzung, payload ohne BSS, Zielvalidierung, AMD64 ELF64 Serialisierung und wählbare Wave Datensatzimplementierung zur Überprüfung der Größe.|Verschiebbare Objekte sind keine ausführbaren Dateien|
| IR | Konstruktion, Druck, signaturgeprüfte direkte/indirekte Aufrufe, versionierte AST-Senkung, Zielvalidierung und geprüfte Typlayouts | AST format 2 / typed IR format 4, bitgenaue Konstanten sowie Textlesen, Prüfung und Round trips verfügbar; Maschinenaufrufausgabe fehlt weiterhin |
|Linker|Symbolinterpretation, Überprüfung des Eingabeziels, Überprüfung der Platzierung von Datei-/Speicherabschnitten|Die Anwendung einer vollständigen Verschiebung und die Ausgabe ausführbarer Dateien werden nicht unterstützt.|
| Ausführung und Debugging | Ganzzahlen/Bool, Kontrollfluss, verfolgter Stapelspeicher und Initialisierungsprüfung | Float, Aufrufe, globale Adressen, native Codeerzeugung und DWARF fehlen |

Für verifizierte IR- und Trace-Speicher gelten Regeln, die undefiniertes Verhalten verbieten. Die experimentelle Implementierung implementiert noch nicht alle Laufzeitprüfungen in der Speicher-/Ausführungsreferenzdokumentation.

## Bauen und zusammenbauen

Erstellen Sie mit Rust 1.86.0 oder höher.

```sh
git clone https://github.com/wavefnd/Whale.git
cd Whale
cargo build --release --locked
```

Speichern Sie die folgende Baugruppe als `answer.asm`.

```asm
section .text
global answer

answer:
    mov eax, 42
    ret
```

Erstellen Sie ein Objekt ELF64.

```sh
./target/release/whale asm --amd64 answer.asm -o answer.o
```

Es verwendet den eigenen Assembler von Whale, sodass kein externer Assembler erforderlich ist. Die Ausgabe enthält eine aufrufbare Funktion und keinen Prozessstartcode.

Die experimentellen AST→IR-Befehle sind durch Erstellen mit `--features socket-cli` verfügbar. Informationen zum Befehl CLI finden Sie unter [Befehlsreferenz](/docs/de/whale/whale-cli).

## Bibliotheksnutzung und Diagnose

IR muss vor der Ausführung oder Codegenerierung überprüft werden. Eingabefehler und Missbrauch von builder führen zu Strukturfehlern. Eingabefehler in der Bibliothek dürfen den Hostprozess nicht beenden oder vorhandene Inhalte überschreiben. Tools, die nicht vertrauenswürdige Eingaben akzeptieren, sollten in der Lage sein, Ressourcenlimits anzupassen.

Dieselbe Toolchain-Version, Eingabe, dasselbe Ziel und dieselbe Konfiguration sollten eine deterministische Ausgabe erzeugen. Die Metadaten der Distribution identifizieren die Version·commit·und die verfügbaren Funktionen. CI prüft auf gute Eingaben, abzulehnende Eingaben, trap, O0 Aufbewahrung, round-trip, native Ausführungssemantik für jede Schnittstelle. Informationen zu Entwicklungs-/Verifizierungsbefehlen finden Sie unter [Whale Beitragsinformationen](https://github.com/wavefnd/Whale/blob/master/CONTRIBUTING.md).
