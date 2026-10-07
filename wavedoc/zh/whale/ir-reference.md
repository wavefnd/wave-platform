---
translation_set_id: whale-ir-reference
path: whale/ir-reference
locale: zh
group: whale
group_order: 1
order: 6
title: Whale IR 参考
summary: 描述类型、标识符、函数有效性、评估顺序和交换格式。
---

## 模块和标识符

模块由目标信息、全局定义和函数组成。值具有显式类型。前端解析源语言的名称、类型、重载和泛型并生成typedIR。

函数和全局变量使用不同的内部命名空间。因此，函数和变量可以具有相同的名称。内部标识符与外部连接名称`link_name`不同，外部名称由前端指定。 Whale 不会通过自动生成新名称来解决外部冲突。请参阅[符号和链接](assembler-linker)。

每个值定义都有一个标识符。定义不能重复，并且类型元数据必须与定义中指定的类型匹配。名称本身并不能识别定义，即使具有相同名称的声明彼此混淆。

## IR 配置与读取

下面是一个完整的 Rust 示例，它使用 `ir` 箱构建并验证函数，然后输出 typed IR。

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

打印机输出IR：

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

将输出保存为 `answer.wir`，即可用文本解析器或 CLI 读取并验证。下面介绍标量整数执行。

将上面的输出保存为 `answer.wir`，即可用文本解析器和 CLI 读取并验证。尚不支持执行 IR。

## 类型

|类型|意义|
| --- | --- |
| `bool` |逻辑值 false 或 true|
| `i1`, `i8`, `i16`, `i32`, `i64`, `i128` |指定位宽的有符号整数|
| `u1`, `u8`, `u16`, `u32`, `u64`, `u128` |指定位宽的无符号整数|
| `f16`, `f32`, `f64` |具有指定位宽的浮点值|
| `ptr<T>` |T 指向类型值的指针|
| `fnptr<signature>` | 具有精确参数/结果类型和调用约定的可调用指针 |
| `array<T, N>` |N 相同类型的元素|
| `struct{T, ...}` |有序结构域|
| `tuple<T, ...>` |有序元组元素|
| `void` |没有结果|

`bool`、`i1` 和 `u1` 是不同的类型。 signed `i1` 表示-1 和0，unsigned `u1` 表示0 和1。整数1 不是隐式逻辑条件。条件分支，Select、`trap_if`的条件需要操作数Bool。

存储大小不仅仅由值中的位数决定，并且遵循[目标布局](memory-model)。例如，`i1`的值为1位，但在内存中至少占用1个字节。

## 函数和调用

该函数指定所有参数、结果类型、调用约定和linkage。直接和间接调用必须与调用者的签名匹配。调用void没有结果ID。即使 O0 不使用结果，对 nonvoid 的调用也会保留结果定义。

返回值必须与函数的结果类型匹配。 void 返回不携带值，nonvoid 返回携带声明的结果类型的值。

### 声明、身份和呼吁

`Module.declarations`记录了每个函数的`FunctionId`、名称、完整签名、链接和外部链接名称。定义指的是这个身份；参数和返回类型必须与其声明相匹配。相同的重复声明通过`declare_function`解析为相同的ID；冲突和重复定义都是错误。内部声明需要模块中的主体。外部声明可能在链接之前无法解析，或者具有导出的主体。内部函数没有`link_name`；外部函数需要一个不带NUL的显式非空名称。两个不同的函数声明不能​​声明相同的外部名称。全局变量和函数仍然使用单独的内部命名空间。

在构造主体之前使用`begin_declared_function`注册声明以支持前向调用和递归。 `begin_function` 仍然为新的内部 Whale 功能提供便利。已检查的`declare_function`、`begin_declared_function`、`function_addr`、`null_function`、`call`接口返回`Result`；被拒绝的调用不会附加指令或分配其结果 ID。

以下完整的 Rust 程序声明了一个外部函数，获取其类型化地址，并发出直接和间接调用：

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

`Callee::Direct(FunctionId)`通过声明表解析； `Callee::Indirect(ValueId)` 需要 `Type::FnPtr(FunctionSignature)` 值。签名包括所有参数类型、结果类型和`CallingConvention::{Whale, SysV64}`。它通过复制、存储、参数、返回、phi 和选择来保留。数据指针和整数值不可调用。涉及函数指针类型的强制转换被拒绝；更改类型注释不能更改可调用签名。函数指针在此目标上有 64 位地址存储；这本身并不实现运行时影子元数据。

数量、确切的参数/结果类型、结果 ID 存在和调用约定必须匹配。没有隐式转换。间接被调用者必须像其参数一样主导该调用。 `variadic: true`、void 类型的参数和 SysV64聚合参数/结果签名被拒绝。 Whale聚合签名可以用IR表示；本机 ABI 分类和机器调用发出尚不适用于这两种约定。

`null_function(signature)` 表示类型化的空函数指针。调用它是类型良好的IR，在进入被调用者之前需要一个运行时陷阱。无效、过期或与检查签名不兼容的非空目标也必须捕获。这些运行时检查和外部回调生命周期管理等待解释器/本机执行层；验证者成功并不意味着任意外部地址都是安全的。

### AST 调用形式

这些是 AST 格式 2 程序内的表达式片段：

```json
{"Call":{"callee":{"Direct":"increment"},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

```json
{"Call":{"callee":{"Indirect":{"FunctionRef":"increment"}},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

即使变量具有相同的名称，`Direct` 和 `FunctionRef` 也使用函数命名空间。 `Indirect` 首先计算其表达式，然后从左到右计算参数。 void 调用作为 `ExprStmt` 有效，但不能作为变量初始值设定项、参数、操作数或返回值。调用和函数引用不是编译时数值常量表达式。 `NullFunction` 采用带有 `params`、`ret`、`convention` 和 `variadic` 字段的签名对象。

[完整的JSON示例](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/ast-v2-calls.json)存储回调并在外部调用之前调用它。降低它：

```sh
cargo run --locked --features socket-cli -- ir lower ir/tests/fixtures/ast-v2-calls.json
```

其[预期IR](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/calls-v3.wir)在降低测试中被检查。函数标识和链接名称在IR边界处表示；通过本机对象生成和链接来保留它们仍然是单独的工作。

## 块和值的可用性

每个块都有一个唯一的标识符和恰好一个terminator。分支目标必须属于同一函数。入口块必须存在并且不能有前沿和phi。创建循环时，从入口块分支到单独的循环头。

可执行路径中值的定义应控制其典型使用点。这意味着从入口点到使用点的所有路径都必须经过该定义。在同一块中，定义必须位于使用之前。块的存储顺序并不决定主导地位。

假设入口点分支到left或right，然后在join加入。仅在left中定义的值不能用作join中的通用值。这是因为通过right的路径未定义。前面每个块的值必须组合成phi，作为输入接收。

无法到达的块也保留在模块中。验证者不断检查块的标识符、类型、操作数和分支结构。不可达块的定义不能为可达路径的正常使用提供价值。

## Phi命令

phi 放置在块中所有常规命令之前。每个不同的前面的块只需要一个输入。输入值的类型必须为phi，并且必须在相应的前一个块的末尾可用。

即使前一个块中有多个边，也只有一个输入。循环phi可以引用模块存储顺序中较晚出现的块的值，只要它是在重复边缘上计算的值即可。缺失、重复或不相关的前面的块以及错误键入的输入都是验证错误。

## 评估与选择

Whale AST 从指定的左侧开始计算调用目标和子表达式。前端将短路评估表示为控制流分支。

Select 选择已计算的值之一。它不会忽略任一输入的计算。例如，即使选择安全值，也无法避免在计算其他输入时遇到trap。仅需要在特定路径上执行的计算应放置在条件块内。

## 验证部trap

错误IR是验证错误。验证器拒绝legacy `undef`，应从AST重新生成。未初始化声明使用format 4的`uninit`和实际读取检查，不替换为零或任意值。运行条件违反产生带IR位置的已定义trap。

当前 `InterpreterTrap` 报告原因、已执行步数与 `ExecutionSite`：函数 ID、块 ID、从 0 开始的指令索引及可选结果值 ID。terminator 的索引紧接指令序列。CLI 诊断还显示输入文件。typed IR 尚无源码 span，因此这是 IR 位置而非源码行号。trap 返回错误并停止后续执行；库不会终止宿主进程。

此保证适用于经过验证的IR 和跟踪内存。外部C·原始地址·内联汇编具有单独的契约，并且并不总是检测其边界之外的违规行为。请参阅[内存模型](memory-model)。

## 交换格式和文本表示

AST 和 typed IR 使用各自的 format version 和通用 semantics version。读者应该拒绝未版本化/未知版本/字段/功能/重复的JSON密钥。构造函数不应假设不支持的属性将被默默忽略。

整数作为位宽·signedness·字符串数字传递。浮点常量作为宽度和精确位字符串传递。 IR 中的文本round-trip 必须保留名称·ID·类型·常量·序列·属性·元数据。空格和评论位置不受保留。

可以使用下面的 AST JSON 契约以及 typed IR format 4 的读取、验证和往返打印。

### 输出标识符与带引号的名称

typed IR format 4 用 `@fN` 表示函数、`@gN` 表示全局变量、`%vN` 表示值、`%bN` 表示基本块。函数和全局 ID 属于模块；值和基本块 ID 属于所在函数。即使编号不连续，也保留提供的 ID。带引号的名称只是说明，不用于解析引用。函数明确记录 `entry %bN`，不依赖基本块的存储顺序。

以下完整模块通过 Rust IR API 验证并输出。两个分支块都叫 `"branch"`，但 ID 可以区分定义和 phi 输入。

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

`%v0` 在参数列表中定义。`%b1` 和 `%b2` 即使同名也不同，phi 通过 ID 指定各前驱块。分支和 switch 目标使用同样的基本块 ID 语法。打印器不会重新编号显式提供的 `%v99`。

所有名称和字符串字段都使用双引号，包括目标、函数、全局变量、参数和基本块名称、外部链接名称、常量声明名称及 trap 原因。可打印的 Unicode 字符保持原样。转义序列为 `\"`、`\\`、`\n`、`\r`、`\t`、`\0`；其他控制字符以及 U+2028/U+2029 使用小写十六进制的 `\u{hex}`。包含换行、制表符、引号、反斜杠和韩文的名称也会输出为一条记录。

```text
"line\ncolumn\tquote\"slash\\한글"
```

读取器接受typed IR format 3和4、semantics version 1，并始终输出format 4。`uninit`只允许在format 4中使用。原有有效format 3指令仍可读取，但legacy `undef`在两种格式中都是验证错误，必须从AST重新生成。format 2文本需要手动迁移为显式ID、带引号名称和入口引用。AST JSON使用独立的format 2。

### 读取并验证文本 IR

`ir::parse_module` 读取 typed IR、验证后返回 `Module`。它接受当前打印器的标量、控制流、内存、直接/间接调用和常量表达式语法。保留 ID、名称、类型、精确整数和 float 位、表达式树与求值结果、基本块顺序、入口、对齐、签名和 link_name。空白与 `//` 行注释会整理为规范输出。将上面的完整 IR 保存为 `answer.wir`，在默认构建中运行以下命令。

```sh
cargo run --locked -- ir verify answer.wir
cargo run --locked -- ir print answer.wir -o canonical.wir
```

未知版本、字段、指令、转义、越界字面量、重复 ID、冲突的类型标注和尾随输入都会被拒绝。`ParseError` 提供字节偏移以及从 1 开始的行和 Unicode scalar 列。验证错误尽可能定位到所属函数或全局声明。`print` 也验证输入，失败时保留已有输出。新增语法或语义需要更新相应的 format/semantics version；未知版本是错误。

### 验证资源限制

`IrLimits` 默认限制输入为 8 MiB、令牌为 1,000,000 个、遍历节点为 1,000,000 个，类型和表达式深度各为 128。根深度为 0，每个子节点增加 1。可降低深度或提高至 `MAX_IR_NESTING` 的 256。超限返回 `LimitError`、`VerifyError::ResourceLimit`、`ConstEvalError::ResourceLimit` 或 `CallError::ResourceLimit`。节点数量统计每个输入边界的遍历工作，包括类型标注和表达式节点，不是耗时限制。

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

还可向 `verify_module_with_limits`、`ConstExpr::evaluate_with_limits`、`validate_signature_with_limits` 和 `ModuleBuilder::declare_function_with_limits` 传入限制。在递归 clone、比较和诊断之前迭代遍历类型；常量表达式用工作栈求值。借用的 Rust 树及其 Drop 仍由调用者管理。任意未验证树的 clone/Drop 仍是递归的；checked 声明 API 拒绝的自有签名会迭代释放。 错误IR是验证错误。验证器拒绝legacy `undef`，应从AST重新生成。未初始化声明使用format 4的`uninit`和实际读取检查，不替换为零或任意值。运行条件违反产生带IR位置的已定义trap。

### 指定版本 AST JSON

将以下内容另存为`program.json`。所有四个信封字段都是必需的。 `program`包含所需的`declarations`、`globals`和`functions`数组，这些数组可能为空。函数名称、参数、返回类型、函数体、`convention` 和 `linkage` 为必填项。对于内部函数，`link_name` 可以不存在/为空，对于外部函数，NUL 必须是非空字符串。每个枚举使用其单元名称或单个变量键对象。单位变体还接受空值对象，例如`{"Void":null}`；编码器发出单元名称`"Void"`。 `VarDecl.init` 可能不存在或为空；必须存在其他必填字段。

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

整数`value`是一个十进制字符串。 Signed 整数的可选减号后使用数字，不允许使用空格、加号、指数和分隔符。允许的范围由声明的宽度和signedness决定。上述`u128::MAX`通过JSON和lowering原样保留。负unsigned值或超出范围的值不是wrap，而是错误。 Float 值使用[数值运算](numeric-operations) 中描述的精确宽度的十六进制位串。

对于此 AST 格式，`format_version` 为 2； `semantics_version` 为 1。`features` 必须是空数组。未知字段、版本、功能、重复的原始 JSON 键（包括转义的等效键）和尾​​随值都是错误，即使使用 `--no-verify` 也是如此。库入口点是`ir::lower_ast::interchange::decode`； `encode` 发出信封。 `decode` 默认为 8 MiB 源字节限制； `decode_with_limit` 接受呼叫者限制。 JSON 嵌套是有界的。使用此原始解码器，而不是解析为可能已经丢弃重复键的通用映射。

[完整的JSON架构](https://github.com/wavefnd/Whale/blob/master/ir/schema/ast-v2.schema.json)指定形状、必填字段和变体。范围/类型检查和重复键检测还适用。标量降低子集包括文字、变量/常量、add/sub/mul、比较、赋值、if/while、返回和中断/继续。支持函数引用、直接调用和间接调用；不支持聚合表达式。 `Opaque` 在模式中可以表示，但不支持降低。

迁移时，需要将旧的裸 Program 包装为 envelope，并将 JSON 数字替换为十进制整数字符串或浮点位字符串。无版本输入会被拒绝。format 1 必须迁移到 format 2，添加 `program.declarations`（未使用时为空数组）以及定义中的显式 `convention` 和 `linkage`。AST 和 typed IR 的版本独立：AST format 为 2，typed IR format 为 3，semantics version 为 1。

### 拒绝输入和 CLI 恢复

将以下完整输入保存为`invalid.json`。

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

该命令以非零状态退出，并且不会创建新的输出或覆盖现有文件。在发布输出之前，类型不匹配也会失败。没有 `socket-cli` 构建的二进制文件以状态 2 退出并输出包含 `--features socket-cli` 的恢复命令。


## 标量整数解释器

默认解释器执行整数、Bool、控制流、栈分配、数据指针存取、typed GEP、memcpy和memset。checked二元组读取不检查填充字节。地址是合成的64位值，不会解引用主机内存。参数和返回仍限于整数、Bool或void；float、调用、函数指针、通用聚合值、全局地址及native执行不受支持。栈分配保持有效直到函数返回。词法生命周期结束、调用和返回中的指针传递、外部内存适配器及native shadow metadata仍需实现。

`InterpreterOptions::max_steps` 默认为 1,000,000。每条执行指令（包括 phi）与每个 terminator 各消耗一步。限制为 0 时在首个操作前停止；无限分支循环返回 `InterpreterError::StepLimit`。`ir_limits` 单独限制验证工作量。未使用的算术也会执行并可能 trap。checked overflow 是 Bool 结果，只有显式 trap_if 才使其 trap。

将此完整模块保存为 `swap-loop.wir`。即使退出块存储在前面，入口块仍被显式指定。进入块时，先从前一个块的值读取全部 phi 输入，再一次性写入结果。三次循环交换 11 和 22 三次，返回 22。

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

Rust API 返回值与步数，或结构化的验证、不支持操作、参数、步数限制或 trap 错误。此完整程序读取同一个 `swap-loop.wir` 文件。

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
