---
translation_set_id: program-structure
path: language/program-structure
locale: ja
group: language
group_order: 2
order: 1
title: 1. ソースファイルから実行プログラムまで
summary: ソースファイル、関数、出力、検査、実行を学びます。
---

## この章を始める前に

[設置ガイド](/docs/ja/getting-started/install)に従ってコンパイラと標準ライブラリを準備します。端末で`wavec --version`を実行できる場合は起動できます。エディタはプレーンテキストファイルを保存できるものです。

この章では、出力1行を作成する際に出発し、ソースファイル、関数、コンパイル、実行、終了コードの関係を学びます。コマンドをコピーするだけでなく、どの段階で何が起こるのかを説明できることが目標です。

## 作業ディレクトリの作成

プログラムごとに別々のディレクトリを使用すると、ソースと生成された実行可能ファイルを見つけるのは簡単です。端末でディレクトリを作成して移動します。

```shell
mkdir wave-study
cd wave-study
```

エディタでこのディレクトリに`main.wave`を作成します。ファイル名が`main.wave.txt`にならないように拡張子を確認してください。以下は関数の内部フラグメントではなく、ファイル全体です。

<!-- wave-example: book-hello -->
```wave playground
fun main() {
    println("Hello, Wave!");
}
```

実行結果：

```text
Hello, Wave!
```

次のコマンドで実行します。

```shell
wavec run main.wave
```

端末コマンドとWaveコードを混ぜないでください。 `wavec run`は端末に入力し、`fun main`はソースファイルに書き込みます。プログラムが出力した`Hello, Wave!`を再びソースに貼り付けるわけでもありません。

## 一行ずつ読む

`fun`は関数を宣言するキーワードです。関数は名前付きジョブの束であり、`main`は実行可能ファイルの始点です。他の関数は後で直接定義します。

`main`の後ろの`()`は、パラメータを書き込む場所です。このプログラムのmainはパラメータを受け取らないので空です。 `{`と`}`の間に関数の作業を書きます。インデントはブロックを読みやすくし、ブロック境界自体は中括弧で表されます。

`println("Hello, Wave!");`は文字列を出力する文です。二重引用符は、文字列リテラルの始まりと終わりを示します。引用符自体が出力に含まれるわけではありません。セミコロンは、この文が終わったことを示します。

## 文は作成した順序で実行されます

出力を3回変更するようにしてください。前のプログラムと合わせないで、main.waveの内容を以下のプログラム全体に置き換えてください。

<!-- wave-example: book-sequence -->
```wave playground
fun main() {
    println("start");
    println("working");
    println("done");
}
```

実行結果：

```text
start
working
done
```

最初の文を終えたら、次の文に進みます。ここには同時に実行されるジョブはありません。出力順序を変更したい場合は、文の順序を変更してください。後で条件文と反復文を学ぶことで、この進捗順序を制御できます。

## printとprintln

`println`は最後に改行を追加します。 `print`は自動的に行を変更しません。小さな部分を続けて一行を作るのに違いがあります。

<!-- wave-example: book-print-lines -->
```wave playground
fun main() {
    print("Wave");
    print(" ");
    println("study");
    println("second line");
}
```

実行結果：

```text
Wave study
second line
```

3回出力しても必ず3行が出るわけではありません。出力呼び出し数と行数を区切ります。改行を文字列内に直接書き込むときは、`\n`escapeを使用します。文字列escapeとソースの改行は[文字列の章](/docs/ja/language/strings)で詳しく説明されています。

## 値を文字列に入れる

計算結果を文字列に入れるには、`{}`桁に対応する値を渡します。

<!-- wave-example: book-format-first -->
```wave playground
fun main() {
    println("{} + {} = {}", 2, 3, 2 + 3);
}
```

実行結果：

```text
2 + 3 = 5
```

最初の`{}`に2、2番目に3、3番目に計算結果5が入ります。フォーマット文字列と値はカンマで区切ります。値の個数を変えると、placeholder個数も合わせなければなりません。これは、文字列を加算する演算とは異なる出力文法です。

## 検査・ビルド・実行を分ける

これまで使用していたrunはビルドと実行を続けて行います。どのステップが失敗したのかを知りたい場合は、次のように分割できます。

```shell
wavec check main.wave
wavec build main.wave -o hello
```

checkは文法やタイプなどをチェックしますが、プログラムのすべての入力をテストするわけではありません。たとえば、ファイルが実行時に存在するかどうかは、ソースチェックだけではわかりません。 buildは実行ファイルを生成します。 Linux/macOSでは、次のように実行します。

```shell
./hello
```

Windowsでは、出力ファイル名を`hello.exe`に指定し、PowerShellから`.\hello.exe`として実行します。実行可能ファイルの作成後にソースを変更した場合は、再構築する必要がある変更内容が反映されます。

## 終了コードも結果です

人は出力文を読みますが、シェルや他のプログラムは終了コードで成功したかどうかを判断できます。次のmainは、i32を返すことを指定します。

<!-- wave-example: book-exit-success -->
```wave playground
fun main() -> i32 {
    println("completed");
    return 0;
}
```

実行結果：

```text
completed
```

0 は正常終了を示す慣例です。直接失敗を知らせるときは、ゼロ以外のコードを返します。 Linux/macOSシェルでは実行直後に`echo $?`、PowerShellでは`$LASTEXITCODE`でコードを確認します。その間に別のコマンドを実行すると、確認するターゲットが異なる場合があります。

`return`を実行すると関数が終了します。 mainはパラメータを宣言せず、デフォルト値があっても許可されません。 mainの戻りタイプは省略するか、i32を使用します。

## 最初のエラーを読み取る方法

次のコードは意図的に間違っています。実行例とは異なり、checkで失敗する必要があります。

```wave
fun main() {
    println("hello")
}
```

文末のセミコロンはありません。診断が表示する行と直前の文を見てください。コンパイラが表示する場所は、エラーが発生したポイントではなく、間違った構造を解釈できなくなったポイントである可能性があります。

最初は最初のエラーを修正してからもう一度確認してください。前の括弧や引用符が閉じられていないと、後ろの通常のコードでもいくつかのエラーが発生する可能性があります。すべての行を同時に修正しようとすると、元の原因を見逃しやすいです。

## 練習問題

1. 自己紹介を3行出力するプログラムを作成してください。
2. 12と8をフォーマット引数として渡し、`12 * 8 = 96`を出力します。
3. 成功メッセージを出力し、終了コード 0 を返すように作成します。
4. printのみを3回使用する場合は、出力行数がどのようになるかを実行する前に予測してください。

### プール：計算プロセス出力

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

実行結果：

```text
Learning Wave
My first program
Ready to calculate
12 * 8 = 96
```

計算結果を文字列に直接`96`と書く代わりに式で渡しました。入力値が変わっても計算結果と表示が別々に遊ばないようにする最初のステップです。次の章では、同じ値を何度も書き込まないように変数に名前を付けます。


## ソースファイルの最上位項目

Waveソースは、次の最上位項目を設定できます。

- `import(...)`
- `const`, `static`
- `type`
- `struct`, `enum`, `proto`
- `extern(...)`, `export(...)`
- `fun`
- サポートされる項目の前の`#[target(...)]`条件

インポートできる宣言の前には`pub`を付けます。地域`var`宣言は関数またはブロックの中に置きます。

## フリースタンディングプログラム

カーネル、ブートコード、またはランタイムを持たないターゲットは、プレスタンディングビルドオプションを使用できます。

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files
```

`--freestanding`はデフォルトのライブラリ依存関係をオフにするビルドプランに関連付けられ、`--entry`はリンカエントリシンボルを設定します。実際の起動可能な出力を作成するには、ターゲットアーキテクチャ、リンカスクリプト、およびオブジェクトフォーマットまで一緒に設計する必要があります。

## 意図的に失敗するエントリポイント

デフォルト値があっても、mainにパラメータを入れることはできません。次のファイルをcheckするとエラーです。

<!-- wave-example: reject-main-argument -->
```wave
fun main(value: i32 = 0) -> i32 {
    return value;
}
```
