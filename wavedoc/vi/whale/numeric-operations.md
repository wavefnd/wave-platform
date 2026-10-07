---
translation_set_id: whale-numeric-operations
path: whale/numeric-operations
locale: vi
group: whale
group_order: 1
order: 7
title: Các phép toán số
summary: Mô tả các thao tác bao bọc và kiểm tra số nguyên, các ca, lỗi chuyển đổi kiểu và kết quả dấu phẩy động.
---

## biểu diễn số nguyên

Số nguyên bit N có các bit giá trị N. Số nguyên không dấu nằm trong khoảng từ 0 đến 2^N − 1 và số nguyên có dấu nằm trong khoảng từ −2^(N−1) đến 2^(N−1) − 1. signedness của thao tác xác định cách diễn giải chuỗi bit.

Bảng dưới đây giải thích kết quả tính toán. IR Ví dụ về mã sử dụng cách trình bày của máy in hiện tại.

## Cộng, trừ, nhân

Các số nguyên cơ bản add·sub·mul giữ các bit N thấp hơn của kết quả. overflow không gây ra trap. Thao tác checked trả về cùng một kết quả wrap, cũng như Bool, cho biết kết quả toán học có nằm ngoài phạm vi của signed hoặc unsigned tương ứng hay không.

|hoạt động|Wrap Kết quả| Checked overflow |
| --- | --- | --- |
| u8: 255 + 1 | 0 | true |
| i8: 127 + 1 | −128 | true |
| u8: 0 − 1 | 255 | true |
| i8: 12 × 3 | 36 | false |

Giao diện người dùng cho các ngôn ngữ phá vỡ quá trình thực thi tại overflow phải sử dụng `trap_if` rõ ràng cho kết quả overflow của thao tác checked. Thao tác mặc định không ngầm áp dụng chính sách overflow của ngôn ngữ nguồn.

### Wrap và IR thể hiện sự kiểm tra rõ ràng

Lưu mô-đun hoàn chỉnh định dạng 3 này thành `integer-operations.wir`; bộ đọc văn bản chấp nhận và trình thông dịch vô hướng chạy được.

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "add_u8": whale () -> u8, linkage internal
  declare @f1 "require_no_overflow": whale () -> u8, linkage internal

  declare @f2 "minimum_division": whale () -> i8, linkage internal
  declare @f3 "minimum_remainder": whale () -> i8, linkage internal
  declare @f4 "negative_shift": whale () -> i8, linkage internal

  fn @f0 "add_u8"() -> u8, entry %b0 {
  %b0 "entry":
    %v0: u8 = const u8 255
    %v1: u8 = const u8 1
    %v2: u8 = add u8 %v0, %v1
    ret u8 %v2
  }

  fn @f1 "require_no_overflow"() -> u8, entry %b1 {
  %b1 "entry":
    %v3: u8 = const u8 255
    %v4: u8 = const u8 1
    %v5: tuple<u8, bool> = uadd_chk u8 %v3, %v4
    %v6: u8 = extract %v5, 0
    %v7: bool = extract %v5, 1
    trap_if bool %v7, reason="integer overflow"
    ret u8 %v6
  }

  fn @f2 "minimum_division"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -128
    %v1: i8 = const i8 -1
    %v2: i8 = sdiv i8 %v0, %v1
    ret i8 %v2
  }

  fn @f3 "minimum_remainder"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -128
    %v1: i8 = const i8 -1
    %v2: i8 = srem i8 %v0, %v1
    ret i8 %v2
  }

  fn @f4 "negative_shift"() -> i8, entry %b0 {
  %b0 "entry":
    %v0: i8 = const i8 -1
    %v1: i8 = const i8 -1
    %v2: i8 = shl i8 %v0, %v1
    ret i8 %v2
  }

}
```

`add_u8` trả về 0, tám bit thấp của 256. `minimum_division` trả về −128, `minimum_remainder` trả về 0. `negative_shift` diễn giải −1 thành unsigned count 255 rồi dùng 255 mod 8 = 7. Chạy bằng ID hàm tường minh:

```shell
whale ir run integer-operations.wir --function @f0
whale ir run integer-operations.wir --function @f2
whale ir run integer-operations.wir --function @f3
whale ir run integer-operations.wir --function @f4
```

```text
u8 0
i8 -128
i8 0
i8 -128
```

`require_no_overflow` lấy kết quả wrap tại chỉ số 0 và cờ Bool overflow tại chỉ số 1. trap_if tường minh dừng trước return. Lệnh sau thoát với trạng thái 1 và ghi chẩn đoán vào stderr:

```shell
whale ir run integer-operations.wir --function @f1
```

```text
Error: integer-operations.wir: trap at @f1 %b1 instruction 5: "integer overflow" (after 6 steps)
```

## Phép chia và số dư

Phép chia số nguyên hoặc số dư cho 0 gây ra bẫy. Chia giá trị có dấu tối thiểu cho −1 sẽ bao bọc giá trị tối thiểu. Phần còn lại trong trường hợp này là bằng không.

|hoạt động|kết quả|
| --- | --- |
| i8: −128 / −1 | −128 |
| i8: −128 % −1 | 0 |
|Chia cho số nguyên 0| trap |
|phần dư của số nguyên 0| trap |

## sự thay đổi

Khi dịch chuyển giá trị bit N, chuỗi bit của count được hiểu là unsigned và phần còn lại chia cho N được sử dụng. Nó không tạo ra trap chỉ vì count nằm ngoài phạm vi từ 0 đến N−1.

Trong các giá trị 8 bit, count 0·8·16 đều là các dịch chuyển 0 bit. Chuỗi bit `11111111` gồm 8 bit count được dịch chuyển 7 bit. Điều này tương tự ngay cả khi chuỗi bit này đại diện cho signed −1. unsigned Quá trình phân tích diễn ra trước các phép tính còn lại.

Các ngôn ngữ nguồn từ chối tiêu cực hoặc quá mức count phải kiểm tra rõ ràng trước ca làm việc.

## chuyển đổi kiểu

|chuyển đổi|ý nghĩa|
| --- | --- |
| Zero extension |Tăng chiều rộng bằng cách điền các bit bậc cao bằng 0|
| Sign extension |Tăng chiều rộng bằng cách nhân đôi bit dấu|
|cắt bit|Chỉ giữ lại các bit bậc thấp tương ứng với độ rộng đích|
|Beat diễn giải lại|Giải thích cùng một chuỗi bit với các loại khác nhau|
|Chuyển đổi số không mất dữ liệu|Nếu nó không thể được biểu thị trong khi vẫn giữ nguyên giá trị số, trap|

Ví dụ: chuyển đổi 8 bit `11111111` thành 16 bit zero extension trở thành `0000000011111111` và sign extension trở thành `1111111111111111`. Ngay cả khi các bit đầu vào giống nhau thì chúng vẫn là các hoạt động khác nhau.

chuyển đổi float-to-int bỏ phần thập phân theo hướng về 0, sau đó kiểm tra phạm vi số nguyên. NaN và vô cực gây ra bẫy. Khi chuyển đổi sang i8, 127,9 trở thành 127, trong khi 128,0 bẫy. Bool chuyển đổi thành số nguyên 0 hoặc 1. Đã ký i1 không thể đại diện cho 1, vì vậy nó không thể là đích đến của chuyển đổi này.

Việc chuyển đổi một địa chỉ thành một số nguyên không khôi phục quyền truy cập của con trỏ vào số nguyên đó. Vui lòng tham khảo [tính hợp lệ của con trỏ](memory-model).

### Dạng cast và phép toán checked được kiểm tra

Trình kiểm tra so sánh kiểu thực của toán hạng với `src_ty`, rồi kiểm tra loại và độ rộng nguồn/đích mà opcode cho phép. Chú thích kiểu kết quả cũng phải khớp định nghĩa. Không có chuyển đổi ngầm định.

| Opcode | Kiểu được chấp nhận |
| --- | --- |
| `zext`, `sext` | Số nguyên khác Bool; độ rộng đích phải lớn hơn |
| `zext` (Bool) | Bool sang số nguyên 0/1; mọi đích số nguyên trừ `i1` có dấu, bao gồm `u1` |
| `trunc` | Số nguyên khác Bool; độ rộng đích phải nhỏ hơn |
| `fext`, `ftrunc` | Số thực; độ rộng đích phải lớn hơn / nhỏ hơn |
| `itof_s`, `itof_u` | Số nguyên có dấu / không dấu sang số thực |
| `ftoi_s`, `ftoi_u` | Số thực sang số nguyên có dấu / không dấu |
| `bitcast` | Giá trị vô hướng nguyên/thực cùng độ rộng, hoặc con trỏ dữ liệu sang con trỏ dữ liệu |
| `ptrtoint`, `inttoptr` | Con trỏ dữ liệu sang số nguyên / số nguyên sang con trỏ dữ liệu |

Mở rộng và cắt số nguyên là phép toán bit; toán hạng số nguyên có thể khác tính có dấu. Bool là kiểu logic riêng: không cho phép mở rộng dấu, cắt hoặc bitcast sang `i1`/`u1`. Cast của kiểu tổng hợp và con trỏ hàm bị từ chối. Kiểm tra loại con trỏ/số nguyên không xác lập vùng cấp phát hợp lệ hay khôi phục quyền truy cập. Kiểm tra chuyển đổi lúc chạy và lowering sang mã máy vẫn là phần việc riêng chưa hoàn tất.

Chương trình Rust hoàn chỉnh sau in hai chuyển đổi hợp lệ, rồi xác nhận rằng thay `zext` bằng `fext` bị từ chối.

```rust
use ir::*;

fn main() {
    let mut builder = ModuleBuilder::new("x86_64-whale-linux", DataLayout::default_64bit_le());
    let mut function = builder.begin_function("widen", vec![("byte".into(), Type::U8)], Type::Void);
    function.ret(None);
    function.finish();
    let mut module = builder.finish();
    let function = &mut module.functions[0];
    function.value_types.extend([(ValueId(1), Type::U32), (ValueId(2), Type::I32)]);
    function.blocks[0].instructions.extend([
        Instruction::Cast {
            dst: ValueId(1), op: CastOp::ZExt,
            src_ty: Type::U8, src: ValueId(0), dst_ty: Type::U32,
        },
        Instruction::Cast {
            dst: ValueId(2), op: CastOp::Bitcast,
            src_ty: Type::U32, src: ValueId(1), dst_ty: Type::I32,
        },
    ]);
    verify_module(&module).unwrap();
    print!("{}", print_module(&module));
    // An integer source cannot be annotated as a floating widening operation.
    if let Instruction::Cast { op, .. } = &mut module.functions[0].blocks[0].instructions[0] {
        *op = CastOp::FExt;
    }
    assert!(matches!(verify_module(&module), Err(VerifyError::InvalidCast { .. })));
}
```

```text
module {
  format_version 4
  semantics_version 1
  target "x86_64-whale-linux"
  datalayout { ptr=64, endian=little }

  declare @f0 "widen": whale (u8) -> void, linkage internal

  fn @f0 "widen"(%v0 "byte": u8) -> void, entry %b0 {
  %b0 "entry":
    %v1: u32 = zext u8 %v0 to u32
    %v2: i32 = bitcast u32 %v1 to i32
    ret void
  }

}
```

Các đoạn lệnh sau cố ý là IR không hợp lệ.

```text
%v1: i32 = ftoi_s f64 %v0 to i32   // Invalid when %v0 is actually Bool.
%v2: i16 = zext i32 %v3 to i16     // zext cannot narrow.
%v4: i1 = zext bool %v5 to i1      // Signed i1 cannot represent true as 1.
```

Kiểu nguồn thực không khớp tạo `OperandTypeMismatch`; cặp kiểu không hợp lệ tạo `InvalidCast` chứa opcode và cả hai kiểu. Lỗi kiểm tra không chèn chuyển đổi hoặc sửa mô-đun.

`sadd_chk`, `ssub_chk`, `smul_chk` yêu cầu số nguyên có dấu; `uadd_chk`, `usub_chk`, `umul_chk` yêu cầu số nguyên không dấu. Cả hai toán hạng phải đúng kiểu phép toán `T`, kết quả phải là `tuple<T, bool>`. Trích xuất yêu cầu tuple, chỉ số trường tồn tại và đúng kiểu trường đó. Đặc biệt, trích xuất trường tràn dưới kiểu `i1` bị từ chối. Kết quả wrap/tràn đã nêu vẫn là hợp đồng thực thi; các kiểm thử này xác nhận cấu trúc, không phải kết quả chạy bằng backend đã được triển khai.

## số học dấu phẩy động

Các giá trị dấu phẩy động có chuỗi bit chính xác f16·f32·f64. Thao tác làm tròn đến giá trị gần nhất trong chiều rộng đã khai báo và, nếu chính xác ở giữa, sử dụng ties-to-even, chọn một giá trị có số bit có ý nghĩa nhỏ nhất trong số các chữ số có nghĩa.

Hoạt động mặc định là fast-math, ẩn FMA, không cho phép xử lý bắt buộc bằng 0 đối với các giá trị nhỏ. Phép nhân theo sau phép cộng sẽ giữ nguyên các bước làm tròn tương ứng của chúng và phần phụ trợ không được ngầm kết hợp chúng thành một thao tác.

Kết quả của phép toán số có thể là NaN hoặc vô cùng. NaN từ các phép toán và thay đổi độ rộng được chuẩn hóa thành một lượng NaN tĩnh cố định trên mỗi chiều rộng. Lưu và sao chép giữ nguyên các bit NaN gốc. Do đó, hành vi sẽ khác khi chuyển tải trọng NaN vào bộ nhớ mà không cần thực hiện các phép toán số học và khi tính toán nó.

Nó không hiển thị cờ trạng thái dấu phẩy động. Mặc dù số học dấu phẩy động cơ bản cho phép NaN·kết quả vô cực, nhưng phép chuyển đổi float→int áp dụng các quy tắc trap ở trên.


### Lưu trữ các hằng số chính xác

Sử dụng `FloatBits` variant hoặc bất kỳ chuỗi bit thập lục phân nào có chiều rộng chính xác. Đẳng thức của các giá trị được lưu trữ được so sánh với chuỗi bit chứa âm 0 và NaN payload. Trình xác minh sẽ từ chối nếu độ rộng của loại IR và payload khác nhau.

```rust
use ir::{FloatBits, ModuleBuilder, Target, Type};
fn main() {
    let bits = FloatBits::parse(32, "0xffc01234").unwrap();
    assert_eq!(bits, FloatBits::F32(0xffc01234));
    let target = Target::X86_64WhaleLinux;
    let mut module = ModuleBuilder::new(target.name(), target.data_layout());
    let mut function = module.begin_function("payload", vec![], Type::F32);
    let value = function.const_float_bits(Type::F32, bits);
    function.ret(Some(value));
    function.finish();
    let module = module.finish();
    ir::verify_module(&module).unwrap();
    assert!(ir::print_module(&module).contains("const f32 0xffc01234"));
    println!("{}", bits);
}
```

```text
0xffc01234
```

Chuỗi bit f16, f32 hoặc f64 bắt đầu bằng `0x`, theo sau là chính xác 4, 8 hoặc 16 chữ số thập lục phân tương ứng. `0x80000000` đại diện cho f32 số 0 âm; `0x7f800000` đại diện cho cực dương. `const_float` chuyển đổi giá trị máy chủ f64 bằng số; sử dụng `const_float_bits` để bảo toàn các bit gốc. Biểu diễn lưu trữ chính xác không ngụ ý một chương trình phụ trợ thực thi dấu phẩy động hoàn chỉnh. Số học tại thời điểm biên dịch vẫn sử dụng máy chủ trung gian f64, do đó, hợp đồng làm tròn đầy đủ cho mỗi chiều rộng được khai báo vẫn chưa được triển khai.
