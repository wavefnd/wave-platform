---
translation_set_id: whale-overview
path: whale/overview
locale: vi
group: whale
group_order: 1
order: 1
title: Tài liệu Whale
summary: Hướng dẫn về các thành phần của chuỗi công cụ, các lệnh có sẵn và tài liệu tham khảo.
---

## Giới thiệu

Whale là chuỗi công cụ biên dịch có mục đích chung để triển khai ngôn ngữ lập trình và các công cụ biên dịch. Cung cấp các biểu thức trung gian đã nhập (IR), trình biên dịch AMD64, thư viện tệp đối tượng và các hàm dựa trên trình liên kết. Mỗi thành phần được sử dụng thông qua thư viện Rust và lệnh `whale`.

IR thể hiện ngữ nghĩa của một phép tính độc lập với mã hóa ngôn ngữ máy. Trình biên dịch mã hóa các lệnh máy để tạo ra các đối tượng có thể định vị lại được. Thư viện đối tượng đại diện cho các phần, ký hiệu và sự sắp xếp lại. Trình liên kết giải quyết các tham chiếu giữa các đối tượng và đặt tệp thực thi.

## Xây dựng và sử dụng các công cụ

|làm việc|tài liệu|
| --- | --- |
|Mỗi công cụ làm gì| [Thành phần chuỗi công cụ](/docs/vi/whale/ecosystem) |
|Wave Xây dựng chương trình/liên kết/lựa chọn mục tiêu| [Xây dựng và liên kết](/docs/vi/whale/build-link-targets) |
|Quản lý gói/phụ thuộc| [Vex](/docs/vi/whale/vex-package-manager) |
|Whale Thực thi lệnh| [Whale CLI](/docs/vi/whale/whale-cli) |

## Hướng dẫn tài liệu

|tài liệu tham khảo|nội dung|
| --- | --- |
| [Xem IR](ir-reference) |Loại, giá trị, chức năng, luồng điều khiển, xác thực, định dạng trao đổi|
| [Các phép toán số](numeric-operations) |Số học số nguyên, dịch chuyển, chuyển đổi kiểu, dấu phẩy động|
| [Mô hình bộ nhớ](memory-model) |Khởi tạo, xác thực con trỏ, tính toán địa chỉ, bố cục, chuỗi|
| [Gỡ lỗi bằng O0](o0-debugging) |Lưu giữ tính toán, lưu trữ biến, thông tin gỡ lỗi|
| [AMD64 Mục tiêu](amd64-target) |Mã định danh mục tiêu, quy ước gọi, phạm vi chức năng native|
| [Trình biên dịch và trình liên kết](assembler-linker) |Tập hợp các toán hạng, phần, ký hiệu, liên kết tĩnh|

Tài liệu tham khảo xác định các quy tắc ngữ nghĩa cho Whale. Các tính năng có sẵn theo bảng dưới đây. Các thao tác được mô tả trong tài liệu tham khảo có thể không có sẵn trong tất cả các bản dựng.

## Trạng thái hỗ trợ

|thành phần|Giao diện được cung cấp|giới hạn|
| --- | --- | --- |
|người lắp ráp|Tạo đối tượng có thể định vị lại ELF64 từ tập hợp AMD64|Phạm vi bao phủ không đầy đủ của các lệnh và chỉ thị|
|thư viện đối tượng|Thành phần đối tượng, payload không có BSS, xác thực mục tiêu, AMD64 ELF64 tuần tự hóa và triển khai bản ghi Wave có thể chọn để kiểm tra kích thước.|Các đối tượng có thể định vị lại không thể thực thi được|
| IR | Xây dựng, in ấn, các cuộc gọi trực tiếp/gián tiếp được kiểm tra chữ ký, hạ thấp phiên bản AST, xác thực mục tiêu và bố cục loại đã kiểm tra | Có AST format 2 / typed IR format 4, hằng bit chính xác, đọc, kiểm tra và in khứ hồi văn bản; chưa có phát sinh lời gọi máy |
|người liên kết|Giải thích biểu tượng, xác minh mục tiêu đầu vào, kiểm tra vị trí phần tệp/bộ nhớ|Việc áp dụng chuyển vị trí đầy đủ và xuất ra các tập tin thực thi không được hỗ trợ.|
| Thực thi và gỡ lỗi | Số nguyên/Bool, luồng điều khiển, ngăn xếp theo dõi và kiểm tra khởi tạo | Float, lời gọi, địa chỉ global, sinh mã native và DWARF chưa hỗ trợ |

Các quy tắc cấm hành vi không xác định áp dụng cho IR đã được xác minh và theo dõi ký ức. Việc triển khai thử nghiệm chưa triển khai tất cả các kiểm tra thời gian chạy trong tài liệu tham khảo bộ nhớ/thực thi.

## Xây dựng và lắp ráp

Xây dựng với Rust 1.86.0 trở lên.

```sh
git clone https://github.com/wavefnd/Whale.git
cd Whale
cargo build --release --locked
```

Lưu tập hợp sau dưới dạng `answer.asm`.

```asm
section .text
global answer

answer:
    mov eax, 42
    ret
```

Tạo một đối tượng ELF64.

```sh
./target/release/whale asm --amd64 answer.asm -o answer.o
```

Nó sử dụng trình biên dịch riêng của Whale nên không cần trình biên dịch bên ngoài. Đầu ra chứa một hàm có thể gọi được và không chứa bất kỳ mã khởi động quy trình nào.

Các lệnh AST→IR thử nghiệm có sẵn bằng cách xây dựng với `--features socket-cli`. Để biết lệnh CLI, hãy tham khảo [Tham khảo lệnh](/docs/vi/whale/whale-cli).

## Sử dụng và chẩn đoán thư viện

IR phải được xác minh trước khi thực thi hoặc tạo mã. Lỗi nhập và sử dụng sai builder sẽ trả về lỗi cấu trúc. Lỗi đầu vào trong thư viện không được phép chấm dứt quá trình lưu trữ hoặc ghi đè lên nội dung hiện có. Các công cụ chấp nhận đầu vào không đáng tin cậy sẽ có thể điều chỉnh giới hạn tài nguyên.

Phiên bản, đầu vào, mục tiêu và cấu hình của chuỗi công cụ giống nhau sẽ tạo ra đầu ra xác định. Siêu dữ liệu của bản phân phối xác định phiên bản·commit·và các tính năng có sẵn. CI kiểm tra đầu vào tốt, đầu vào bị từ chối, trap, O0 lưu giữ, round-trip, native ngữ nghĩa thực thi cho từng giao diện. Để biết các lệnh phát triển/xác minh, vui lòng tham khảo [Whale Thông tin đóng góp](https://github.com/wavefnd/Whale/blob/master/CONTRIBUTING.md).
