---
translation_set_id: whale-amd64-target
path: whale/amd64-target
locale: de
group: whale
group_order: 1
order: 10
title: AMD64-Ziele und ABI
summary: Linux AMD64 Beschreibt Zieleigenschaften, Aufrufgrenzen und native Funktionsumfang.
---

## Zielkennung

Die Zielkennung für Profil native ist `x86_64-whale-linux`.

|Attribut|Wert|
| --- | --- |
|Betriebssystem| Linux |
|Befehlssatz| AMD64 |
|Byte-Reihenfolge| Little endian |
|Native Adressbreite|64-Bit|
|Objektformat| ELF64 |
|C Anrufprotokoll| SysV AMD64 ABI |
|statisches ausführbares Dateiformat| ELF ET_EXEC |

Build-Host und Ausgabeziel sind unterschiedliche Konzepte. Nur weil Sie Whale auf einem anderen Host ausführen können, heißt das nicht, dass es den Befehlssatz oder das Objektformat dieses Hosts ausgeben kann. Den implementierten Kompilierungspfad finden Sie unter [Supportstatus](overview).

## Zielauswahl und -überprüfung

Der experimentelle Befehl `ir lower` verwendet `x86_64-whale-linux` als standardmäßiges und einziges unterstütztes Ziel. Um diesen Befehl zu verwenden, erstellen Sie ihn als `--features socket-cli`.

```sh
whale ir lower program.json --target x86_64-whale-linux
```

Das Ausgabeziel stellt unabhängig vom Build-Host ein 64-Bit-Little-Endian-Datenlayout bereit. Unbekannte Bezeichner oder nicht unterstützte Kombinationen wie „aarch64-whale-linux“ und „x86_64-whale-windows“ schlagen mit einem Fehler fehl, der das unterstützte Ziel anweist, bevor die Eingabe gelesen oder die Ausgabedatei ersetzt wird. Die Prüfung der Zielauswahl kann auch mit „--no-verify“ nicht ausgeschaltet werden.

Wenn Sie das Leerzeichen AST(`{"format_version":2,"semantics_version":1,"features":[],"program":{"declarations":[],"globals":[],"functions":[]}}`) eingeben, werden die folgenden Header und Module ausgegeben:

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

}
```

Wenn Sie ein nicht unterstütztes Ziel angeben, wird es im Status „Fehler“ beendet.

```sh
whale ir lower program.json --target banana
```

```text
Error: unsupported target "banana"; supported targets: x86_64-whale-linux
```

In Rust können Sie mit `ir::Target::lookup("x86_64-whale-linux")` das Ziel auswählen und `name()` und `data_layout()` nach `lower_o0` übertragen. Lowering lehnt ab, wenn das gelieferte Layout vom ausgewählten Ziel abweicht. Überprüfen Sie die selbstkonfigurierten Versionen IR und `verify_module` auf nicht unterstützte Zielnamen und Layout-Inkonsistenzen.

Das Objektmodell speichert `format`, `machine`, `endian` und `address_bits` in `ObjectTarget`. `ObjectFile::with_target` behält die angegebenen Identifikationsinformationen bei und die Serialisierung erlaubt nur die Kombination von AMD64·little-endian·64-Bit ELF64. Das Vorhandensein der Kennung machine für eine andere Architektur bedeutet nicht, dass dieser Encoder unterstützt wird. Beide Einstiegspunkte ELF writer prüfen Metadaten, und die Linker-Eingabe wird vor der Symbolauflösung oder Verknüpfung geprüft. Der ELF-Header für unterstützte Objekte behält `EM_X86_64` bei.

`ObjectFile::new(ObjectFormat::ELF64)` ist ein praktischer Konstruktor, der diesen AMD64-Bezeichner wie zuvor verwendet. Codes, die zuvor auf `object.format` zugegriffen haben, müssen `object.target.format` verwenden.

Diese Inspektion dient der Zielauswahl und Objektidentifizierung. Die Größe von Strukturen, Arrays usw., Feldern offset, stride kann mit IR Layout API gesucht werden. Das Lesen von Objektdateien, native ABI lowering, das Verknüpfen mit ausführbaren Dateien wird noch nicht unterstützt. Der Skalar lowering legt weiterhin die explizite Ausrichtung fest, und die vollständigen Layoutregeln für komplexe Typen sind in [Speicherreferenzdokument](memory-model) definiert.

## Rufen Sie an und unterschreiben Sie

Der IR-Deklarations- und Aufrufprüfer akzeptiert explizite `Whale`- und `SysV64`-Konventionen. Die Konvention ist Teil eines Typs `fnptr` und muss am Anrufort übereinstimmen. Variadische Signaturen und SysV64-Aggregatsignaturen werden abgelehnt. Siehe das [Beispiel zum Aufbau eines ausführbaren Aufrufs](ir-reference). Dadurch werden IR-Verträge validiert; Maschinen-ABI-Klassifizierung, Argumentregister/Stack-Platzierung und native Anrufausgabe sind noch nicht implementiert.

Unterstützte Aufrufe erfordern eine explizite Signatur und Aufrufkonvention. Eine nicht unterstützte Signatur ist ein Fehler. Das Backend darf keine ähnlichen Aufrufe durchführen, indem Argumente fehlen oder durch andere Ausdrücke ersetzt werden.

Die Signaturunterstützung ist in grundlegende Ganzzahlen/Zeiger, f32/f64, Strukturen/weite Werte und variable Argumente unterteilt. Nur weil es einen Typ im Typsystem IR gibt, heißt das nicht, dass die Übergabe oder Rückgabe von Argumenten dieses Typs in ABI unterstützt wird. Bei der Auswahl eines Backends sollten Sie prüfen, ob jedes die von Ihnen benötigten Kategorien unterstützt.

Beispielsweise sollten i32-Strukturrückgabesignaturen, die nicht auch vom Backend unterstützt werden, das die Rückgabe verarbeitet, abgelehnt werden. Die Skalarrückgabekonvention kann nicht angewendet werden, nur weil ein Teil der Struktur in ein Skalarregister geht.

## Interne Anrufe und C Grenzen

native Die Zeigeradresse beträgt 64 Bit. shadow metadata des Tracking-Zeigers müssen auch nach dem Kopieren, Speichern, Argumentieren und Zurückgeben gemeinsam übergeben werden.

C ABI und das interne Metadaten-Übermittlungsprotokoll sind unterschiedlich. C Grenzüberschreitende Anrufe erfordern einen expliziten Adapter. Nur weil Sie eine numerische Adresse an C ABI übergeben haben, sollte dies nicht als Beibehaltung der Zuweisungsidentität, der Lebensdauer, des Umfangs oder der Zugriffsrechte angesehen werden.

## Stapelrahmen

Behält den Frame-Zeiger bei und verwendet nicht red zone. In O0 wird der Speicherplatz verschiedener lokaler Variablen nicht wiederverwendet. Die Call-Frame-Debug-Informationen beschreiben die Entsprechung des native-Frames mit dem Originalprogramm.

Diese Regel dient nur Debugging-Zwecken und ist nicht zur Unterstützung der Ausnahme unwinding gedacht. Bitte beachten Sie [Debuggen mit O0](o0-debugging).

## Profilbeschränkungen

Die ersten O0 native-Profile enthalten die folgenden Funktionen nicht:

- O1 Optimierung, Vektorisierung, LTO.
- Gemeinsamer Speicher und atomic-Operationen.
- Ausnahme unwinding, async Funktion, coroutine.
- GC und sprachspezifische Besitzüberprüfung.
- Eigentumsübertragung eines beliebigen externen Speichers.
- Vollständige Unterstützung für Inline-Assembly-Einschränkungen.
- Dynamische Verknüpfung, TLS, zusätzlicher Befehlssatz, zusätzliche Objekttypen.

Anfragen, die außerhalb des Supportumfangs liegen, sollten abgelehnt werden und nicht stillschweigend durch eine andere Funktion ersetzt werden. Diese Einschränkung gilt für den Kompilierungspfad native und bedeutet nicht, dass andere unabhängig bereitgestellte Toolchain-Komponenten entfernt werden.
