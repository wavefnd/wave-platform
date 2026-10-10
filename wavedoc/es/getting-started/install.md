---
translation_set_id: install
path: getting-started/install
locale: es
group: getting-started
group_order: 1
order: 2
title: Instalar Wave
summary: Instala Wave en Linux, macOS o Windows y ejecuta tu primer programa.
---

## Política de instalación

Los instaladores solo instalan la última versión pública, incluidas las versiones preliminares con número de versión. Se excluyen Draft y Nightly. Para versiones anteriores y Nightly, siga la instalación manual. Si la última versión no tiene un paquete para su plataforma, la instalación se detiene sin elegir una versión anterior.

| Sistema operativo | Arquitecturas |
| --- | --- |
| Linux | amd64, arm64, riscv64, loong64 |
| macOS | amd64, arm64 |
| Windows | amd64, arm64 |
| FreeBSD | amd64 |

## Linux, macOS y FreeBSD

Se requieren Bash, curl, jq, tar y una herramienta SHA-256. En FreeBSD, instale primero bash, curl y jq. Linux requiere glibc y bibliotecas del sistema compatibles; macOS requiere Apple Command Line Tools; FreeBSD requiere un sistema base compatible.

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash
```

## Windows

Ejecute en PowerShell. El instalador detecta amd64 o arm64 y selecciona el paquete MSVC. Se requieren el runtime de Visual C++, Windows SDK y las bibliotecas MSVC/UCRT. Una consola de desarrollador de Visual Studio puede proporcionar el entorno para comprobar la compilación.

```powershell
irm https://wave-lang.dev/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

## Opciones de instalación

De forma predeterminada se instala Wave y la última versión de Vex si existe un paquete para la plataforma. Si no existe, solo se instala Wave y se informa de ello. Los errores de descarga o verificación de Vex no se omiten.

| | Opción Bash | Opción PowerShell |
| --- | --- | --- |
| Exigir Vex | `--with-vex` | `-WithVex` |
| Instalar solo Wave | `--without-vex` | `-WithoutVex` |
| No modificar PATH | `--no-modify-path` | `-NoModifyPath` |

Use el mismo comando para actualizar. La instalación anterior se conserva hasta verificar el nuevo compilador y la std incluida; si falla, se restaura. Las rutas predeterminadas son ~/.wave/bin en Unix y %LOCALAPPDATA%\Wave\bin en Windows. WAVE_INSTALL_DIR permite elegir un directorio exclusivo. Abra una terminal nueva tras configurar PATH.

## Instalar manualmente versiones anteriores y Nightly

Descargue el .tar.gz o .zip de la versión, sistema y arquitectura deseados desde [GitHub Releases](https://github.com/wavefnd/Wave/releases). Extraiga conservando las rutas relativas de wavec, std, llvm y los demás archivos incluidos. Añada el directorio de wavec a PATH. Consulte las notas de cada versión anterior para conocer su estructura y requisitos. No se admiten --version, --vex-version, -Version ni -VexVersion.

## Primera ejecución

El instalador compila y ejecuta un programa pequeño con la std incluida. No necesita descargar otra std. Ejecute el ejemplo siguiente tras instalar. Si se instaló Vex, compruébelo también con vex --version.

```shell
wavec --version
```

Cree un archivo de texto sin formato llamado `main.wave` en el directorio que prefiera y guarde el programa completo que aparece a continuación. Compruebe que el nombre no sea `main.wave.txt`.

<!-- wave-example: install-stdlib -->
```wave
import("std::string::len")::{
    len
};

fun main() {
    println("Wave: {} bytes", len("Wave"));
}
```

Abra una terminal en el directorio que contiene `main.wave` o acceda a él con `cd`. Después, ejecute:

```shell
wavec run main.wave
```

Salida esperada:

```text
Wave: 4 bytes
```
