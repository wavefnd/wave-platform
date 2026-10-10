---
translation_set_id: install
path: getting-started/install
locale: ru
group: getting-started
group_order: 1
order: 2
title: Установка Wave
summary: Установите Wave в Linux, macOS или Windows и запустите первую программу.
---

## Правила установки

Установщики устанавливают только последний публичный выпуск с номером версии, включая предварительные версии. Draft и Nightly исключены. Старые версии и Nightly устанавливаются вручную по инструкции ниже. Если в последнем выпуске нет пакета для платформы, установка завершается без перехода на старую версию.

| Операционная система | Архитектуры |
| --- | --- |
| Linux | amd64, arm64, riscv64, loong64 |
| macOS | amd64, arm64 |
| Windows | amd64, arm64 |
| FreeBSD | amd64 |

## Linux, macOS и FreeBSD

Требуются Bash, curl, jq, tar и средство SHA-256. В FreeBSD сначала установите bash, curl и jq. Linux нужны совместимые glibc и системные библиотеки; macOS — Apple Command Line Tools; FreeBSD — совместимая базовая система.

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash
```

## Windows

Запустите в PowerShell. Установщик определяет amd64 или arm64 и выбирает пакет MSVC. Нужны среда Visual C++, Windows SDK и библиотеки MSVC/UCRT. Среду для проверки компиляции можно настроить через оболочку разработчика Visual Studio.

```powershell
irm https://wave-lang.dev/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

## Параметры установки

По умолчанию устанавливаются Wave и последний Vex, если для платформы есть пакет Vex. Иначе устанавливается только Wave с явным уведомлением. Ошибки загрузки или проверки Vex не игнорируются.

| | Параметр Bash | Параметр PowerShell |
| --- | --- | --- |
| Требовать Vex | `--with-vex` | `-WithVex` |
| Установить только Wave | `--without-vex` | `-WithoutVex` |
| Не менять PATH | `--no-modify-path` | `-NoModifyPath` |

Для обновления повторите команду. Предыдущая установка сохраняется до успешной проверки нового компилятора и встроенной std; при ошибке она восстанавливается. Пути по умолчанию: ~/.wave/bin в Unix и %LOCALAPPDATA%\Wave\bin в Windows. WAVE_INSTALL_DIR задаёт отдельный каталог установки. После настройки PATH откройте новый терминал.

## Ручная установка старых версий и Nightly

Скачайте .tar.gz или .zip нужной версии, ОС и архитектуры из [GitHub Releases](https://github.com/wavefnd/Wave/releases). Распакуйте, сохранив относительные пути wavec, std, llvm и остальных файлов. Добавьте каталог wavec в PATH. Для старых версий следуйте их описанию структуры и требований. Параметры --version, --vex-version, -Version и -VexVersion не поддерживаются.

## Первый запуск

Установщик компилирует и запускает небольшую программу со встроенной std. Отдельно скачивать последнюю std не нужно. После установки запустите пример ниже. Если Vex установлен, его можно проверить командой vex --version.

```shell
wavec --version
```

Создайте в выбранном каталоге обычный текстовый файл `main.wave` и сохраните в нём всю программу ниже. Убедитесь, что файл не называется `main.wave.txt`.

<!-- wave-example: install-stdlib -->
```wave
import("std::string::len")::{
    len
};

fun main() {
    println("Wave: {} bytes", len("Wave"));
}
```

Откройте терминал в каталоге с `main.wave` или перейдите туда командой `cd`, затем выполните:

```shell
wavec run main.wave
```

Ожидаемый вывод:

```text
Wave: 4 bytes
```
