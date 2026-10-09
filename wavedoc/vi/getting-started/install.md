---
translation_set_id: install
path: getting-started/install
locale: vi
group: getting-started
group_order: 1
order: 2
title: Cài đặt Wave
summary: Cài đặt Wave trên Linux, macOS hoặc Windows và chạy chương trình đầu tiên.
---

## Chính sách cài đặt

Trình cài đặt chỉ cài bản phát hành công khai mới nhất có số phiên bản, kể cả bản phát hành trước có số phiên bản. Không chọn Draft hoặc Nightly. Với bản cũ và Nightly, hãy cài thủ công theo hướng dẫn bên dưới. Nếu bản mới nhất không có gói cho nền tảng của bạn, quá trình sẽ dừng thay vì chọn bản cũ.

| Hệ điều hành | Kiến trúc |
| --- | --- |
| Linux | amd64, arm64, riscv64, loong64 |
| macOS | amd64, arm64 |
| Windows | amd64, arm64 |
| FreeBSD | amd64 |

## Linux, macOS và FreeBSD

Cần Bash, curl, jq, tar và công cụ SHA-256. Trên FreeBSD, hãy cài bash, curl và jq trước. Linux cần glibc và thư viện hệ thống tương thích; macOS cần Apple Command Line Tools; FreeBSD cần hệ thống cơ sở tương thích.

```shell
curl -fsSL https://wave-lang.dev/install.sh | bash
```

## Windows

Chạy trong PowerShell. Trình cài đặt nhận diện amd64 hoặc arm64 và chọn gói MSVC. Cần Visual C++ runtime, Windows SDK và thư viện MSVC/UCRT. Shell dành cho nhà phát triển Visual Studio có thể cung cấp môi trường kiểm tra biên dịch.

```powershell
irm https://wave-lang.dev/install.ps1 -OutFile install.ps1
powershell -ExecutionPolicy Bypass -File .\install.ps1
```

## Tùy chọn cài đặt

Mặc định cài Wave và Vex mới nhất nếu có gói Vex cho nền tảng. Nếu không có, chỉ cài Wave và thông báo rõ. Lỗi tải xuống hoặc xác minh Vex được báo lỗi, không bị bỏ qua âm thầm.

| | Tùy chọn Bash | Tùy chọn PowerShell |
| --- | --- | --- |
| Yêu cầu Vex | `--with-vex` | `-WithVex` |
| Chỉ cài Wave | `--without-vex` | `-WithoutVex` |
| Không sửa PATH | `--no-modify-path` | `-NoModifyPath` |

Chạy lại cùng lệnh để cập nhật. Giữ bản cài cũ cho đến khi trình biên dịch mới và std đi kèm vượt qua kiểm tra thực thi; khôi phục khi thất bại. Thư mục mặc định là ~/.wave/bin trên Unix và %LOCALAPPDATA%\Wave\bin trên Windows. Dùng WAVE_INSTALL_DIR để chọn thư mục cài riêng. Mở terminal mới sau khi cấu hình PATH.

## Cài thủ công bản cũ và Nightly

Tải .tar.gz hoặc .zip đúng phiên bản, hệ điều hành và kiến trúc từ [GitHub Releases](https://github.com/wavefnd/Wave/releases). Giải nén và giữ đường dẫn tương đối của wavec, std, llvm cùng các tệp đi kèm. Thêm thư mục chứa wavec vào PATH. Với bản cũ, làm theo ghi chú về cấu trúc và yêu cầu của bản đó. Không hỗ trợ --version, --vex-version, -Version và -VexVersion.

## Chạy lần đầu

Trình cài đặt biên dịch và chạy một chương trình nhỏ dùng std đi kèm. Không cần tải std mới nhất riêng. Sau khi cài, chạy ví dụ bên dưới. Nếu có Vex, cũng có thể kiểm tra bằng vex --version.

```shell
wavec --version
```

Tạo tệp văn bản thuần có tên `main.wave` trong thư mục tùy chọn và lưu toàn bộ chương trình bên dưới. Kiểm tra để bảo đảm tên tệp không phải là `main.wave.txt`.

<!-- wave-example: install-stdlib -->
```wave
import("std::string::len")::{
    len
};

fun main() {
    println("Wave: {} bytes", len("Wave"));
}
```

Mở terminal trong thư mục chứa `main.wave` hoặc dùng `cd` để chuyển đến thư mục đó, rồi chạy:

```shell
wavec run main.wave
```

Kết quả mong đợi:

```text
Wave: 4 bytes
```
