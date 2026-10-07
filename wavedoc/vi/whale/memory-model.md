---
translation_set_id: whale-memory-model
path: whale/memory-model
locale: vi
group: whale
group_order: 1
order: 8
title: Mô hình bộ nhớ
summary: Theo dõi phân bổ, đọc các giá trị khởi tạo, số học con trỏ, bố cục và quy tắc lưu trữ chuỗi.
---

## Theo dõi phân bổ

Một con trỏ được theo dõi liên kết một địa chỉ với ID phân bổ, thế hệ, giới hạn, độ lệch và quyền truy cập. Mô hình bộ nhớ cũng theo dõi thời gian tồn tại và trạng thái khởi tạo của phân bổ. Quyền truy cập phải đáp ứng các điều kiện này; vi phạm gây ra bẫy.

Mặc dù cùng một địa chỉ vật lý được sử dụng lại nhưng các thế hệ vẫn tách biệt nhau trong suốt cuộc đời. Sự hiện diện đơn thuần của một địa chỉ không xác định được con trỏ hợp lệ hay người gọi có quyền truy cập vào không gian lưu trữ đó.

Địa chỉ native duy trì 64 bit. Riêng shadow metadata được chuyển cùng với con trỏ thông qua sao chép/lưu/gọi/trả lại. Phạm vi bộ nhớ native ban đầu là ngăn xếp có thể theo dõi và phân bổ toàn cầu. C Các ranh giới yêu cầu bộ điều hợp rõ ràng và việc chuyển quyền sở hữu bộ nhớ ngoài tùy ý không được bao gồm trong phạm vi này.

## Khởi tạo và đọc

Khai báo không gian lưu trữ không khởi tạo giá trị. Kiểm tra xem phạm vi byte thực sự được đọc đã được khởi tạo hay chưa. Nếu thậm chí một phần của giá trị được đọc ở trạng thái chưa được khởi tạo thì đó là trap. Nó không thay thế kết quả đọc bằng 0 hoặc giá trị không xác định.

Kiểm tra khởi tạo trên giá trị đọc không bao gồm byte padding. Ví dụ: nếu tất cả các trường của cấu trúc được khởi tạo, các giá trị đọc sẽ không sai đơn giản vì các byte trống do căn chỉnh trường chưa được khởi tạo.

Bản sao bộ nhớ mang trạng thái khởi tạo cùng với các byte. Sao chép không gian lưu trữ chưa được khởi tạo sẽ không thay đổi nó thành không gian lưu trữ được khởi tạo. Khi đọc các giá trị từ đích sau này, các bước kiểm tra tương tự sẽ được áp dụng như khi đọc nguồn.

Thực tế là byte vật lý của BSS là 0 không cho phép khởi tạo biến IR.

### IR trong không gian lưu trữ vô hướng được khởi tạo

Mô-đun sau được định cấu hình là builder và đã vượt qua trình xác minh. `store` trước khi đọc giá trị và cả ba hướng dẫn bộ nhớ đều chỉ định cách sắp xếp lũy thừa 2 khác 0.

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

Lưu mô-đun đầy đủ này thành `initialized.wir`; thực thi in `i32 42`. Bỏ store khiến load trap do chưa khởi tạo. Alignment 3 là lỗi kiểm tra. Byte không vật lý từ alloca không có nghĩa đã khởi tạo.

## So sánh với số học con trỏ

Tính toán địa chỉ overflow là trap. Bạn có thể tạo con trỏ one-past trỏ ngay sau phạm vi phân bổ, nhưng bạn không thể truy cập bộ nhớ thông qua nó.

Đẳng thức con trỏ sử dụng nhận dạng phân bổ chứ không chỉ là địa chỉ số. Việc sắp xếp hoặc trừ đi các con trỏ từ các lần phân bổ khác nhau sẽ gây ra bẫy. Việc xây dựng lại địa chỉ từ một số nguyên không khôi phục quyền truy cập.

### GEP

GEP tính toán địa chỉ trên cơ sở phần tử và trường. Đây không phải là lệnh để đọc giá trị.

Chỉ mục đầu tiên là đơn vị phần tử offset của loại được trỏ đến bởi con trỏ cơ sở. Chỉ mục tiếp theo chọn một phần tử mảng hoặc trường cấu trúc/bộ dữ liệu. Chỉ mục trường của cấu trúc/bộ dữ liệu là số thứ tự trường tại thời điểm biên dịch, không phải byte offset. Kiểu được chọn sẽ xác định kiểu con trỏ kết quả.

Khi loại cơ sở là `ptr<array<i32, 4>>`, chỉ mục `[0, 2]` chọn phần tử i32 thứ ba của mảng, dẫn đến `ptr<i32>`. Chỉ định chỉ mục đầu tiên là 1 di chuyển i32 theo một phần tử, thay vì theo một phần tử trong mảng. Ví dụ này minh họa ngữ nghĩa chỉ mục chứ không phải cú pháp lệnh văn bản.

Mục tiêu native ban đầu từ chối số học con trỏ trên các phần tử có kích thước 0. Việc tính toán địa chỉ không loại bỏ thời gian tồn tại, phạm vi, khởi tạo và kiểm tra quyền cần thiết cho lần truy cập tiếp theo.

## bố cục dữ liệu

Mục tiêu đầu ra xác định kích thước·căn chỉnh·trường offset·mảng stride. Thứ tự của các trường trong cấu trúc và bộ dữ liệu giữ nguyên thứ tự khai báo của chúng. Bố cục của máy chủ đang chạy trình biên dịch không được coi là bố cục của mục tiêu đầu ra.

|giá trị|Lưu quy tắc|
| --- | --- |
| Bool, signed i1, unsigned u1 |Ít nhất 1 byte|
|Cấu trúc/bộ dữ liệu trống|Kích thước 0, căn chỉnh 1|
|Mảng|Bố cục mục tiêu xác định phần tử stride|
|Cấu trúc/bộ dữ liệu|Giữ nguyên thứ tự khai báo và yêu cầu căn chỉnh của mục tiêu|

Căn chỉnh hoàn chỉnh của IR là lũy thừa của 2, không phải 0. Việc tự động căn chỉnh phải được xác định trước khi tạo IR này. Bố cục packed·union·bitfield không được hỗ trợ trong hồ sơ này và nên bị từ chối.

### Yêu cầu bố cục đầu ra

Rust API tính toán bố cục lưu trữ độc lập với máy chủ xây dựng. Trong ví dụ sau, có 7 byte padding trước trường u64 và 6 byte ở cuối padding.

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

Kích thước được trả về bởi `layout_of` và bước tiến của mảng bao gồm phần đệm ở cuối. Cấu trúc và bộ dữ liệu sử dụng cùng một quy tắc thứ tự trường. Bool, i1 và u1 mỗi cái chiếm một byte. Các cấu trúc và bộ dữ liệu trống có kích thước 0 và căn chỉnh 1; mảng có độ dài bằng 0 vẫn giữ được sự liên kết tự nhiên của phần tử. `void` không có bố cục lưu trữ, trong khi `ptr<void>` chiếm 8 byte.

Sử dụng tính năng sắp xếp tự nhiên cho các trường và phần tử mảng. `allocation_align` áp dụng quy tắc SysV AMD64, quy tắc này yêu cầu căn chỉnh ít nhất 16 byte, cho các mảng cục bộ/toàn cầu độc lập từ 16 byte trở lên. Nó không làm tăng sự liên kết của các trường mảng hoặc phần tử stride. AST lowering sử dụng tính năng tra cứu sắp xếp hàng loạt này cho bộ nhớ cục bộ.

Phép nhân cường độ, phép cộng trường offset, overflow trong phép tính padding trả về `LayoutError::Overflow`. `layout_of` có giới hạn lồng kiểu phức tạp là 128 cấp độ và `layout_of_with_limit` cho phép người gọi chỉ định giới hạn. Nó không phân bổ không gian lưu trữ bằng số phần tử mảng. `pointer_stride` từ chối pointee với kích thước 0 trong số học con trỏ native, nhưng bản thân bố cục lưu trữ của loại đó là hợp lệ. Packed·union·bitfield không có biểu thức loại được hỗ trợ. Tra cứu bố cục được lưu trữ này không triển khai các quy ước gọi kiểu phức tạp hoặc kiểm tra giới hạn thời gian chạy.

## Chuỗi và ranh giới C

Chuỗi là một chuỗi byte bất biến có độ dài được chỉ định. Mã hóa mặc định là UTF-8. Cho phép NUL nội bộ và không tự động thêm phần kết thúc NUL. O0 không tự động kết hợp các đối tượng chuỗi có cùng nội dung.

Do đó, độ dài của chuỗi byte bao gồm `A`, NUL và `B` là 3. Không thể sử dụng cho các chuyển đổi chuỗi C rõ ràng, từ chối NUL nội bộ. Giao diện người dùng không được cắt ngắn phần này thành `A`.

Bên ngoài C·Địa chỉ thô·Hội nội tuyến là ranh giới hợp đồng riêng biệt. Việc kiểm tra thời gian chạy của bộ nhớ theo dõi không được đảm bảo để phát hiện tất cả hành vi không chính xác trong mã bên ngoài.

## Thực thi bộ nhớ ngăn xếp được theo dõi

Bộ thông dịch mặc định thực thi số nguyên/Bool, luồng điều khiển, cấp phát ngăn xếp, lưu và đọc con trỏ dữ liệu, typed GEP, memcpy và memset. Đọc cặp checked bỏ qua padding. Địa chỉ là giá trị tổng hợp 64 bit, không giải tham chiếu bộ nhớ máy chủ. Đối số và kết quả vẫn giới hạn ở số nguyên/Bool hoặc void; float, lời gọi, con trỏ hàm, giá trị aggregate tổng quát, địa chỉ global và thực thi native chưa hỗ trợ. Cấp phát ngăn xếp tồn tại đến khi hàm trả về. Kết thúc vòng đời theo phạm vi, truyền con trỏ qua lời gọi/kết quả, bộ chuyển đổi bộ nhớ ngoài và native shadow metadata cần triển khai tiếp.

`--max-memory` đặt ngân sách byte cấp phát logic, mặc định 64 MiB. `InterpreterOptions::memory_limits` còn giới hạn 16384 cấp phát, 262144 mảnh metadata byte con trỏ và 256 Mi đơn vị công việc byte/metadata. Vượt giới hạn trả về `MemoryLimit` có vị trí IR, tách biệt với trap của chương trình. Kiểm tra cũng từ chối overflow kích thước lưu trữ đã biết của đích đầu ra trước khi thực thi.

```shell
whale ir run initialized.wir --function @f0
```

```text
i32 42
```

## Sao chép byte và đặt lại khởi tạo

Lưu mô-đun đầy đủ dưới đây thành `tracked-memory.wir`. @f0 sao chép byte con trỏ và metadata riêng, rồi đọc 42. Trong @f1, uninit đánh dấu vùng lưu trữ của kiểu chưa khởi tạo và xóa metadata con trỏ mà không đổi byte hiện có. memcpy có thể sao chép byte chưa khởi tạo; lần đọc giá trị đích sau đó kiểm tra trạng thái. Sao chép từng phần chỉ giữ quyền khi đủ tám mảnh metadata nhất quán. Ghi bit giống nhau bằng số nguyên hoặc memset không phục hồi quyền.

memcpy/memset dài bằng không thành công mà không kiểm tra truy cập hoặc alignment, kể cả con trỏ null và one-past. memcpy chồng lấn không rỗng gây trap. memset khởi tạo byte đã ghi và xóa metadata con trỏ của chúng. Bool phải lưu là 0 hoặc 1; đọc biểu diễn khác gây trap.

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

## Trách nhiệm kiểm tra, lowering và thực thi

| Giai đoạn | Trách nhiệm |
| --- | --- |
| Verifier | Kiểm tra toán hạng/kiểu con trỏ, dạng alignment và kích thước lưu trữ đích; từ chối undef. |
| O0 lowering | Giữ uninit tại khai báo được thực thi, lưu giá trị khởi tạo và giữ lần đọc. |
| Runtime | Kiểm tra danh tính cấp phát, thế hệ/vòng đời, phạm vi, quyền, alignment, overflow địa chỉ và khởi tạo; truyền trạng thái sao chép. |
