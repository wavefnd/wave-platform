---
translation_set_id: install
path: getting-started/install
locale: id
group: getting-started
group_order: 1
order: 2
title: Menginstal Wave
summary: Instal Wave di Linux, macOS, atau Windows, lalu jalankan program pertama Anda.
---

## Kebijakan instalasi

Installer hanya memasang rilis publik terbaru yang memiliki nomor versi, termasuk prarilis bernomor versi. Draft dan Nightly tidak dipilih. Untuk versi lama dan Nightly, gunakan instalasi manual di bawah. Jika rilis terbaru tidak menyediakan paket platform Anda, instalasi berhenti tanpa memilih versi lama.

| Sistem operasi | Arsitektur |
| --- | --- |
| Linux | amd64, arm64, riscv64, loong64 |
| macOS | amd64, arm64 |
| Windows | amd64, arm64 |
| FreeBSD | amd64 |

## Linux, macOS, dan FreeBSD

Memerlukan Bash, curl, jq, tar, dan alat SHA-256. Di FreeBSD, pasang bash, curl, dan jq terlebih dahulu. Linux membutuhkan glibc dan pustaka sistem yang kompatibel; macOS membutuhkan Apple Command Line Tools; FreeBSD membutuhkan sistem dasar yang kompatibel.

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash
```

## Windows

Jalankan di PowerShell. Installer mendeteksi amd64 atau arm64 dan memilih paket MSVC. Memerlukan runtime Visual C++, Windows SDK, dan pustaka MSVC/UCRT. Shell pengembang Visual Studio dapat menyediakan lingkungan untuk pemeriksaan kompilasi.

```powershell
irm https://wave-lang.dev/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

## Opsi instalasi

Secara bawaan, Wave dan Vex terbaru dipasang jika paket Vex tersedia untuk platform tersebut. Jika tidak, hanya Wave yang dipasang dengan pemberitahuan. Kegagalan unduhan atau verifikasi Vex menjadi kesalahan dan tidak dilewati diam-diam.

| | Opsi Bash | Opsi PowerShell |
| --- | --- | --- |
| Wajibkan Vex | `--with-vex` | `-WithVex` |
| Pasang Wave saja | `--without-vex` | `-WithoutVex` |
| Jangan ubah PATH | `--no-modify-path` | `-NoModifyPath` |

Gunakan perintah yang sama untuk memperbarui. Instalasi lama dipertahankan hingga compiler baru dan std bawaan lulus pemeriksaan eksekusi; jika gagal, instalasi lama dipulihkan. Lokasi bawaan adalah ~/.wave/bin di Unix dan %LOCALAPPDATA%\Wave\bin di Windows. Gunakan WAVE_INSTALL_DIR untuk direktori instalasi khusus. Buka terminal baru setelah PATH dikonfigurasi.

## Instalasi manual versi lama dan Nightly

Unduh .tar.gz atau .zip untuk versi, OS, dan arsitektur yang diinginkan dari [GitHub Releases](https://github.com/wavefnd/Wave/releases). Ekstrak dengan mempertahankan jalur relatif wavec, std, llvm, dan seluruh berkas bawaan. Tambahkan direktori wavec ke PATH. Ikuti catatan rilis lama untuk struktur dan persyaratannya. Opsi --version, --vex-version, -Version, dan -VexVersion tidak didukung.

## Eksekusi pertama

Installer mengompilasi dan menjalankan program kecil dengan std bawaan. Tidak perlu mengunduh std terbaru secara terpisah. Jalankan contoh berikut setelah instalasi. Jika Vex terpasang, periksa juga dengan vex --version.

```shell
wavec --version
```

Buat berkas teks biasa bernama `main.wave` di direktori pilihan Anda dan simpan seluruh program di bawah ini. Pastikan namanya bukan `main.wave.txt`.

<!-- wave-example: install-stdlib -->
```wave
import("std::string::len")::{
    len
};

fun main() {
    println("Wave: {} bytes", len("Wave"));
}
```

Buka terminal di direktori yang berisi `main.wave`, atau pindah ke direktori tersebut dengan `cd`, lalu jalankan:

```shell
wavec run main.wave
```

Keluaran yang diharapkan:

```text
Wave: 4 bytes
```
