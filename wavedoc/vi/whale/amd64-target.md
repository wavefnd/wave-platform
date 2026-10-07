---
translation_set_id: whale-amd64-target
path: whale/amd64-target
locale: vi
group: whale
group_order: 1
order: 10
title: Mục tiêu AMD64 và ABI
summary: Linux AMD64 Mô tả các thuộc tính mục tiêu, ranh giới cuộc gọi và phạm vi chức năng native.
---

## định danh mục tiêu

Mã định danh mục tiêu cho hồ sơ native là `x86_64-whale-linux`.

|thuộc tính|giá trị|
| --- | --- |
|hệ điều hành| Linux |
|tập lệnh| AMD64 |
|thứ tự byte| Little endian |
|Native Độ rộng địa chỉ|64 bit|
|định dạng đối tượng| ELF64 |
|C Giao thức gọi điện| SysV AMD64 ABI |
|định dạng tập tin thực thi tĩnh| ELF ET_EXEC |

Xây dựng máy chủ và mục tiêu đầu ra là các khái niệm khác nhau. Chỉ vì bạn có thể chạy Whale trên máy chủ khác không có nghĩa là nó có thể xuất ra bộ lệnh hoặc định dạng đối tượng của máy chủ đó. Vui lòng tham khảo [Trạng thái hỗ trợ](overview) để biết đường dẫn biên dịch đã triển khai.

## Lựa chọn và xác minh mục tiêu

Lệnh `ir lower` thử nghiệm sử dụng `x86_64-whale-linux` làm mục tiêu mặc định và duy nhất được hỗ trợ. Để sử dụng lệnh này, hãy xây dựng dưới dạng `--features socket-cli`.

```sh
whale ir lower program.json --target x86_64-whale-linux
```

Mục tiêu đầu ra cung cấp bố cục dữ liệu little-endian 64-bit bất kể máy chủ xây dựng. Các mã định danh không xác định hoặc các tổ hợp không được hỗ trợ như `aarch64-whale-linux`, `x86_64-whale-windows` sẽ không thành công kèm theo lỗi hướng dẫn mục tiêu được hỗ trợ trước khi đọc đầu vào hoặc thay thế tệp đầu ra. `--no-verify` không tắt tính năng kiểm tra lựa chọn mục tiêu.

Nhập vào chỗ trống AST(`{"format_version":2,"semantics_version":1,"features":[],"program":{"declarations":[],"globals":[],"functions":[]}}`) sẽ xuất ra các tiêu đề và mô-đun sau:

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

}
```

Nếu bạn chỉ định một mục tiêu không được hỗ trợ, mục tiêu đó sẽ kết thúc ở trạng thái không thành công.

```sh
whale ir lower program.json --target banana
```

```text
Error: unsupported target "banana"; supported targets: x86_64-whale-linux
```

Trong Rust, bạn có thể chọn mục tiêu bằng `ir::Target::lookup("x86_64-whale-linux")` và chuyển `name()` và `data_layout()` sang `lower_o0`. Lowering từ chối nếu bố cục được phân phối khác với mục tiêu đã chọn. Kiểm tra tên mục tiêu không được hỗ trợ và sự không nhất quán về bố cục trong IR tự định cấu hình cũng như `verify_module`.

Mô hình đối tượng lưu trữ `format`, `machine`, `endian` và `address_bits` trong `ObjectTarget`. `ObjectFile::with_target` lưu giữ thông tin nhận dạng đã chỉ định và việc xê-ri hóa chỉ cho phép kết hợp AMD64·little-endian·64-bit ELF64. Sự hiện diện của mã định danh machine cho một kiến ​​trúc khác không ngụ ý hỗ trợ cho bộ mã hóa đó. Cả hai điểm nhập ELF writer đều kiểm tra siêu dữ liệu và đầu vào của trình liên kết được kiểm tra trước khi phân giải hoặc liên kết ký hiệu. Tiêu đề ELF dành cho các đối tượng được hỗ trợ vẫn giữ nguyên `EM_X86_64`.

`ObjectFile::new(ObjectFormat::ELF64)` là một hàm tạo tiện lợi sử dụng mã định danh AMD64 này như trước đây. Các mã đã truy cập trước đây `object.format` phải sử dụng `object.target.format`.

Việc kiểm tra này cung cấp lựa chọn mục tiêu và nhận dạng đối tượng. Kích thước của cấu trúc, mảng, v.v., các trường offset, stride có thể được tìm kiếm bằng bố cục IR API. Đọc tệp đối tượng, native ABI lowering, liên kết đến tệp thực thi chưa được hỗ trợ. Đại lượng vô hướng lowering tiếp tục chỉ định căn chỉnh rõ ràng và các quy tắc bố cục kiểu phức tạp đầy đủ được xác định trong [Tài liệu tham khảo bộ nhớ](memory-model).

## Gọi và ký

Trình xác minh cuộc gọi và khai báo IR chấp nhận các quy ước rõ ràng `Whale` và `SysV64`. Quy ước là một phần của loại `fnptr` và phải khớp với địa điểm cuộc gọi. Chữ ký đa dạng và chữ ký tổng hợp SysV64 bị từ chối. Xem [ví dụ về xây dựng lệnh gọi có thể thực thi](ir-reference). Điều này xác nhận hợp đồng IR; Phân loại ABI của máy, thanh ghi đối số/vị trí ngăn xếp và phát lệnh gọi gốc vẫn chưa được triển khai.

Các cuộc gọi được hỗ trợ yêu cầu chữ ký rõ ràng và quy ước gọi điện. Chữ ký không được hỗ trợ là một lỗi. Phần phụ trợ không được thực hiện các lệnh gọi tương tự do thiếu đối số hoặc thay thế chúng bằng các biểu thức khác.

Hỗ trợ chữ ký được chia thành số nguyên/con trỏ cơ bản, f32/f64, cấu trúc/giá trị rộng và đối số biến. Chỉ vì có một loại trong hệ thống loại IR không có nghĩa là việc truyền hoặc trả về các đối số của loại đó được hỗ trợ trong ABI. Khi chọn một chương trình phụ trợ, bạn nên kiểm tra xem mỗi chương trình có hỗ trợ các danh mục bạn cần hay không.

Ví dụ: i32 chữ ký trả về cấu trúc không được hỗ trợ bởi phần phụ trợ xử lý trả về sẽ bị từ chối. Quy ước trả về vô hướng không thể được áp dụng chỉ vì một phần của cấu trúc được đưa vào một thanh ghi vô hướng.

## Cuộc gọi nội bộ và ranh giới C

native Địa chỉ con trỏ là 64 bit. shadow metadata của con trỏ theo dõi phải được chuyển cùng nhau ngay cả sau khi sao chép, lưu, đối số và trả về.

C ABI và giao thức phân phối siêu dữ liệu nội bộ là khác nhau. C Cuộc gọi xuyên biên giới yêu cầu bộ chuyển đổi rõ ràng. Chỉ vì bạn đã chuyển một địa chỉ số tới C ABI không nên được coi là bảo toàn danh tính, thời gian tồn tại, phạm vi hoặc quyền truy cập của nhiệm vụ.

## khung ngăn xếp

Giữ nguyên con trỏ khung và không sử dụng red zone. Trong O0, không gian lưu trữ của các biến cục bộ khác nhau không được sử dụng lại. Thông tin gỡ lỗi khung cuộc gọi mô tả sự tương ứng của khung native với chương trình gốc.

Quy tắc này chỉ nhằm mục đích gỡ lỗi và không nhằm mục đích hỗ trợ ngoại lệ unwinding. Vui lòng tham khảo [Gỡ lỗi bằng O0](o0-debugging).

## Hạn chế hồ sơ

Hồ sơ O0 native đầu tiên không bao gồm các tính năng sau:

- O1 Tối ưu hóa, vector hóa, LTO.
- Bộ nhớ dùng chung và các hoạt động atomic.
- Ngoại lệ unwinding, chức năng async, coroutine.
- GC và kiểm tra quyền sở hữu theo ngôn ngữ cụ thể.
- Chuyển quyền sở hữu bộ nhớ ngoài tùy ý.
- Hỗ trợ ràng buộc lắp ráp nội tuyến đầy đủ.
- Liên kết động, TLS, bộ lệnh bổ sung, các loại đối tượng bổ sung.

Những yêu cầu nằm ngoài phạm vi hỗ trợ nên bị từ chối thay vì lặng lẽ thay thế bằng một tính năng khác. Hạn chế này dành cho đường dẫn biên dịch native và không có nghĩa là loại bỏ các thành phần chuỗi công cụ được cung cấp độc lập khác.
