---
translation_set_id: program-structure
path: language/program-structure
locale: vi
group: language
group_order: 2
order: 1
title: 1. Từ file nguồn đến chương trình đang chạy
summary: Tìm hiểu về các tệp nguồn, chức năng, đầu ra, kiểm tra và thực thi.
---

## Trước khi bắt đầu chương này

Chuẩn bị trình biên dịch và thư viện chuẩn theo [Hướng dẫn cài đặt](/docs/vi/getting-started/install). Nếu bạn có thể chạy `wavec --version` trong thiết bị đầu cuối của mình, bạn có thể bắt đầu. Bất kỳ trình soạn thảo nào có thể lưu tệp văn bản thuần túy đều có thể thực hiện được.

Trong chương này, chúng ta bắt đầu bằng việc tạo một dòng đầu ra và tìm hiểu mối quan hệ giữa các tệp nguồn, hàm, biên dịch, thực thi và mã thoát. Mục tiêu không chỉ là sao chép hướng dẫn mà còn có thể giải thích những gì xảy ra ở giai đoạn nào.

## Tạo một thư mục làm việc

Việc sử dụng một thư mục riêng cho mỗi chương trình sẽ giúp bạn dễ dàng tìm thấy nguồn và các tệp thực thi được tạo ra hơn. Tạo và điều hướng đến một thư mục trong terminal.

```shell
mkdir wave-study
cd wave-study
```

Tạo `main.wave` trong thư mục này bằng trình chỉnh sửa của bạn. Kiểm tra phần mở rộng để đảm bảo tên tệp không phải là `main.wave.txt`. Dưới đây là toàn bộ tệp, không chỉ một đoạn bên trong hàm.

<!-- wave-example: book-hello -->
```wave playground
fun main() {
    println("Hello, Wave!");
}
```

Kết quả thực hiện:

```text
Hello, Wave!
```

Chạy nó bằng lệnh sau:

```shell
wavec run main.wave
```

Không kết hợp các lệnh đầu cuối với mã Wave. Nhập `wavec run` vào thiết bị đầu cuối và viết `fun main` vào tệp nguồn. Không cần thiết phải dán đầu ra `Hello, Wave!` của chương trình vào nguồn.

## đọc từng dòng

`fun` là từ khóa khai báo hàm. Hàm là một tập hợp các thao tác được đặt tên và `main` là điểm vào của tệp thực thi. Bạn sẽ tự xác định các chức năng khác sau.

`()` sau `main` là nơi bạn ghi thông số. main trong chương trình này không có tham số và do đó trống. Viết ra các thao tác của hàm giữa `{` và `}`. Việc thụt lề làm cho các khối dễ đọc hơn đối với con người và bản thân các ranh giới của khối được biểu thị bằng dấu ngoặc nhọn.

`println("Hello, Wave!");` là câu xuất ra một chuỗi. Dấu ngoặc kép cho biết phần đầu và phần cuối của một chuỗi ký tự. Bản thân dấu ngoặc kép không được bao gồm trong đầu ra. Dấu chấm phẩy cho biết sự kết thúc của tuyên bố này.

## Các câu lệnh được thực hiện theo thứ tự chúng được viết

Hãy thử thay đổi nó để in ba lần. Không kết hợp với chương trình trước mà thay thế nội dung của main.wave bằng toàn bộ chương trình bên dưới.

<!-- wave-example: book-sequence -->
```wave playground
fun main() {
    println("start");
    println("working");
    println("done");
}
```

Kết quả thực hiện:

```text
start
working
done
```

Sau khi kết thúc câu đầu tiên, hãy chuyển sang câu tiếp theo. Không có nhiệm vụ nào chạy đồng thời ở đây. Nếu muốn thay đổi thứ tự đầu ra chỉ cần thay đổi thứ tự các câu. Bạn có thể kiểm soát trình tự này bằng cách học các câu lệnh điều kiện và vòng lặp sau này.

## print và println

`println` thêm dấu ngắt dòng ở cuối. `print` không tự động chuyển dòng. Sự khác biệt thể hiện rõ ở việc nối các mảnh nhỏ lại để tạo thành một đường nét duy nhất.

<!-- wave-example: book-print-lines -->
```wave playground
fun main() {
    print("Wave");
    print(" ");
    println("study");
    println("second line");
}
```

Kết quả thực hiện:

```text
Wave study
second line
```

In ba lần không phải lúc nào cũng có ba dòng. Phân biệt giữa số lượng cuộc gọi đầu ra và số lượng dòng. Khi viết ngắt dòng trực tiếp trong chuỗi, hãy sử dụng `\n` escape. Chuỗi escape và ngắt dòng trong nguồn được trình bày chi tiết trong [tấm dây](/docs/vi/language/strings).

## Chèn giá trị vào chuỗi

Để đưa kết quả tính toán thành một chuỗi, truyền giá trị tương ứng với `{}`.

<!-- wave-example: book-format-first -->
```wave playground
fun main() {
    println("{} + {} = {}", 2, 3, 2 + 3);
}
```

Kết quả thực hiện:

```text
2 + 3 = 5
```

`{}` đầu tiên chứa 2, chuỗi thứ hai chứa 3 và chuỗi thứ ba chứa 5. Chuỗi định dạng và giá trị được phân tách bằng dấu phẩy. Nếu bạn thay đổi số lượng giá trị thì số placeholder cũng phải khớp. Đây là cú pháp đầu ra khác với thao tác cộng chuỗi.

## Phân chia kiểm tra, xây dựng và thực hiện

run được sử dụng cho đến nay sẽ tiến hành xây dựng và thực thi một cách tuần tự. Nếu bạn muốn biết bước nào không thành công, bạn có thể chia nhỏ nó như thế này:

```shell
wavec check main.wave
wavec build main.wave -o hello
```

check kiểm tra ngữ pháp, loại, v.v., nhưng không kiểm tra tất cả đầu vào của chương trình. Ví dụ: chỉ kiểm tra nguồn không thể xác định liệu tệp có tồn tại trong thời gian chạy hay không. build tạo một tệp thực thi. Trong Linux/macOS, thực hiện như sau.

```shell
./hello
```

Trong Windows, đặt tên tệp đầu ra là `hello.exe` và chạy trong PowerShell là `.\hello.exe`. Nếu bạn sửa đổi nguồn sau khi tạo một tệp thực thi, bạn phải xây dựng lại nó để những thay đổi được phản ánh.

## Mã thoát cũng là kết quả

Con người đọc câu lệnh đầu ra, nhưng shell hoặc chương trình khác có thể xác định thành công bằng mã thoát. Phần sau đây chỉ định rằng main trả về i32.

<!-- wave-example: book-exit-success -->
```wave playground
fun main() -> i32 {
    println("completed");
    return 0;
}
```

Kết quả thực hiện:

```text
completed
```

0 là quy ước để biểu thị việc tắt máy bình thường. Khi báo hiệu lỗi trực tiếp, hãy trả về mã khác 0. Trong shell Linux/macOS, mã được kiểm tra là `echo $?` ngay sau khi thực thi và trong PowerShell, mã được kiểm tra là `$LASTEXITCODE`. Nếu bạn chạy các lệnh khác trong lúc chờ đợi, nội dung bạn kiểm tra có thể thay đổi.

Việc thực thi `return` sẽ kết thúc hàm. main không khai báo bất kỳ tham số nào và ngay cả khi chúng có giá trị mặc định cũng không được phép. Bỏ qua kiểu trả về main hoặc sử dụng i32.

## Cách đọc lỗi đầu tiên

Mã sau đây cố tình không chính xác: Không giống như ví dụ đang chạy, mã này sẽ bị lỗi ở check.

```wave
fun main() {
    println("hello")
}
```

Cuối câu không có dấu chấm phẩy. Nhìn vào dòng chẩn đoán hiển thị và câu ngay trước nó. Vị trí được trình biên dịch đánh dấu có thể không phải là nơi bắt nguồn lỗi mà là nơi cấu trúc không chính xác không thể diễn giải được nữa.

Đầu tiên, sửa lỗi đầu tiên và sau đó kiểm tra lại. Nếu dấu ngoặc đơn hoặc dấu ngoặc kép trước đó không được đóng, nhiều lỗi khác nhau có thể xảy ra ngay cả trong đoạn mã thông thường tiếp theo. Nếu cố gắng sửa tất cả các dòng cùng lúc thì rất dễ bỏ sót nguyên nhân ban đầu.

## vấn đề thực hành

1. Viết chương trình in ra 3 dòng giới thiệu bản thân.
2. Chuyển 12 và 8 làm đối số định dạng để in `12 * 8 = 96`.
3. Viết nó để in thông báo thành công và trả về mã thoát là 0.
4. Dự đoán trước khi thực hiện số dòng đầu ra sẽ là bao nhiêu nếu bạn chỉ sử dụng print ba lần.

### Lời giải: Kết quả quá trình tính toán

<!-- wave-example: book-first-solution -->
```wave playground
fun main() -> i32 {
    println("Learning Wave");
    println("My first program");
    println("Ready to calculate");
    println("{} * {} = {}", 12, 8, 12 * 8);
    return 0;
}
```

Kết quả thực hiện:

```text
Learning Wave
My first program
Ready to calculate
12 * 8 = 96
```

Thay vì viết kết quả tính toán trực tiếp vào chuỗi dưới dạng `96`, tôi chuyển nó dưới dạng biểu thức. Đây là bước đầu tiên để đảm bảo kết quả tính toán và hiển thị không thay đổi ngay cả khi giá trị đầu vào thay đổi. Trong chương tiếp theo, chúng ta sẽ đặt tên biến để tránh viết cùng một giá trị nhiều lần.


## Mục cấp cao nhất trong tệp nguồn

Nguồn Wave có thể bao gồm các mục cấp cao nhất sau:

- `import(...)`
- `const`, `static`
- `type`
- `struct`, `enum`, `proto`
- `extern(...)`, `export(...)`
- `fun`
- `#[target(...)]` tình trạng trước mục được hỗ trợ

Tiền tố khai báo có thể nhập bằng `pub`. Đặt khai báo `var` cục bộ bên trong một hàm hoặc khối.

## Chương trình độc lập

Các mục tiêu không có kernel, mã khởi động hoặc thời gian chạy có thể sử dụng tùy chọn xây dựng độc lập.

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files
```

`--freestanding` liên kết đến một kế hoạch xây dựng để tắt các phần phụ thuộc thư viện mặc định và `--entry` đặt biểu tượng mục nhập trình liên kết. Để tạo đầu ra có khả năng khởi động thực tế, bạn cần thiết kế kiến ​​trúc đích, tập lệnh liên kết và thậm chí cả định dạng đối tượng.

## Điểm vào cố tình thất bại

Bạn không thể có tham số trong main ngay cả khi nó có giá trị mặc định. Đã xảy ra lỗi khi sao chép tệp sau check.

<!-- wave-example: reject-main-argument -->
```wave
fun main(value: i32 = 0) -> i32 {
    return value;
}
```
