---
translation_set_id: install
path: getting-started/install
locale: zh
group: getting-started
group_order: 1
order: 2
title: 安装 Wave
summary: 在 Linux、macOS 或 Windows 上安装 Wave，并运行第一个程序。
---

## 安装策略

安装器仅安装最新公开的版本化发布，包括带版本号的预发布版本。Draft 和 Nightly 不在自动选择范围内。旧版本与 Nightly 请按下文手动安装。若最新版本没有当前平台的包，安装器将退出，不会回退到旧版本。

| 操作系统 | 架构 |
| --- | --- |
| Linux | amd64, arm64, riscv64, loong64 |
| macOS | amd64, arm64 |
| Windows | amd64, arm64 |
| FreeBSD | amd64 |

## Linux、macOS 和 FreeBSD

需要 Bash、curl、jq、tar 和 SHA-256 工具。FreeBSD 请先安装 bash、curl 和 jq。Linux 需要兼容的 glibc 与系统库，macOS 需要 Apple Command Line Tools，FreeBSD 需要兼容的基本系统。

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash
```

## Windows

在 PowerShell 中运行。安装器自动识别 Windows amd64 或 arm64，并选择 MSVC 包。需要 Visual C++ 运行库、Windows SDK 和 MSVC/UCRT 库。也可通过 Visual Studio 开发者终端配置编译检查所需的环境。

```powershell
irm https://wave-lang.dev/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

## 安装选项

默认安装 Wave，并在当前平台存在包时安装最新的 Vex。若 Vex 没有对应包，则仅安装 Wave 并明确提示。Vex 下载或校验失败会报错，不会被静默忽略。

| | Bash 选项 | PowerShell 选项 |
| --- | --- | --- |
| 必须安装 Vex | `--with-vex` | `-WithVex` |
| 仅安装 Wave | `--without-vex` | `-WithoutVex` |
| 不修改 PATH | `--no-modify-path` | `-NoModifyPath` |

更新时运行相同命令。新编译器与附带 std 通过运行检查后才清理旧安装；失败则恢复。默认目录为 Unix 的 ~/.wave/bin 和 Windows 的 %LOCALAPPDATA%\Wave\bin。可通过 WAVE_INSTALL_DIR 指定专用安装目录。自动配置 PATH 后请打开新终端。

## 手动安装旧版本和 Nightly

从 [GitHub Releases](https://github.com/wavefnd/Wave/releases) 下载所需版本、操作系统及架构的 .tar.gz 或 .zip。解压时保留 wavec、std、llvm 及其他附带文件的相对路径，将 wavec 所在目录加入 PATH。旧版本的布局或依赖要求请参照该版本说明。安装器不支持 --version、--vex-version、-Version 和 -VexVersion 选项。

## 首次运行

安装器会使用附带 std 编译并运行一个小程序，无需另外下载最新 std。安装后可运行下例。如果同时安装了 Vex，也可使用 vex --version 检查。

```shell
wavec --version
```

在任意目录中创建名为 `main.wave` 的纯文本文件，保存下面的完整程序。请确认文件名不是 `main.wave.txt`。

<!-- wave-example: install-stdlib -->
```wave
import("std::string::len")::{
    len
};

fun main() {
    println("Wave: {} bytes", len("Wave"));
}
```

在包含 `main.wave` 的目录中打开终端，或使用 `cd` 切换到该目录，然后运行：

```shell
wavec run main.wave
```

预期输出：

```text
Wave: 4 bytes
```
