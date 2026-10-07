---
translation_set_id: whale-ir-reference
path: whale/ir-reference
locale: id
group: whale
group_order: 1
order: 6
title: Referensi Whale IR
summary: Menjelaskan jenis, pengidentifikasi, validitas fungsi, urutan evaluasi, dan format pertukaran.
---

## Modul dan Pengidentifikasi

Modul terdiri dari informasi target, definisi global, dan fungsi. Nilai memiliki tipe eksplisit. Frontend menyelesaikan nama, jenis, kelebihan beban, dan generik bahasa sumber dan menghasilkan typed IR.

Fungsi dan variabel global menggunakan namespace internal yang berbeda. Oleh karena itu, fungsi dan variabel dapat memiliki nama yang sama. Pengidentifikasi internal berbeda dari nama koneksi eksternal `link_name`, dan nama eksternal ditentukan oleh frontend. Whale tidak menyelesaikan konflik eksternal dengan membuat nama baru secara otomatis. Silakan merujuk ke [Simbol dan Tautan](assembler-linker).

Setiap definisi nilai memiliki pengidentifikasi. Definisi tidak dapat diduplikasi, dan tipe metadata harus cocok dengan tipe yang ditentukan dalam definisi. Nama saja tidak dapat mengidentifikasi definisi, meskipun deklarasi dengan nama yang sama mengaburkan satu sama lain.

## IR Konfigurasi dan pembacaan

Di bawah ini adalah contoh lengkap Rust yang membangun dan memverifikasi fungsi dengan peti `ir`, lalu menghasilkan keluaran typed IR.

```rust
use ir::{ModuleBuilder, Target, Type};

fn main() {
    let target = Target::lookup("x86_64-whale-linux").unwrap();
    let mut module = ModuleBuilder::new(target.name(), target.data_layout());
    let mut function = module.begin_function("answer", vec![], Type::I32);
    let left = function.const_i32(40);
    let right = function.const_i32(2);
    let answer = function.add(Type::I32, left, right);
    function.ret(Some(answer));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    print!("{}", ir::print_module(&module));
}
```

Output printer IR:

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "answer": whale () -> i32, linkage internal

  fn @f0 "answer"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 40
    %v1: i32 = const i32 2
    %v2: i32 = add i32 %v0, %v1
    ret i32 %v2
  }

}
```

Simpan keluaran sebagai `answer.wir` untuk dibaca dan diverifikasi melalui parser teks atau CLI. Eksekusi integer skalar dijelaskan di bawah.

Simpan keluaran sebagai `answer.wir` untuk dibaca dan diverifikasi dengan parser teks dan CLI. Eksekusi IR belum tersedia.

## mengetik

|mengetik|artinya|
| --- | --- |
| `bool` |Nilai logika false atau true|
| `i1`, `i8`, `i16`, `i32`, `i64`, `i128` |bilangan bulat bertanda dengan lebar bit tertentu|
| `u1`, `u8`, `u16`, `u32`, `u64`, `u128` |bilangan bulat tak bertanda dengan lebar bit tertentu|
| `f16`, `f32`, `f64` |Nilai floating point dengan lebar bit tertentu|
| `ptr<T>` |T Penunjuk untuk mengetik nilai|
| `fnptr<signature>` | Pointer yang dapat dipanggil dengan parameter/tipe hasil yang tepat dan konvensi pemanggilan |
| `array<T, N>` |N elemen bertipe sama|
| `struct{T, ...}` |Bidang struktur yang dipesan|
| `tuple<T, ...>` |elemen tupel yang dipesan|
| `void` |Tidak ada hasil|

`bool`, `i1`, dan `u1` adalah tipe yang berbeda. signed `i1` mewakili −1 dan 0, dan unsigned `u1` mewakili 0 dan 1. Bilangan bulat 1 bukan merupakan kondisi logika implisit. Cabang bersyarat, kondisi Select, `trap_if` memerlukan operan Bool.

Ukuran penyimpanan tidak hanya ditentukan oleh jumlah bit dalam nilainya dan mengikuti [tata letak sasaran](memory-model). Misalnya, nilai `i1` adalah 1 bit, tetapi memakan setidaknya 1 byte di memori.

## Fungsi dan Panggilan

Fungsi ini menentukan semua parameter, tipe hasil, konvensi pemanggilan, dan linkage. Panggilan langsung dan tidak langsung harus sesuai dengan tanda tangan penelepon. Panggilan void tidak membuahkan hasil ID. Panggilan ke nonvoid mempertahankan definisi hasil meskipun O0 tidak menggunakan hasilnya.

Pengembaliannya harus sesuai dengan jenis hasil fungsi. Pengembalian void tidak membawa nilai, dan pengembalian nonvoid membawa nilai dari tipe hasil yang dideklarasikan.

### Deklarasi, identitas dan panggilan

`Module.declarations` mencatat `FunctionId` setiap fungsi, nama, tanda tangan lengkap, tautan, dan nama tautan eksternal. Definisi mengacu pada identitas ini; parameter dan tipe kembalian harus sesuai dengan deklarasinya. Deklarasi berulang yang identik diselesaikan ke ID yang sama melalui `declare_function`; konflik dan definisi duplikat adalah kesalahan. Deklarasi internal memerlukan isi modul. Deklarasi eksternal mungkin belum terselesaikan hingga tertaut, atau memiliki badan yang diekspor. Fungsi internal tidak memiliki `link_name`; fungsi eksternal memerlukan nama kosong yang eksplisit tanpa NUL. Dua deklarasi fungsi yang berbeda tidak dapat mengklaim nama eksternal yang sama. Global dan fungsi masih menggunakan namespace internal yang terpisah.

Daftarkan deklarasi sebelum membuat badan dengan `begin_declared_function` untuk mendukung panggilan penerusan dan rekursi. `begin_function` tetap menjadi kemudahan untuk fungsi Whale internal baru. API `declare_function`, `begin_declared_function`, `function_addr`, `null_function` dan `call` yang dicentang akan menghasilkan `Result`; panggilan yang ditolak tidak menambahkan instruksi atau mengalokasikan ID hasilnya.

Program Rust lengkap berikut ini mendeklarasikan fungsi eksternal, mengambil alamat yang diketiknya, dan melakukan panggilan langsung dan tidak langsung:

```rust
use ir::{Callee, CallingConvention, DataLayout, FunctionSignature, Linkage, ModuleBuilder, Type};

fn main() {
    let mut module = ModuleBuilder::new("x86_64-whale-linux", DataLayout::default_64bit_le());
    let signature = FunctionSignature {
        params: vec![Type::I32], ret: Type::I32,
        convention: CallingConvention::SysV64, variadic: false,
    };
    let identity = module.declare_function(
        "identity", signature, Linkage::External, Some("identity_i32".into()),
    ).unwrap();
    let mut function = module.begin_function("answer", vec![], Type::I32);
    let input = function.const_i32(42);
    let callback = function.function_addr(identity).unwrap();
    // The direct call's result remains defined even though it is unused.
    function.call(Callee::Direct(identity), vec![input]).unwrap();
    let result = function.call(Callee::Indirect(callback), vec![input]).unwrap().unwrap();
    function.ret(Some(result));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    print!("{}", ir::print_module(&module));
}
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "identity": sysv64 (i32) -> i32, linkage external, link_name "identity_i32"
  declare @f1 "answer": whale () -> i32, linkage internal

  fn @f1 "answer"() -> i32, entry %b0 {
  %b0 "entry":
    %v0: i32 = const i32 42
    %v1: fnptr<sysv64 (i32) -> i32> = function_addr @f0
    %v2: i32 = call sysv64 i32 @f0(%v0)
    %v3: i32 = call sysv64 i32 indirect %v1(%v0)
    ret i32 %v3
  }

}
```

`Callee::Direct(FunctionId)` diselesaikan melalui tabel deklarasi; `Callee::Indirect(ValueId)` memerlukan nilai `Type::FnPtr(FunctionSignature)`. Tanda tangan mencakup semua tipe parameter, tipe hasil, dan `CallingConvention::{Whale, SysV64}`. Itu disimpan melalui salinan, penyimpanan, parameter, pengembalian, phi dan pilih. Pointer data dan nilai integer tidak dapat dipanggil. Pemeran yang melibatkan tipe penunjuk fungsi ditolak; mengubah anotasi tipe tidak dapat mengubah tanda tangan yang dapat dipanggil. Penunjuk fungsi memiliki penyimpanan alamat 64-bit pada target ini; ini sendiri tidak mengimplementasikan metadata bayangan runtime.

Arity, tipe argumen/hasil yang tepat, keberadaan ID hasil, dan konvensi pemanggilan harus cocok. Tidak ada konversi implisit. Pihak yang dipanggil tidak langsung harus mendominasi panggilan tersebut seperti halnya argumennya. `variadic: true`, parameter bertipe void dan SysV64 parameter agregat/tanda tangan hasil ditolak. Whale tanda tangan agregat dapat direpresentasikan dalam IR; klasifikasi ABI asli dan emisi panggilan mesin belum tersedia untuk kedua konvensi tersebut.

`null_function(signature)` mewakili penunjuk fungsi nol yang diketik. Memanggilnya diketik dengan baik IR dengan runtime trap yang diperlukan sebelum memasukkan callee. Target nonnull yang tidak valid, kedaluwarsa, atau tidak kompatibel dengan tanda tangan yang dicentang juga harus dijebak. Pemeriksaan runtime dan manajemen seumur hidup panggilan balik asing ini menunggu lapisan penerjemah/eksekusi asli; keberhasilan pemverifikasi tidak berarti alamat eksternal yang sewenang-wenang aman.

### Bentuk pemanggilan AST

Ini adalah fragmen ekspresi di dalam program format 2 AST:

```json
{"Call":{"callee":{"Direct":"increment"},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

```json
{"Call":{"callee":{"Indirect":{"FunctionRef":"increment"}},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

`Direct` dan `FunctionRef` menggunakan namespace fungsi meskipun variabel memiliki nama yang sama. `Indirect` mengevaluasi ekspresinya terlebih dahulu, lalu mengevaluasi argumen dari kiri ke kanan. Panggilan batal valid sebagai `ExprStmt`, tetapi tidak sebagai penginisialisasi variabel, argumen, operan, atau nilai yang dikembalikan. Panggilan dan referensi fungsi bukanlah ekspresi konstanta numerik pada waktu kompilasi. `NullFunction` mengambil objek tanda tangan dengan bidang `params`, `ret`, `convention` dan `variadic`.

[Contoh JSON lengkap](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/ast-v2-calls.json) menyimpan panggilan balik dan memanggilnya sebelum panggilan eksternal. Turunkan dengan:

```sh
cargo run --locked --features socket-cli -- ir lower ir/tests/fixtures/ast-v2-calls.json
```

[yang diharapkan IR](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/calls-v3.wir) diperiksa dalam pengujian penurunan. Identitas fungsi dan nama tautan direpresentasikan pada batas IR; melestarikannya melalui pembuatan objek asli dan menghubungkannya masih merupakan pekerjaan terpisah.

## Ketersediaan blok dan nilai

Setiap blok memiliki pengidentifikasi unik dan tepat satu terminator. Target cabang harus memiliki fungsi yang sama. Blok entri harus ada dan tidak boleh memiliki tepi depan dan phi. Saat membuat loop, lakukan cabang dari blok entri ke header loop terpisah.

Definisi nilai di jalur yang dapat dieksekusi harus mengatur titik penggunaannya. Artinya semua jalur dari titik masuk ke titik penggunaan harus melewati definisi tersebut. Di blok yang sama, definisi harus didahulukan sebelum penggunaan. Urutan penyimpanan blok tidak menentukan dominasi.

Asumsikan titik masuk bercabang ke left atau right lalu bergabung di join. Nilai yang ditentukan hanya di left tidak dapat digunakan sebagai nilai umum di join. Hal ini karena jalur yang melewati right tidak terdefinisi. Nilai dari setiap blok sebelumnya harus digabungkan menjadi phi, yang diterima sebagai masukan.

Blok yang tidak dapat dijangkau juga disimpan dalam modul. Verifikator terus-menerus memeriksa pengidentifikasi, jenis, operan, dan struktur cabang blok. Definisi blok yang tidak dapat dijangkau tidak dapat memberikan nilai pada penggunaan normal jalur yang dapat dijangkau.

## Perintah Phi

phi ditempatkan sebelum semua perintah reguler di blok. Tepat satu masukan diperlukan untuk setiap blok sebelumnya yang berbeda. Nilai input harus bertipe phi dan harus tersedia di akhir blok sebelumnya yang sesuai.

Bahkan jika ada beberapa sisi dalam satu blok sebelumnya, hanya ada satu masukan. Loop phi dapat merujuk pada nilai blok yang muncul kemudian dalam urutan penyimpanan modul selama itu adalah nilai yang dihitung pada tepi berulang. Blok sebelumnya yang hilang, terduplikasi, atau tidak relevan dan input yang salah diketik merupakan kesalahan validasi.

## Evaluasi dan Seleksi

Whale AST mengevaluasi target panggilan dan subekspresi dari kiri di tempat yang ditentukan. Frontend mengungkapkan evaluasi hubung singkat sebagai cabang aliran kontrol.

Select memilih salah satu nilai yang sudah dihitung. Itu tidak menghilangkan perhitungan input mana pun. Misalnya, meskipun Anda memilih nilai aman, Anda tidak dapat menghindari trap ditemui saat menghitung input lainnya. Perhitungan yang perlu dijalankan hanya pada jalur tertentu harus ditempatkan di dalam blok bersyarat.

## Departemen Verifikasi trap

IR tidak sah adalah kesalahan verifikasi. Verifier menolak legacy `undef`; hasilkan ulang dari AST. Deklarasi tanpa inisialisasi memakai format 4 `uninit` dan pembacaan aktual yang diperiksa, tanpa penggantian nol atau nilai sembarang. Pelanggaran syarat runtime menghasilkan trap terdefinisi dengan lokasi IR.

`InterpreterTrap` melaporkan alasan, langkah yang dijalankan dan `ExecutionSite`: ID fungsi, ID blok, indeks instruksi mulai nol dan ID hasil opsional. Indeks terminator berada setelah instruksi. CLI juga menyebut file masukan. Typed IR belum memiliki span sumber, jadi ini lokasi IR, bukan nomor baris sumber. Trap mengembalikan error dan menghentikan eksekusi selanjutnya; pustaka tidak menghentikan proses host.

Jaminan ini berlaku untuk IR terverifikasi dan memori penelusuran. Eksternal C·alamat mentah·perakitan inline memiliki kontrak terpisah dan tidak selalu mendeteksi pelanggaran di luar batasannya. Silakan merujuk ke [Model memori](memory-model).

## Format Pertukaran dan Representasi Teks

AST dan typed IR masing-masing menggunakan format version dan semantics version yang umum. Pembaca harus menolak kunci JSON yang tidak berversi/tidak diketahui versi/bidang/fungsi/duplikat. Konstruktor tidak boleh berasumsi bahwa properti yang tidak didukung akan diabaikan secara diam-diam.

Bilangan bulat diteruskan sebagai nomor string lebar bit·signedness·. Konstanta floating point dilewatkan sebagai string bit lebar dan tepat. Teks round-trip di IR harus mempertahankan nama·ID·tipe·konstan·urutan·properti·metadata. Spasi dan penempatan komentar tidak dapat dipertahankan.

Kontrak AST JSON berikut serta pembacaan, verifikasi dan pencetakan bolak-balik typed IR format 4 tersedia.

### Identitas tercetak dan nama dalam tanda kutip

Typed IR format 4 mencetak fungsi sebagai `@fN`, global sebagai `@gN`, nilai sebagai `%vN`, dan blok sebagai `%bN`. ID fungsi dan global berada dalam lingkup modul; ID nilai dan blok berada dalam fungsi yang memuatnya. ID yang diberikan dipertahankan, termasuk celah nomornya. Nama dalam tanda kutip hanya keterangan, bukan dasar penyelesaian referensi. Fungsi menyatakan `entry %bN` secara eksplisit tanpa bergantung pada urutan penyimpanan blok.

Modul lengkap berikut diverifikasi dan dicetak melalui API IR Rust. Kedua blok cabang bernama `"branch"`, tetapi ID membedakan definisi dan masukan phi.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "choose": whale (bool) -> i32, linkage internal

  fn @f0 "choose"(%v0 "condition": bool) -> i32, entry %b0 {
  %b0 "entry":
    cbr bool %v0, label %b1, label %b2
  %b1 "branch":
    %v1: i32 = const i32 1
    br label %b3
  %b2 "branch":
    %v2: i32 = const i32 2
    br label %b3
  %b3 "join":
    %v99: i32 = phi i32 [ %v1, %b1 ], [ %v2, %b2 ]
    ret i32 %v99
  }

}
```

`%v0` didefinisikan dalam daftar parameter. `%b1` dan `%b2` tetap berbeda meskipun namanya sama; phi menunjukkan setiap pendahulu dengan ID. Cabang dan tujuan switch memakai sintaks ID blok yang sama. Pencetak tidak menomori ulang `%v99` yang diberikan secara eksplisit.

Semua bidang nama dan string memakai tanda kutip ganda: target, nama fungsi, global, parameter dan blok, nama tautan eksternal, nama deklarasi konstanta, serta alasan trap. Unicode yang dapat dicetak dipertahankan. Escape yang dipakai ialah `\"`, `\\`, `\n`, `\r`, `\t`, `\0`, dan `\u{hex}` dengan digit heksadesimal huruf kecil untuk karakter kontrol lainnya serta U+2028/U+2029. Nama dengan baris baru, tab, tanda kutip, garis miring terbalik dan teks Korea tetap dicetak sebagai satu rekaman.

```text
"line\ncolumn\tquote\"slash\\한글"
```

Pembaca menerima typed IR formats 3 dan 4 dengan semantics version 1 dan selalu mencetak format 4. `uninit` memerlukan format 4. Instruksi format 3 yang sah tetap terbaca, tetapi legacy `undef` adalah kesalahan verifikasi dalam kedua format dan harus dihasilkan ulang dari AST. Teks format 2 memerlukan migrasi manual ke ID eksplisit, nama bertanda kutip dan referensi masuk. AST JSON memiliki format 2 tersendiri.

### Membaca dan memverifikasi IR teks

`ir::parse_module` membaca typed IR, memverifikasinya dan mengembalikan `Module`. Sintaks printer saat ini untuk skalar, alur kontrol, memori, panggilan langsung/tidak langsung dan ekspresi konstan diterima. ID, nama, tipe, integer dan bit float yang tepat, pohon ekspresi dan hasil evaluasi, urutan blok, masuk, alignment, signature dan link_name dipertahankan. Spasi dan komentar baris `//` menjadi keluaran kanonis. Simpan IR lengkap di atas sebagai `answer.wir` lalu jalankan perintah ini pada build bawaan.

```sh
cargo run --locked -- ir verify answer.wir
cargo run --locked -- ir print answer.wir -o canonical.wir
```

Versi, field, instruksi dan escape tidak dikenal, literal di luar rentang, ID duplikat, anotasi tipe bertentangan dan masukan tambahan ditolak. `ParseError` memberikan offset byte dan baris/kolom Unicode scalar mulai 1. Kesalahan verifikasi dikaitkan dengan fungsi atau deklarasi global terkait jika tersedia. `print` juga memverifikasi dan mempertahankan keluaran yang ada jika gagal. Sintaks atau semantik baru memerlukan perubahan format/semantics version terkait; versi tidak dikenal adalah kesalahan.

### Batas sumber daya verifikasi

Nilai bawaan `IrLimits` adalah input 8 MiB, 1,000,000 token, 1,000,000 node traversal, serta kedalaman tipe dan ekspresi masing-masing 128. Kedalaman akar 0 dan tiap anak menambah 1. Kedalaman dapat diturunkan atau dinaikkan hingga `MAX_IR_NESTING`, 256. Kelebihan mengembalikan `LimitError`, `VerifyError::ResourceLimit`, `ConstEvalError::ResourceLimit` atau `CallError::ResourceLimit`. Node menghitung traversal pada tiap batas input, termasuk anotasi tipe dan ekspresi, bukan waktu berlalu.

```rust
use ir::{parse_module_with_limits, print_module, IrLimits};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let source = std::fs::read_to_string("answer.wir")?;
    let limits = IrLimits {
        max_type_depth: 32,
        max_const_depth: 32,
        max_nodes: 10_000,
        ..IrLimits::default()
    };
    let module = parse_module_with_limits(&source, limits)?;
    let canonical = print_module(&module);
    let reread = parse_module_with_limits(&canonical, limits)?;
    assert_eq!(print_module(&reread), canonical);
    let invalid = source.replacen("format_version 4", "format_version 99", 1);
    assert!(parse_module_with_limits(&invalid, limits).is_err());
    Ok(())
}
```

Batas juga dapat diberikan kepada `verify_module_with_limits`, `ConstExpr::evaluate_with_limits`, `validate_signature_with_limits` dan `ModuleBuilder::declare_function_with_limits`. Traversal tipe iteratif mendahului clone, perbandingan dan diagnostik rekursif; konstanta dievaluasi dengan stack kerja. Pohon Rust yang dipinjam beserta Drop tetap dimiliki pemanggil. Pohon sembarang yang belum diverifikasi masih memiliki clone/Drop rekursif; signature milik API deklarasi checked yang ditolak dibuang iteratif. IR tidak sah adalah kesalahan verifikasi. Verifier menolak legacy `undef`; hasilkan ulang dari AST. Deklarasi tanpa inisialisasi memakai format 4 `uninit` dan pembacaan aktual yang diperiksa, tanpa penggantian nol atau nilai sembarang. Pelanggaran syarat runtime menghasilkan trap terdefinisi dengan lokasi IR.

### Versi yang ditentukan AST JSON

Simpan yang berikut ini sebagai `program.json`. Keempat bidang amplop wajib diisi. `program` berisi array `declarations`, `globals` dan `functions` yang diperlukan, yang mungkin kosong. Nama fungsi, parameter, tipe kembalian, isi, `convention` dan `linkage` wajib diisi. `link_name` mungkin tidak ada/null untuk fungsi internal dan harus berupa string yang tidak kosong tanpa NUL untuk fungsi eksternal. Setiap enum menggunakan nama unitnya atau objek kunci varian tunggal. Varian unit juga menerima objek bernilai null, seperti `{"Void":null}`; encoder mengeluarkan nama unit `"Void"`. `VarDecl.init` mungkin tidak ada atau nol; bidang wajib lainnya harus ada.

```json
{
  "format_version": 2,
  "semantics_version": 1,
  "features": [],
  "program": {
    "globals": [],
    "functions": [
      {
        "name": "answer",
        "parameters": [],
        "return_type": {
          "Int": {
            "bits": 128,
            "signed": false
          }
        },
        "body": [
          {
            "Return": {
              "Lit": {
                "Int": {
                  "bits": 128,
                  "signed": false,
                  "value": "340282366920938463463374607431768211455"
                }
              }
            }
          }
        ],
        "convention": "Whale",
        "linkage": "Internal",
        "link_name": null
      }
    ],
    "declarations": []
  }
}
```

```sh
cargo run --locked --features socket-cli -- ir lower program.json
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "answer": whale () -> u128, linkage internal

  fn @f0 "answer"() -> u128, entry %b0 {
  %b0 "entry":
    %v0: u128 = const u128 340282366920938463463374607431768211455
    ret u128 %v0
  }

}
```

Bilangan bulat `value` adalah string desimal. Signed Angka digunakan setelah minus opsional dari bilangan bulat, dan spasi, plus, eksponen, dan pemisah tidak diperbolehkan. Kisaran yang diperbolehkan ditentukan oleh lebar yang dinyatakan dan signedness. `u128::MAX` di atas dipertahankan sebagaimana adanya melalui JSON dan lowering. Nilai unsigned negatif atau nilai di luar rentang bukan wrap dan merupakan kesalahan. Nilai Float menggunakan string bit heksadesimal dengan lebar persis seperti yang dijelaskan dalam [Operasi numerik](numeric-operations).

`format_version` adalah 2 untuk format AST ini; `semantics_version` adalah 1. `features` harus berupa array kosong. Bidang, versi, fitur, duplikat kunci JSON mentah yang tidak diketahui (termasuk kunci setara yang di-escape), dan nilai akhir merupakan kesalahan, bahkan dengan `--no-verify`. Titik masuk perpustakaan adalah `ir::lower_ast::interchange::decode`; `encode` mengeluarkan amplop. `decode` defaultnya adalah batas byte sumber 8 MiB; `decode_with_limit` menerima batas penelepon. JSON sarang dibatasi. Gunakan dekoder mentah ini daripada menguraikannya ke dalam peta umum yang sudah dapat membuang kunci duplikat.

[Skema JSON lengkap](https://github.com/wavefnd/Whale/blob/master/ir/schema/ast-v2.schema.json) menentukan bentuk, bidang yang wajib diisi, dan varian. Pemeriksaan rentang/jenis dan deteksi kunci duplikat juga berlaku. Subset penurun skalar mencakup literal, variabel/konstanta, tambah/sub/mul, perbandingan, penugasan, jika/sementara, kembali dan putus/lanjutkan. Referensi fungsi, panggilan langsung dan panggilan tidak langsung didukung; ekspresi agregat tidak didukung. `Opaque` dapat diwakili dalam skema tetapi tidak didukung oleh penurunan.

Migrasi memerlukan pembungkusan Program lama tanpa envelope dan penggantian angka JSON dengan string integer desimal atau string bit float. Masukan tanpa versi ditolak. Format 1 harus dimigrasikan ke format 2 dengan menambahkan `program.declarations` (array kosong bila tidak dipakai) dan `convention`/`linkage` eksplisit pada definisi. Versinya independen: AST format 2, typed IR format 4 dan semantics version 1.

### Masukan yang ditolak dan pemulihan CLI

Simpan masukan lengkap berikut sebagai `invalid.json`.

```json
{
  "format_version": 99,
  "semantics_version": 1,
  "features": [],
  "program": {
    "globals": [],
    "functions": [],
    "declarations": []
  }
}
```

```sh
cargo run --locked --features socket-cli -- ir lower invalid.json -o rejected.wir
```

```text
Failed to parse socket JSON: unsupported AST format_version 99; expected 2
```

Perintah keluar dengan status bukan nol dan tidak membuat keluaran baru atau menimpa file yang sudah ada. Jenis ketidakcocokan juga akan gagal sebelum menerbitkan keluaran. Biner dibuat tanpa keluar `socket-cli` dengan status 2 dan mengeluarkan perintah pemulihan yang berisi `--features socket-cli`.


## Interpreter bilangan bulat skalar

Interpreter bawaan menjalankan integer/Bool, alur kontrol, alokasi stack, penyimpanan dan pembacaan pointer data, typed GEP, memcpy dan memset. Pembacaan pasangan checked tidak memeriksa padding. Alamat berupa nilai sintetis 64 bit, tanpa dereferensi memori host. Argumen dan hasil tetap integer/Bool atau void; float, panggilan, pointer fungsi, nilai agregat umum, alamat global dan eksekusi native belum didukung. Alokasi stack hidup sampai fungsi kembali. Akhir masa hidup leksikal, penerusan pointer melalui panggilan/hasil, adaptor memori asing dan native shadow metadata memerlukan implementasi lanjutan.

Nilai bawaan `InterpreterOptions::max_steps` adalah 1,000,000. Setiap instruksi yang dijalankan, termasuk phi, dan setiap terminator menghabiskan satu langkah. Nol berhenti sebelum operasi pertama; loop tanpa akhir mengembalikan `InterpreterError::StepLimit`. `ir_limits` membatasi verifikasi secara terpisah. Aritmetika tanpa penggunaan tetap dijalankan dan dapat trap. Overflow checked adalah hasil Bool; hanya trap_if eksplisit yang menjadikannya trap.

Simpan modul lengkap ini sebagai `swap-loop.wir`. Entry tetap eksplisit meskipun blok keluar disimpan pertama. Saat memasuki blok, semua input phi dibaca dari nilai blok sebelumnya sebelum hasil phi ditulis bersama. Tiga iterasi menukar 11 dan 22 tiga kali dan mengembalikan 22.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f7 "swap_loop": whale (u32) -> i32, linkage internal

  fn @f7 "swap_loop"(%v0 "iterations": u32) -> i32, entry %b11 {
  %b90 "exit":
    ret i32 %v5
  %b11 "entry":
    %v1: i32 = const i32 11
    %v2: i32 = const i32 22
    %v3: u32 = const u32 0
    %v4: u32 = const u32 1
    br label %b20
  %b20 "loop":
    %v5: i32 = phi i32 [ %v1, %b11 ], [ %v6, %b30 ]
    %v6: i32 = phi i32 [ %v2, %b11 ], [ %v5, %b30 ]
    %v7: u32 = phi u32 [ %v3, %b11 ], [ %v9, %b30 ]
    %v8: bool = icmp ult u32 %v7, %v0
    cbr bool %v8, label %b30, label %b90
  %b30 "next":
    %v9: u32 = add u32 %v7, %v4
    br label %b20
  }

}
```

```shell
whale ir run swap-loop.wir --function @f7 --arg 3 --max-steps 100
```

```text
i32 22
```

API Rust mengembalikan nilai dan jumlah langkah atau error terstruktur untuk verifikasi, operasi tidak didukung, argumen, batas langkah atau trap. Program lengkap ini membaca file `swap-loop.wir` yang sama:

```rust
use ir::{interpret_with_options, parse_module, ConstValue, FunctionId,
         InterpreterError, InterpreterOptions};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let source = std::fs::read_to_string("swap-loop.wir")?;
    let module = parse_module(&source)?;
    let options = InterpreterOptions {
        max_steps: 100,
        ..InterpreterOptions::default()
    };
    let result = interpret_with_options(
        &module, FunctionId(7), &[ConstValue::U(3)], options,
    )?;
    assert_eq!(result.value, Some(ConstValue::I(22)));
    assert_eq!(result.steps, 32);
    let stopped = interpret_with_options(
        &module, FunctionId(7), &[ConstValue::U(3)],
        InterpreterOptions { max_steps: 0, ..options },
    );
    assert!(matches!(stopped, Err(InterpreterError::StepLimit { .. })));
    println!("i32 22");
    Ok(())
}
```
