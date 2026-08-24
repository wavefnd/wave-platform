---
translation_set_id: system-io
path: reference/system-io-network-process
locale: ko
group: reference
group_order: 3
order: 6
title: 시스템 입출력, 파일, 네트워크와 프로세스
summary: fd 기반 I/O, 파일 시스템, 소켓/TCP/UDP와 프로세스 API의 실패 처리 방식을 설명합니다.
---

## 파일 열기와 닫기

```wave
import("std::fs::file");

fun main() {
    var fd: i64 = fs_open_read("input.txt");
    if (fd >= 0) {
        fs_close(fd);
    }
}
```

`fs_open_read`는 실패 시 음수 값을 그대로 반환할 수 있으므로 사용 전에 검사합니다. `fs_close`도 반환값을 가지므로 종료 오류를 중요하게 다루는 코드에서는 결과를 확인하십시오.

## std::io

`std::io`는 파일 디스크립터 기반의 읽기·쓰기, 정확한 길이 읽기/쓰기, seek와 복사 같은 동작을 제공합니다. 버퍼를 받는 API에서는 포인터와 길이를 항상 같은 단위로 관리해야 합니다.

## std::fs

`std::fs::file`에는 파일 존재 확인, 열기, 닫기, 크기 조회, 제거, 디렉터리 생성·제거, 전체 읽기/쓰기와 파일 복사 도우미가 포함되어 있습니다.

```wave
var size: i64 = fs_file_size("input.txt");
if (size < 0) {
    println("file error");
}
```

## 네트워크

`std::net`은 주소, 기본 소켓, 소켓 옵션, poll, TCP, UDP 영역으로 나뉩니다. 네트워크 코드를 작성할 때는 다음을 명시적으로 관리하십시오.

- 주소 패밀리와 소켓 종류
- 포트와 정수의 바이트 순서
- blocking/non-blocking 상태
- 부분 읽기와 부분 쓰기
- 종료와 오류 반환값

## 프로세스

`std::process`는 프로세스 생성, 대기, 상수와 표준 스트림 재지정에 필요한 소스 단위를 제공합니다. 자식 프로세스의 생성 성공 여부와 종료 상태를 별개로 처리해야 합니다.

## 플랫폼 의존성

이 모듈들은 내부적으로 `std::sys`와 네이티브 호출을 사용합니다. 특정 오류 코드나 플래그 값에 의존하는 프로그램은 대상 운영체제와 ABI를 명확히 정해 테스트하십시오.
