---
translation_set_id: system-io
path: reference/system-io-network-process
locale: ko
group: reference
group_order: 3
order: 6
title: 시스템 입출력, 파일, 네트워크와 프로세스
summary: 릴리스의 fd, 파일 시스템, 소켓, TCP, UDP와 프로세스 모듈을 사용합니다.
---

## 파일 디스크립터와 파일

```wave
import("std::fs::file");
import("std::io::fd");

var fd: i64 = fs_open_read("input.txt");
if (fd >= 0) {
    fs_close(fd);
}
```

std::io는 디스크립터 단위 읽기, 쓰기, seek, 복사 도우미를 제공합니다. std::fs는 존재 확인, 열기, 크기, 전체 읽기·쓰기, 복사, 메타데이터, 디렉터리와 제거 연산을 제공합니다.

## 네트워크

std::net은 주소, poll, 기본 소켓, 소켓 옵션, TCP, UDP를 분리합니다. 네트워크 주소와 포트는 호스트·네트워크 바이트 순서 변환을 명시합니다.

## 프로세스

std::process는 핵심 프로세스 연산, 상수, spawn, wait, 표준 스트림 재지정을 제공합니다. 반환값은 플랫폼 방식 오류를 유지하므로 반드시 확인해야 합니다.

