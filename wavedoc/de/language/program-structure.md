---
translation_set_id: program-structure
path: language/program-structure
locale: de
group: language
group_order: 2
order: 1
title: 1. Von der Quelldatei zum laufenden Programm
summary: Erfahren Sie mehr über Quelldateien, Funktionen, Ausgabe, Inspektion und Ausführung.
---

## Bevor Sie mit diesem Kapitel beginnen

Bereiten Sie Compiler und Standardbibliothek gemäß [Installationsanleitung](/docs/de/getting-started/install) vor. Wenn Sie `wavec --version` in Ihrem Terminal ausführen können, können Sie loslegen. Jeder Editor, der reine Textdateien speichern kann, reicht aus.

In diesem Kapitel beginnen wir mit der Erstellung einer einzelnen Ausgabezeile und lernen die Beziehungen zwischen Quelldateien, Funktionen, Kompilierung, Ausführung und Exit-Code kennen. Das Ziel besteht nicht nur darin, Anweisungen zu kopieren, sondern auch erklären zu können, was in welcher Phase passiert.

## Erstellen Sie ein Arbeitsverzeichnis

Die Verwendung eines separaten Verzeichnisses für jedes Programm erleichtert das Auffinden der Quelle und der generierten ausführbaren Dateien. Erstellen Sie ein Verzeichnis im Terminal und navigieren Sie zu diesem.

```shell
mkdir wave-study
cd wave-study
```

Erstellen Sie in diesem Verzeichnis mit Ihrem Editor `main.wave`. Überprüfen Sie die Erweiterung, um sicherzustellen, dass der Dateiname nicht `main.wave.txt` lautet. Unten finden Sie die gesamte Datei, nicht nur ein Fragment innerhalb der Funktion.

<!-- wave-example: book-hello -->
```wave playground
fun main() {
    println("Hello, Wave!");
}
```

Ausführungsergebnis:

```text
Hello, Wave!
```

Führen Sie es mit dem folgenden Befehl aus:

```shell
wavec run main.wave
```

Vermischen Sie keine Terminalbefehle mit dem Code Wave. Geben Sie `wavec run` in das Terminal ein und schreiben Sie `fun main` in die Quelldatei. Es ist nicht erforderlich, die vom Programm ausgegebene `Hello, Wave!` wieder in die Quelle einzufügen.

## Zeile für Zeile lesen

`fun` ist ein Schlüsselwort, das eine Funktion deklariert. Eine Funktion ist eine benannte Menge von Operationen und `main` ist der Einstiegspunkt der ausführbaren Datei. Weitere Funktionen definieren Sie später selbst.

In `()` nach `main` schreiben Sie die Parameter. main in diesem Programm benötigt keine Parameter und ist daher leer. Notieren Sie die Operationen der Funktion zwischen `{` und `}`. Durch die Einrückung sind Blöcke für Menschen leichter lesbar und die Blockgrenzen selbst werden durch geschweifte Klammern angezeigt.

`println("Hello, Wave!");` ist ein Satz, der eine Zeichenfolge ausgibt. Doppelte Anführungszeichen geben den Anfang und das Ende eines Zeichenfolgenliterals an. Die Anführungszeichen selbst werden nicht in die Ausgabe einbezogen. Das Semikolon gibt das Ende dieser Anweisung an.

## Anweisungen werden in der Reihenfolge ausgeführt, in der sie geschrieben werden

Versuchen Sie, es dreimal so zu ändern, dass es gedruckt wird. Nicht mit dem vorherigen Programm kombinieren, sondern den Inhalt von main.wave durch das gesamte Programm unten ersetzen.

<!-- wave-example: book-sequence -->
```wave playground
fun main() {
    println("start");
    println("working");
    println("done");
}
```

Ausführungsergebnis:

```text
start
working
done
```

Nachdem Sie den ersten Satz beendet haben, fahren Sie mit dem nächsten Satz fort. Hier werden keine Aufgaben gleichzeitig ausgeführt. Wenn Sie die Ausgabereihenfolge ändern möchten, ändern Sie einfach die Reihenfolge der Sätze. Sie können diese Reihenfolge steuern, indem Sie später bedingte Anweisungen und Schleifen lernen.

## print und println

`println` fügt am Ende einen Zeilenumbruch hinzu. `print` ändert die Zeilen nicht automatisch. Der Unterschied wird deutlich, wenn kleine Teile zu einer einzigen Linie verbunden werden.

<!-- wave-example: book-print-lines -->
```wave playground
fun main() {
    print("Wave");
    print(" ");
    println("study");
    println("second line");
}
```

Ausführungsergebnis:

```text
Wave study
second line
```

Dreimaliges Drucken führt nicht immer zu drei Zeilen. Unterscheiden Sie zwischen der Anzahl der Ausgabeaufrufe und der Anzahl der Zeilen. Wenn Sie einen Zeilenumbruch direkt in eine Zeichenfolge schreiben, verwenden Sie `\n` escape. Die Zeichenfolge escape und Zeilenumbrüche in der Quelle werden ausführlich in [Saitenblatt](/docs/de/language/strings) behandelt.

## Wert in String einfügen

Um das Berechnungsergebnis in eine Zeichenfolge einzufügen, übergeben Sie den Wert, der `{}` entspricht.

<!-- wave-example: book-format-first -->
```wave playground
fun main() {
    println("{} + {} = {}", 2, 3, 2 + 3);
}
```

Ausführungsergebnis:

```text
2 + 3 = 5
```

Der erste `{}` enthält 2, der zweite enthält 3 und der dritte enthält 5. Formatzeichenfolge und Wert werden durch Kommas getrennt. Wenn Sie die Anzahl der Werte ändern, muss auch die Anzahl von placeholder übereinstimmen. Dies ist eine andere Ausgabesyntax als die String-Additionsoperation.

## Teilen Sie Inspektion, Bau und Ausführung auf

Das bisher verwendete run führt den Build und die Ausführung nacheinander durch. Wenn Sie wissen möchten, welcher Schritt fehlgeschlagen ist, können Sie ihn wie folgt aufschlüsseln:

```shell
wavec check main.wave
wavec build main.wave -o hello
```

check prüft Grammatik, Typen usw., testet jedoch nicht alle Eingaben in das Programm. Beispielsweise kann die Quelleninspektion allein nicht feststellen, ob eine Datei zur Laufzeit vorhanden ist. build erstellt eine ausführbare Datei. Führen Sie in Linux/macOS Folgendes aus.

```shell
./hello
```

Benennen Sie in Windows die Ausgabedatei `hello.exe` und führen Sie sie in PowerShell als `.\hello.exe` aus. Wenn Sie die Quelle ändern, nachdem Sie eine ausführbare Datei erstellt haben, müssen Sie sie neu erstellen, damit die Änderungen wirksam werden.

## Der Exit-Code ist ebenfalls ein Ergebnis

Menschen lesen die Ausgabeanweisung, aber eine Shell oder ein anderes Programm kann den Erfolg anhand des Exit-Codes bestimmen. Im Folgenden wird angegeben, dass main i32 zurückgibt.

<!-- wave-example: book-exit-success -->
```wave playground
fun main() -> i32 {
    println("completed");
    return 0;
}
```

Ausführungsergebnis:

```text
completed
```

0 ist eine Konvention, die ein normales Herunterfahren anzeigt. Wenn Sie einen Fehler direkt melden, geben Sie einen Code ungleich Null zurück. In der Shell Linux/macOS wird der Code unmittelbar nach der Ausführung als `echo $?` überprüft, und in der Shell PowerShell wird er als `$LASTEXITCODE` überprüft. Wenn Sie in der Zwischenzeit andere Befehle ausführen, kann sich das, was Sie überprüfen, ändern.

Durch Ausführen von `return` wird die Funktion beendet. main deklariert keine Parameter und selbst wenn sie Standardwerte haben, sind sie nicht zulässig. Lassen Sie den Rückgabetyp main weg oder verwenden Sie i32.

## So lesen Sie den ersten Fehler

Der folgende Code ist absichtlich falsch: Im Gegensatz zum laufenden Beispiel sollte er bei check fehlschlagen.

```wave
fun main() {
    println("hello")
}
```

Am Ende des Satzes steht kein Semikolon. Schauen Sie sich die Zeile an, die in der Diagnose angezeigt wird, und den Satz direkt davor. Der vom Compiler markierte Ort ist möglicherweise nicht der Ort, an dem der Fehler entstanden ist, sondern der Ort, an dem die fehlerhafte Struktur nicht mehr interpretiert werden kann.

Beheben Sie zunächst den ersten Fehler und überprüfen Sie ihn dann erneut. Wenn die vorangehenden Klammern oder Anführungszeichen nicht geschlossen werden, kann es auch im folgenden normalen Code zu verschiedenen Fehlern kommen. Wenn Sie versuchen, alle Leitungen gleichzeitig zu beheben, kann es leicht passieren, dass die ursprüngliche Ursache übersehen wird.

## Übungsprobleme

1. Erstellen Sie ein Programm, das drei Zeilen Selbstvorstellung ausgibt.
2. Übergeben Sie 12 und 8 als Formatargumente zum Drucken von `12 * 8 = 96`.
3. Schreiben Sie es, um eine Erfolgsmeldung auszudrucken und den Exit-Code 0 zurückzugeben.
4. Sagen Sie vor der Ausführung voraus, wie viele Ausgabezeilen es geben wird, wenn Sie print nur dreimal verwenden.

### Lösung: Ausgabe des Berechnungsprozesses

<!-- wave-example: book-first-solution -->
```wave playground
fun main() -> i32 {
    println("Learning Wave");
    println("My first program");
    println("Ready to calculate");
    println("{} * {} = {}", 12, 8, 12 * 8);
    return 0;
}
```

Ausführungsergebnis:

```text
Learning Wave
My first program
Ready to calculate
12 * 8 = 96
```

Anstatt das Berechnungsergebnis direkt als `96` in den String zu schreiben, habe ich es als Ausdruck übergeben. Dies ist der erste Schritt, um sicherzustellen, dass sich die Berechnungsergebnisse und die Anzeige auch dann nicht ändern, wenn sich der Eingabewert ändert. Im nächsten Kapitel werden wir Variablen Namen geben, um zu vermeiden, dass derselbe Wert mehrmals geschrieben wird.


## Element der obersten Ebene in der Quelldatei

Die Quelle Wave kann aus den folgenden Elementen der obersten Ebene bestehen:

- `import(...)`
- `const`, `static`
- `type`
- `struct`, `enum`, `proto`
- `extern(...)`, `export(...)`
- `fun`
- `#[target(...)]` Zustand vor dem unterstützten Artikel

Stellen Sie importierbaren Deklarationen `pub` voran. Platzieren Sie die lokale `var`-Deklaration in einer Funktion oder einem Block.

## Freistehendes Programm

Ziele ohne Kernel, Bootcode oder Laufzeit können die freistehende Build-Option verwenden.

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files
```

`--freestanding` verweist auf einen Build-Plan, der Standardbibliotheksabhängigkeiten deaktiviert, und `--entry` legt das Linker-Eintragssymbol fest. Um eine tatsächlich bootfähige Ausgabe zu erstellen, müssen Sie die Zielarchitektur, das Linker-Skript und sogar das Objektformat entwerfen.

## Einstiegspunkte, die absichtlich scheitern

Sie können in main keinen Parameter haben, selbst wenn dieser einen Standardwert hat. Es ist ein Fehler, die folgende Datei check zu kopieren.

<!-- wave-example: reject-main-argument -->
```wave
fun main(value: i32 = 0) -> i32 {
    return value;
}
```
