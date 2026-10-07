---
translation_set_id: whale-ir-reference
path: whale/ir-reference
locale: vi
group: whale
group_order: 1
order: 6
title: Tài liệu tham khảo Whale IR
summary: Mô tả các loại, mã định danh, tính hợp lệ của hàm, thứ tự đánh giá và định dạng trao đổi.
---

## Mô-đun và mã định danh

Một mô-đun bao gồm thông tin mục tiêu, định nghĩa chung và các chức năng. Các giá trị có loại rõ ràng. Giao diện người dùng phân giải tên, loại, tình trạng quá tải và tên chung của ngôn ngữ nguồn và tạo ra typed IR.

Các hàm và biến toàn cục sử dụng các không gian tên nội bộ khác nhau. Vì vậy, hàm và biến có thể có cùng tên. Mã định danh bên trong khác với tên kết nối bên ngoài `link_name` và tên bên ngoài do giao diện người dùng chỉ định. Whale không giải quyết xung đột bên ngoài bằng cách tự động tạo tên mới. Vui lòng tham khảo [Biểu tượng và liên kết](assembler-linker).

Mỗi định nghĩa giá trị có một mã định danh. Các định nghĩa không thể trùng lặp và siêu dữ liệu loại phải khớp với loại được chỉ định trong định nghĩa. Chỉ riêng tên không xác định được các định nghĩa, ngay cả khi các khai báo có cùng tên sẽ che khuất lẫn nhau.

## IR Cấu hình và đọc

Dưới đây là một ví dụ Rust hoàn chỉnh giúp xây dựng và xác minh hàm với thùng `ir`, sau đó xuất ra typed IR.

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

Máy in xuất ra IR:

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

Lưu đầu ra thành `answer.wir` để đọc và xác minh bằng bộ phân tích văn bản hoặc CLI. Việc chạy số nguyên vô hướng được mô tả bên dưới.

Lưu đầu ra thành `answer.wir` để đọc và kiểm tra bằng bộ phân tích văn bản và CLI. Chưa có khả năng thực thi IR.

## loại

|loại|ý nghĩa|
| --- | --- |
| `bool` |Giá trị logic false hoặc true|
| `i1`, `i8`, `i16`, `i32`, `i64`, `i128` |số nguyên có dấu của độ rộng bit được chỉ định|
| `u1`, `u8`, `u16`, `u32`, `u64`, `u128` |số nguyên không dấu có độ rộng bit được chỉ định|
| `f16`, `f32`, `f64` |Giá trị dấu phẩy động với độ rộng bit được chỉ định|
| `ptr<T>` |T Con trỏ để gõ giá trị|
| `fnptr<signature>` | Con trỏ có thể gọi được với các loại tham số/kết quả chính xác và quy ước gọi |
| `array<T, N>` |N phần tử cùng loại|
| `struct{T, ...}` |Trường cấu trúc có thứ tự|
| `tuple<T, ...>` |các phần tử tuple được sắp xếp|
| `void` |Không có kết quả|

`bool`, `i1` và `u1` là các loại khác nhau. signed `i1` đại diện cho −1 và 0, và unsigned `u1` đại diện cho 0 và 1. Số nguyên 1 không phải là điều kiện logic ẩn. Nhánh có điều kiện, điều kiện Select, `trap_if` yêu cầu toán hạng Bool.

Kích thước lưu trữ không chỉ được xác định bởi số bit trong giá trị và tuân theo [bố trí mục tiêu](memory-model). Ví dụ: giá trị của `i1` là 1 bit nhưng chiếm ít nhất 1 byte trong bộ nhớ.

## Chức năng và cuộc gọi

Hàm chỉ định tất cả các tham số, loại kết quả, quy ước gọi và linkage. Cuộc gọi trực tiếp và gián tiếp phải trùng với chữ ký của người gọi. Cuộc gọi void không có kết quả ID. Lệnh gọi tới nonvoid sẽ giữ nguyên định nghĩa kết quả ngay cả khi O0 không sử dụng kết quả.

Kết quả trả về phải khớp với loại kết quả của hàm. Trả về void không mang giá trị và trả về nonvoid mang giá trị của loại kết quả được khai báo.

### Tuyên bố, nhận dạng và kêu gọi

`Module.declarations` ghi lại `FunctionId`, tên, chữ ký đầy đủ, liên kết và tên liên kết bên ngoài của từng chức năng. Một định nghĩa đề cập đến danh tính này; kiểu tham số và trả về phải khớp với khai báo của nó. Các nội dung khai báo lặp lại giống hệt nhau sẽ chuyển thành cùng một ID thông qua `declare_function`; xung đột và định nghĩa trùng lặp là lỗi. Một khai báo nội bộ cần có phần thân trong mô-đun. Một khai báo bên ngoài có thể chưa được giải quyết cho đến khi liên kết hoặc có nội dung được xuất. Các chức năng bên trong không có `link_name`; các hàm bên ngoài yêu cầu tên không trống rõ ràng và không có NUL. Hai khai báo hàm riêng biệt không thể có cùng tên bên ngoài. Toàn cầu và hàm vẫn sử dụng các không gian tên nội bộ riêng biệt.

Đăng ký các khai báo trước khi xây dựng các phần thân với `begin_declared_function` để hỗ trợ các lệnh gọi chuyển tiếp và đệ quy. `begin_function` vẫn là sự tiện lợi cho chức năng Whale nội bộ mới. Các API đã kiểm tra `declare_function`, `begin_declared_function`, `function_addr`, `null_function` và `call` trả về `Result`; cuộc gọi bị từ chối không nối thêm lệnh hoặc phân bổ ID kết quả của nó.

Chương trình Rust hoàn chỉnh sau đây khai báo một hàm bên ngoài, lấy địa chỉ đã nhập của nó và thực hiện cả lệnh gọi trực tiếp và gián tiếp:

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

`Callee::Direct(FunctionId)` giải quyết thông qua bảng khai báo; `Callee::Indirect(ValueId)` yêu cầu giá trị `Type::FnPtr(FunctionSignature)`. Chữ ký bao gồm tất cả các loại tham số, loại kết quả và `CallingConvention::{Whale, SysV64}`. Nó được giữ lại thông qua các bản sao, lưu trữ, tham số, trả về, phi và chọn. Con trỏ dữ liệu và giá trị số nguyên không thể gọi được. Các kiểu ép kiểu liên quan đến kiểu con trỏ hàm bị từ chối; thay đổi chú thích loại không thể thay đổi chữ ký có thể gọi được. Một con trỏ hàm có bộ lưu trữ địa chỉ 64-bit trên mục tiêu này; bản thân điều này không triển khai siêu dữ liệu bóng thời gian chạy.

Tính chất, loại đối số/kết quả chính xác, sự hiện diện của ID kết quả và quy ước gọi hàm phải khớp. Không có chuyển đổi tiềm ẩn. Người được gọi gián tiếp phải thống trị cuộc gọi giống như các đối số của nó. `variadic: true`, tham số void và SysV64 chữ ký tham số/kết quả tổng hợp bị từ chối. Whale chữ ký tổng hợp có thể được thể hiện trong IR; Phân loại ABI gốc và phát lệnh gọi máy vẫn chưa có sẵn cho cả hai quy ước.

`null_function(signature)` đại diện cho một con trỏ hàm null được gõ. Gọi nó là gõ đúng IR với bẫy thời gian chạy bắt buộc trước khi vào callee. Mục tiêu không có giá trị không hợp lệ, hết hạn hoặc không tương thích với chữ ký đã kiểm tra cũng phải bẫy. Các hoạt động kiểm tra thời gian chạy và quản lý thời gian gọi lại nước ngoài này đang chờ lớp trình thông dịch/thực thi gốc; trình xác minh thành công không có nghĩa là các địa chỉ bên ngoài tùy ý là an toàn.

### Dạng lời gọi trong AST

Đây là các đoạn biểu thức bên trong chương trình định dạng AST 2:

```json
{"Call":{"callee":{"Direct":"increment"},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

```json
{"Call":{"callee":{"Indirect":{"FunctionRef":"increment"}},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

`Direct` và `FunctionRef` sử dụng không gian tên hàm ngay cả khi một biến có cùng tên. `Indirect` đánh giá biểu thức của nó trước, sau đó đánh giá các đối số từ trái sang phải. Lệnh gọi void có giá trị dưới dạng `ExprStmt` nhưng không hợp lệ dưới dạng giá trị khởi tạo biến, đối số, toán hạng hoặc giá trị trả về. Các lệnh gọi và tham chiếu hàm không phải là các biểu thức hằng số tại thời điểm biên dịch. `NullFunction` lấy đối tượng chữ ký với các trường `params`, `ret`, `convention` và `variadic`.

[Ví dụ hoàn chỉnh JSON](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/ast-v2-calls.json) lưu trữ lệnh gọi lại và gọi nó trước lệnh gọi bên ngoài. Hạ nó xuống bằng:

```sh
cargo run --locked --features socket-cli -- ir lower ir/tests/fixtures/ast-v2-calls.json
```

[Dự kiến ​​IR](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/calls-v3.wir) của nó đã được kiểm tra trong các thử nghiệm hạ thấp. Danh tính hàm và tên liên kết được thể hiện ở ranh giới IR; bảo tồn chúng thông qua việc tạo và liên kết đối tượng gốc vẫn là công việc riêng biệt.

## Sự sẵn có của các khối và giá trị

Mỗi khối có một mã định danh duy nhất và chính xác là một terminator. Các mục tiêu nhánh phải thuộc cùng một chức năng. Khối nhập phải tồn tại và không được có cạnh đầu và phi. Khi tạo một vòng lặp, hãy phân nhánh từ khối đầu vào tới một tiêu đề vòng lặp riêng biệt.

Định nghĩa về giá trị trong đường dẫn thực thi sẽ chi phối điểm sử dụng thông thường của nó. Điều này có nghĩa là tất cả các đường dẫn từ điểm vào đến điểm sử dụng đều phải đi qua định nghĩa đó. Trong cùng một khối, định nghĩa phải có trước khi sử dụng. Thứ tự lưu trữ của các khối không quyết định sự thống trị.

Giả sử rằng điểm vào phân nhánh thành left hoặc right rồi nối tại join. Các giá trị chỉ được xác định trong left không thể được sử dụng làm giá trị chung trong join. Điều này là do đường dẫn đi qua right không được xác định. Các giá trị của mỗi khối trước phải được kết hợp thành phi, được nhận làm đầu vào.

Các khối không thể truy cập cũng được bảo tồn trong mô-đun. Trình xác minh liên tục kiểm tra mã định danh, loại, toán hạng và cấu trúc nhánh của khối. Định nghĩa về khối không thể truy cập không thể cung cấp giá trị cho việc sử dụng thông thường đường dẫn có thể truy cập.

## lệnh Phi

phi được đặt trước tất cả các lệnh thông thường trong khối. Cần có chính xác một đầu vào cho mỗi khối trước đó. Giá trị đầu vào phải thuộc loại phi và phải có sẵn ở cuối khối tương ứng trước đó.

Ngay cả khi có nhiều cạnh trong một khối trước đó thì cũng chỉ có một đầu vào. Vòng lặp phi có thể tham chiếu đến giá trị của khối xuất hiện sau trong thứ tự lưu trữ mô-đun miễn là đó là giá trị được tính ở cạnh lặp lại. Các khối trước bị thiếu, trùng lặp hoặc không liên quan và đầu vào được nhập không chính xác là lỗi xác thực.

## Đánh giá và lựa chọn

Whale AST đánh giá mục tiêu cuộc gọi và biểu thức con từ bên trái nơi chúng được chỉ định. Giao diện người dùng thể hiện đánh giá ngắn mạch như một nhánh luồng điều khiển.

Select chọn một trong các giá trị đã được tính toán. Nó không bỏ qua việc tính toán của một trong hai đầu vào. Ví dụ: ngay cả khi bạn chọn một giá trị an toàn, bạn không thể tránh gặp phải trap khi tính toán các đầu vào khác. Các phép tính chỉ cần được thực hiện trên một đường dẫn cụ thể phải được đặt bên trong một khối có điều kiện.

## Phòng Kiểm định trap

IR không hợp lệ là lỗi kiểm tra. Trình kiểm tra từ chối legacy `undef`; hãy tạo lại từ AST. Khai báo chưa khởi tạo dùng format 4 `uninit` và kiểm tra lần đọc thực tế, không thay bằng không hoặc giá trị tùy ý. Vi phạm điều kiện thực thi tạo trap xác định có vị trí IR.

`InterpreterTrap` báo lý do, số bước đã chạy và `ExecutionSite`: ID hàm, ID khối, chỉ số lệnh từ 0 và ID giá trị kết quả tùy chọn. Chỉ số terminator nằm sau các lệnh. CLI cũng nêu tệp đầu vào. Typed IR chưa có span nguồn nên đây là vị trí IR, không phải số dòng nguồn. Trap trả lỗi và dừng chạy tiếp; thư viện không kết thúc tiến trình chủ.

Bảo đảm này áp dụng cho IR đã được xác minh và bộ nhớ theo dõi. Bên ngoài C·địa chỉ thô·tổ hợp nội tuyến có hợp đồng riêng và không phải lúc nào cũng phát hiện ra các vi phạm bên ngoài ranh giới của nó. Vui lòng tham khảo [Mô hình bộ nhớ](memory-model).

## Định dạng trao đổi và trình bày văn bản

AST và typed IR sử dụng format version tương ứng và semantics version chung tương ứng. Người đọc nên từ chối khóa JSON chưa được phiên bản/không xác định. Các nhà xây dựng không nên cho rằng các thuộc tính không được hỗ trợ sẽ bị bỏ qua một cách âm thầm.

Các số nguyên được truyền dưới dạng độ rộng bit·signedness·số chuỗi. Các hằng số dấu phẩy động được truyền dưới dạng chuỗi bit có chiều rộng và chính xác. Văn bản round-trip trong IR phải giữ nguyên tên·ID·loại·hằng·chuỗi·thuộc tính·siêu dữ liệu. Không gian và vị trí bình luận không được bảo tồn.

Có thể dùng hợp đồng AST JSON bên dưới cùng việc đọc, kiểm tra và in khứ hồi typed IR format 4.

### Định danh được in và tên trong dấu ngoặc kép

Typed IR format 4 in hàm dưới dạng `@fN`, giá trị toàn cục là `@gN`, giá trị là `%vN`, khối là `%bN`. ID hàm và toàn cục thuộc mô-đun; ID giá trị và khối thuộc hàm chứa chúng. Các ID được cung cấp được giữ nguyên, kể cả khoảng trống trong số thứ tự. Tên trong ngoặc kép chỉ là chú thích, không dùng để phân giải tham chiếu. Hàm ghi rõ `entry %bN`, độc lập với thứ tự lưu các khối.

Mô-đun hoàn chỉnh sau đã được kiểm tra và in qua API IR Rust. Cả hai khối nhánh đều có tên `"branch"`; ID phân biệt định nghĩa và đầu vào phi.

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

`%v0` được định nghĩa trong danh sách tham số. `%b1` và `%b2` vẫn khác nhau dù cùng tên; phi chỉ rõ từng khối tiền nhiệm bằng ID. Nhánh và đích switch dùng cùng cú pháp ID khối. Bộ in không đánh số lại `%v99` được chỉ định rõ ràng.

Mọi trường tên và chuỗi đều dùng dấu ngoặc kép: đích, tên hàm, toàn cục, tham số và khối, tên liên kết ngoài, tên khai báo hằng và lý do trap. Unicode có thể in được giữ nguyên. Các escape là `\"`, `\\`, `\n`, `\r`, `\t`, `\0` và `\u{hex}` với chữ số thập lục phân viết thường cho các ký tự điều khiển khác cùng U+2028/U+2029. Tên chứa xuống dòng, tab, dấu ngoặc kép, dấu gạch chéo ngược và tiếng Hàn vẫn được in trong một bản ghi.

```text
"line\ncolumn\tquote\"slash\\한글"
```

Bộ đọc nhận typed IR formats 3 và 4 với semantics version 1 và luôn in format 4. `uninit` yêu cầu format 4. Lệnh format 3 hợp lệ vẫn đọc được, nhưng legacy `undef` là lỗi kiểm tra trong cả hai định dạng và phải tạo lại từ AST. Văn bản format 2 cần chuyển thủ công sang ID tường minh, tên trong ngoặc kép và tham chiếu khối vào. AST JSON có format 2 độc lập.

### Đọc và kiểm tra IR văn bản

`ir::parse_module` đọc typed IR, kiểm tra rồi trả về `Module`. Nó nhận cú pháp hiện tại của bộ in cho vô hướng, luồng điều khiển, bộ nhớ, lời gọi trực tiếp/gián tiếp và biểu thức hằng. ID, tên, kiểu, số nguyên và bit float chính xác, cây biểu thức và kết quả tính, thứ tự khối, khối vào, căn chỉnh, chữ ký và link_name được giữ nguyên. Khoảng trắng và chú thích dòng `//` được chuẩn hóa. Lưu toàn bộ IR ở trên thành `answer.wir` và chạy các lệnh sau trong bản dựng mặc định.

```sh
cargo run --locked -- ir verify answer.wir
cargo run --locked -- ir print answer.wir -o canonical.wir
```

Từ chối phiên bản, trường, lệnh và escape không biết, literal vượt phạm vi, ID trùng, chú thích kiểu mâu thuẫn và đầu vào dư. `ParseError` cung cấp độ lệch byte và dòng/cột Unicode scalar bắt đầu từ 1. Lỗi kiểm tra được gắn với hàm hoặc khai báo toàn cục liên quan khi có thể. `print` cũng kiểm tra và giữ đầu ra cũ khi thất bại. Cú pháp hoặc ngữ nghĩa mới cần đổi format/semantics version tương ứng; phiên bản không biết là lỗi.

### Giới hạn tài nguyên kiểm tra

Mặc định `IrLimits` cho phép đầu vào 8 MiB, 1,000,000 token, 1,000,000 nút duyệt, độ sâu kiểu và biểu thức mỗi loại 128. Gốc có độ sâu 0, mỗi con tăng 1. Có thể giảm độ sâu hoặc tăng tới `MAX_IR_NESTING`, 256. Vượt giới hạn trả về `LimitError`, `VerifyError::ResourceLimit`, `ConstEvalError::ResourceLimit` hoặc `CallError::ResourceLimit`. Số nút đếm việc duyệt tại mỗi ranh giới đầu vào, gồm chú thích kiểu và biểu thức, không phải thời gian trôi qua.

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

Có thể truyền giới hạn vào `verify_module_with_limits`, `ConstExpr::evaluate_with_limits`, `validate_signature_with_limits` và `ModuleBuilder::declare_function_with_limits`. Duyệt kiểu lặp diễn ra trước clone, so sánh và chẩn đoán đệ quy; hằng được tính bằng ngăn xếp công việc. Cây Rust mượn và Drop vẫn thuộc bên gọi. Cây tùy ý chưa kiểm tra vẫn có clone/Drop đệ quy; chữ ký sở hữu bị API khai báo checked từ chối được giải phóng lặp. IR không hợp lệ là lỗi kiểm tra. Trình kiểm tra từ chối legacy `undef`; hãy tạo lại từ AST. Khai báo chưa khởi tạo dùng format 4 `uninit` và kiểm tra lần đọc thực tế, không thay bằng không hoặc giá trị tùy ý. Vi phạm điều kiện thực thi tạo trap xác định có vị trí IR.

### Phiên bản được chỉ định AST JSON

Lưu phần sau dưới dạng `program.json`. Tất cả bốn trường phong bì là bắt buộc. `program` chứa các mảng `declarations`, `globals` và `functions` bắt buộc, có thể trống. Tên hàm, tham số, kiểu trả về, nội dung, `convention` và `linkage` là bắt buộc. `link_name` có thể vắng mặt/null đối với các hàm bên trong và phải là một chuỗi khác trống không có NUL đối với các hàm bên ngoài. Mỗi enum sử dụng tên đơn vị của nó hoặc một đối tượng khóa biến thể duy nhất. Các biến thể đơn vị cũng chấp nhận đối tượng có giá trị null, chẳng hạn như `{"Void":null}`; bộ mã hóa phát ra tên đơn vị `"Void"`. `VarDecl.init` có thể vắng mặt hoặc vô hiệu; các trường bắt buộc khác phải có mặt.

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

Số nguyên `value` là một chuỗi thập phân. Signed Một số được sử dụng sau dấu trừ tùy chọn của một số nguyên và không được phép có dấu cách, dấu cộng, số mũ và dấu phân cách. Phạm vi cho phép được xác định bởi chiều rộng đã khai báo và signedness. `u128::MAX` ở trên được giữ nguyên thông qua JSON và lowering. Các giá trị unsigned âm hoặc các giá trị ngoài phạm vi không phải là wrap và là lỗi. Giá trị Float sử dụng chuỗi bit thập lục phân có chiều rộng chính xác được mô tả trong [Các phép toán số](numeric-operations).

`format_version` là 2 cho định dạng AST này; `semantics_version` là 1. `features` phải là một mảng trống. Các trường, phiên bản, tính năng không xác định, khóa JSON thô trùng lặp (bao gồm cả các khóa tương đương đã thoát) và các giá trị theo sau đều là lỗi, ngay cả với `--no-verify`. Điểm vào thư viện là `ir::lower_ast::interchange::decode`; `encode` phát ra phong bì. `decode` mặc định có giới hạn 8 byte nguồn MiB; `decode_with_limit` chấp nhận giới hạn người gọi. JSON lồng nhau bị giới hạn. Sử dụng bộ giải mã thô này thay vì phân tích cú pháp thành một bản đồ chung có thể loại bỏ các khóa trùng lặp.

[Lược đồ JSON hoàn chỉnh](https://github.com/wavefnd/Whale/blob/master/ir/schema/ast-v2.schema.json) chỉ định hình dạng, trường bắt buộc và biến thể. Việc kiểm tra phạm vi/loại và phát hiện khóa trùng lặp cũng được áp dụng. Tập hợp con hạ thấp vô hướng bao gồm hằng số, biến/hằng, cộng/phụ/mul, so sánh, gán, if/while, trả về và ngắt/tiếp tục. Hỗ trợ tham chiếu chức năng, gọi trực tiếp và gọi gián tiếp; biểu thức tổng hợp không được hỗ trợ. `Opaque` có ​​thể biểu thị trong lược đồ nhưng không được hỗ trợ bằng cách hạ thấp.

Di chuyển yêu cầu bọc Program cũ chưa có envelope và thay số JSON bằng chuỗi số nguyên thập phân hoặc chuỗi bit số thực. Đầu vào không có phiên bản bị từ chối. Format 1 phải chuyển sang format 2, thêm `program.declarations` (mảng rỗng nếu không dùng) và `convention`/`linkage` rõ ràng trong định nghĩa. Các phiên bản độc lập: AST format 2, typed IR format 4 và semantics version 1.

### Đầu vào bị từ chối và khôi phục CLI

Lưu thông tin đầu vào hoàn chỉnh sau đây dưới dạng `invalid.json`.

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

Lệnh thoát với trạng thái khác 0 và không tạo đầu ra mới hoặc ghi đè lên các tệp hiện có. Loại không khớp cũng sẽ không thành công trước khi xuất bản đầu ra. Các tệp nhị phân được tạo không có `socket-cli` thoát với trạng thái 2 và xuất ra lệnh khôi phục có chứa `--features socket-cli`.


## Trình thông dịch số nguyên vô hướng

Bộ thông dịch mặc định thực thi số nguyên/Bool, luồng điều khiển, cấp phát ngăn xếp, lưu và đọc con trỏ dữ liệu, typed GEP, memcpy và memset. Đọc cặp checked bỏ qua padding. Địa chỉ là giá trị tổng hợp 64 bit, không giải tham chiếu bộ nhớ máy chủ. Đối số và kết quả vẫn giới hạn ở số nguyên/Bool hoặc void; float, lời gọi, con trỏ hàm, giá trị aggregate tổng quát, địa chỉ global và thực thi native chưa hỗ trợ. Cấp phát ngăn xếp tồn tại đến khi hàm trả về. Kết thúc vòng đời theo phạm vi, truyền con trỏ qua lời gọi/kết quả, bộ chuyển đổi bộ nhớ ngoài và native shadow metadata cần triển khai tiếp.

`InterpreterOptions::max_steps` mặc định là 1,000,000. Mỗi lệnh được chạy, gồm phi, và mỗi terminator tiêu thụ một bước. Giới hạn 0 dừng trước thao tác đầu tiên; vòng lặp vô hạn trả về `InterpreterError::StepLimit`. `ir_limits` giới hạn xác minh riêng. Phép toán không có nơi dùng vẫn chạy và có thể trap. Overflow checked là kết quả Bool; chỉ trap_if tường minh mới biến nó thành trap.

Lưu mô-đun hoàn chỉnh này thành `swap-loop.wir`. Khối vào được chỉ rõ dù khối thoát được lưu trước. Khi vào khối, mọi đầu vào phi được đọc từ giá trị của khối trước rồi mới ghi đồng thời các kết quả phi. Ba vòng lặp đổi 11 và 22 ba lần và trả về 22.

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

API Rust trả về giá trị và số bước hoặc lỗi có cấu trúc về xác minh, thao tác chưa hỗ trợ, đối số, giới hạn bước hay trap. Chương trình hoàn chỉnh đọc cùng tệp `swap-loop.wir`:

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
