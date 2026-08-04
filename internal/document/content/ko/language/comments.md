---
translation_set_id: comments
path: language/comments
locale: ko
group: language
group_order: 2
order: 11
title: 주석
summary: 한 줄 주석, 중첩 블록 주석과 진단 동작을 설명합니다.
---

## 한 줄 주석

```wave
var count: i32 = 10; // 줄 끝까지 무시
```

## 블록 주석

```wave
/* 바깥 주석
   /* 중첩 주석 */
   바깥 주석 계속
*/
```

문자열 리터럴 안의 주석 기호는 문자열 내용으로 유지됩니다. 닫히지 않은 블록 주석은 컴파일 오류이며 진단은 시작 구분자의 위치를 가리킵니다.

