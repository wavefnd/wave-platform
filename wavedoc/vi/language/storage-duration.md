---
translation_set_id: storage-duration
path: language/storage-duration
locale: vi
group: language
group_order: 2
order: 18
title: Thời gian lưu trữ và khả năng thay đổi
summary: Phân biệt giữa phạm vi và khả năng ghi của var, const và static.
---

## Ý nghĩa của từng tuyên bố

|định dạng|Vị trí được phép|gán lại|sử dụng|
| --- | --- | --- | --- |
| `var` |Chức năng/khối|có thể|Các biến cục bộ có thể thay đổi phổ biến|
| `const` |hàng đầu|Không thể|Khai báo hằng số toàn cầu|
| `static` |hàng đầu|có thể|Các khai báo được lưu trữ tĩnh tồn tại trong suốt vòng đời của chương trình|

```wave
const PAGE_SIZE: i32 = 4096;
static request_count: i64 = 0;

fun main() {
    var limit: i32 = 4;
    var current: i32 = 0;
    var retries: i32 = 0;

    current += 1;
    retries += 1;
    println("{} {} {}", limit, current, retries);
}
```

## Quy tắc khai báo cục bộ

```wave
var value: i32 = 1;
value = 2;
```

Một giá trị mới có thể được gán cho biến cục bộ được khai báo là `var`. Các hằng số được sử dụng trong suốt chương trình được khai báo ở cấp cao nhất là `const`.

## Sử dụng const và static trong phạm vi cục bộ

`const` và `static` là các khai báo cấp cao nhất. Phần thân hàm và quá trình khởi tạo `for` sử dụng khai báo cục bộ `var`.

## Tuổi thọ và con trỏ

Bạn có thể lấy địa chỉ của biến cục bộ là `&`, nhưng loại `ptr<T>` không theo dõi thời gian tồn tại thực tế của bộ lưu trữ được con trỏ trỏ tới. Khi chuyển một địa chỉ lưu trữ cục bộ ra khỏi một hàm, cấu trúc chương trình phải trực tiếp đảm bảo rằng địa chỉ đó vẫn hợp lệ.

## Phạm vi học tập và ví dụ

[Thực hành với toàn bộ chương trình](/docs/vi/getting-started/overview) · [Thư viện chuẩn](/docs/vi/stdlib)
