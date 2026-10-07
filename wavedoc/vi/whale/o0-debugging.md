---
translation_set_id: whale-o0-debugging
path: whale/o0-debugging
locale: vi
group: whale
group_order: 1
order: 9
title: O0 và gỡ lỗi
summary: Các quy tắc để bảo toàn các phép tính, biểu thức hằng số, dung lượng lưu trữ và thông tin phản hồi gỡ lỗi.
---

## mô hình bảo tồn

O0 bảo toàn các phép tính, biến và luồng điều khiển của typed IR ban đầu cho mục đích gỡ lỗi. Nó cũng giữ lại các kết quả không được sử dụng và các khối không thể truy cập được về mặt cấu trúc. Quá trình xác minh sẽ chẩn đoán lỗi IR và không xóa các khối hoặc đơn giản hóa các thao tác.

Các chuyển đổi cần thiết để tạo ngôn ngữ máy được thực hiện trong một tiểu mục phụ riêng biệt IR. Ngay cả sau khi chuyển đổi, sự tương ứng với bản gốc ID phải được duy trì. Việc bảo tồn typed IR không ngụ ý sự tương ứng một-một của tất cả các hoạt động và hướng dẫn máy IR.

## Chuyển đổi không được thực hiện bởi O0

|chuyển đổi|O0 Vận hành|
| --- | --- |
|nội tuyến|Duy trì ranh giới cuộc gọi và chức năng|
|Chuyển đổi Tail-call|Duy trì cấu trúc cuộc gọi/trả lời chung|
|loại bỏ mã chết|Duy trì các tính toán không sử dụng và các khối không thể truy cập|
|Thu gọn các hằng số thời gian chạy|giữ nguyên hoạt động ban đầu|
|Tổng hợp các phép tính thông thường|Giữ các tính toán riêng biệt|
|Tái sử dụng không gian lưu trữ biến cục bộ|Duy trì từng không gian lưu trữ|
|Bỏ qua con trỏ khung|duy trì con trỏ khung|
|Tự động hợp nhất chuỗi|Duy trì các đối tượng chuỗi riêng biệt|

Ví dụ: một thao tác thời gian chạy thêm hai hằng số vẫn là phép cộng ngay cả khi kết quả không được sử dụng. Các nhánh có điều kiện không đổi cũng duy trì cấu trúc luồng điều khiển ban đầu.

## hằng số thời gian biên dịch

Khai báo hằng số trong thời gian biên dịch bảo toàn cả biểu thức khởi tạo typed và kết quả đánh giá. Điều này khác với việc gấp liên tục trong các hướng dẫn thời gian chạy thông thường.

Một khai báo có biểu thức khởi tạo là `1 + 2` còn lại một biểu thức bổ sung và kết quả là `3`. Ví dụ này minh họa ý nghĩa của biểu thức chứ không phải ngữ pháp khai báo của ngôn ngữ nguồn cụ thể. Kết quả có thể được sử dụng khi xây dựng dữ liệu tĩnh và không thay thế biểu thức khởi tạo bằng phần bổ sung thời gian chạy.

Các khai báo không được sử dụng và không thể truy cập cũng phải được xác định. Ngay cả khi các khai báo có cùng tên che khuất nhau thì chúng vẫn phải được phân tách bằng tên·khai báo ID·tham chiếu. Xác thực từ chối các kết quả được lưu trữ có tham chiếu không hợp lệ, chu kỳ phụ thuộc, loại không hợp lệ hoặc không khớp với biểu thức gốc.

## tuyên bố nguồn không thể truy cập

Các câu sau return·break·continue cũng nằm trong khối không liên kết. Vì tuyên bố này, bạn không nên thay đổi terminator trước đó hoặc tạo đường dẫn thực thi mới. Nó cũng chẩn đoán các biểu thức không hợp lệ trong các câu không thể truy cập được.

Quy tắc bảo toàn này cho phép kiểm tra cấu trúc chương trình gốc. Điều này không có nghĩa là câu lệnh sau terminator sẽ thực sự được thực thi.

## Ví dụ IR được bảo tồn

Dưới đây là đầu ra máy in hiện tại cho mô-đun đã vượt qua trình xác minh. Hiển thị cùng nhau các phép tính không được sử dụng, khai báo thời gian biên dịch và các khối không được liên kết.

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

`%v2` vẫn là `add` ngay cả khi nó không được sử dụng. `%v3` là một `const_decl` riêng biệt, bảo toàn biểu thức `add(i32 1, i32 2)` và kết quả đánh giá 3 cùng nhau. Vì `unreachable.cont` không có cạnh vào nên chúng ta có thể kiểm tra việc thêm `%v6` mà không cần thêm đường dẫn thực thi sau `ret void`.

Việc xác minh sẽ bảo tồn các hướng dẫn và khối này. Ví dụ này hiển thị cấu hình/xác minh/đầu ra IR và không ngụ ý rằng việc thực thi native hoặc đầu ra DWARF được cung cấp.

## Thông tin nguồn và ngăn xếp

Giao diện gỡ lỗi sử dụng thông tin hàm, dòng nguồn, biến cục bộ mặc định và thông tin khung cuộc gọi từ DWARF 5. Ngay cả khi được chuyển đổi thành biểu thức phụ, sự tương ứng giữa thông tin biến hàm/cục bộ và mã định danh IR ban đầu vẫn phải được duy trì.

Cấu hình AMD64 giữ nguyên con trỏ khung và không sử dụng red zone. Thông tin khung cuộc gọi được sử dụng để kiểm tra ngăn xếp và không ngụ ý hỗ trợ ngoại lệ unwinding. trap chấm dứt thực thi mà không đảm bảo hàm hủy hoặc unwinding.

Vui lòng tham khảo [Tổng quan về chuỗi công cụ](overview) để biết tính khả dụng của đầu ra DWARF và khả năng thực thi native. Hành vi của các tối ưu hóa O1 trở lên nằm ngoài phạm vi của tài liệu tham khảo O0 này.

## Khai báo lặp lại trong vòng lặp

Lưu mô-đun đầy đủ này thành `initialization-loop.wir`. Dù lần lặp đầu lưu 42, uninit của khai báo thứ hai đặt lại trạng thái khởi tạo nên lần đọc gây trap. O0 giữ uninit tại khai báo được thực thi dù alloca nằm ở khối vào. Lần đọc không sử dụng nhưng được thực thi vẫn kiểm tra; khối không thể tới được giữ lại không thực thi.

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
