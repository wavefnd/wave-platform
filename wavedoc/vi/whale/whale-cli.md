---
translation_set_id: whale-cli
path: whale/whale-cli
locale: vi
group: whale
group_order: 1
order: 3
title: Tài liệu tham khảo lệnh Whale
summary: Mô tả Whale assembler, object wrapper, đầu ra chẩn đoán và các lệnh IR tùy chọn.
---

## Whale Xây dựng

Trong kho lưu trữ Whale, hãy chạy:

```shell
cargo build --release
```

Một tệp thực thi cấp cao nhất có bốn nhóm lệnh:

```text
whale asm [--amd64 | --aarch64] <input> -o <output>
whale object <input> -o <output>
whale ir <subcommand> [options]
```

## AMD64 assembler

```shell
whale asm --amd64 input.asm -o output.o
```

AMD64 assembler nhận đường dẫn `.o` làm đầu ra và ELF64 relocatable chứa section, symbol và relocation Tạo object.

Bật đầu ra chẩn đoán chi tiết bằng `--debug-whale`.

```shell
whale asm --amd64 input.asm -o output.o \
  --debug-whale --token --ast --bytes --dump-hex --stats
```

Cờ chẩn đoán bao gồm `--token`, `--ast`, `--bytes`, `--dump-hex`, `--dump-bin`, `--dump-json` và `--stats`. `--trace` in tiến trình xử lý.

## Object wrapper

```shell
whale object input.bin -o output.o
```

Lệnh `object` đặt các byte thô vào phần ELF64 `.text` và thêm biểu tượng `start` toàn cầu ở độ lệch 0. Lệnh này bao bọc mã máy thô trong tệp đối tượng ELF.

## Kiểm tra và in IR văn bản

Bản dựng mặc định đọc và kiểm tra typed IR format 4. Lưu ví dụ đầy đủ trong [tham chiếu IR](ir-reference) thành `answer.wir`. `print` kiểm tra trước khi in dạng chuẩn và giữ tệp cũ khi thất bại. Nó không thực thi IR hay sinh mã native.

```shell
whale ir verify answer.wir
whale ir print answer.wir -o canonical.wir
```

## Tùy chọn IR socket

AST JSON `ir lower` cần feature `socket-cli`. `verify` và `print` cho IR văn bản không cần.

```shell
cargo run -p whale --features socket-cli -- ir lower program.json
cargo run -p whale --features socket-cli -- ir lower program.json -o program.wir
```

`ir lower` đọc JSON của Whale socket schema, chuyển đổi nó thành Whale IR và xác minh mô-đun. Văn bản IR được xuất ra đường dẫn stdout hoặc `-o`. `--target <triple>` thay thế chuỗi đích và `--no-verify` bỏ qua xác thực.

Dựng với `socket-cli` để dùng `ir lower`. Bên tạo Socket JSON và Whale phải dùng cùng AST schema version.


## Trình thông dịch số nguyên vô hướng

Bản dựng mặc định cũng chạy IR số nguyên/Bool vô hướng. Bắt buộc có `--function @fN`; lặp `--arg` cho đối số thập phân chính xác hoặc true/false cho Bool. `--max-steps` đếm lệnh và terminator, mặc định 1,000,000. run không nhận -o hay --no-verify. Lưu vòng lặp hoàn chỉnh từ [tham chiếu IR](ir-reference) thành `swap-loop.wir`:

```shell
whale ir run swap-loop.wir --function @f7 --arg 3 --max-steps 100
```

```text
i32 22
```

## Thực thi bộ nhớ ngăn xếp được theo dõi

Bộ thông dịch mặc định thực thi số nguyên/Bool, luồng điều khiển, cấp phát ngăn xếp, lưu và đọc con trỏ dữ liệu, typed GEP, memcpy và memset. Đọc cặp checked bỏ qua padding. Địa chỉ là giá trị tổng hợp 64 bit, không giải tham chiếu bộ nhớ máy chủ. Đối số và kết quả vẫn giới hạn ở số nguyên/Bool hoặc void; float, lời gọi, con trỏ hàm, giá trị aggregate tổng quát, địa chỉ global và thực thi native chưa hỗ trợ. Cấp phát ngăn xếp tồn tại đến khi hàm trả về. Kết thúc vòng đời theo phạm vi, truyền con trỏ qua lời gọi/kết quả, bộ chuyển đổi bộ nhớ ngoài và native shadow metadata cần triển khai tiếp.

`--max-memory` đặt ngân sách byte cấp phát logic, mặc định 64 MiB. `InterpreterOptions::memory_limits` còn giới hạn 16384 cấp phát, 262144 mảnh metadata byte con trỏ và 256 Mi đơn vị công việc byte/metadata. Vượt giới hạn trả về `MemoryLimit` có vị trí IR, tách biệt với trap của chương trình. Kiểm tra cũng từ chối overflow kích thước lưu trữ đã biết của đích đầu ra trước khi thực thi.

[Mô hình bộ nhớ](memory-model): `tracked-memory.wir`.

```shell
whale ir run tracked-memory.wir --function @f0 --max-memory 20
whale ir run tracked-memory.wir --function @f0 --max-memory 3
```

```text
u32 42
Error: tracked-memory.wir: interpreter memory Bytes limit 3 reached at @f0 %b0 instruction 0 (%v0)
```
