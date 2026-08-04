---
translation_set_id: overview
path: getting-started/overview
locale: ko
group: getting-started
group_order: 1
order: 1
title: Wave 언어 개요
summary: Wave의 목적과 이 문서가 다루는 릴리스 범위를 설명합니다.
---

## 문서 범위

Wave는 명시적인 제어, 읽을 수 있는 저수준 코드, 네이티브 라이브러리 연동을 지향하는 정적 타입 시스템 프로그래밍 언어입니다. 이 문서는 소스 리비전 bd5549b에 해당하는 Wave 0.2.0-pre-beta를 기준으로 합니다.

> **릴리스 기준**
> 
> 이 문서의 문법 예제는 Wave 0.2.0-pre-beta를 대상으로 합니다. 이후 개발 버전은 다를 수 있습니다.

## 첫 프로그램

```wave
fun main() {
    println("Hello, Wave!");
}
```

main.wave로 저장한 뒤 wavec로 컴파일합니다. Wave 프로그램은 main에서 시작합니다.

## 콘솔 입출력

```wave
var name: str;
input("{}", name);
print("Hello, ");
println("{}", name);
```

print는 줄을 끝내지 않고 출력하고, println은 형식화한 한 줄을 출력하며, input은 변수로 형식화된 입력을 읽습니다.

## 권장 학습 순서

- 컴파일러를 설치하고 버전을 확인합니다.
- 선언, 타입, 표현식, 제어 흐름을 익힙니다.
- 함수, 구조체, 열거형, 모듈과 명시적 메모리 타입 모델을 학습합니다.
- 문법이 생각나지 않을 때 빠른 참조를 사용합니다.

