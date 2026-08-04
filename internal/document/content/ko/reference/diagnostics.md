---
translation_set_id: diagnostics
path: reference/diagnostics
locale: ko
group: reference
group_order: 3
order: 2
title: 진단과 문제 해결
summary: 컴파일러 오류를 해석하고 재현 가능한 문제 보고서를 만듭니다.
---

## 진단 절차

- 연쇄 오류보다 첫 진단과 소스 범위를 먼저 읽습니다.
- wavec --version이 문서 버전과 일치하는지 확인합니다.
- 오류를 재현하는 가장 작은 완전한 소스 파일로 줄입니다.
- --emit=check로 프런트엔드 검사와 링크·실행 문제를 구분합니다.
- FFI 오류는 심볼, ABI, 라이브러리 경로, 대상 아키텍처를 각각 확인합니다.

## 좋은 버그 보고서

정확한 컴파일러 버전, 운영체제, 대상, 실행 명령, 최소 소스, 전체 진단, 기대 결과, 실제 결과를 포함하십시오. 비밀 정보와 개인 경로는 제거하십시오.

