---
translation_set_id: install
path: getting-started/install
locale: ko
group: getting-started
group_order: 1
order: 2
title: Wave 설치
summary: Linux, macOS, Windows에 Wave를 설치하고 첫 프로그램을 실행합니다.
---

## 설치 정책

설치기는 이 플랫폼에서 고정한 Wave 릴리스를 설치합니다. Draft와 Nightly는 자동 선택하지 않습니다. 다른 릴리스는 아래 수동 설치 절차를 사용하세요. 고정된 릴리스에 해당 플랫폼의 패키지가 없으면 다른 버전을 선택하지 않고 종료합니다.

| 운영체제 | 아키텍처 |
| --- | --- |
| Linux | amd64, arm64, riscv64, loong64 |
| macOS | amd64, arm64 |
| Windows | amd64, arm64 |
| FreeBSD | amd64 |

## Linux, macOS, FreeBSD

Bash, curl, jq, tar와 SHA-256 도구가 필요합니다. FreeBSD에서는 bash, curl, jq를 먼저 설치하세요. Linux는 호환되는 glibc와 시스템 라이브러리, macOS는 Apple Command Line Tools, FreeBSD는 패키지와 호환되는 기본 시스템이 필요합니다.

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash
```

## Windows

PowerShell에서 실행합니다. Windows amd64와 arm64를 자동 구분하며 MSVC 패키지를 설치합니다. Visual C++ 런타임, Windows SDK, MSVC/UCRT 라이브러리가 필요합니다. 컴파일 검증에 필요한 환경은 Visual Studio 개발자 셸에서도 구성할 수 있습니다.

```powershell
irm https://wave-lang.dev/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

## 설치 옵션

기본적으로 Wave와 해당 플랫폼에서 제공되는 최신 Vex를 설치합니다. Vex 패키지가 없으면 Wave만 설치하고 안내합니다. Vex 다운로드나 검증 실패는 생략하지 않고 오류로 처리합니다.

| | Bash 옵션 | PowerShell 옵션 |
| --- | --- | --- |
| Vex 필수 | `--with-vex` | `-WithVex` |
| Wave만 설치 | `--without-vex` | `-WithoutVex` |
| PATH 수정 안 함 | `--no-modify-path` | `-NoModifyPath` |

같은 명령으로 업데이트합니다. 기존 설치는 새 compiler·std 실행 검증이 성공한 뒤 정리하며, 실패하면 복구합니다. 기본 경로는 Unix의 ~/.wave/bin, Windows의 %LOCALAPPDATA%\Wave\bin입니다. 전용 설치 경로는 WAVE_INSTALL_DIR로 지정할 수 있습니다. PATH를 자동 등록한 경우 새 터미널을 여세요.

## 이전 버전과 Nightly 수동 설치

[GitHub Releases](https://github.com/wavefnd/Wave/releases)에서 원하는 버전과 OS·아키텍처의 .tar.gz 또는 .zip을 다운로드합니다. 압축을 풀고 wavec, std, llvm 및 나머지 동봉 파일의 상대 경로를 유지하세요. wavec가 있는 디렉터리를 PATH에 추가합니다. 이전 릴리즈의 구성과 요구 사항은 해당 릴리즈 안내를 따르세요. 설치기의 --version, --vex-version, -Version, -VexVersion 옵션은 지원하지 않습니다.

## 첫 실행

설치기는 동봉 std를 사용하는 작은 프로그램을 컴파일하고 실행합니다. 별도로 최신 std를 내려받을 필요가 없습니다. 설치 후 다음 예제를 실행하세요. Vex가 함께 설치된 경우 vex --version으로도 확인할 수 있습니다.

```shell
wavec --version
```

원하는 디렉터리에 일반 텍스트 파일 `main.wave`를 만들고 아래 전체 프로그램을 저장하세요. 파일 이름이 `main.wave.txt`가 아닌지 확인하세요.

<!-- wave-example: install-stdlib -->
```wave
import("std::string::len")::{
    len
};

fun main() {
    println("Wave: {} bytes", len("Wave"));
}
```

`main.wave`가 있는 디렉터리에서 터미널을 열거나 `cd`로 이동한 뒤 다음 명령을 실행하세요:

```shell
wavec run main.wave
```

예상 출력:

```text
Wave: 4 bytes
```
