---
translation_set_id: overview
path: getting-started/overview
locale: vi
group: getting-started
group_order: 1
order: 1
title: Tổng quan về ngôn ngữ Wave
summary: Giới thiệu ngắn về cấu trúc chương trình, biến và trình biên dịch Wave.
---

## Về Wave

Wave là ngôn ngữ lập trình hệ thống có kiểu tĩnh, hướng đến sinh mã máy và khả năng điều khiển cấp thấp một cách tường minh.

## Chương trình đầu tiên

```wave
fun main() {
    println("Hello, Wave!");
}
```

Lưu tệp thành `main.wave` và chạy:

```shell
wavec run main.wave
```

## Khai báo biến

Biến cục bộ được khai báo bằng `var`.

```wave
var count: i32 = 1;
count += 1;
```

Hai dạng cũ `let` và `let mut` đã bị loại bỏ và hiện là lỗi cú pháp. Dùng `const` và `static` cho khai báo ở cấp cao nhất.

## Trạng thái bản dịch

Các trang chưa được dịch sang tiếng Việt sẽ được ghi rõ và hiển thị bằng tiếng Anh.
