---
translation_set_id: whale-amd64-target
path: whale/amd64-target
locale: id
group: whale
group_order: 1
order: 10
title: Target AMD64 dan ABI
summary: Linux AMD64 Menjelaskan properti target, batas panggilan, dan cakupan fungsi native.
---

## pengenal sasaran

Pengidentifikasi target untuk profil native adalah `x86_64-whale-linux`.

|atribut|nilai|
| --- | --- |
|sistem operasi| Linux |
|set instruksi| AMD64 |
|urutan byte| Little endian |
|Native Lebar Alamat|64 sedikit|
|format objek| ELF64 |
|C Protokol Panggilan| SysV AMD64 ABI |
|format file statis yang dapat dieksekusi| ELF ET_EXEC |

Membangun host dan target keluaran adalah konsep yang berbeda. Hanya karena Anda dapat menjalankan Whale di host lain tidak berarti ia dapat menampilkan kumpulan perintah atau format objek host tersebut. Silakan merujuk ke [Status dukungan](overview) untuk jalur kompilasi yang diterapkan.

## Pemilihan dan verifikasi target

Perintah eksperimental `ir lower` menggunakan `x86_64-whale-linux` sebagai target default dan hanya didukung. Untuk menggunakan perintah ini, buatlah sebagai `--features socket-cli`.

```sh
whale ir lower program.json --target x86_64-whale-linux
```

Target output menyediakan tata letak data 64-bit little-endian terlepas dari host build. Pengidentifikasi yang tidak dikenal atau kombinasi yang tidak didukung seperti `aarch64-whale-linux`, `x86_64-whale-windows` akan gagal dengan kesalahan saat memandu target yang didukung sebelum membaca input atau mengganti file output. `--no-verify` tidak menonaktifkan pemeriksaan pemilihan target.

Memasukkan bagian kosong AST(`{"format_version":2,"semantics_version":1,"features":[],"program":{"declarations":[],"globals":[],"functions":[]}}`) akan menampilkan header dan modul berikut:

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

}
```

Jika Anda menentukan target yang tidak didukung, target tersebut akan berakhir dalam keadaan gagal.

```sh
whale ir lower program.json --target banana
```

```text
Error: unsupported target "banana"; supported targets: x86_64-whale-linux
```

Di Rust, Anda dapat memilih target dengan `ir::Target::lookup("x86_64-whale-linux")` dan mentransfer `name()` dan `data_layout()` ke `lower_o0`. Lowering ditolak jika tata letak yang dikirimkan berbeda dengan target yang dipilih. Periksa nama target yang tidak didukung dan ketidakkonsistenan tata letak di IR yang dikonfigurasi sendiri serta `verify_module`.

Model objek menyimpan `format`, `machine`, `endian`, dan `address_bits` di `ObjectTarget`. `ObjectFile::with_target` mempertahankan informasi identifikasi yang ditentukan, dan serialisasi hanya mengizinkan kombinasi AMD64·little-endian·64-bit ELF64. Kehadiran pengidentifikasi machine untuk arsitektur yang berbeda tidak berarti dukungan untuk pembuat enkode tersebut. Kedua titik masuk ELF writer memeriksa metadata, dan masukan tautan diperiksa sebelum resolusi simbol atau penautan. Header ELF untuk objek yang didukung tetap mempertahankan `EM_X86_64`.

`ObjectFile::new(ObjectFormat::ELF64)` adalah konstruktor praktis yang menggunakan pengenal AMD64 ini seperti sebelumnya. Kode yang sebelumnya diakses `object.format` harus menggunakan `object.target.format`.

Inspeksi ini menyediakan pemilihan target dan identifikasi objek. Ukuran struktur, array, dll., bidang offset, stride dapat dicari dengan tata letak IR API. Membaca file objek, native ABI lowering, menghubungkan ke file yang dapat dieksekusi belum didukung. Skalar lowering terus menentukan perataan eksplisit, dan aturan tata letak tipe kompleks lengkap ditentukan di [Dokumen Referensi Memori](memory-model).

## Panggil dan Tandatangani

Deklarasi IR dan pemverifikasi panggilan menerima konvensi `Whale` dan `SysV64` yang eksplisit. Konvensi ini merupakan bagian dari tipe `fnptr` dan harus cocok di lokasi panggilan. Tanda tangan variadik dan tanda tangan agregat SysV64 ditolak. Lihat [contoh konstruksi panggilan yang dapat dieksekusi](ir-reference). Ini memvalidasi kontrak IR; klasifikasi ABI mesin, register argumen/penempatan tumpukan, dan emisi panggilan asli belum diterapkan.

Panggilan yang didukung memerlukan tanda tangan dan konvensi panggilan yang eksplisit. Tanda tangan yang tidak didukung adalah kesalahan. Backend tidak boleh melakukan panggilan serupa dengan melewatkan argumen atau menggantinya dengan ekspresi berbeda.

Dukungan tanda tangan dibagi menjadi bilangan bulat/penunjuk dasar, f32/f64, struktur/nilai lebar, dan argumen variabel. Hanya karena ada tipe dalam sistem tipe IR tidak berarti meneruskan atau mengembalikan argumen tipe tersebut didukung di ABI. Saat memilih backend, Anda harus memeriksa apakah masing-masing mendukung kategori yang Anda perlukan.

Misalnya, tanda tangan pengembalian struct i32 yang tidak juga didukung oleh backend yang menangani pengembalian harus ditolak. Konvensi pengembalian skalar tidak dapat diterapkan hanya karena sebagian struktur masuk ke register skalar.

## Panggilan internal dan batasan C

native Alamat penunjuk adalah 64 bit. shadow metadata dari penunjuk pelacakan harus diteruskan bersama-sama bahkan setelah menyalin, menyimpan, argumen, dan mengembalikan.

C ABI dan protokol pengiriman metadata internal berbeda. C Panggilan lintas batas memerlukan adaptor eksplisit. Hanya karena Anda memberikan alamat numerik ke C ABI tidak boleh dianggap mempertahankan identitas penugasan, masa pakai, cakupan, atau hak akses.

## bingkai tumpukan

Mempertahankan penunjuk bingkai dan tidak menggunakan red zone. Di O0, ruang penyimpanan variabel lokal yang berbeda tidak digunakan kembali. Informasi debug bingkai panggilan menjelaskan korespondensi bingkai native dengan program aslinya.

Aturan ini hanya untuk tujuan debugging dan tidak dimaksudkan untuk mendukung pengecualian unwinding. Silakan merujuk ke [Men-debug dengan O0](o0-debugging).

## Batasan Profil

Profil O0 native pertama tidak menyertakan fitur berikut:

- O1 Optimasi, vektorisasi, LTO.
- Memori bersama dan operasi atomic.
- Pengecualian fungsi unwinding, async, coroutine.
- GC dan pemeriksaan kepemilikan khusus bahasa.
- Pengalihan kepemilikan memori eksternal secara sewenang-wenang.
- Dukungan kendala perakitan inline penuh.
- Tautan dinamis, TLS, kumpulan perintah tambahan, tipe objek tambahan.

Permintaan yang berada di luar cakupan dukungan harus ditolak daripada diganti secara diam-diam dengan fitur lain. Pembatasan ini ditujukan untuk jalur kompilasi native dan tidak berarti menghapus komponen rantai alat lain yang disediakan secara independen.
