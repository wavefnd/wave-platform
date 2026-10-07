---
translation_set_id: whale-overview
path: whale/overview
locale: id
group: whale
group_order: 1
order: 1
title: Dokumentasi Whale
summary: Panduan untuk komponen rantai alat, perintah yang tersedia, dan dokumentasi referensi.
---

## Pendahuluan

Whale adalah rantai alat kompiler tujuan umum untuk implementasi bahasa pemrograman dan alat kompiler. Menyediakan ekspresi perantara yang diketik (IR), assembler AMD64, pustaka file objek, dan fungsi berbasis linker. Setiap komponen digunakan melalui perpustakaan Rust dan perintah `whale`.

IR mengekspresikan semantik komputasi secara independen dari pengkodean bahasa mesin. Assembler mengkodekan instruksi mesin untuk membuat objek yang dapat direlokasi. Pustaka objek mewakili bagian, simbol, dan penataan ulang. Linker menyelesaikan referensi antar objek dan menempatkan file yang dapat dieksekusi.

## Membangun dan menggunakan alat

|bekerja|dokumen|
| --- | --- |
|Apa yang dilakukan setiap alat| [Komponen rantai alat](/docs/id/whale/ecosystem) |
|Wave Pembuatan program/tautan/pemilihan target| [Bangun dan Tautkan](/docs/id/whale/build-link-targets) |
|Manajemen paket/ketergantungan| [Vex](/docs/id/whale/vex-package-manager) |
|Whale Jalankan perintah| [Whale CLI](/docs/id/whale/whale-cli) |

## Panduan Dokumen

|dokumen referensi|konten|
| --- | --- |
| [Lihat IR](ir-reference) |Jenis, nilai, fungsi, aliran kontrol, validasi, format pertukaran|
| [Operasi numerik](numeric-operations) |Aritmatika bilangan bulat, pergeseran, konversi tipe, floating point|
| [Model memori](memory-model) |Inisialisasi, validasi penunjuk, perhitungan alamat, tata letak, string|
| [Men-debug dengan O0](o0-debugging) |Retensi komputasi, penyimpanan variabel, informasi debug|
| [AMD64 Sasaran](amd64-target) |Pengidentifikasi target, konvensi pemanggilan, cakupan fungsi native|
| [Perakit dan Penghubung](assembler-linker) |Operan perakitan, bagian, simbol, tautan statis|

Dokumen referensi mendefinisikan aturan semantik untuk Whale. Fitur yang tersedia ikuti tabel di bawah ini. Operasi yang dijelaskan dalam dokumentasi referensi mungkin tidak tersedia di semua build.

## Status dukungan

|komponen|Antarmuka disediakan|batas|
| --- | --- | --- |
|perakit|Buat ELF64 objek yang dapat direlokasi dari perakitan AMD64|Cakupan perintah dan arahan yang tidak lengkap|
|perpustakaan objek|Komposisi objek, payload tanpa BSS, validasi target, AMD64 ELF64 serialisasi dan implementasi rekaman Wave yang dapat dipilih untuk memeriksa ukuran.|Objek yang dapat direlokasi bukanlah objek yang dapat dieksekusi|
| IR | Konstruksi, pencetakan, panggilan langsung/tidak langsung yang diperiksa tanda tangan, penurunan versi AST, validasi target, dan tata letak tipe yang diperiksa | AST format 2 / typed IR format 4, konstanta bit tepat serta parsing, verifikasi dan pencetakan bolak-balik teks tersedia; emisi panggilan mesin belum tersedia |
|penghubung|Interpretasi simbol, verifikasi target input, inspeksi penempatan bagian file/memori|Menerapkan relokasi penuh dan mengeluarkan file yang dapat dieksekusi tidak didukung.|
| Eksekusi dan debugging | Integer/Bool, alur kontrol, stack terlacak dan pemeriksaan inisialisasi | Float, panggilan, alamat global, pembangkitan kode native dan DWARF belum didukung |

Aturan yang melarang perilaku tidak terdefinisi berlaku untuk IR terverifikasi dan melacak memori. Implementasi eksperimental belum mengimplementasikan semua pemeriksaan runtime dalam dokumentasi referensi memori/eksekusi.

## Membangun dan Merakit

Bangun dengan Rust 1.86.0 atau lebih tinggi.

```sh
git clone https://github.com/wavefnd/Whale.git
cd Whale
cargo build --release --locked
```

Simpan perakitan berikut sebagai `answer.asm`.

```asm
section .text
global answer

answer:
    mov eax, 42
    ret
```

Buat objek ELF64.

```sh
./target/release/whale asm --amd64 answer.asm -o answer.o
```

Ia menggunakan assembler Whale sendiri, jadi tidak diperlukan assembler eksternal. Outputnya berisi fungsi yang dapat dipanggil dan tidak berisi kode startup proses apa pun.

Perintah eksperimental AST→IR tersedia dengan membangun dengan `--features socket-cli`. Untuk perintah CLI, lihat [Referensi Perintah](/docs/id/whale/whale-cli).

## Penggunaan dan diagnosis perpustakaan

IR harus diverifikasi sebelum eksekusi atau pembuatan kode. Kesalahan input dan penyalahgunaan builder akan mengembalikan kesalahan struktural. Kesalahan input di perpustakaan tidak boleh menghentikan proses host atau menimpa konten yang ada. Alat yang menerima masukan tidak tepercaya harus dapat menyesuaikan batas sumber daya.

Versi rantai alat, masukan, target, dan konfigurasi yang sama harus menghasilkan keluaran deterministik. Metadata distribusi mengidentifikasi versi·commit·dan fitur yang tersedia. CI memeriksa masukan yang baik, masukan yang ditolak, trap, O0 retensi, round-trip, native semantik eksekusi untuk setiap antarmuka. Untuk perintah pengembangan/verifikasi, silakan merujuk ke [Whale Informasi Kontribusi](https://github.com/wavefnd/Whale/blob/master/CONTRIBUTING.md).
