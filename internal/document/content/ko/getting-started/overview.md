---
translation_set_id: overview
path: getting-started/overview
locale: ko
group: getting-started
group_order: 1
order: 1
title: Wave 언어 개요
summary: Wave 0.2.0-pre-beta의 성격, 기본 프로그램 구조와 문서 범위를 빠르게 파악합니다.
---

## 문서 기준

Wave는 정적 타입과 네이티브 코드 생성을 기반으로 저수준 제어를 제공하는 시스템 프로그래밍 언어입니다. 이 문서는 **Wave v0.2.0-pre-beta** 태그와 해당 태그의 커밋 `bd5549bd99a6cd8372b6542b4170a2221bac85d0`을 기준으로 합니다.

릴리스 이후 `master`에서 추가되거나 변경된 문법은 이 문서의 보장 범위에 포함하지 않습니다. 예제가 현재 설치된 컴파일러와 다르게 동작한다면 먼저 `wavec --version`으로 버전을 확인하십시오.

## 첫 프로그램

```wave
fun main() {
    println("Hello, Wave!");
}
```

소스를 `main.wave`로 저장한 뒤 다음과 같이 실행할 수 있습니다.

```shell
wavec run main.wave
```

실행 파일만 만들려면 `build`를 사용합니다.

```shell
wavec build main.wave -o app
```

일반적인 호스트 실행 파일은 `main` 함수를 진입점으로 사용합니다. 프리스탠딩 환경은 별도의 진입 심볼과 링크 설정을 지정할 수 있습니다.

## 문장의 기본 형태

Wave의 변수와 함수 인자는 타입을 명시합니다.

```wave
fun add(left: i32, right: i32) -> i32 {
    let result: i32 = left + right;
    return result;
}

fun main() {
    var count: i32 = 1;
    count += 1;
    println("count = {}", count);
}
```

`var`는 재대입 가능한 지역 변수이고 `let`은 재대입할 수 없는 지역 바인딩입니다. `let mut`은 명시적으로 가변인 `let` 바인딩입니다. `const`와 `static`은 이 릴리스에서 최상위 선언으로 사용합니다.

## 콘솔 입출력

v0.2.0-pre-beta의 `print`, `println`, `input`은 일반적인 표준 라이브러리 함수 호출이 아니라 파서가 직접 인식하는 입출력 문장입니다. 첫 번째 인자는 문자열 리터럴이어야 하며 `{}` 자리표시자의 수와 뒤따르는 인자의 수가 일치해야 합니다.

```wave
fun main() {
    var value: i32 = 0;
    input("{}", value);
    print("value = ");
    println("{}", value);
}
```

## 저수준 기능

Wave는 `ptr<T>`, 주소 연산자 `&`, 명시적 역참조 `deref`, C ABI용 `extern(c)`와 `export(c)`, 인라인 `asm`을 제공합니다. 이런 기능은 자동 소유권이나 경계 검사를 대신하지 않으므로 포인터의 유효성, 크기, 정렬과 수명은 호출 계약에 맞게 관리해야 합니다.

## 권장 학습 순서

1. 설치 후 `wavec --version`과 `wavec --help`를 확인합니다.
2. 선언과 타입, 표현식, 제어 흐름을 익힙니다.
3. 함수와 제네릭, 구조체·열거형·`proto`를 익힙니다.
4. 포인터, import, FFI와 표준 라이브러리를 학습합니다.
5. 필요할 때 툴체인·빠른 참조 문서를 사용합니다.
