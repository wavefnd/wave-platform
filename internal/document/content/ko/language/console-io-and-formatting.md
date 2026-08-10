---
translation_set_id: console-io-formatting
path: language/console-io-and-formatting
locale: ko
group: language
group_order: 2
order: 13
title: 콘솔 입출력과 포매팅
summary: 파서가 직접 인식하는 print, println, input 문장과 자리표시자 규칙을 설명합니다.
---

## 입출력 문장

현재 Wave는 `print`, `println`, `input`을 언어 문장으로 인식합니다. 호출처럼 보이지만 일반 함수로 해석하지 않고 파서가 직접 처리합니다.

```wave
fun main() {
    var count: i32 = 0;
    input("{}", count);
    print("count = ");
    println("{}", count);
}
```

각 문장은 `;`로 끝납니다. 첫 번째 인자는 반드시 문자열 리터럴이어야 합니다. 변수나 계산한 문자열을 format 인자로 사용할 수 없습니다.

## 자리표시자

정확히 두 글자인 `{}`만 자리표시자입니다.

```wave
println("name = {}, score = {}", name, score);
```

파서는 자리표시자 수와 뒤따르는 식의 수가 정확히 같은지 검사합니다. 수가 다르면 컴파일 오류입니다.

```wave
println("{} {}", one);       // 오류: 자리표시자 2개, 값 1개
println("plain text", one); // 오류: 자리표시자 없음, 값이 남음
```

다른 형태의 중괄호는 일반 텍스트로 남습니다. 이 문법에는 이름이나 번호가 붙은 자리표시자가 없습니다.

## print와 println

`print`는 포매팅한 텍스트를 그대로 출력하고 `println`은 줄바꿈을 덧붙입니다.

```wave
print("loading...");
println("done");
```

코드 생성 단계에서는 Wave 타입을 기준으로 포맷을 결정합니다. 정수, 부동소수점, 포인터, 문자열 계열과 aggregate 값은 타입과 hosted C runtime 인터페이스에 맞게 낮아집니다.

## input 대상

runtime이 주소를 통해 읽은 값을 저장하므로 `input` format 뒤의 모든 식은 쓰기 가능한 lvalue여야 합니다.

```wave
var number: i32 = 0;
input("{}", number);
```

변수, 지원되는 field access와 역참조 형태를 대상으로 사용할 수 있습니다. 리터럴과 계산된 rvalue는 사용할 수 없습니다.

생성된 hosted 구현은 runtime이 변환한 필드 수를 확인합니다. 요청한 값을 모두 변환하지 못하면 프로그램은 실패 상태로 종료합니다.

## 런타임 경계

이 문장들은 현재 hosted C의 `printf`, `scanf` 계열 기능으로 낮아집니다. 따라서 일반 hosted runtime이 필요하며 프리스탠딩 환경에서 이식 가능한 입출력 수단은 아닙니다. 커널과 임베디드 프로그램은 대상별 입출력을 명시적 함수나 FFI 경계 뒤에 정의해야 합니다.
