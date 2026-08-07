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

Unix 계열 환경에서는 공식 설치 스크립트로 최신 릴리스를 설치할 수 있습니다.

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash -s -- latest
```

설치가 끝난 뒤 현재 셸에 PATH가 반영되지 않았다면 설치 프로그램이 출력한 안내를 적용하거나 새 셸을 여십시오.

## 설치 확인

```shell
wavec --version
wavec --help
```

문서와 정확히 같은 동작을 확인하려면 `v0.2.0-pre-beta`가 설치되어 있는지 확인합니다.

## v0.2.0-pre-beta 제공 바이너리

이 릴리스의 GitHub Release에는 다음 사전 빌드 아카이브가 게시되어 있습니다.

| 플랫폼 | 릴리스 아카이브 |
| --- | --- |
| Linux x86-64 GNU | `wave-v0.2.0-pre-beta-x86_64-linux-gnu.tar.gz` |
| Windows x86-64 GNU | `wave-v0.2.0-pre-beta-x86_64-pc-windows-gnu.zip` |
| macOS Apple Silicon | `wave-v0.2.0-pre-beta-aarch64-apple-darwin.tar.gz` |

릴리스에는 `SHA256SUMS`도 함께 제공되므로 직접 아카이브를 내려받는 경우 체크섬을 검증할 수 있습니다.

## 소스에서 빌드

컴파일러 저장소를 직접 빌드하려면 Rust 툴체인이 필요합니다.

```shell
git clone https://github.com/wavefnd/Wave.git
cd Wave
git checkout v0.2.0-pre-beta
cargo build
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
> 보안 정책상 원격 스크립트를 바로 셸에 전달할 수 없는 환경에서는 설치 스크립트를 먼저 내려받아 내용을 검토한 뒤 실행하십시오.
