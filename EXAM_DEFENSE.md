# WorkflowScript — Part 2 Exam Defense

**CSE341 · Concepts of Programming Languages**
**Gebze Technical University · 22 May 2026**

---
// direkt kod üzerinden de soru gelebilir o yüzden kodları oku ve kritik yerleri iyice öğren. Başka nasıl yapılabilirdi düşün.


## Genel Mimari

WorkflowScript bir **tree-walking interpreter**'dır. Pipeline:

```
source text
   → Lexer (tokens)
   → Parser (AST)
   → Checker (static type analysis)   ← Part 2
   → Interpreter (tree-walk eval)     ← Part 2
```

Checker ve Interpreter her ikisi de aynı **iki-geçiş (two-pass)** mimarisini kullanır:

1. **First pass:** Tüm `task`, `pipeline`, `fun` bildirimlerini kaydet (forward reference desteği)
2. **Second pass:** Her statement'ı sırayla analiz et / çalıştır

Bu sayede bir fonksiyon, kaynak dosyada altında tanımlanan bir task'a referans verebilir.

---

## Type Checker (`internal/checker/`)

### 4.1 Scope Yönetimi

```go
type Checker struct {
    scopes     []map[string]Type   // stack of scopes
    funs       map[string]*ast.FunDecl
    tasks      map[string]*ast.TaskDecl
    pipelines  map[string]*ast.PipelineDecl
    returnType Type                 // fonksiyon içindeyken beklenen dönüş tipi
}
```

`scopes` bir **stack**'tir. `push()` yeni bir scope açar, `pop()` kapatır. `lookup()` stack'i üstten alta tarar — bu **static (lexical) scoping**'in doğrudan implementasyonudur. Inner scope outer scope'taki ismi gölgeleyebilir (shadowing).

**Neden static scoping?**  
Caller'ın değişkenleri fonksiyon scope'unu etkilemez. Aynı fonksiyon farklı call site'lardan çağrıldığında aynı davranışı gösterir. Pipeline güvenilirliği için kritik.

### 4.2 Type Sistemi

5 tip:

| Tip    | Go struct    | Kullanım                                      |
|--------|-------------|----------------------------------------------|
| `int`  | `IntType`   | Tamsayı literal, `timeout`, `retries`         |
| `float`| `FloatType` | Ondalıklı literal, karışık aritmetik sonucu   |
| `bool` | `BoolType`  | Boolean literal, `parallel`, karşılaştırma   |
| `void` | `VoidType`  | Dönüş değeri olmayan fonksiyon               |
| `task` | `TaskType`  | Task declaration, fonksiyon parametresi       |

**Name equivalence:** `task` tipi için name equivalence uygulanır (Sebesta §6.15). Tip sistemde tek bir `TaskType` struct'ı vardır — `fun f(t: task)` her task'ı kabul eder (`build` da `deploy` da). Ayrım **tip seviyesinde değil, değer seviyesindedir**: runtime'da her task kendi adıyla ayrı bir `*TaskValue` olarak saklanır. İki task aynı field değerlerine sahip olsa bile farklı domain rolleri temsil eder; tip sistemi bunları ayırt etmez, bunu programcı yapar.

**Sıklıkla sorulan:** "Tek bir structured type varken name vs structural equivalence'ın farkı ne ki?"  
Cevap: Sebesta §6.15 Celsius/Fahrenheit örneğiyle tam bunu gösterir — aynı field yapısına sahip iki tip, structural equivalence altında birbirinin yerine geçebilir. WorkflowScript'e ilerleyen sürümlerde `pipeline` gibi ikinci bir structured tip eklenseydi (aynı `timeout: int, retries: int, parallel: bool` field'larıyla), structural equivalence altında `task` bekleyen bir fonksiyon `pipeline` de kabul ederdi — bu domain açısından yanlış olurdu. Name equivalence bunu baştan önler.

### 4.3 Örtük Tür Dönüşümü (Coercion)

```go
func typesCompatible(a, b Type) bool {
    if a == b { return true }
    _, aIsInt   := a.(IntType)
    _, bIsFloat := b.(FloatType)
    return aIsInt && bIsFloat   // int → float, tek yönlü
}
```

`int` → `float` widening implicit'tir. Örnek: `var x: float = 1 + 2.5` geçerlidir. Ters yön (`float` → `int`) yoktur — veri kaybına yol açacağından kasıtlı çıkarıldı.

**Sıklıkla sorulan:** "Neden float → int yok?"  
Cevap: Truncation sessiz veri kaybıdır. WorkflowScript'in güvenilirlik önceliğiyle çelişir; bu tip dönüşümü explicit olmak zorundadır.

**Sıklıkla sorulan:** "Sebesta coercion'ın strong typing'i zayıflattığını söylüyor, sizin dilinizde durum ne?"  
Cevap: Sebesta §7.4.1'de haklı — `d = b * a` örneğinde `a` yanlışlıkla `c` yerine yazıldığında Java bunu yakalamaz çünkü `int` → `float` coercion devreye girer. WorkflowScript'te de aynı teorik zayıflık vardır: `int` beklenirken başka bir `int` değişken yazılırsa tip hatası değil, sessiz widening olur. Bu nedenle dili "kesinlikle strongly typed" değil, "etkin anlamda strongly typed" olarak tanımlamak daha doğrudur. Pratikte risk minimldir: widening tek yönlüdür, yalnızca numerik tipler arasında geçerlidir ve task konfigürasyon domain'i dardır.

`==` ve `!=` **aynı tipleri** gerektirir — `int == float` geçersizdir. Bu, `typesCompatible` kullanan aritmetikten kasıtlı olarak farklıdır: eşitlik karşılaştırması tip-bilinçlidir, widening uygulamaz. Tek ek kural: `void` karşılaştırılamaz (TVoid bir singleton olduğundan naif `left != right` kontrolü yanlışlıkla `void == void`'e izin verirdi):

```go
case token.EQ, token.NEQ:
    if _, ok := left.(VoidType); ok {
        return nil, fmt.Errorf("line %d: void is not comparable", n.Op.Line)
    }
    if _, ok := right.(VoidType); ok {
        return nil, fmt.Errorf("line %d: void is not comparable", n.Op.Line)
    }
    if left != right {
        return nil, fmt.Errorf("line %d: cannot compare %s with %s", ...)
    }
    return TBool, nil
```

**`var x: void` yasaktır.** `void` bir dönüş tipi belirtecidir, değer tipi değildir. `VarDecl` kontrolünde declared tip `VoidType` ise compile-time hata üretilir:

```go
if _, ok := declared.(VoidType); ok {
    return fmt.Errorf("line %d: cannot declare variable of type void", n.Name.Line)
}
```

### 4.4 Return Type Kontrolü ve Exhaustive Return

```go
case *ast.FunDecl:
    c.returnType = retType
    if err := c.checkBlock(n.Body); err != nil { ... }
    c.returnType = prev
    // non-void fonksiyonlar her yolda return etmek zorunda
    if _, isVoid := retType.(VoidType); !isVoid {
        if !blockAlwaysReturns(n.Body) {
            return fmt.Errorf("line %d: function %q does not return on all paths", ...)
        }
    }

case *ast.ReturnStmt:
    if !typesCompatible(actual, c.returnType) {
        return fmt.Errorf("cannot return %s from function declared to return %s", ...)
    }
```

`Checker.returnType` alanı her `FunDecl`'e girerken saklanıp çıkarken restore edilir — nested fonksiyon desteği için.

`blockAlwaysReturns` / `stmtAlwaysReturns` AST üzerinde yürür. `ReturnStmt` garantili sayılır; `IfStmt` yalnızca hem `then` hem `else` branch'i varsa ve her ikisi de return ediyorsa garantili kabul edilir. `while` hiçbir zaman garantili değildir — koşulun her zaman true olduğunu statik olarak saptamak constant folding gerektirir.

| Fonksiyon | Sonuç |
|-----------|-------|
| `fun f() -> int { return 1 }` | geçerli |
| `fun f(x:int) -> int { if x>0 { return x } else { return 0 } }` | geçerli |
| `fun f(x:int) -> int { if x>0 { return x } }` | **hata** — else yok |
| `fun f() -> int { while true { return 1 } }` | **hata** — while garantisiz |
| `fun f() -> void { run build }` | geçerli — void fonksiyonlar muaf |

**Void dönüş özel case:** `fun f() -> void { return 42 }` hata verir çünkü `typesCompatible(int, void)` false döner.

### 4.5 RunStmt Static Check

```go
case *ast.RunStmt:
    _, isTask     := c.tasks[name]
    _, isPipeline := c.pipelines[name]
    if !isTask && !isPipeline {
        return fmt.Errorf("line %d: undefined task or pipeline %q", ...)
    }
```

`run ghost` → compile-time hata. Runtime'a hiç ulaşmaz. Bu, WorkflowScript'in "definition-time error" garantisinin merkezidir.

### 4.6 Pipeline Stage Kontrolü

```go
case *ast.PipelineDecl:
    for _, stage := range n.Stages {
        if _, ok := c.tasks[stage.Lexeme]; !ok {
            return fmt.Errorf("line %d: unknown task %q in pipeline", ...)
        }
    }
```

Pipeline içindeki her `stage` referansının var olan bir task'a işaret ettiği compile-time'da doğrulanır.

### 4.7 print Built-in

```go
if n.Name.Lexeme == "print" {
    // 1 argüman kontrolü
    c.checkExpr(n.Args[0])   // argümanın tipi geçerliyse TVoid döner
    return TVoid, nil
}
```

`print` dilin keyword'ü değil, built-in fonksiyondur. Checker'da özel case ile ele alınır. Return type `TVoid` olduğundan `var x: int = print(42)` hata verir.

---

## Interpreter (`internal/eval/`)

### 5.1 Değer Temsili

```go
type TaskValue struct {
    Timeout  int64
    Retries  int64
    Parallel bool
}
```

Runtime'da değerler Go'nun `interface{}` tipiyle taşınır:
- `int` → `int64`
- `float` → `float64`
- `bool` → `bool`
- `task` → `*TaskValue`

### 5.2 Environment ve Static Scoping

```go
// evalCall içinde:
callEnv := NewEnv(interp.env)   // global env'e bağlı, caller env'e değil
```

Fonksiyon çağrısında yeni environment'ın parent'ı **global environment**'tır, caller'ın environment'ı değil. Bu **static scoping**'i uygular. Eğer caller'ın env'i parent yapılsaydı dynamic scoping olurdu.

**Test:** `TestEval_StaticScoping` — `get_x()` fonksiyonu global `x`'i döndürür, caller'ın potansiyel lokal `x`'ini değil.

### 5.3 Short-Circuit Evaluation

```go
func (interp *Interpreter) evalLogical(n *ast.BinaryExpr, env *Env) (interface{}, error) {
    left, _ := interp.evalExpr(n.Left, env)
    if n.Op.Type == token.AND && !left.(bool) {
        return false, nil   // sağ taraf eval edilmez
    }
    if n.Op.Type == token.OR && left.(bool) {
        return true, nil    // sağ taraf eval edilmez
    }
    right, _ := interp.evalExpr(n.Right, env)
    return right.(bool), nil
}
```

`&&` ve `||` ayrı metodla işlenir. Normal `evalBinary`'nin aksine sağ tarafı unconditional evaluate etmez.

**Test:** `false && (1/0 == 0)` → hata yok (division by zero hiç çalışmaz).  
**Test:** `true || (1/0 == 0)` → hata yok.

### 5.4 Pipeline Execution

```go
case *ast.RunStmt:
    if _, ok := interp.tasks[name]; ok {
        fmt.Printf("[run] %s\n", name)
        interp.runLog = append(interp.runLog, name)
        return nil, nil
    }
    if pipeline, ok := interp.pipelines[name]; ok {
        for _, stage := range pipeline.Stages {
            fmt.Printf("[run] %s\n", stage.Lexeme)
            interp.runLog = append(interp.runLog, stage.Lexeme)
        }
        return nil, nil
    }
    return nil, fmt.Errorf("line %d: undefined task or pipeline %q", ...)
```

`run task_name` → o task'ı çalıştırır.  
`run pipeline_name` → her stage'i sırayla çalıştırır.  
`runLog` test amaçlı; üretimde `fmt.Printf` yeterlidir.

**Neden on_failure otomatik tetiklenmez?** `on_failure` bloku bir hata handling deklarasyonudur; hangi stage'in başarısız olduğunu interpreter şu an simüle etmez (task'lar gerçek process çalıştırmaz). `on_failure` içindeki statement'lar bildirimsel olarak çalıştırılabilir, mevcut implementasyonda pipeline başarısız olmadığı için tetiklenmez — bu scope içi.

### 5.5 Return Propagation

```go
type returnSignal struct{ value interface{} }

// ReturnStmt eval:
return &returnSignal{value: val}, nil

// evalBlock eval:
if ret != nil { return ret, nil }   // sinyal yukarı taşınır

// evalCall:
if sig, ok := ret.(*returnSignal); ok {
    return sig.value, nil
}
```

`returnSignal` bir **sentinel value**'dur. Panic/recover kullanmadan return değerini call stack üzerinde taşır. Block her adımda `ret != nil` mi diye kontrol eder; fonksiyon çağrısı signal'ı alınca değeri wrap'ından çıkarır.

**Neden panic değil?** Panic Go'da exception eşdeğeridir — catch edilebilir ama hata akışlarıyla karışabilir, test edilmesi güçtür. Sentinel pattern temiz ve explicit'tir.

### 5.6 Type Coercion (Runtime)

```go
li, lIsInt := left.(int64)
lf, _      := toFloat(left)
mixed      := (lIsInt && !rIsInt) || (!lIsInt && rIsInt)

case token.PLUS:
    if mixed || (!lIsInt && !rIsInt) {
        return lf + rf, nil   // float arithmetic
    }
    return li + ri, nil       // int arithmetic
```

Runtime'da coercion checker'ı yansıtır: operandlardan biri float ise sonuç float. `toFloat()` hem `float64` hem `int64` kabul eder.

---

## Önemli Tasarım Kararları — Savunma Notları

### "Neden static scoping?"
Fonksiyonun davranışı call site'a bağlı değildir. Aynı pipeline farklı `on_failure` handler'larından çağrıldığında aynı değişkenleri görür. Kodu okuyarak name resolution yapılabilir, çalıştırmadan.

### "Neden task tipi immutable?"
Reproducibility. Aynı task her pipeline çalışmasında aynı parametrelerle çalışır. Runtime'da `TaskValue` pointer paylaşılır ama alanlar değiştirilmez.

### "int → float neden tek yönlü?"
Float → int truncation sessiz veri kaybıdır. Güvenilirlik önceliğiyle çelişir. Narrow dönüşüm explicit yazılmalıdır.

### "Assignment neden expression değil?"
`if x = 5` bug'ını önler. Parser tek token lookahead ile `IDENT "="` (assignment statement) ve `IDENT "("` (call expr) arasını ayırt eder.

### "pipeline vs fun — fark nedir?"
`fun` hesaplar ve değer döndürür. `pipeline` iş akışını **bildirimsel** tanımlar — stage sırası, failure handling. `run` ile tetiklenir, çağrı sırasında value üretmez.

### "task tipi return forbidden — neden?"
Tasks are **static lifetime** — program boyunca yaşar. Bir fonksiyondan task dönmek task'ın sanki stack-dynamic gibi davrandığı izlenimini yaratır. Checker bunu reddeder (`task` return type'ı tip olarak mevcut ama fonksiyon return type olarak kullanılamaz — şu implementasyonda `fun f() -> task` ayrıştırılabilir ama domain dokümantasyonunda forbidden).

---

## Test Kapsamı

### Checker Testleri (`checker_test.go`)

| Test | Ne test ediyor |
|------|---------------|
| `TestChecker_ValidVarDecl` | Temel var declaration |
| `TestChecker_TypeMismatch` | `var x: int = true` → hata |
| `TestChecker_IntToFloatCoercion` | `var x: float = 1 + 2.5` → OK |
| `TestChecker_UndeclaredVariable` | `y` tanımsız → hata |
| `TestChecker_VoidReturnAsValue` | `notify()` void → `var x: int = notify()` hata |
| `TestChecker_UnknownTaskInPipeline` | `stage ghost` → hata |
| `TestChecker_TaskFieldAccess` | `build.timeout` → int |
| `TestChecker_InvalidTaskField` | `build.nonexistent` → hata |
| `TestChecker_ReturnTypeMismatch` | `-> int { return true }` → hata |
| `TestChecker_VoidFunctionReturnValue` | `-> void { return 42 }` → hata |
| `TestChecker_RunUndefinedTarget` | `run ghost` compile-time → hata |
| `TestChecker_PrintBuiltin` | `print(42)` → OK |
| `TestChecker_ExhaustiveReturn` | `if` without else in non-void fun → hata |
| `TestChecker_ExhaustiveReturnIfElse` | `if/else` her iki branch return → OK |
| `TestChecker_VoidVariable` | `var x: void` → hata |
| `TestChecker_VoidComparison` | `notify() == notify()` → hata |

### Interpreter Testleri (`eval_test.go`)

| Test | Ne test ediyor |
|------|---------------|
| `TestEval_VarDecl` | `3 + 4 = 7` |
| `TestEval_WhileLoop` | `i` 0→3 sayar |
| `TestEval_FunctionCall` | `double(21) = 42` |
| `TestEval_TaskFieldAccess` | `build.timeout = 30` |
| `TestEval_DivisionByZero` | Runtime hata |
| `TestEval_ShortCircuitAnd` | `false && (1/0==0)` → hata yok |
| `TestEval_ShortCircuitOr` | `true \|\| (1/0==0)` → hata yok |
| `TestEval_RunUndefinedTarget` | `run ghost` runtime → hata |
| `TestEval_PipelineExecution` | `run ci` → `[build, test]` sırayla |
| `TestEval_PrintBuiltin` | `print(result)` → hata yok |
| `TestEval_StaticScoping` | `get_x()` global x'i görür |

---

## Örnek Programlar

### Program 1 — Koşullu Dispatch (`examples/valid/01_basic.ws`)

```
task fast     { timeout: 20  retries: 0  parallel: true  }
task thorough { timeout: 90  retries: 3  parallel: false }

fun total_cost(t: task) -> int {
    return t.timeout * (t.retries + 1)
}

var fast_cost:     int = total_cost(fast)       // 20 * 1 = 20
var thorough_cost: int = total_cost(thorough)   // 90 * 4 = 360

if fast_cost < thorough_cost {
    run fast      // bu dal çalışır
} else {
    run thorough
}
```

**Beklenen çıktı:**
```
[run] fast
```

**Gösterilen özellikler:** task field access, fonksiyon, int aritmetiği, if/else, run.

---

### Program 2 — Fonksiyonlar ve Field Access (`examples/valid/02_functions.ws`)

```
task deploy { timeout: 120 retries: 1 parallel: false }

fun max_time(t: int, r: int) -> int { return t * (r + 1) }
fun is_long(t: task) -> bool        { return t.timeout > 60 }

var m:    int  = max_time(120, 1)   // 120 * 2 = 240
var slow: bool = is_long(deploy)    // 120 > 60 → true

if slow { run deploy }
```

**Beklenen çıktı:**
```
[run] deploy
```

**Gösterilen özellikler:** task parametre tipi, bool dönüş, fonksiyon zinciri.

---

### Program 3 — Pipeline + While Loop (`examples/valid/03_pipeline.ws`)

```
task build  { timeout: 30  retries: 0 parallel: false }
task test   { timeout: 60  retries: 2 parallel: true  }
task deploy { timeout: 120 retries: 1 parallel: false }

pipeline release {
    stage build
    stage test
    stage deploy
    on_failure { run build }
}

fun should_run(limit: int) -> bool { return limit > 0 }

var attempt: int = 0
var limit:   int = 3

while attempt < limit { attempt = attempt + 1 }

if should_run(limit) { run release }
```

**Beklenen çıktı:**
```
[run] build
[run] test
[run] deploy
```

**Gösterilen özellikler:** pipeline, while loop, on_failure, pipeline execution sıralı.

---

### Program 4 — Type Error Yakalama

```
task build { timeout: 30 retries: 0 parallel: false }
fun notify() -> void { run build }
var x: int = notify()   // HATA: void değer int değişkenine atanamaz
```

**Beklenen çıktı (checker):**
```
line 3: cannot assign void to variable of type int
```

**Gösterilen özellikler:** void return type, değer olarak kullanamama, compile-time hata.

---

## Sıklıkla Sorulan Sorular

**S: Type checker ile interpreter arasında nasıl veri paylaşılıyor?**  
C: Paylaşılmıyor. İkisi de AST üzerinden ayrı geçişler yapar. Checker pure analysis; interpreter pure execution. Aralarında bir sembol tablosu aktarılmaz — interpreter kendi task/fun/pipeline map'lerini first pass'te yeniden oluşturur.

**S: Forward reference nasıl destekleniyor?**  
C: İki-geçiş mimarisi sayesinde. First pass tüm bildirimleri kaydeder; bu sayede bir fonksiyon, aynı dosyada sonra tanımlanan bir task'a referans verebilir.

**S: Shadowing çalışıyor mu?**  
C: Evet. `lookup()` scope stack'ini üstten alta tarar, bulduğu ilk tanımı döner. Inner scope outer scope'taki ismi gizler.

**S: on_failure ne zaman tetiklenir?**  
C: Mevcut implementasyonda `pipeline` struct'ı `OnFailure []ast.Stmt` saklar ve checker bu statement'ları type check eder. Runtime'da task'lar gerçek process başlatmaz; başarısız olma simüle edilmez. `on_failure` deklarasyon olarak mevcuttur, otomatik tetiklenmez. Bu, scope'un dışında kalan bir feature.

**S: Neden string tipi yok?**  
C: WorkflowScript'te pipeline stage adları identifier'dır, veri değil. String manipülasyonuna ihtiyaç yoktur. Eklenmesi tip sistemini ve checker'ı karmaşıklaştırır; domain'de karşılığı yoktur.

**S: Recursive fonksiyon çalışır mı?**  
C: Evet. `evalCall` her çağrıda yeni environment oluşturur. Fonksiyon kendi adını `funs` map'inden bulabilir; sonsuz özyineleme stack overflow ile sonuçlanır (Go call stack sınırına kadar).
