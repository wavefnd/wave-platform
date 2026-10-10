---
translation_set_id: language-variants
path: language/variants
locale: vi
group: language
group_order: 2
order: 17
title: Các biến thể và khớp mẫu
summary: Lưu trữ dữ liệu dành riêng cho từng trường hợp cùng nhau và tách biệt an toàn khỏi match.
---

## Phân biệt với enum

enum đại diện cho một giá trị số nguyên được đặt tên và variant chứa payload, giá trị này khác nhau trong mỗi trường hợp. Result bên dưới lưu trữ giá trị i32 trong trường hợp thành công và số lỗi i32 trong trường hợp thất bại. Ngay cả khi lỗi và kết quả có cùng loại số nguyên, ý nghĩa của chúng có thể được phân biệt theo tên.

## Khai báo, lập, kiểm tra

Lưu nó vào `main.wave` và chạy nó.

<!-- wave-example: variant-api -->
```wave playground
variant Result {
    Value(i32), Error(i32)
}

fun calculate(valid: bool) -> Result {
    if (!valid) {
        return Result::Error(1);
    }

    return Result::Value(42);
}

fun main() {
    var result: Result = calculate(true);
    match (result) {
        Result::Value(value) => {
            println("value={}", value);
        }
        Result::Error(code) => {
            println("error={}", code);
        }
    }
}
```

Kết quả thực hiện:

```text
value=42
```

Giả sử `Result::Value(42)` có payload. Chỉ sử dụng payload dưới tên value trong mẫu tương ứng của `match`. Nó không bắt buộc phải đọc payload trong các trường hợp khác. arm Văn bản được viết theo khối. Xử lý tất cả các trường hợp hoặc phần còn lại bằng `_`. Tốt hơn là nên chia nhỏ từng trường hợp một cách rõ ràng để biết liệu có cần xử lý khi thêm một trường hợp mới hay không.

## Thuốc generic và tuổi thọ

Các tham số loại có thể được sử dụng như trong `variant Optional<T> { Some(T), None }`. Chỉ định loại cụ thể cho các biến cục bộ, chẳng hạn như `Optional<i32>`. Biến thể chứa con trỏ không tự động quản lý quyền sở hữu bộ nhớ. Việc sao chép giá trị không trùng lặp với phân bổ trỏ tới.

Đừng cho rằng cách biểu diễn bộ nhớ của variant giống hệt với bất kỳ C union tùy ý nào. Đối với dữ liệu được gửi ra bên ngoài ABI, một biểu thức riêng biệt được xác định và phân phối dưới dạng loại FFI được phép.

## luyện tập

Nếu bạn đổi nó thành calculate(false), bạn sẽ nhận được `error=1`. Hãy thử thêm trường hợp Empty không có payload, sau đó xử lý trường hợp đó trong match.

[Cấu trúc và enum](/docs/vi/language/structures-enums-and-aliases) · [Lớp xử lý lỗi](/docs/vi/language/errors)
