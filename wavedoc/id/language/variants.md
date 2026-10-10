---
translation_set_id: language-variants
path: language/variants
locale: id
group: language
group_order: 2
order: 17
title: Varian dan pencocokan pola
summary: Simpan data khusus kasus bersama-sama dan terpisah dengan aman dari match.
---

## Dibedakan dari enum

enum mewakili nilai bilangan bulat bernama, dan variant menampung payload, yang berbeda dalam setiap kasus. Result di bawah menyimpan nilai i32 jika berhasil dan nomor kesalahan i32 jika gagal. Meskipun error dan hasilnya bertipe integer yang sama, maknanya dapat dibedakan berdasarkan namanya.

## Deklarasi, penciptaan, inspeksi

Simpan ke `main.wave` dan jalankan.

<!-- wave-example: variant-api -->
```wave playground
variant Result {
    Value(i32), Error(i32)
}

fun calculate(valid: bool) -> Result {
    if (!valid) {
        return Result::Error(1);
    }

    return Result::Value(42);
}

fun main() {
    var result: Result = calculate(true);
    match (result) {
        Result::Value(value) => {
            println("value={}", value);
        }
        Result::Error(code) => {
            println("error={}", code);
        }
    }
}
```

Hasil eksekusi:

```text
value=42
```

Menyatakan bahwa `Result::Value(42)` memiliki payload. Gunakan payload di bawah nama value hanya dalam pola yang sesuai `match`. Ini tidak memaksa pembacaan payload dalam kasus lain. arm Teks ditulis dalam bentuk blok. Tangani semua kasus atau sisanya dengan `_`. Lebih baik jika secara eksplisit mengelompokkan setiap kasus untuk mengungkapkan apakah pemrosesan diperlukan ketika kasus baru ditambahkan.

## Generik dan umur panjang

Parameter tipe dapat digunakan seperti pada `variant Optional<T> { Some(T), None }`. Tentukan tipe konkret untuk variabel lokal, seperti `Optional<i32>`. Varian yang berisi pointer tidak secara otomatis mengelola kepemilikan memori. Menyalin nilai tidak menduplikasi alokasi yang ditunjuk.

Jangan berasumsi bahwa representasi memori variant identik dengan C union apa pun. Agar data dikirim secara eksternal ABI, ekspresi terpisah ditentukan dan dikirimkan sebagai jenis FFI yang diizinkan.

## berlatih

Jika Anda mengubahnya menjadi calculate(false), Anda akan mendapatkan `error=1`. Coba tambahkan kasus Empty tanpa payload, lalu tangani kasus tersebut di match juga.

[Struktur dan enum](/docs/id/language/structures-enums-and-aliases) · [Kelas penanganan kesalahan](/docs/id/language/errors)
