---
translation_set_id: program-structure
path: language/program-structure
locale: id
group: language
group_order: 2
order: 1
title: 1. Dari file sumber hingga program yang sedang berjalan
summary: Pelajari tentang file sumber, fungsi, keluaran, inspeksi, dan eksekusi.
---

## Sebelum memulai bab ini

Siapkan compiler dan pustaka standar sesuai [Petunjuk instalasi](/docs/id/getting-started/install). Jika Anda dapat menjalankan `wavec --version` di terminal Anda, Anda dapat memulai. Editor apa pun yang dapat menyimpan file teks biasa bisa digunakan.

Dalam bab ini, kita mulai dengan membuat satu baris keluaran dan mempelajari hubungan antara file sumber, fungsi, kompilasi, eksekusi, dan kode keluar. Tujuannya bukan sekadar menyalin instruksi, tapi mampu menjelaskan apa yang terjadi pada tahap apa.

## Buat direktori kerja

Menggunakan direktori terpisah untuk setiap program memudahkan untuk menemukan sumber dan menghasilkan file yang dapat dieksekusi. Buat dan navigasikan ke direktori di terminal.

```shell
mkdir wave-study
cd wave-study
```

Buat `main.wave` di direktori ini dengan editor Anda. Periksa ekstensi untuk memastikan nama file bukan `main.wave.txt`. Di bawah ini adalah keseluruhan file, bukan hanya sebagian di dalam fungsinya.

<!-- wave-example: book-hello -->
```wave playground
fun main() {
    println("Hello, Wave!");
}
```

Hasil eksekusi:

```text
Hello, Wave!
```

Jalankan dengan perintah berikut:

```shell
wavec run main.wave
```

Jangan mencampur perintah terminal dengan kode Wave. Masukkan `wavec run` ke terminal, dan tulis `fun main` ke dalam file sumber. Tidak perlu menempelkan kembali keluaran `Hello, Wave!` oleh program ke sumbernya.

## membaca baris demi baris

`fun` adalah kata kunci yang mendeklarasikan suatu fungsi. Fungsi adalah kumpulan operasi bernama, dan `main` adalah titik masuk dari eksekusi. Anda akan menentukan sendiri fungsi lainnya nanti.

`()` setelah `main` adalah tempat Anda menulis parameter. main dalam program ini tidak menggunakan parameter dan karenanya kosong. Tuliskan operasi fungsi antara `{` dan `}`. Indentasi membuat blok lebih mudah dibaca manusia, dan batas blok itu sendiri ditandai dengan tanda kurung kurawal.

`println("Hello, Wave!");` adalah kalimat yang menghasilkan string. Tanda kutip ganda menunjukkan awal dan akhir string literal. Tanda kutip itu sendiri tidak disertakan dalam keluaran. Titik koma menunjukkan akhir dari pernyataan ini.

## Pernyataan dieksekusi sesuai urutan penulisannya

Coba ubah untuk mencetak tiga kali. Jangan gabungkan dengan program sebelumnya, tapi ganti isi main.wave dengan keseluruhan program di bawah ini.

<!-- wave-example: book-sequence -->
```wave playground
fun main() {
    println("start");
    println("working");
    println("done");
}
```

Hasil eksekusi:

```text
start
working
done
```

Setelah menyelesaikan kalimat pertama, lanjutkan ke kalimat berikutnya. Tidak ada tugas yang berjalan secara bersamaan di sini. Jika ingin mengubah urutan keluaran, ubah saja urutan kalimatnya. Anda dapat mengontrol urutan ini dengan mempelajari pernyataan kondisional dan loop nanti.

## print dan println

`println` menambahkan jeda baris di akhir. `print` tidak otomatis mengubah garis. Perbedaannya terlihat jelas dalam menghubungkan potongan-potongan kecil menjadi satu garis.

<!-- wave-example: book-print-lines -->
```wave playground
fun main() {
    print("Wave");
    print(" ");
    println("study");
    println("second line");
}
```

Hasil eksekusi:

```text
Wave study
second line
```

Mencetak tiga kali tidak selalu menghasilkan tiga baris. Bedakan antara jumlah panggilan keluaran dan jumlah saluran. Saat menulis jeda baris langsung dalam string, gunakan `\n` escape. String escape dan jeda baris pada sumber dibahas secara rinci di [lembaran tali](/docs/id/language/strings).

## Masukkan nilai ke dalam string

Untuk memasukkan hasil perhitungan ke dalam string, berikan nilai yang sesuai dengan `{}`.

<!-- wave-example: book-format-first -->
```wave playground
fun main() {
    println("{} + {} = {}", 2, 3, 2 + 3);
}
```

Hasil eksekusi:

```text
2 + 3 = 5
```

`{}` pertama berisi 2, yang kedua berisi 3, dan yang ketiga berisi 5. Format string dan nilai dipisahkan dengan koma. Jika Anda mengubah jumlah nilai, jumlah placeholder juga harus cocok. Ini adalah sintaks keluaran yang berbeda dari operasi penambahan string.

## Bagilah inspeksi, pembangunan, dan eksekusi

run yang digunakan sejauh ini menjalankan pembangunan dan eksekusi secara berurutan. Jika Anda ingin mengetahui langkah mana yang gagal, Anda dapat menguraikannya seperti ini:

```shell
wavec check main.wave
wavec build main.wave -o hello
```

check memeriksa tata bahasa, jenis, dll., tetapi tidak menguji semua masukan ke program. Misalnya, pemeriksaan sumber saja tidak dapat menentukan apakah suatu file ada saat runtime. build membuat file yang dapat dieksekusi. Di Linux/macOS, jalankan sebagai berikut.

```shell
./hello
```

Di Windows, beri nama file output `hello.exe` dan jalankan di PowerShell sebagai `.\hello.exe`. Jika Anda mengubah sumber setelah membuat file yang dapat dieksekusi, Anda harus membangunnya kembali agar perubahan dapat diterapkan.

## Kode keluar juga merupakan hasil

Manusia membaca pernyataan keluaran, tetapi shell atau program lain dapat menentukan keberhasilan melalui kode keluar. Berikut ini menetapkan bahwa main mengembalikan i32.

<!-- wave-example: book-exit-success -->
```wave playground
fun main() -> i32 {
    println("completed");
    return 0;
}
```

Hasil eksekusi:

```text
completed
```

0 adalah konvensi untuk menunjukkan pematian normal. Saat memberi sinyal kegagalan secara langsung, kembalikan kode bukan nol. Di shell Linux/macOS, kode diperiksa sebagai `echo $?` segera setelah eksekusi, dan di PowerShell, kode diperiksa sebagai `$LASTEXITCODE`. Jika Anda menjalankan perintah lain sementara itu, apa yang Anda periksa mungkin berubah.

Mengeksekusi `return` akan mengakhiri fungsinya. main tidak mendeklarasikan parameter apa pun dan meskipun memiliki nilai default, parameter tersebut tidak diperbolehkan. Hilangkan jenis pengembalian main atau gunakan i32.

## Cara membaca kesalahan pertama

Kode berikut ini sengaja salah: Berbeda dengan contoh yang berjalan, kode ini akan gagal di check.

```wave
fun main() {
    println("hello")
}
```

Tidak ada titik koma di akhir kalimat. Lihatlah baris yang ditampilkan diagnosis dan kalimat sebelum itu. Lokasi yang ditandai oleh kompiler mungkin bukan di mana kesalahan berasal, melainkan di mana struktur yang salah tidak lagi dapat diinterpretasikan.

Pertama, perbaiki kesalahan pertama lalu periksa kembali. Jika tanda kurung atau tanda kutip sebelumnya tidak ditutup, berbagai kesalahan mungkin terjadi bahkan pada kode normal berikutnya. Jika Anda mencoba memperbaiki semua lini secara bersamaan, penyebab aslinya akan mudah terlewatkan.

## soal latihan

1. Buat program yang mencetak tiga baris perkenalan diri.
2. Lewati 12 dan 8 sebagai argumen format untuk mencetak `12 * 8 = 96`.
3. Tulis untuk mencetak pesan sukses dan mengembalikan kode keluar 0.
4. Prediksikan sebelum eksekusi berapa jumlah jalur keluaran jika Anda menggunakan print hanya tiga kali.

### Solusi : Perhitungan keluaran proses

<!-- wave-example: book-first-solution -->
```wave playground
fun main() -> i32 {
    println("Learning Wave");
    println("My first program");
    println("Ready to calculate");
    println("{} * {} = {}", 12, 8, 12 * 8);
    return 0;
}
```

Hasil eksekusi:

```text
Learning Wave
My first program
Ready to calculate
12 * 8 = 96
```

Daripada menulis hasil penghitungan langsung dalam string sebagai `96`, saya meneruskannya sebagai ekspresi. Ini merupakan langkah awal untuk memastikan hasil perhitungan dan tampilan tidak berubah meskipun nilai inputnya berubah. Pada bab selanjutnya, kita akan memberi nama variabel untuk menghindari penulisan nilai yang sama berkali-kali.


## Item tingkat atas dalam file sumber

Sumber Wave dapat terdiri dari item tingkat atas berikut:

- `import(...)`
- `const`, `static`
- `type`
- `struct`, `enum`, `proto`
- `extern(...)`, `export(...)`
- `fun`
- `#[target(...)]` kondisi sebelum item yang didukung

Awali deklarasi yang dapat diimpor dengan `pub`. Tempatkan deklarasi `var` lokal di dalam fungsi atau blok.

## Program Berdiri Bebas

Target tanpa kernel, kode boot, atau runtime dapat menggunakan opsi build yang berdiri sendiri.

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files
```

`--freestanding` menautkan ke rencana pembangunan yang mematikan dependensi perpustakaan default, dan `--entry` menyetel simbol entri tautan. Untuk membuat keluaran aktual yang dapat di-boot, Anda perlu merancang arsitektur target, skrip linker, dan bahkan format objek.

## Titik masuk yang sengaja gagal

Anda tidak dapat memiliki parameter di main meskipun parameter tersebut memiliki nilai default. Terjadi kesalahan saat menyalin file berikut check.

<!-- wave-example: reject-main-argument -->
```wave
fun main(value: i32 = 0) -> i32 {
    return value;
}
```
