---
translation_set_id: vex-package-manager
path: toolchain/vex-package-manager
locale: ko
group: toolchain
group_order: 4
order: 3
title: Vex 패키지 매니저
summary: manifest 기반 Wave 프로젝트, Git·경로 의존성, lockfile, 오프라인 빌드와 wavec 경계를 설명합니다.
---

## 역할

Vex는 Wave의 패키지 매니저이자 빌드 도구입니다. Vex는 `wavec` 위에서 동작합니다. 프로젝트 구조와 의존성 해석은 Vex가 담당하고, 컴파일러 플래그와 컴파일 파이프라인은 `wavec`가 담당합니다.

Vex 명령은 manifest 기반입니다. `vex build`, `vex check`, `vex run`은 raw `wavec` 플래그를 의도적으로 받지 않습니다.

## 패키지 만들기

```shell
vex init
vex init --lib
```

애플리케이션은 `src/main.wave`, 라이브러리는 `src/lib.wave`를 사용합니다. 패키지 루트 구조는 다음과 같습니다.

```text
my_project/
├── src/
│   └── main.wave
├── vex.ws
├── vex.lock
└── .vex/
    └── deps/
```

`vex.ws`가 manifest입니다. Vex는 `.wson` 확장자의 manifest를 사용하지 않습니다.

```wson
{
    name = "my_project",
    version = 0.1.0,
    lib = false,
    description = "my_project Project",
    author = "unknown",
    license = "Unknown",
    dependencies = []
}
```

## 빌드 명령

```shell
vex build [--target <triple>] [--release] [--dry-run] [--locked] [--offline]
vex check [--target <triple>] [--release] [--dry-run] [--locked] [--offline]
vex run   [--target <triple>] [--release] [--dry-run] [--locked] [--offline] [-- <args...>]
```

Vex의 옵션은 작게 유지됩니다. emit, linker, CPU, ABI 또는 debug처럼 컴파일러에 종속적인 제어가 필요하면 `wavec`를 직접 사용하십시오. 특정 컴파일러를 사용해야 할 때는 `VEX_WAVEC=/path/to/wavec`를 설정합니다.

`Resolving`, `Fetching`, `Compiling`, `Checking`, `Running`, `Finished` 같은 진행 단계는 stderr에 출력되고 프로그램 출력은 stdout에 유지됩니다.

## Git 중심 의존성

Vex 의존성은 로컬 `path` 또는 Git URL로 지정합니다. 의존성 하나에는 두 방식 중 하나만 사용할 수 있습니다.

```wson
{
    name = "app",
    version = 0.1.0,
    dependencies = [
        { name = "local_math", path = "../local_math" },
        { name = "remote_math", git = "https://github.com/example/math.git", tag = "v0.1.0" }
    ]
}
```

Git 의존성에는 `branch`, `tag`, `rev` 중 최대 하나만 지정할 수 있습니다. 모든 의존성 루트에는 자체 `vex.ws`가 있어야 합니다. Vex는 의존성 manifest를 재귀적으로 해석하고 충돌하는 패키지 정체성을 거부하며, 관리하는 Git checkout을 `.vex/deps/<name>`에 저장합니다.

## lockfile 계약

스키마 v2 `vex.lock`은 전체 전이 의존성 그래프와 정확한 Git commit을 기록합니다. manifest와 함께 커밋하십시오. 같은 manifest와 유효한 lockfile을 사용하면 branch나 tag를 다시 따라가지 않고 같은 의존성 그래프를 선택합니다.

의존성이 필요한 명령은 자동으로 해석하며 다음 명령으로 미리 준비할 수도 있습니다.

```shell
vex fetch
vex update
vex update math shared_core
```

`vex update`는 모든 Git 패키지를 갱신하거나, 지정한 패키지와 영향을 받는 전이 그래프만 갱신합니다. 관련 없는 잠긴 패키지는 선택된 commit을 유지합니다.

## locked와 offline 작업 흐름

`--locked`는 `vex.lock` 생성과 변경을 금지합니다. 파일이 없거나 지원하지 않는 스키마이거나 manifest 그래프와 맞지 않으면 실패합니다. lockfile에 이미 고정된 commit은 필요할 때 가져올 수 있습니다.

`--offline`은 모든 Git 네트워크 작업을 금지합니다. 필요한 checkout과 commit이 로컬에 이미 있어야 합니다.

```shell
vex fetch --locked
vex build --locked --offline
```

이 두 명령은 엄격한 CI 작업 흐름입니다. 네트워크를 쓸 수 있을 때 정확히 잠긴 commit을 준비하고, 그다음 네트워크나 lockfile 변경 없이 컴파일합니다. dry-run은 의존성을 가져오거나 lockfile을 다시 쓰지 않습니다.

## 컴파일러 설정과 정보

```shell
vex info
vex setup wavec
vex setup wavec --version <version>
vex --version
```

Vex는 실제 빌드 전에 `wavec` dry-run JSON 스키마를 검증합니다. 필요한 스키마를 구현하지 않은 컴파일러는 알 수 없는 계획으로 실행하지 않고 호환성 오류로 거부합니다.
