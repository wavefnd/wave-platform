---
translation_set_id: whale-ir-reference
path: whale/ir-reference
locale: ja
group: whale
group_order: 1
order: 6
title: Whale IR リファレンス
summary: 型、識別子、関数の妥当性、評価順序、および交換形式について説明します。
---

## モジュールと識別子

モジュールは、ターゲット情報、グローバル定義、関数で構成されます。値には明示的な型があります。フロントエンドはソース言語の名前・タイプ・オーバーロード・ジェネリックを解決した後、typedIRを生成します。

関数とグローバル変数は異なる内部名前空間を使用します。したがって、関数と変数は同じ名前を持つことができます。内部識別子は外部接続名`link_name`と区別され、外部名はフロントエンドによって指定されます。 Whaleは新しい名前を自動生成して外部競合を解決しません。 [シンボルとリンク](assembler-linker)を参照してください。

各値定義には識別子があります。定義は重複することはできず、型メタデータは定義に指定された型と一致する必要があります。同じ名前の宣言が互いを隠す場合でも、名前だけで定義を識別しません。

## IR構成と読み取り

以下は、`ir`クレートで関数を構成して検証した後、typedIRを出力する完全なRust例です。

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

プリンタは次のIRを出力します。

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

出力を `answer.wir` に保存するとテキストパーサーと CLI で読込み・検証できます。スカラー整数実行は以下で説明します。

上の出力を `answer.wir` に保存すると、テキストパーサと CLI で読み取り・検証できます。IR の実行は未対応です。

## タイプ

|タイプ|意味|
| --- | --- |
| `bool` |論理値falseまたはtrue|
| `i1`, `i8`, `i16`, `i32`, `i64`, `i128` |指定されたビット幅の符号付き整数|
| `u1`, `u8`, `u16`, `u32`, `u64`, `u128` |指定されたビット幅の符号なし整数|
| `f16`, `f32`, `f64` |指定されたビット幅の浮動小数点値|
| `ptr<T>` |Tタイプ値を指すポインタ|
| `fnptr<signature>` |正確なパラメーター/結果の型と呼び出し規約を持つ呼び出し可能なポインター |
| `array<T, N>` |同じタイプの元素N個|
| `struct{T, ...}` |順序付き構造体フィールド|
| `tuple<T, ...>` |順序付きタプル要素|
| `void` |結果なし|

`bool`、`i1`、`u1`は異なるタイプです。 signed`i1`は−1と0、unsigned`u1`は0と1を表します。整数 1 は暗黙的に論理条件にはなりません。条件分岐、Selectの条件、`trap_if`はBoolオペランドを要求します。

記憶サイズは値のビット数だけで決まらず、[ターゲットレイアウト](memory-model)に従います。たとえば、`i1`の値は1ビットですが、メモリでは少なくとも1バイトを占めます。

## 関数と呼び出し

関数には、全パラメータ・結果タイプ、呼び出し規約、linkageを指定します。直接呼び出しと間接呼び出しは、呼び出し先の署名と一致する必要があります。 void呼び出しには結果IDはありません。 nonvoid呼び出しは、O0で結果を使用しなくても結果定義を保持します。

戻り値は関数の結果型と一致する必要があります。 void戻りは値を渡さず、nonvoid戻りは宣言された結果型の値を渡します。

### 宣言、識別子、呼び出し

`Module.declarations` は、各関数の `FunctionId`、名前、完全な署名、リンク、外部リンク名を記録します。定義はこのアイデンティティを指します。パラメータと戻り値の型はその宣言と一致する必要があります。同一の繰り返し宣言は、`declare_function` を通じて同じ ID に解決されます。競合や重複した定義はエラーです。内部宣言にはモジュール内に本体が必要です。外部宣言は、リンクするまで未解決であるか、エクスポートされた本文を持つ可能性があります。内部関数には `link_name` はありません。外部関数には、NUL のない明示的な空ではない名前が必要です。 2 つの異なる関数宣言が同じ外部名を要求することはできません。グローバルと関数は依然として別の内部名前空間を使用します。

前方呼び出しと再帰をサポートするには、`begin_declared_function` で本体を構築する前に宣言を登録します。 `begin_function` は、新しい内部 Whale 関数を作るための便利な API です。チェックされた `declare_function`、`begin_declared_function`、`function_addr`、`null_function`、`call` API は `Result` を返します。拒否された呼び出しでは、命令は追加されず、結果 ID も割り当てられません。

次の完全な Rust プログラムは、外部関数を宣言し、その型指定されたアドレスを取得し、直接呼び出しと間接呼び出しの両方を発行します。

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

`Callee::Direct(FunctionId)` は宣言テーブルを通じて解決されます。 `Callee::Indirect(ValueId)` には `Type::FnPtr(FunctionSignature)` の値が必要です。署名には、すべてのパラメータのタイプ、結果のタイプ、および `CallingConvention::{Whale, SysV64}` が含まれます。これは、コピー、ストレージ、パラメータ、リターン、ファイ、選択を通じて保持されます。データ ポインターと整数値は呼び出すことができません。関数ポインター型を含むキャストは拒否されます。型アノテーションを変更しても、呼び出し可能なシグネチャを変更することはできません。関数ポインタには、このターゲット上に 64 ビットのアドレス ストレージがあります。これ自体はランタイム シャドウ メタデータを実装しません。

アリティ、正確な引数/結果の型、結果 ID の存在、および呼び出し規約が一致する必要があります。暗黙的な変換はありません。間接的な呼び出し先は、その引数と同じように呼び出しを支配する必要があります。 `variadic: true`、void 型のパラメータおよび SysV64 集約パラメータ/結果の署名は拒否されます。 Whale 集約署名は IR で表すことができます。ネイティブ ABI 分類とマシン コールの発行は、どちらの規則でもまだ利用できません。

`null_function(signature)` は、型情報を持つ null 関数ポインターです。この値を呼び出す IR は型としては有効ですが、実行時には呼び出し先へ入る前に必ず trap しなければなりません。null でなくても、無効な対象、有効期間が終了した対象、検査済みのシグネチャと互換性のない対象への呼び出しは trap が必要です。これらの実行時検査と外部コールバックの有効期間管理は、インタプリタ・ネイティブ実行層で今後実装する機能です。検証に成功しても、任意の外部アドレスが安全になるわけではありません。

### AST 呼び出しの形式

これらは、AST フォーマット 2 プログラム内の式フラグメントです。

```json
{"Call":{"callee":{"Direct":"increment"},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

```json
{"Call":{"callee":{"Indirect":{"FunctionRef":"increment"}},"args":[{"Lit":{"Int":{"bits":32,"signed":true,"value":"41"}}}]}}
```

`Direct` と `FunctionRef` は、変数が同じ名前であっても関数名前空間を使用します。 `Indirect` は最初に式を評価し、次に引数を左から右に評価します。 void 呼び出しは、`ExprStmt` として有効ですが、変数初期化子、引数、オペランド、または戻り値としては有効ではありません。呼び出しと関数参照はコンパイル時の数値定数式ではありません。 `NullFunction` は、`params`、`ret`、`convention`、`variadic` フィールドを持つ署名オブジェクトを受け取ります。

[完全な JSON の例](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/ast-v2-calls.json) はコールバックを保存し、外部呼び出しの前にそれを呼び出します。次のように下げます。

```sh
cargo run --locked --features socket-cli -- ir lower ir/tests/fixtures/ast-v2-calls.json
```

その[予想IR](https://github.com/wavefnd/Whale/blob/master/ir/tests/fixtures/calls-v3.wir)は、降下テストでチェックされます。関数 ID とリンク名は IR 境界で表されます。ネイティブ オブジェクトの生成とリンクを通じてそれらを保存することは、依然として別の作業です。

## ブロックと値の利用可能性

各ブロックには一意の識別子と正確に1つのterminatorがあります。分岐先は同じ関数に属する必要があります。エントリブロックは必ず存在し、先行幹線とphiを持つことはできません。ループを作成するときは、エントリブロックから別のループヘッダに分岐します。

実行可能パス内の値の定義は、一般的な使用ポイントを支配する必要があります。つまり、エントリポイントから使用ポイントへのすべてのパスがその定義を通過する必要があります。同じブロックでは、定義は使用の前になければなりません。ブロックの記憶順序は支配関係を決定しない。

エントリーポイントからleftまたはrightに分岐し、joinに参加するとします。 leftでのみ定義した値は、joinで一般値として使用できません。 rightを通るパスには定義がないからです。各先行ブロックの値を入力として受け取るphiで合計する必要があります。

到達不能なブロックもモジュールに保存されます。検証器は、該当ブロックの識別子・タイプ・オペランド・分岐構造を検査し続けます。到達不能なブロックの定義は、到達可能なパスの一般的な使用に値を供給できません。

## Phi コマンド

phiは、ブロック内のすべての一般コマンドの前に配置されます。異なる先行ブロックごとに正確に1つの入力が必要です。入力値はphiと同じタイプでなければならず、対応する先行ブロックの終わりで利用可能でなければなりません。

１つの先行ブロックから複数の幹線が入っていても、入力は１つである。ループphiは、モジュールの保存順序で後で出てくるブロックの値でも繰り返し幹線で計算される値であれば参照できます。欠落・重複・無関係な先行ブロックと誤ったタイプの入力は検証エラーです。

## 評価と選択

WhaleASTは、呼び出し先とサブ式を指定された左から評価します。フロントエンドは、短絡評価を制御フロー分岐として表します。

Selectは、すでに計算された値の1つを選択します。どちらの入力の計算も省略しません。たとえば、安全な値を選択しても、他の入力を計算しながら発生したtrapを避けることはできません。特定のパスでのみ実行する必要がある計算は、条件付きブロック内に配置する必要があります。

## 検証課 trap

不正なIRは検証エラーです。検証器はlegacy `undef`を拒否するのでASTから再生成してください。初期化なしの宣言にはformat 4の`uninit`と実際の読取り検査を使用し、0や任意値には置換しません。実行条件違反はIR位置付きの定義されたtrapです。

現在の `InterpreterTrap` は理由、実行段階数、`ExecutionSite` を返します。位置は関数 ID、ブロック ID、0 始まりの命令インデックス、任意の結果値 ID です。terminator は命令列の直後のインデックスです。CLI は入力ファイルも表示します。typed IR にソース span はまだなく、これはソース行番号ではなく IR 位置です。trap はエラーを返して後続実行を停止し、ライブラリはホストプロセスを終了しません。

この保証は、実績のあるIRおよびトレースメモリに適用されます。外部C・生アドレス・インラインアセンブリは別途契約を有し、その境界外の違反まで常に検出するわけではありません。 [メモリモデル](memory-model)を参照してください。

## 交換形式とテキスト表現

ASTとtypedIRはそれぞれのformatversionと共通semanticsversion読む方は、無バージョン・不明なバージョン・フィールド・機能・重複JSONキーを拒否しなければなりません。コンストラクタは、サポートしていない属性が静かに無視されると仮定してはいけません。

整数はビット幅・signedness・文字列数で渡します。浮動小数点定数は幅と正確なビット列で渡されます。テキストIRのround-tripは、名前・ID・タイプ・定数・順序・属性・メタデータを保存する必要があります。スペースとコメントの配置は保存対象ではありません。

以下の AST JSON 契約と typed IR format 4 の読み取り・検証・往復出力を利用できます。

### 出力される識別子と引用符付きの名前

typed IR format 4 は関数を `@fN`、グローバルを `@gN`、値を `%vN`、ブロックを `%bN` で出力します。関数・グローバル ID はモジュールに、値・ブロック ID は所属する関数に属します。番号に空きがあっても指定された ID を保存します。引用符付きの名前は説明用であり、参照の解決には使いません。関数はブロックの保存順序とは独立に `entry %bN` を明示します。

次の完全なモジュールは Rust IR API で検証して出力しました。両方の分岐ブロックは `"branch"` という名前ですが、ID が定義と phi 入力を区別します。

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

`%v0` はパラメータリストで定義されます。`%b1` と `%b2` は同名でも別のブロックであり、phi は各先行ブロックを ID で指定します。分岐と switch の宛先も同じブロック ID 構文を使います。プリンタは明示した `%v99` の番号を変更しません。

すべての名前・文字列フィールドは二重引用符を使います。対象はターゲット、関数・グローバル・パラメータ・ブロックの名前、外部リンク名、定数宣言名、trap の理由です。表示可能な Unicode はそのまま保存します。エスケープは `\"`、`\\`、`\n`、`\r`、`\t`、`\0` です。その他の制御文字と U+2028/U+2029 は小文字の16進数による `\u{hex}` で表します。改行、タブ、引用符、バックスラッシュ、韓国語を含む名前も1つのレコードとして出力できます。

```text
"line\ncolumn\tquote\"slash\\한글"
```

読取りはtyped IR format 3・4とsemantics version 1に対応し、出力は常にformat 4です。`uninit`はformat 4でのみ使用できます。既存の有効なformat 3命令は読めますが、legacy `undef`は両形式で検証エラーとなりASTから再生成が必要です。format 2のテキストは明示ID・引用名・入口参照への手動移行が必要です。AST JSONは独立したformat 2です。

### テキスト IR の読み取りと検証

`ir::parse_module` は typed IR を読み取り、検証して `Module` を返します。現在のプリンタのスカラー・制御フロー・メモリ・直接/間接呼び出し・定数式の構文を受け付けます。ID、名前、型、正確な整数と float ビット、式の木と評価結果、ブロック順序、入口、整列、署名、link_name を保存します。空白と `//` 行コメントは標準出力に整理されます。上の完全な IR を `answer.wir` に保存し、標準ビルドで次を実行してください。

```sh
cargo run --locked -- ir verify answer.wir
cargo run --locked -- ir print answer.wir -o canonical.wir
```

未知のバージョン・フィールド・命令・エスケープ、範囲外リテラル、重複 ID、矛盾する型表記、末尾の余分な入力は拒否されます。`ParseError` はバイト位置と、1 から始まる行・Unicode scalar 単位の列を返します。検証エラーは可能なら該当する関数またはグローバル宣言に結び付けられます。`print` も検証し、失敗時には既存の出力を保護します。新しい構文や意味には対応する format/semantics version の変更が必要で、未知のバージョンはエラーです。

### 検証の資源制限

`IrLimits` の既定値は入力 8 MiB、トークン 1,000,000 個、走査ノード 1,000,000 個、型と定数式の深さは各 128 です。根の深さは 0、子ごとに 1 増えます。深さは小さくするか `MAX_IR_NESTING` の 256 まで大きくできます。超過は `LimitError`、`VerifyError::ResourceLimit`、`ConstEvalError::ResourceLimit` または `CallError::ResourceLimit` を返します。ノード数は型表記と式を含む入力境界ごとの走査量であり、経過時間ではありません。

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

`verify_module_with_limits`、`ConstExpr::evaluate_with_limits`、`validate_signature_with_limits`、`ModuleBuilder::declare_function_with_limits` にも制限を渡せます。型は再帰的 clone・比較・診断の前に反復走査し、定数式は作業スタックで評価します。借用した Rust の木の所有権と Drop は呼び出し側にあります。任意の未検証の木の clone/Drop は再帰的なままですが、checked 宣言 API が所有する拒否された署名は反復的に解放します。 不正なIRは検証エラーです。検証器はlegacy `undef`を拒否するのでASTから再生成してください。初期化なしの宣言にはformat 4の`uninit`と実際の読取り検査を使用し、0や任意値には置換しません。実行条件違反はIR位置付きの定義されたtrapです。

### バージョンが指定された AST JSON

以下を`program.json`として保存します。 4 つのエンベロープ フィールドはすべて必須です。 `program` には必須の `declarations`、`globals`、`functions` の配列が含まれていますが、空の場合もあります。関数名、パラメータ、戻り値の型、本体、`convention` および `linkage` は必須です。内部関数の場合、`link_name` は存在しないか null である可能性があり、外部関数の場合は、NUL のない空でない文字列である必要があります。各列挙型は、そのユニット名または単一のバリアント キー オブジェクトのいずれかを使用します。ユニットのバリアントは、`{"Void":null}` などの null 値のオブジェクトも受け入れます。エンコーダーはユニット名 `"Void"` を出力します。 `VarDecl.init` は存在しないか null の可能性があります。他の必須フィールドが存在する必要があります。

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

整数`value`は10進文字列です。 Signed 整数の選択的なマイナスの後に数字を書き、空白・プラス・指数・区切り文字は許可しません。許容範囲は、宣言した幅とsignednessで決定します。上記の`u128::MAX`はJSONとloweringを経てそのまま保存されます。負のunsigned値または範囲外の値はwrapなしでエラーです。 Float値には、[数値演算](numeric-operations)で説明されている正確な幅の16進ビット文字列を使用してください。

`format_version` は、この AST 形式では 2 です。 `semantics_version` は 1 です。 `features` は空の配列でなければなりません。不明なフィールド、バージョン、機能、重複した生の JSON キー (エスケープされた等価キーを含む)、および末尾の値は、`--no-verify` であってもエラーとなります。ライブラリのエントリ ポイントは `ir::lower_ast::interchange::decode` です。 `encode` はエンベロープを放射します。 `decode` のデフォルトのソースバイト制限は 8 MiB です。 `decode_with_limit` は発信者制限を受け入れます。 JSON ネストは制限されています。すでに重複キーを破棄している可能性がある汎用マップに解析するのではなく、この生のデコーダーを使用します。

[完全な JSON スキーマ](https://github.com/wavefnd/Whale/blob/master/ir/schema/ast-v2.schema.json) は、形状、必須フィールド、およびバリアントを指定します。範囲/タイプのチェックと重複キーの検出も追加で適用されます。スカラーを下げるサブセットには、リテラル、変数/定数、加算/減算/乗算、比較、代入、if/while、return、break/Continue が含まれます。関数参照、直接呼び出し、間接呼び出しがサポートされています。集計式はサポートされていません。 `Opaque` はスキーマで表現できますが、引き下げることはサポートされていません。

移行には旧 bare Program を envelope で包み、JSON の数値を10進整数文字列または浮動小数点ビット文字列に置き換える必要があります。無バージョン入力は拒否します。format 1 は format 2 に移行し、`program.declarations`（未使用なら空配列）と定義の明示的な `convention`・`linkage` を追加します。AST と typed IR のバージョンは独立しています。AST format は 2、typed IR format は 3、semantics version は 1 です。

### 拒否される入力とCLI回復

次の完全な入力を`invalid.json`として保存します。

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

コマンドはゼロ以外の状態で終了し、新しい出力を作成したり既存のファイルを上書きしたりしません。型の不一致も出力発行前に失敗します。 `socket-cli`なしでビルドしたバイナリは状態2で終了し、`--features socket-cli`を含むリカバリコマンドを出力します。


## スカラー整数インタプリタ

標準インタープリタは整数・Bool、制御フロー、スタック割り当て、データポインタの保存と読取り、typed GEP、memcpy・memsetを実行します。checked ペアの読取りではパディングを除外します。アドレスは合成64ビット値で、ホストメモリを参照しません。引数と戻り値は整数・Boolまたはvoidに限定され、float、呼出し、関数ポインタ、一般aggregate値、グローバルアドレス、native実行は未対応です。割り当ては関数のreturnまで有効です。字句スコープの寿命終了、呼出し・returnでのポインタ転送、外部メモリアダプタ、native shadow metadataは今後の実装です。

`InterpreterOptions::max_steps` の既定値は 1,000,000 です。phi を含む実行命令と terminator はそれぞれ一段階を消費します。0 なら最初の操作の前に停止し、無限分岐ループは `InterpreterError::StepLimit` を返します。`ir_limits` は検証量を別に制限します。未使用の算術も実行され trap し得ます。checked overflow は Bool 結果であり、明示的 trap_if がある場合だけ trap になります。

この完全なモジュールを `swap-loop.wir` に保存してください。終了ブロックを先に保存しても entry は明示されています。ブロックに入るとすべての phi 入力を前のブロックの値から読み、その後で結果を一括して書きます。3 回の反復で 11 と 22 を 3 回交換して 22 を返します。

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

Rust API は値と段階数、または構造化された検証・未対応操作・引数・段階制限・trap エラーを返します。この完全なプログラムは同じ `swap-loop.wir` を読みます。

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
