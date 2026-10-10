---
translation_set_id: program-structure
path: language/program-structure
locale: ko
group: language
group_order: 2
order: 1
title: 1. 소스 파일에서 실행 결과까지
summary: 소스 파일, 함수, 출력, 검사와 실행을 배웁니다.
---

## 이 장을 시작하기 전에

[설치 안내](/docs/ko/getting-started/install)에 따라 컴파일러와 표준 라이브러리를 준비하십시오. 터미널에서 `wavec --version`을 실행할 수 있으면 시작할 수 있습니다. 편집기는 일반 텍스트 파일을 저장할 수 있는 것이면 됩니다.

이 장에서는 출력 한 줄을 만드는 데서 출발해 소스 파일, 함수, 컴파일, 실행, 종료 코드의 관계를 익힙니다. 명령을 복사하는 것뿐 아니라 어느 단계에서 무엇이 일어나는지 설명할 수 있는 것이 목표입니다.

## 작업 디렉터리 만들기

프로그램마다 별도 디렉터리를 사용하면 소스와 생성된 실행 파일을 찾기 쉽습니다. 터미널에서 디렉터리를 만들고 이동합니다.

```shell
mkdir wave-study
cd wave-study
```

편집기로 이 디렉터리에 `main.wave`를 만드십시오. 파일 이름이 `main.wave.txt`가 되지 않도록 확장자를 확인합니다. 아래는 함수 내부 조각이 아니라 파일 전체입니다.

<!-- wave-example: book-hello -->
```wave playground
fun main() {
    println("Hello, Wave!");
}
```

실행 결과:

```text
Hello, Wave!
```

다음 명령으로 실행합니다.

```shell
wavec run main.wave
```

터미널 명령과 Wave 코드를 섞지 마십시오. `wavec run`은 터미널에 입력하고, `fun main`은 소스 파일에 작성합니다. 프로그램이 출력한 `Hello, Wave!`를 다시 소스에 붙여 넣는 것도 아닙니다.

## 한 줄씩 읽기

`fun`은 함수를 선언하는 키워드입니다. 함수는 이름을 붙인 작업 묶음이고, `main`은 실행 파일의 시작점입니다. 다른 함수는 나중에 직접 정의합니다.

`main` 뒤의 `()`는 매개변수를 적는 자리입니다. 이 프로그램의 main은 매개변수를 받지 않으므로 비어 있습니다. `{`와 `}` 사이에 함수의 작업을 적습니다. 들여쓰기는 사람이 블록을 읽기 쉽게 해 주며, 블록 경계 자체는 중괄호가 나타냅니다.

`println("Hello, Wave!");`는 문자열을 출력하는 문장입니다. 큰따옴표는 문자열 리터럴의 시작과 끝을 나타냅니다. 따옴표 자체가 출력에 포함되는 것은 아닙니다. 세미콜론은 이 문장이 끝났음을 나타냅니다.

## 문장은 작성한 순서대로 실행됨

출력을 세 번 하도록 바꿔 봅니다. 앞 프로그램과 합치지 말고 main.wave의 내용을 아래 전체 프로그램으로 교체합니다.

<!-- wave-example: book-sequence -->
```wave playground
fun main() {
    println("start");
    println("working");
    println("done");
}
```

실행 결과:

```text
start
working
done
```

첫 문장을 마친 뒤 다음 문장으로 이동합니다. 여기에는 동시에 실행되는 작업이 없습니다. 출력 순서를 바꾸고 싶다면 문장의 순서를 바꾸면 됩니다. 뒤에서 조건문과 반복문을 배우면 이 진행 순서를 제어할 수 있습니다.

## print와 println

`println`은 마지막에 줄바꿈을 추가합니다. `print`는 자동으로 줄을 바꾸지 않습니다. 작은 조각을 이어서 한 줄을 만드는 데 차이가 드러납니다.

<!-- wave-example: book-print-lines -->
```wave playground
fun main() {
    print("Wave");
    print(" ");
    println("study");
    println("second line");
}
```

실행 결과:

```text
Wave study
second line
```

세 번 출력했다고 항상 세 줄이 생기는 것은 아닙니다. 출력 호출 수와 줄 수를 구분하십시오. 줄바꿈을 문자열 안에 직접 적을 때는 `\n` escape를 사용합니다. 문자열 escape와 소스의 줄바꿈은 [문자열 장](/docs/ko/language/strings)에서 자세히 다룹니다.

## 값을 문자열 안에 넣기

계산 결과를 문자열에 넣으려면 `{}` 자리에 대응하는 값을 전달합니다.

<!-- wave-example: book-format-first -->
```wave playground
fun main() {
    println("{} + {} = {}", 2, 3, 2 + 3);
}
```

실행 결과:

```text
2 + 3 = 5
```

첫 `{}`에 2, 두 번째에 3, 세 번째에 계산 결과 5가 들어갑니다. 포맷 문자열과 값은 쉼표로 구분합니다. 값의 개수를 바꾸면 placeholder 개수도 맞춰야 합니다. 이것은 문자열을 더하는 연산과 다른 출력 문법입니다.

## 검사·빌드·실행을 나누기

지금까지 사용한 run은 빌드와 실행을 이어서 수행합니다. 어느 단계가 실패했는지 알고 싶을 때는 다음처럼 나눌 수 있습니다.

```shell
wavec check main.wave
wavec build main.wave -o hello
```

check는 문법과 타입 등을 검사하지만 프로그램의 모든 입력을 시험하지는 않습니다. 예를 들어 파일이 실행 시 존재하는지는 소스 검사만으로 알 수 없습니다. build는 실행 파일을 생성합니다. Linux/macOS에서는 다음과 같이 실행합니다.

```shell
./hello
```

Windows에서는 출력 파일 이름을 `hello.exe`로 지정하고 PowerShell에서 `.\hello.exe`로 실행합니다. 실행 파일을 생성한 뒤 소스를 수정했다면 다시 빌드해야 바뀐 내용이 반영됩니다.

## 종료 코드도 결과임

사람은 출력 문장을 읽지만, 셸이나 다른 프로그램은 종료 코드로 성공 여부를 판단할 수 있습니다. 다음 main은 i32를 반환한다고 명시합니다.

<!-- wave-example: book-exit-success -->
```wave playground
fun main() -> i32 {
    println("completed");
    return 0;
}
```

실행 결과:

```text
completed
```

0은 정상 종료를 나타내는 관례입니다. 직접 실패를 알릴 때는 0이 아닌 코드를 반환합니다. Linux/macOS 셸에서는 실행 직후 `echo $?`, PowerShell에서는 `$LASTEXITCODE`로 코드를 확인합니다. 그 사이에 다른 명령을 실행하면 확인하는 대상이 달라질 수 있습니다.

`return`을 실행하면 함수가 끝납니다. main에는 매개변수를 선언하지 않으며 기본값이 있어도 허용되지 않습니다. main의 반환 타입은 생략하거나 i32를 사용합니다.

## 첫 오류를 읽는 방법

다음 코드는 의도적으로 잘못되었습니다. 실행 예제와 달리 check에서 실패해야 합니다.

```wave
fun main() {
    println("hello")
}
```

문장 끝의 세미콜론이 없습니다. 진단이 표시하는 줄과 바로 앞 문장을 살펴보십시오. 컴파일러가 표시한 위치는 오류가 시작된 지점이 아니라 잘못된 구조를 더 이상 해석할 수 없게 된 지점일 수도 있습니다.

처음에는 첫 오류 하나를 고친 뒤 다시 검사합니다. 앞부분의 괄호나 따옴표가 닫히지 않으면 뒤의 정상 코드에서도 여러 오류가 이어질 수 있습니다. 모든 줄을 동시에 고치려 하면 원래 원인을 놓치기 쉽습니다.

## 연습 문제

1. 자기소개를 세 줄 출력하는 프로그램을 만드십시오.
2. 12와 8을 포맷 인자로 전달해 `12 * 8 = 96`을 출력하십시오.
3. 성공 메시지를 출력하고 종료 코드 0을 반환하도록 작성하십시오.
4. print만 세 번 사용하면 출력 줄 수가 어떻게 되는지 실행 전에 예측하십시오.

### 풀이: 계산 과정 출력

<!-- wave-example: book-first-solution -->
```wave playground
fun main() -> i32 {
    println("Learning Wave");
    println("My first program");
    println("Ready to calculate");
    println("{} * {} = {}", 12, 8, 12 * 8);
    return 0;
}
```

실행 결과:

```text
Learning Wave
My first program
Ready to calculate
12 * 8 = 96
```

계산 결과를 문자열에 직접 `96`이라고 적는 대신 식으로 전달했습니다. 입력 값이 바뀌어도 계산 결과와 표시가 따로 놀지 않게 하는 첫 단계입니다. 다음 장에서는 같은 값을 여러 번 적지 않도록 변수에 이름을 붙입니다.


## 소스 파일의 최상위 항목

Wave 소스는 다음과 같은 최상위 항목을 구성할 수 있습니다.

- `import(...)`
- `const`, `static`
- `type`
- `struct`, `enum`, `proto`
- `extern(...)`, `export(...)`
- `fun`
- 지원되는 항목 앞의 `#[target(...)]` 조건

가져올 수 있는 선언 앞에는 `pub`을 붙입니다. 지역 `var` 선언은 함수나 블록 안에 둡니다.

## 프리스탠딩 프로그램

커널, 부트 코드 또는 런타임이 없는 대상은 프리스탠딩 빌드 옵션을 사용할 수 있습니다.

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files
```

`--freestanding`은 기본 라이브러리 의존을 끄는 빌드 계획과 연결되며, `--entry`는 링커 엔트리 심볼을 설정합니다. 실제 부팅 가능한 산출물을 만들려면 대상 아키텍처, 링커 스크립트와 오브젝트 형식까지 함께 설계해야 합니다.

## 의도적으로 실패하는 진입점

기본값이 있더라도 main에 매개변수를 둘 수 없습니다. 다음 파일을 check하면 오류입니다.

<!-- wave-example: reject-main-argument -->
```wave
fun main(value: i32 = 0) -> i32 {
    return value;
}
```
