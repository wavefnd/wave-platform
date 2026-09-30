---
translation_set_id: storage-duration
path: language/storage-duration
locale: id
group: language
group_order: 2
order: 18
title: Durasi penyimpanan dan mutabilitas
summary: Bedakan antara cakupan dan kemampuan menulis var, const dan static.
---

## Arti setiap deklarasi

|memformat|Lokasi yang Diizinkan|penugasan ulang|Gunakan|
| --- | --- | --- | --- |
| `var` |Fungsi/blok|mungkin|Variabel lokal umum yang bisa berubah|
| `const` |atas|Tidak mungkin|Deklarasi konstanta global|
| `static` |atas|mungkin|Deklarasi tersimpan statis yang ada selama masa program|

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;

fun main() {
    var limit: i32 = 4;
    var current: i32 = 0;
    var retries: i32 = 0;

    current += 1;
    retries += 1;
    println("{} {} {}", limit, current, retries);
}
```

## Aturan deklarasi lokal

```wave
var value: i32 = 1;
value = 2;
```

Nilai baru dapat ditetapkan ke variabel lokal yang dideklarasikan sebagai `var`. Konstanta yang akan digunakan sepanjang program dideklarasikan di tingkat atas sebagai `const`.

## Penggunaan lokal const dan static

`const` dan `static` adalah deklarasi tingkat atas. Badan fungsi dan inisialisasi `for` menggunakan deklarasi lokal `var`.

## Umur dan Pointer

Anda bisa mendapatkan alamat variabel lokal sebagai `&`, tetapi tipe `ptr<T>` tidak melacak masa penyimpanan sebenarnya yang ditunjukkan oleh penunjuk. Saat meneruskan alamat penyimpanan lokal dari suatu fungsi, struktur program harus secara langsung memastikan bahwa alamat tersebut tetap valid.

## Rentang Pembelajaran dan Contoh

[Berlatihlah dengan program penuh](/docs/id/getting-started/overview) · [Perpustakaan standar](/docs/id/stdlib)
