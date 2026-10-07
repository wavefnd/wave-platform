---
translation_set_id: whale-memory-model
path: whale/memory-model
locale: id
group: whale
group_order: 1
order: 8
title: Model memori
summary: Pelacakan alokasi, pembacaan nilai yang diinisialisasi, aritmatika penunjuk, tata letak, dan aturan penyimpanan string.
---

## Pelacakan Alokasi

Penunjuk yang dilacak mengaitkan alamat dengan ID alokasi, pembuatan, batasan, offset, dan izin akses. Model memori juga melacak masa pakai alokasi dan status inisialisasi. Akses harus memenuhi persyaratan berikut; pelanggaran menyebabkan jebakan.

Meskipun alamat fisik yang sama digunakan kembali, generasi-generasi akan terpisah setiap masa hidupnya. Kehadiran alamat saja tidak menentukan apakah penunjuknya valid atau pemanggil memiliki akses ke ruang penyimpanan tersebut.

Alamat native mempertahankan 64 bit. Pisahkan shadow metadata diteruskan bersama dengan penunjuk melalui salin/simpan/panggilan/pengembalian. Cakupan memori awal native adalah tumpukan yang dapat dilacak dan alokasi global. C Batasan memerlukan adaptor eksplisit dan transfer kepemilikan memori eksternal arbitrer tidak termasuk dalam cakupan ini.

## Inisialisasi dan Membaca

Mendeklarasikan ruang penyimpanan tidak menginisialisasi nilai. Memeriksa apakah rentang byte yang sebenarnya sedang dibaca telah diinisialisasi. Bahkan jika sebagian nilai dibaca dalam keadaan tidak diinisialisasi, nilainya adalah trap. Itu tidak menggantikan hasil baca dengan 0 atau nilai yang tidak ditentukan.

Pemeriksaan inisialisasi pada nilai baca tidak termasuk byte padding. Misalnya, jika semua bidang struktur diinisialisasi, pembacaan nilai tidak akan salah hanya karena byte kosong yang dihasilkan dari penyelarasan bidang tidak diinisialisasi.

Salinan memori membawa status inisialisasi bersama dengan byte. Menyalin ruang penyimpanan yang tidak diinisialisasi tidak mengubahnya menjadi ruang penyimpanan yang diinisialisasi. Saat membaca nilai dari tujuan nanti, pemeriksaan yang sama diterapkan seperti saat membaca sumbernya.

Fakta bahwa byte fisik BSS adalah 0 tidak memungkinkan inisialisasi variabel IR.

### IR dalam ruang penyimpanan skalar yang diinisialisasi

Modul berikut dikonfigurasi sebagai builder dan lolos verifikasi. `store` sebelum membaca nilainya, dan ketiga instruksi memori menentukan pengurutan pangkat 2 bukan nol.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "initialized_local": whale () -> i32, linkage internal

  fn @f0 "initialized_local"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: ptr<i32> = alloca i32, align 4
    %v1: i32 = const i32 42
    store i32 %v1, ptr<i32> %v0, align 4
    %v2: i32 = load i32, ptr<i32> %v0, align 4
    ret i32 %v2
  }

}
```

Simpan modul lengkap ini sebagai `initialized.wir`; eksekusi mencetak `i32 42`. Menghapus store menyebabkan trap pembacaan belum diinisialisasi pada load. Alignment 3 adalah kesalahan verifikasi. Nol fisik dari alloca bukan inisialisasi.

## Perbandingan dengan aritmatika pointer

Perhitungan alamat overflow adalah trap. Anda dapat membuat penunjuk one-past yang menunjuk tepat setelah rentang alokasi, tetapi Anda tidak dapat mengakses memori melalui rentang tersebut.

Kesetaraan penunjuk menggunakan identitas alokasi, bukan hanya alamat numerik. Mengurutkan atau mengurangi pointer dari alokasi berbeda menyebabkan jebakan. Merekonstruksi alamat dari bilangan bulat tidak mengembalikan izin akses.

### GEP

GEP menghitung alamat berdasarkan elemen dan bidang. Ini bukan perintah untuk membaca suatu nilai.

Indeks pertama adalah unit elemen offset dari tipe yang ditunjukkan oleh penunjuk dasar. Indeks berikutnya memilih elemen array atau bidang struktur/tupel. Indeks bidang suatu struktur/tupel adalah nomor urutan bidang pada saat kompilasi, bukan byte offset. Tipe yang dipilih menentukan tipe penunjuk yang dihasilkan.

Jika tipe dasarnya adalah `ptr<array<i32, 4>>`, indeks `[0, 2]` memilih elemen i32 ketiga dari array, sehingga menghasilkan `ptr<i32>`. Menentukan indeks pertama sebagai 1 gerakan i32 berdasarkan satu elemen, bukan satu elemen dalam array. Contoh ini mengilustrasikan semantik indeks dan bukan sintaks perintah teks.

Target awal native menolak aritmatika penunjuk pada elemen ukuran 0. Menghitung alamat tidak menghilangkan pemeriksaan seumur hidup, cakupan, inisialisasi, dan izin yang diperlukan untuk akses selanjutnya.

## tata letak data

Target keluaran menentukan ukuran·penyelarasan·bidang offset·array stride. Urutan bidang dalam struktur dan tupel mempertahankan urutan deklarasinya. Tata letak host yang menjalankan kompiler tidak boleh dianggap sebagai tata letak target keluaran.

|nilai|Simpan Aturan|
| --- | --- |
| Bool, signed i1, unsigned u1 |Setidaknya 1 byte|
|Struktur/tupel kosong|Ukuran 0, penyelarasan 1|
|Himpunan|Tata letak target menentukan elemen stride|
|Struktur/tupel|Pertahankan urutan deklarasi dan persyaratan penyelarasan target|

Penyelarasan yang telah selesai dari IR adalah pangkat 2, bukan 0. Penyelarasan otomatis harus ditentukan sebelum menghasilkan IR ini. Tata letak packed·union·bitfield tidak didukung di profil ini dan harus ditolak.

### Permintaan Tata Letak Keluaran

Rust API menghitung tata letak penyimpanan yang tidak bergantung pada host build. Pada contoh berikut, terdapat 7 byte padding sebelum kolom u64 dan 6 byte di akhir padding.

```rust
use ir::{allocation_align, layout_of, Target, Type};

fn main() {
    let target = Target::X86_64WhaleLinux;
    let record = Type::Struct(vec![Type::U8, Type::U64, Type::U16]);
    let layout = layout_of(&record, target).unwrap();
    assert_eq!((layout.size, layout.align), (24, 8));
    assert_eq!(layout.field_offsets, [0, 8, 16]);

    let array = Type::Array(Box::new(record), 3);
    let layout = layout_of(&array, target).unwrap();
    assert_eq!((layout.size, layout.align), (72, 8));
    assert_eq!(layout.element_stride, Some(24));
    assert_eq!(allocation_align(&array, target).unwrap(), 16);
}
```

```text
struct{u8, u64, u16}: size 24, natural alignment 8
field 0: byte 0
field 1: byte 8
field 2: byte 16
array of 3: size 72, element stride 24
standalone array placement alignment: 16
```

Ukuran yang dikembalikan oleh `layout_of` dan langkah array menyertakan padding tambahan. Struktur dan tupel menggunakan aturan urutan bidang yang sama. Bool, i1, dan u1 masing-masing menempati satu byte. Struct dan tupel kosong memiliki ukuran 0 dan perataan 1; array dengan panjang nol mempertahankan keselarasan alami elemennya. `void` tidak memiliki tata letak penyimpanan, sedangkan `ptr<void>` menempati 8 byte.

Gunakan penyortiran alami untuk bidang dan elemen array. `allocation_align` menerapkan aturan SysV AMD64, yang memerlukan setidaknya penyelarasan 16 byte, ke array lokal/global independen berukuran 16 byte atau lebih. Itu tidak meningkatkan perataan bidang array atau elemen stride. AST lowering menggunakan pencarian pengurutan batch ini untuk penyimpanan lokal.

Perkalian besaran, penambahan bidang offset, overflow dalam perhitungan padding menghasilkan hasil `LayoutError::Overflow`. `layout_of` memiliki batas sarang tipe kompleks sebanyak 128 level, dan `layout_of_with_limit` memungkinkan pemanggil menentukan batasnya. Itu tidak mengalokasikan ruang penyimpanan sama dengan jumlah elemen array. `pointer_stride` menolak pointee dengan ukuran 0 dalam aritmatika penunjuk native, tetapi tata letak penyimpanan jenis itu sendiri valid. Packed·union·bitfield tidak memiliki ekspresi tipe yang didukung. Pencarian tata letak tersimpan ini tidak mengimplementasikan konvensi pemanggilan tipe kompleks atau pemeriksaan batas waktu proses.

## String dan batas C

String adalah string byte yang tidak dapat diubah dengan panjang tertentu. Pengkodean defaultnya adalah UTF-8. Mengizinkan NUL internal dan tidak secara otomatis menambahkan akhiran NUL. O0 tidak secara otomatis menggabungkan objek string dengan konten yang sama.

Oleh karena itu, panjang string byte yang terdiri dari `A`, NUL, dan `B` adalah 3. Tidak dapat digunakan untuk konversi string C eksplisit, yang menolak NUL internal. Frontend tidak boleh memotongnya secara diam-diam menjadi `A`.

Eksternal C·Alamat Mentah·Perakitan Inline adalah batas kontrak terpisah. Pemeriksaan runtime memori jejak tidak dijamin mendeteksi semua perilaku yang salah dalam kode eksternal.

## Eksekusi memori stack terlacak

Interpreter bawaan menjalankan integer/Bool, alur kontrol, alokasi stack, penyimpanan dan pembacaan pointer data, typed GEP, memcpy dan memset. Pembacaan pasangan checked tidak memeriksa padding. Alamat berupa nilai sintetis 64 bit, tanpa dereferensi memori host. Argumen dan hasil tetap integer/Bool atau void; float, panggilan, pointer fungsi, nilai agregat umum, alamat global dan eksekusi native belum didukung. Alokasi stack hidup sampai fungsi kembali. Akhir masa hidup leksikal, penerusan pointer melalui panggilan/hasil, adaptor memori asing dan native shadow metadata memerlukan implementasi lanjutan.

`--max-memory` mengatur anggaran byte alokasi logis, bawaan 64 MiB. `InterpreterOptions::memory_limits` juga membatasi jumlah alokasi 16384, metadata byte pointer 262144 fragmen dan kerja byte/metadata 256 Mi unit. Batas terlampaui mengembalikan `MemoryLimit` dengan lokasi IR, terpisah dari trap program. Verifikasi juga menolak overflow ukuran penyimpanan yang diketahui untuk target keluaran sebelum eksekusi.

```shell
whale ir run initialized.wir --function @f0
```

```text
i32 42
```

## Salinan byte dan reset inisialisasi

Simpan modul lengkap berikut sebagai `tracked-memory.wir`. @f0 menyalin byte pointer serta metadata terpisah, lalu membaca 42. Pada @f1, uninit menandai rentang penyimpanan tipe belum diinisialisasi dan menghapus metadata pointer tanpa mengubah byte yang ada. memcpy dapat menyalin byte belum diinisialisasi; pembacaan nilai tujuan berikutnya memeriksa statusnya. Salinan parsial mempertahankan wewenang hanya jika kedelapan fragmen metadata konsisten tersedia. Menulis bit identik dengan integer atau memset tidak memulihkan wewenang.

memcpy/memset berpanjang nol berhasil tanpa pemeriksaan akses atau alignment, termasuk pointer null dan one-past. memcpy tumpang tindih yang tidak kosong menghasilkan trap. memset menginisialisasi byte tertulis dan menghapus metadata pointer-nya. Penyimpanan Bool harus 0 atau 1; membaca representasi lain menghasilkan trap.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "copied_pointer": whale () -> u32, linkage internal
  declare @f1 "uninitialized": whale () -> u32, linkage internal

  fn @f0 "copied_pointer"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    %v1: u32 = const u32 42
    store u32 %v1, ptr<u32> %v0, align 4
    %v2: ptr<ptr<u32>> = alloca ptr<u32>, align 8
    %v3: ptr<ptr<u32>> = alloca ptr<u32>, align 8
    store ptr<u32> %v0, ptr<ptr<u32>> %v2, align 8
    %v4: ptr<u8> = bitcast ptr<ptr<u32>> %v2 to ptr<u8>
    %v5: ptr<u8> = bitcast ptr<ptr<u32>> %v3 to ptr<u8>
    %v6: u64 = const u64 8
    memcpy ptr<u8> %v5, ptr<u8> %v4, u64 %v6, align 8
    %v7: ptr<u32> = load ptr<u32>, ptr<ptr<u32>> %v3, align 8
    %v8: u32 = load u32, ptr<u32> %v7, align 4
    ret u32 %v8
  }

  fn @f1 "uninitialized"() -> u32, entry %b0 {
  %b0 "entry":
    %v0: ptr<u32> = alloca u32, align 4
    uninit u32, ptr<u32> %v0, align 4
    %v1: u32 = load u32, ptr<u32> %v0, align 4
    ret u32 %v1
  }

}
```

```shell
whale ir run tracked-memory.wir --function @f0
whale ir run tracked-memory.wir --function @f1
```

```text
u32 42
Error: tracked-memory.wir: trap at @f1 %b0 instruction 2 (%v1): uninitialized byte at allocation offset 0 (after 3 steps)
```

## Kewajiban verifikasi, lowering dan eksekusi

| Tahap | Kewajiban |
| --- | --- |
| Verifier | Periksa operand/tipe pointer, bentuk alignment dan ukuran penyimpanan target; tolak undef. |
| O0 lowering | Pertahankan uninit di deklarasi yang dieksekusi, simpan nilai awal dan pertahankan pembacaan. |
| Runtime | Periksa identitas alokasi, generasi/masa hidup, rentang, izin, alignment, overflow alamat dan inisialisasi; teruskan status salinan. |
