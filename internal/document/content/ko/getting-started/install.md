---
translation_set_id: install
path: getting-started/install
locale: ko
group: getting-started
group_order: 1
order: 2
title: Wave 설치
summary: 공식 설치 스크립트, 릴리스 아카이브와 소스 빌드 방법을 설명합니다.
---

## 공식 설치 스크립트

Linux와 macOS에서는 공식 셸 설치 프로그램으로 최신 릴리스를 설치할 수 있습니다.

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash -s -- latest
```

Windows x86-64에서는 PowerShell 설치 프로그램을 실행합니다.

```powershell
irm https://wave-lang.dev/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1 -Latest
```

설치가 끝난 뒤 현재 셸에 PATH가 반영되지 않았다면 설치 프로그램이 출력한 안내를 적용하거나 새 셸을 여십시오. Windows 설치 프로그램은 기본적으로 `%LOCALAPPDATA%\Wave\bin`을 사용하고 사용자 PATH에 이 경로를 추가합니다.

## 설치 확인

```shell
wavec --version
wavec --help
```

문서와 다른 동작을 보고할 때 이 출력을 함께 기록하십시오. 문서는 현재 컴파일러 계약을 설명하므로 오래된 설치 바이너리는 더 작은 명령 또는 문법 범위를 제공할 수 있습니다.

## 릴리스 아카이브

릴리스 asset 이름에는 선택한 버전이 들어갑니다. 정확한 버전 문자열과 제공 플랫폼은 현재 릴리스 페이지에서 확인하십시오.

| 플랫폼 | 아카이브 이름 형태 |
| --- | --- |
| Linux x86-64 GNU | `wave-<version>-x86_64-linux-gnu.tar.gz` |
| Windows x86-64 GNU | `wave-<version>-x86_64-pc-windows-gnu.zip` |
| macOS Apple Silicon | `wave-<version>-aarch64-apple-darwin.tar.gz` |

릴리스에는 `SHA256SUMS`도 함께 제공되므로 직접 아카이브를 내려받는 경우 체크섬을 검증할 수 있습니다.

## 소스에서 빌드

컴파일러 저장소를 직접 빌드하려면 Rust 툴체인이 필요합니다.

```shell
git clone https://github.com/wavefnd/Wave.git
cd Wave
cargo build
```

위 명령은 기본 브랜치를 빌드합니다. 게시된 릴리스를 재현하려면 Cargo를 실행하기 전에 해당 태그를 checkout합니다.

```shell
git checkout <release-tag>
```

최적화된 컴파일러가 필요하면 Cargo의 release 프로필을 사용할 수 있습니다.

```shell
cargo build --release
```

빌드 산출물은 일반적으로 `target/debug/wavec` 또는 `target/release/wavec`에 생성됩니다.

## 설치 뒤 확인할 것

- `wavec --version`이 예상 버전을 출력하는지 확인합니다.
- `wavec print supported-targets`로 현재 컴파일러가 인식하는 대상 목록을 확인합니다.
- 표준 라이브러리 경로가 필요하면 `wavec print std-path`를 사용합니다.
- 셸에서 `wavec`를 찾지 못하면 PATH와 설치 위치를 먼저 확인합니다.

> **스크립트 검토**
>
> 보안 정책에서 요구하는 경우 `install.sh` 또는 `install.ps1`을 먼저 내려받아 내용을 검토한 뒤 실행하십시오.
