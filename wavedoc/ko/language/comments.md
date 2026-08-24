---
translation_set_id: comments
path: language/comments
locale: ko
group: language
group_order: 2
order: 11
title: 주석
summary: 한 줄 주석, 중첩 가능한 블록 주석과 닫히지 않은 주석 진단을 설명합니다.
---

## 한 줄 주석

`//` 뒤의 내용은 줄 끝까지 주석입니다.

```wave
var count: i32 = 10; // 현재 요청 수
```

## 블록 주석

`/*`와 `*/` 사이를 블록 주석으로 처리합니다.

```wave
/* 여러 줄에 걸친
   설명을 작성할 수 있습니다. */
```

lexer는 블록 주석의 깊이를 추적하므로 **중첩 블록 주석도 지원**합니다.

```wave
/* 바깥 주석
   /* 안쪽 주석 */
   다시 바깥 주석
*/
```

## 문자열과 주석 기호

lexer는 문자열·문자 리터럴을 토큰으로 처리하므로 리터럴 내용에 들어 있는 `//`, `/*`, `*/`는 주석 시작이나 끝으로 취급되지 않습니다.

```wave
var text: str = "https://wave-lang.dev";
```

## 닫히지 않은 블록 주석

파일 끝까지 `*/`를 찾지 못하면 lexer가 `UnterminatedComment` 진단을 생성합니다. 현재 이 오류에는 `E1002` 코드가 사용됩니다.

긴 블록을 임시로 비활성화할 때도 중첩 깊이가 맞는지 확인하십시오.
