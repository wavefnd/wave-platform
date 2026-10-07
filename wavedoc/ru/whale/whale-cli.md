---
translation_set_id: whale-cli
path: whale/whale-cli
locale: ru
group: whale
group_order: 1
order: 3
title: Справочник по командам Whale
summary: Описывает Whale assembler, object wrapper, диагностические выходные данные и дополнительные команды IR.
---

## Whale Сборка

В репозитории Whale запустите:

```shell
cargo build --release
```

Исполняемый файл верхнего уровня имеет четыре семейства команд:

```text
whale asm [--amd64 | --aarch64] <input> -o <output>
whale object <input> -o <output>
whale ir <subcommand> [options]
```

## AMD64 assembler

```shell
whale asm --amd64 input.asm -o output.o
```

AMD64 assembler получает путь `.o` в качестве выходных данных, а ELF64 relocatable содержит section, symbol и relocation Создать object.

Включите подробную диагностику с помощью `--debug-whale`.

```shell
whale asm --amd64 input.asm -o output.o \
  --debug-whale --token --ast --bytes --dump-hex --stats
```

Диагностические флаги включают `--token`, `--ast`, `--bytes`, `--dump-hex`, `--dump-bin`, `--dump-json` и `--stats`. `--trace` печатает ход обработки.

## Object wrapper

```shell
whale object input.bin -o output.o
```

Команда `object` помещает необработанные байты в секцию `.text` ELF64 и добавляет глобальный символ `start` со смещением 0. Она оборачивает необработанный машинный код в объектный файл ELF.

## Проверка и печать текстового IR

Стандартная сборка читает и проверяет typed IR format 4. Сохраните полный пример из [справочника IR](ir-reference) как `answer.wir`. `print` проверяет перед канонической печатью и при ошибке сохраняет прежний файл. Он не исполняет IR и не генерирует native код.

```shell
whale ir verify answer.wir
whale ir print answer.wir -o canonical.wir
```

## Дополнительно IR socket

Для AST JSON `ir lower` нужна feature `socket-cli`. Для текстового IR `verify` и `print` она не нужна.

```shell
cargo run -p whale --features socket-cli -- ir lower program.json
cargo run -p whale --features socket-cli -- ir lower program.json -o program.wir
```

`ir lower` считывает JSON из Whale socket schema, преобразует его в Whale IR и проверяет модуль. Текст IR выводится по пути stdout или `-o`. `--target <triple>` заменяет целевую строку, а `--no-verify` пропускает проверку.

Для `ir lower` соберите с `socket-cli`. Производители Socket JSON и Whale должны использовать одинаковую AST schema version.


## Интерпретатор скалярных целых чисел

Обычная сборка также исполняет скалярный целый/Bool IR. `--function @fN` обязателен; повторяйте `--arg` для точных десятичных аргументов, для Bool — true/false. `--max-steps` считает инструкции и terminators, по умолчанию 1,000,000. run не принимает -o или --no-verify. Сохраните полный цикл из [справочника IR](ir-reference) как `swap-loop.wir`:

```shell
whale ir run swap-loop.wir --function @f7 --arg 3 --max-steps 100
```

```text
i32 22
```

## Исполнение с отслеживаемой стековой памятью

Интерпретатор по умолчанию исполняет целые/Bool, поток управления, стековые выделения, хранение и чтение указателей данных, typed GEP, memcpy и memset. Чтение checked-пар исключает padding. Адреса — синтетические 64-битные значения без разыменования памяти хоста. Аргументы и результаты ограничены целыми/Bool или void; float, вызовы, указатели функций, общие агрегатные значения, глобальные адреса и native-исполнение не поддержаны. Стековые выделения живут до возврата. Завершение лексического времени жизни, передача указателей при вызове/возврате, адаптеры внешней памяти и native shadow metadata требуют дальнейшей реализации.

`--max-memory` задаёт бюджет логических байтов выделения, по умолчанию 64 MiB. `InterpreterOptions::memory_limits` также ограничивает число выделений до 16384, метаданные байтов указателя до 262144 фрагментов и обработку байтов/метаданных до 256 Mi единиц. Превышение возвращает `MemoryLimit` с позицией IR отдельно от trap программы. Проверка также отвергает overflow известных размеров хранения целевой платформы до исполнения.

[Модель памяти](memory-model): `tracked-memory.wir`.

```shell
whale ir run tracked-memory.wir --function @f0 --max-memory 20
whale ir run tracked-memory.wir --function @f0 --max-memory 3
```

```text
u32 42
Error: tracked-memory.wir: interpreter memory Bytes limit 3 reached at @f0 %b0 instruction 0 (%v0)
```
