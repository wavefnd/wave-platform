---
translation_set_id: program-structure
path: language/program-structure
locale: es
group: language
group_order: 2
order: 1
title: 1. Del archivo fuente al programa en ejecución
summary: Obtenga información sobre archivos fuente, funciones, resultados, inspección y ejecución.
---

## Antes de comenzar este capítulo

Prepare el compilador y la biblioteca estándar según [Instrucciones de instalación](/docs/es/getting-started/install). Si puede ejecutar `wavec --version` en su terminal, puede comenzar. Cualquier editor que pueda guardar archivos de texto sin formato servirá.

En este capítulo, comenzamos creando una sola línea de salida y aprendemos las relaciones entre los archivos fuente, las funciones, la compilación, la ejecución y el código de salida. El objetivo no es sólo copiar instrucciones, sino poder explicar qué sucede y en qué etapa.

## Crear un directorio de trabajo

El uso de un directorio separado para cada programa facilita la búsqueda de los archivos fuente y ejecutables generados. Cree y navegue hasta un directorio en la terminal.

```shell
mkdir wave-study
cd wave-study
```

Cree `main.wave` en este directorio con su editor. Verifique la extensión para asegurarse de que el nombre del archivo no sea `main.wave.txt`. A continuación se muestra el archivo completo, no solo un fragmento dentro de la función.

<!-- wave-example: book-hello -->
```wave playground
fun main() {
    println("Hello, Wave!");
}
```

Resultado de la ejecución:

```text
Hello, Wave!
```

Ejecútelo con el siguiente comando:

```shell
wavec run main.wave
```

No mezcle comandos de terminal con el código Wave. Ingrese `wavec run` en la terminal y escriba `fun main` en el archivo fuente. No es necesario volver a pegar la salida `Hello, Wave!` del programa en la fuente.

## leer línea por línea

`fun` es una palabra clave que declara una función. Una función es un conjunto de operaciones con nombre y `main` es el punto de entrada del ejecutable. Otras funciones las definirá usted mismo más adelante.

`()` después de `main` es donde escribes los parámetros. main en este programa no toma parámetros y, por lo tanto, está vacío. Escriba las operaciones de la función entre `{` y `}`. La sangría hace que los bloques sean más fáciles de leer para los humanos, y los propios límites del bloque se indican mediante llaves.

`println("Hello, Wave!");` es una oración que genera una cadena. Las comillas dobles indican el principio y el final de una cadena literal. Las comillas en sí no se incluyen en el resultado. El punto y coma indica el final de esta declaración.

## Las declaraciones se ejecutan en el orden en que se escriben.

Intente cambiarlo para imprimir tres veces. No lo combine con el programa anterior, sino reemplace el contenido de main.wave con el programa completo a continuación.

<!-- wave-example: book-sequence -->
```wave playground
fun main() {
    println("start");
    println("working");
    println("done");
}
```

Resultado de la ejecución:

```text
start
working
done
```

Después de terminar la primera oración, pase a la siguiente. Aquí no hay tareas ejecutándose simultáneamente. Si desea cambiar el orden de salida, simplemente cambie el orden de las oraciones. Puede controlar esta secuencia aprendiendo declaraciones condicionales y bucles más adelante.

## print y println

`println` agrega un salto de línea al final. `print` no cambia las líneas automáticamente. La diferencia es evidente al conectar piezas pequeñas para crear una sola línea.

<!-- wave-example: book-print-lines -->
```wave playground
fun main() {
    print("Wave");
    print(" ");
    println("study");
    println("second line");
}
```

Resultado de la ejecución:

```text
Wave study
second line
```

Imprimir tres veces no siempre da como resultado tres líneas. Distinga entre número de llamadas de salida y número de líneas. Al escribir un salto de línea directamente en una cadena, utilice `\n` escape. La cadena escape y los saltos de línea en la fuente se tratan en detalle en [hoja de hilo](/docs/es/language/strings).

## Insertar valor en cadena

Para poner el resultado del cálculo en una cadena, pase el valor correspondiente a `{}`.

<!-- wave-example: book-format-first -->
```wave playground
fun main() {
    println("{} + {} = {}", 2, 3, 2 + 3);
}
```

Resultado de la ejecución:

```text
2 + 3 = 5
```

El primero `{}` contiene 2, el segundo contiene 3 y el tercero contiene 5. La cadena de formato y el valor están separados por comas. Si cambia la cantidad de valores, la cantidad de placeholder también debe coincidir. Esta es una sintaxis de salida diferente a la operación de adición de cadenas.

## Divida la inspección, la construcción y la ejecución

run utilizado hasta ahora realiza la construcción y ejecución de forma secuencial. Si desea saber qué paso falló, puede desglosarlo así:

```shell
wavec check main.wave
wavec build main.wave -o hello
```

check comprueba la gramática, los tipos, etc., pero no prueba todas las entradas del programa. Por ejemplo, la inspección del origen por sí sola no puede determinar si un archivo existe en tiempo de ejecución. build crea un archivo ejecutable. En Linux/macOS, ejecute lo siguiente.

```shell
./hello
```

En Windows, nombre el archivo de salida `hello.exe` y ejecútelo en PowerShell como `.\hello.exe`. Si modifica la fuente después de crear un archivo ejecutable, debe reconstruirla para que se reflejen los cambios.

## El código de salida también es un resultado.

Los humanos leen la declaración de salida, pero un shell u otro programa puede determinar el éxito mediante el código de salida. Lo siguiente especifica que main devuelve i32.

<!-- wave-example: book-exit-success -->
```wave playground
fun main() -> i32 {
    println("completed");
    return 0;
}
```

Resultado de la ejecución:

```text
completed
```

0 es una convención para indicar un apagado normal. Al señalar una falla directamente, devuelve un código distinto de cero. En el shell Linux/macOS, el código se verifica como `echo $?` inmediatamente después de la ejecución, y en PowerShell, se verifica como `$LASTEXITCODE`. Si ejecuta otros comandos mientras tanto, lo que verifique puede cambiar.

Al ejecutar `return` finaliza la función. main no declara ningún parámetro y aunque tengan valores predeterminados, no están permitidos. Omita el tipo de devolución de main o use i32.

## Cómo leer el primer error

El siguiente código es intencionalmente incorrecto: a diferencia del ejemplo en ejecución, debería fallar en check.

```wave
fun main() {
    println("hello")
}
```

No hay punto y coma al final de la oración. Mire la línea que muestra el diagnóstico y la frase que la precede inmediatamente. Es posible que la ubicación marcada por el compilador no sea donde se originó el error, sino donde la estructura incorrecta ya no se puede interpretar.

Primero, corrija el primer error y luego vuelva a verificar. Si los paréntesis o comillas anteriores no están cerrados, pueden producirse varios errores incluso en el código normal que sigue. Si intentas arreglar todas las líneas al mismo tiempo, es fácil pasar por alto la causa original.

## problemas de practica

1. Cree un programa que imprima tres líneas de autopresentación.
2. Pase 12 y 8 como argumentos de formato para imprimir `12 * 8 = 96`.
3. Escríbalo para imprimir un mensaje de éxito y devolver un código de salida de 0.
4. Prediga antes de la ejecución cuál será el número de líneas de salida si usa print solo tres veces.

### Solución: salida del proceso de cálculo

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

Resultado de la ejecución:

```text
Learning Wave
My first program
Ready to calculate
12 * 8 = 96
```

En lugar de escribir el resultado del cálculo directamente en la cadena como `96`, lo pasé como una expresión. Este es el primer paso para garantizar que los resultados del cálculo y la visualización no cambien incluso si cambia el valor de entrada. En el próximo capítulo, daremos nombres a las variables para evitar escribir el mismo valor varias veces.


## Elemento de nivel superior en el archivo fuente

La fuente Wave puede constar de los siguientes elementos de nivel superior:

- `import(...)`
- `const`, `static`
- `type`
- `struct`, `enum`, `proto`
- `extern(...)`, `export(...)`
- `fun`
- `#[target(...)]` condición anterior al artículo admitido

Anteponga las declaraciones importables con `pub`. Coloque la declaración local `var` dentro de una función o bloque.

## Programa independiente

Los objetivos sin kernel, código de arranque o tiempo de ejecución pueden usar la opción de compilación independiente.

```shell
wavec build kernel.wave \
  --freestanding \
  --entry=_start \
  --linker-script=linker.ld \
  --no-start-files
```

`--freestanding` vincula a un plan de compilación que desactiva las dependencias de biblioteca predeterminadas y `--entry` establece el símbolo de entrada del vinculador. Para crear una salida de arranque real, debe diseñar la arquitectura de destino, el script del vinculador e incluso el formato del objeto.

## Puntos de entrada que fallan intencionalmente

No puede tener un parámetro en main incluso si tiene un valor predeterminado. Es un error copiar el siguiente archivo check.

<!-- wave-example: reject-main-argument -->
```wave
fun main(value: i32 = 0) -> i32 {
    return value;
}
```
