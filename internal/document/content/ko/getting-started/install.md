---
translation_set_id: install
path: getting-started/install
locale: ko
group: getting-started
group_order: 1
order: 2
title: Wave 설치
summary: 릴리스된 컴파일러를 설치하고 툴체인을 확인합니다.
---

## 공식 설치 스크립트

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash -s -- latest
```

설치가 끝나면 새 셸을 열거나 설치 프로그램이 출력한 PATH 설정을 적용합니다.

## 설치 확인

```shell
wavec --version
wavec --help
```

> **보안**
> 
> 운영 환경의 정책이 요구한다면 셸로 전달하기 전에 설치 스크립트 내용을 검토하십시오.

## 소스 빌드

빌드하려는 정확한 Wave 소스 리비전에 포함된 안내를 따르십시오. 컴파일러 의존성과 지원 대상은 릴리스마다 달라질 수 있습니다.

