---
translation_set_id: overview
path: getting-started/overview
locale: ko
group: getting-started
group_order: 1
order: 1
title: Wave 언어 개요
summary: Wave의 성격과 기본 프로그램 구조를 빠르게 파악합니다.
---

## Wave란

Wave는 정적 타입과 네이티브 코드 생성을 기반으로 저수준 제어를 제공하는 시스템 프로그래밍 언어입니다. 타입, 메모리 접근과 외부 ABI 경계를 코드에 명시하여 프로그램의 동작과 데이터 표현을 분명하게 드러냅니다.

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

Wave의 함수 인자와 반환값은 타입을 명시합니다. 지역 변수는 타입을 직접 적거나 초깃값에서 추론할 수 있습니다.

```wave
fun add(left: i32, right: i32) -> i32 {
    var result: i32 = left + right;
    return result;
}

fun main() {
    var count: i32 = 1;
    var next = count + 1;
    count += 1;
    println("count = {}, next = {}", count, next);
}
```

`var`는 지역 변수를 선언하는 문법입니다. `var name: Type = value;`는 타입을 직접 지정하고, `var name = value;`는 초깃값의 타입을 사용합니다. `const`와 `static`은 최상위 선언입니다.

## 콘솔 입출력

`print`, `println`, `input`은 Wave의 콘솔 입출력 문장입니다. 첫 번째 인자는 문자열 리터럴이어야 하며 `{}` 자리표시자의 수와 뒤따르는 인자의 수가 일치해야 합니다.

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
5. 필요할 때 컴파일러, Vex, Whale, 대상과 빠른 참조 문서를 사용합니다.
