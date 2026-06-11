# WorkflowScript

Görev (task) ve iş akışı (pipeline) tanımlamak için tasarlanmış, statik tipli küçük bir DSL. Lexer, parser, type checker ve tree-walking interpreter Go ile yazılmıştır.

## Özellikler

- **Tipler:** `int`, `float`, `bool`, built-in `task` kaydı (`timeout: int`, `retries: int`, `parallel: bool`)
- **Kontrol akışı:** `if/else`, `while`
- **Fonksiyonlar:** `fun name(params) -> type { ... }`, `return`
- **Domain yapıları:** `task { }`, `pipeline { stage ... on_failure { } }`, `run`
- **Statik (lexical) scope**, name equivalence, `int → float` implicit coercion (sadece aritmetikte)
- **Short-circuit** `&&` ve `||`, soldan sağa değerlendirme
- **Built-in:** `print(expr)`

## Build

    go build -o wsc ./cmd/wsc/

## Run

    ./wsc <file.ws>              # lex + parse + type check + execute
    ./wsc --dump-ast <file.ws>   # AST'i bas ve çık

Hata durumunda çıkış kodu `1`; mesaj türe göre `type error:` veya `runtime error:` ön ekiyle stderr'e yazılır.

## Örnekler

    ./wsc examples/valid/01_basic.ws
    ./wsc examples/valid/02_functions.ws
    ./wsc examples/valid/03_pipeline.ws
    ./wsc --dump-ast examples/valid/03_pipeline.ws

Hata örnekleri:

    ./wsc examples/invalid/01_missing_brace.ws   # parse error
    ./wsc examples/invalid/06_type_error.ws      # type error

## Kısa örnek

```
task build  { timeout: 30  retries: 0 parallel: false }
task deploy { timeout: 120 retries: 1 parallel: false }

pipeline release {
    stage build
    stage deploy
    on_failure { run build }
}

fun cost(t: task) -> int {
    return t.timeout * (t.retries + 1)
}

if cost(deploy) > 60 {
    run release
}
```

## Proje yapısı

    cmd/wsc/        CLI giriş noktası
    internal/lexer/   tokenizer
    internal/token/   token tanımları
    internal/parser/  recursive-descent parser
    internal/ast/     AST düğümleri
    internal/checker/ statik tip kontrolü
    internal/eval/    tree-walking interpreter
    examples/valid/   çalışan programlar
    examples/invalid/ parse / tip / runtime hata örnekleri
