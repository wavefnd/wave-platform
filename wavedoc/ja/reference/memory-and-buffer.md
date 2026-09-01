---
translation_set_id: memory-buffer
path: reference/memory-and-buffer
locale: ja
group: reference
group_order: 3
order: 5
title: メモリとバッファ
summary: std::mem の手動割り当て、コピー、アラインメント、ページ補助機能と、std::buffer を安全に使うための規則を説明します。
---

## 基本的な手動割り当て

```wave
import("std::mem::alloc")::{mem_alloc, mem_free};
import("std::mem::ops")::{mem_zero};

fun main() {
    var size: i64 = 256;
    var memory: ptr<u8> = mem_alloc(size);

    if (memory != null) {
        mem_zero(memory, size);
        mem_free(memory, size);
    }
}
```

`mem_alloc(size)` は `ptr<u8>` を返し、失敗時には `null` を返すことがあります。`mem_free` にはポインタと元の割り当てサイズの両方を渡します。

## ゼロ初期化と再割り当て

`std::mem::alloc` は次の機能を提供します。

- `mem_alloc`
- `mem_alloc_zeroed`
- `mem_realloc`
- `mem_free`
- ジェネリックな項目の割り当て、再割り当て、解放の補助機能
- ページ数とページアラインメントの補助機能
- アラインメント指定の割り当てと解放の補助機能

`mem_realloc(old_ptr, old_size, new_size)` は `new_size` の大きさのストレージを返し、古いサイズと新しいサイズの小さい方までの内容を保持します。新しいサイズが正でない場合、有効な古いストレージを解放して `null` を返します。新しい割り当てに失敗した場合は `null` を返し、古い割り当ては所有者が引き続き利用できます。

## サイズの単位

主なメモリサイズ引数には、バイト数を表す `i64` を使います。要素数とバイト数を混同しないように、呼び出し側のコードで単位を明示してください。

```wave
var count: i64 = 32;
var elem_size: i64 = 4;
var bytes: i64 = count * elem_size;
```

## ポインタアクセス

手動で割り当てた `ptr<u8>` は境界情報を持ちません。`deref p[index]` にアクセスする前に、インデックスが割り当て範囲内にあることを呼び出し側で保証します。

## std::buffer

`std::buffer` は `std::mem` の上にバッファ補助機能を構築します。その API でも、割り当て失敗、長さと容量、所有権、解放の契約を呼び出し側が守る必要があります。

割り当てと解放を同じ抽象化層にまとめると、リークやサイズ不一致を防ぎ、確認しやすくなります。
