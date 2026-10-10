---
translation_set_id: install
path: getting-started/install
locale: de
group: getting-started
group_order: 1
order: 2
title: Wave installieren
summary: Installiere Wave unter Linux, macOS oder Windows und führe dein erstes Programm aus.
---

## Installationsrichtlinie

Die Installer installieren ausschließlich die neueste öffentlich veröffentlichte Version, einschließlich Vorabversionen mit Versionsnummer. Draft und Nightly sind ausgeschlossen. Ältere Versionen und Nightly werden wie unten beschrieben manuell installiert. Fehlt ein Paket für die Plattform, wird keine ältere Version ausgewählt.

| Betriebssystem | Architekturen |
| --- | --- |
| Linux | amd64, arm64, riscv64, loong64 |
| macOS | amd64, arm64 |
| Windows | amd64, arm64 |
| FreeBSD | amd64 |

## Linux, macOS und FreeBSD

Bash, curl, jq, tar und ein SHA-256-Werkzeug werden benötigt. Installieren Sie unter FreeBSD zuerst bash, curl und jq. Linux benötigt kompatible glibc- und Systembibliotheken, macOS die Apple Command Line Tools und FreeBSD ein kompatibles Basissystem.

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash
```

## Windows

In PowerShell ausführen. Der Installer erkennt amd64 oder arm64 und wählt das MSVC-Paket. Visual C++ Runtime, Windows SDK und MSVC/UCRT-Bibliotheken werden benötigt. Eine Visual-Studio-Entwicklershell kann die Umgebung für den Kompilierungstest bereitstellen.

```powershell
irm https://wave-lang.dev/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

## Installationsoptionen

Standardmäßig werden Wave und die neueste Vex-Version installiert, sofern ein passendes Vex-Paket existiert. Andernfalls wird nur Wave installiert und dies angezeigt. Download- oder Prüffehler bei Vex werden nicht stillschweigend übersprungen.

| | Bash-Option | PowerShell-Option |
| --- | --- | --- |
| Vex voraussetzen | `--with-vex` | `-WithVex` |
| Nur Wave installieren | `--without-vex` | `-WithoutVex` |
| PATH nicht ändern | `--no-modify-path` | `-NoModifyPath` |

Zum Aktualisieren denselben Befehl ausführen. Die bisherige Installation bleibt bis zur erfolgreichen Ausführungsprüfung des neuen Compilers mit enthaltener std erhalten und wird bei Fehlern wiederhergestellt. Standardpfade: ~/.wave/bin unter Unix und %LOCALAPPDATA%\Wave\bin unter Windows. WAVE_INSTALL_DIR legt ein eigenes Installationsverzeichnis fest. Nach der PATH-Konfiguration ein neues Terminal öffnen.

## Ältere Versionen und Nightly manuell installieren

Laden Sie das passende .tar.gz oder .zip für Version, System und Architektur von [GitHub Releases](https://github.com/wavefnd/Wave/releases). Behalten Sie beim Entpacken die relativen Pfade von wavec, std, llvm und allen weiteren Dateien bei. Fügen Sie das wavec-Verzeichnis zu PATH hinzu. Beachten Sie bei älteren Versionen deren Hinweise zu Struktur und Voraussetzungen. --version, --vex-version, -Version und -VexVersion werden nicht unterstützt.

## Erster Start

Der Installer kompiliert und startet ein kleines Programm mit der enthaltenen std. Ein separater Download der neuesten std ist nicht erforderlich. Führen Sie danach das folgende Beispiel aus. Falls Vex installiert wurde, kann es mit vex --version geprüft werden.

```shell
wavec --version
```

Erstellen Sie in einem Verzeichnis Ihrer Wahl eine reine Textdatei namens `main.wave` und speichern Sie darin das vollständige Programm unten. Achten Sie darauf, dass die Datei nicht `main.wave.txt` heißt.

<!-- wave-example: install-stdlib -->
```wave
import("std::string::len")::{
    len
};

fun main() {
    println("Wave: {} bytes", len("Wave"));
}
```

Öffnen Sie ein Terminal im Verzeichnis mit `main.wave` oder wechseln Sie mit `cd` dorthin. Führen Sie dann Folgendes aus:

```shell
wavec run main.wave
```

Erwartete Ausgabe:

```text
Wave: 4 bytes
```
