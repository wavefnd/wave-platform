---
translation_set_id: whale-o0-debugging
path: whale/o0-debugging
locale: id
group: whale
group_order: 1
order: 9
title: O0 dan debugging
summary: Aturan untuk menyimpan penghitungan, ekspresi konstan, ruang penyimpanan, dan informasi respons debug.
---

## model konservasi

O0 mempertahankan perhitungan, variabel, dan aliran kontrol dari typed IR asli untuk tujuan debugging. Itu juga menyimpan hasil yang tidak digunakan dan blok yang secara struktural tidak dapat dijangkau. Verifikasi mendiagnosis kesalahan IR dan tidak menghapus blok atau menyederhanakan operasi.

Konversi yang diperlukan untuk menghasilkan bahasa mesin dilakukan dalam sub-sub terpisah IR. Bahkan setelah konversi, korespondensi dengan ID asli harus dipertahankan. typed IR konservasi tidak berarti korespondensi satu-ke-satu dari semua operasi dan instruksi mesin IR.

## Konversi tidak dilakukan oleh O0

|konversi|O0 Operasi|
| --- | --- |
|sebaris|Pertahankan batasan panggilan dan fungsi|
|Konversi Tail-call|Pertahankan struktur panggilan/pengembalian umum|
|menghapus kode mati|Pertahankan komputasi yang tidak digunakan dan blok yang tidak dapat dijangkau|
|Ciutkan konstanta runtime|pertahankan operasi aslinya|
|Gabungkan perhitungan umum|Simpan perhitungan terpisah|
|Penggunaan kembali ruang penyimpanan variabel lokal|Jaga setiap ruang penyimpanan|
|Hilangkan penunjuk bingkai|mempertahankan penunjuk bingkai|
|Penggabungan string secara otomatis|Pertahankan objek string yang terpisah|

Misalnya, operasi runtime yang menambahkan dua konstanta akan tetap menjadi tambahan meskipun hasilnya tidak digunakan. Cabang-cabang dengan kondisi konstan juga mempertahankan struktur aliran kendali aslinya.

## konstanta waktu kompilasi

Deklarasi konstanta waktu kompilasi mempertahankan ekspresi inisialisasi typed dan hasil evaluasi. Hal ini berbeda dengan pelipatan konstan pada instruksi runtime biasa.

Deklarasi dengan ekspresi inisialisasi `1 + 2` dibiarkan dengan ekspresi penjumlahan dan hasil `3`. Contoh ini mengilustrasikan arti ungkapan dan bukan tata bahasa deklaratif bahasa sumber tertentu. Hasilnya dapat digunakan saat membuat data statis dan tidak menggantikan ekspresi inisialisasi dengan penambahan runtime.

Deklarasi yang tidak digunakan dan tidak dapat dijangkau juga harus diidentifikasi. Meskipun deklarasi dengan nama yang sama mengaburkan satu sama lain, deklarasi tersebut harus dipisahkan berdasarkan nama·deklarasi ID·referensi. Validasi menolak hasil tersimpan yang memiliki referensi tidak valid, siklus ketergantungan, tipe tidak valid, atau tidak cocok dengan ekspresi aslinya.

## pernyataan sumber yang tidak dapat dijangkau

Kalimat setelah return·break·continue juga tetap berada dalam blok yang tidak terhubung. Karena pernyataan ini, Anda tidak boleh mengubah terminator sebelumnya atau membuat jalur eksekusi baru. Ini juga mendiagnosis ekspresi yang tidak valid dalam kalimat yang tidak dapat dijangkau.

Aturan pelestarian ini memungkinkan pemeriksaan struktur program asli. Ini tidak berarti bahwa pernyataan setelah terminator akan benar-benar dieksekusi.

## Contoh IR yang diawetkan

Di bawah ini adalah keluaran printer saat ini untuk modul yang lolos verifikasi. Menampilkan penghitungan yang tidak digunakan, deklarasi waktu kompilasi, dan blok yang tidak tertaut.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "preserved": whale () -> void, linkage internal

  fn @f0 "preserved"() -> void, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 1
    %v1: i32 = const i32 2
    %v2: i32 = add i32 %v0, %v1
    %v3: i32 = const_decl "count" add(i32 1, i32 2) => const i32 3
    ret void
  %b1 "unreachable.cont":
    %v4: i32 = const i32 4
    %v5: i32 = const i32 5
    %v6: i32 = add i32 %v4, %v5
    ret void
  }

}
```

`%v2` tetap menjadi `add` meskipun tidak ada gunanya. `%v3` adalah `const_decl` terpisah, yang mempertahankan ekspresi `add(i32 1, i32 2)` dan hasil evaluasi 3 secara bersamaan. Karena `unreachable.cont` tidak memiliki tepi masuk, kita dapat memeriksa penambahan `%v6` tanpa menambahkan jalur eksekusi setelah `ret void`.

Verifikasi mempertahankan instruksi dan blok ini. Contoh ini menunjukkan IR konfigurasi/verifikasi/output dan tidak berarti bahwa native eksekusi atau output DWARF disediakan.

## Sumber dan Informasi Tumpukan

Antarmuka debug menggunakan informasi fungsi, baris sumber, variabel lokal default, dan informasi bingkai panggilan dari DWARF 5. Bahkan jika dikonversi ke sub-ekspresi, korespondensi antara informasi fungsi/variabel lokal dan pengidentifikasi IR asli harus dipertahankan.

Profil AMD64 mempertahankan penunjuk bingkai dan tidak menggunakan red zone. Informasi bingkai panggilan digunakan untuk pemeriksaan tumpukan dan tidak menyiratkan dukungan untuk pengecualian unwinding. trap menghentikan eksekusi tanpa menjamin destruktor atau unwinding.

Silakan merujuk ke [Ikhtisar rantai alat](overview) untuk ketersediaan keluaran DWARF dan eksekusi native. Perilaku O1 dan optimasi di atas berada di luar cakupan referensi O0 ini.

## Deklarasi berulang dalam loop

Simpan modul lengkap ini sebagai `initialization-loop.wir`. Walau iterasi pertama menyimpan 42, uninit deklarasi kedua mereset inisialisasi sehingga pembacaan menghasilkan trap. O0 mempertahankan uninit pada deklarasi yang dieksekusi meskipun alloca di blok masuk. Pembacaan tak terpakai yang dieksekusi tetap diperiksa; blok tak terjangkau yang dipertahankan tidak dieksekusi.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "redeclaration": whale () -> u32, linkage internal

  fn @f0 "redeclaration"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    %v1: u32 = const u32 0
    %v2: u32 = const u32 1
    %v3: u32 = const u32 42
    br label %b1
  %b1 "declaration":
    %v4: u32 = phi u32 [ %v1, %b0 ], [ %v6, %b2 ]
    uninit u32, ptr<u32> %v0, align 4
    %v5: bool = icmp eq u32 %v4, %v1
    cbr bool %v5, label %b2, label %b3
  %b2 "first_iteration":
    store u32 %v3, ptr<u32> %v0, align 4
    %v6: u32 = add u32 %v4, %v2
    br label %b1
  %b3 "second_iteration":
    %v7: u32 = load u32, ptr<u32> %v0, align 4
    ret u32 %v7
  }

}
```

```shell
whale ir run initialization-loop.wir --function @f0
```

```text
Error: initialization-loop.wir: trap at @f0 %b3 instruction 0 (%v7): uninitialized byte at allocation offset 0 (after 17 steps)
```
