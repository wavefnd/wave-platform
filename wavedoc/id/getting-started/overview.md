---
translation_set_id: overview
path: getting-started/overview
locale: id
group: getting-started
group_order: 1
order: 1
title: Gambaran keseluruhan bahasa Wave
summary: Pengenalan ringkas kepada struktur program, pemboleh ubah, dan pengkompil Wave.
---

## Tentang Wave

Wave ialah bahasa pengaturcaraan sistem bertipe statik untuk penjanaan kod natif dan kawalan aras rendah yang jelas. Lokal `id` digunakan bersama untuk dokumentasi bahasa Indonesia dan bahasa Melayu.

## Program pertama

```wave
fun main() {
    println("Hello, Wave!");
}
```

Simpan sebagai `main.wave`, kemudian jalankan:

```shell
wavec run main.wave
```

## Deklarasi pemboleh ubah

Pemboleh ubah lokal dinyatakan dengan `var`.

```wave
var count: i32 = 1;
count += 1;
```

Pemalar dan storan statik pada aras atas dinyatakan dengan `const` dan `static`.

## Status terjemahan

Halaman yang belum diterjemahkan akan ditandai dengan jelas dan dipaparkan dalam bahasa Inggeris.
