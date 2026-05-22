# Part 2 — Type Checker Düzeltmeleri

Bu belgede Part 2 implementasyonu sırasında tespit edilen ve düzeltilen üç tip sistemi açığı açıklanmaktadır.

---

## 1. Exhaustive Return Check

### Problem

Non-void bir fonksiyon her kod yolunda `return` etmese de checker bunu kabul ediyordu. Aşağıdaki program hatasız derleniyor, `max(1, 2)` çağrısı `nil` döndürüyor ve program `<nil>` yazdırıyordu:

```
fun max(a: int, b: int) -> int {
    if a > b { return a }
    // else branch yok — checker geçiriyor, runtime nil döndürüyor
}
var x: int = max(1, 2)
print(x)   // <nil>
```

Bu, WorkflowScript'in birincil tasarım hedefi olan **güvenilirlik** ilkesiyle doğrudan çelişiyordu. Tip checker'ı geçen bir programın runtime'da tanımsız davranış üretmemesi gerekir.

### Düzeltme

İki yardımcı fonksiyon eklendi. Bunlar AST üzerinde yineleyerek bir bloğun tüm yürütme yollarının `return` ile kapanıp kapanmadığını analiz eder:

```go
func blockAlwaysReturns(b *ast.Block) bool {
    for _, s := range b.Stmts {
        if stmtAlwaysReturns(s) {
            return true
        }
    }
    return false
}

func stmtAlwaysReturns(s ast.Stmt) bool {
    switch n := s.(type) {
    case *ast.ReturnStmt:
        return true
    case *ast.IfStmt:
        return n.Else != nil && blockAlwaysReturns(n.Then) && blockAlwaysReturns(n.Else)
    }
    return false
}
```

`FunDecl` kontrolünde, `checkBlock` başarıyla tamamlandıktan sonra void olmayan fonksiyonlar için bu kontrol çalıştırılır:

```go
if _, isVoid := retType.(VoidType); !isVoid {
    if !blockAlwaysReturns(n.Body) {
        return fmt.Errorf("line %d: function %q does not return on all paths",
            n.Name.Line, n.Name.Lexeme)
    }
}
```

### Kapsam

| Durum | Sonuç |
|-------|-------|
| `fun f() -> int { return 1 }` | geçerli |
| `fun f() -> int { if c { return 1 } else { return 0 } }` | geçerli |
| `fun f() -> int { if c { return 1 } }` | **hata** — else branch yok |
| `fun f() -> int { while c { return 1 } }` | **hata** — while garantili değil |
| `fun f() -> void { run build }` | geçerli — void fonksiyonlar muaf |

`while` döngüsü kasıtlı olarak garantisiz kabul edilir. Koşulun her zaman `true` olduğunu statik olarak saptamak sabit ifade değerlendirmesi (constant folding) gerektirir; bu scope dışı tutuldu.

---

## 2. `void == void` Karşılaştırması

### Problem

`void == void` ifadesi hatasız geçiyordu. Bunun nedeni: eski EQ/NEQ kontrolü `left != right` (Go interface eşitliği) kullanıyordu. `TVoid` bir singleton olduğundan her iki taraf da aynı interface değeri — karşılaştırma `true` döndürüyor, hata üretilmiyordu:

```go
// Eski kod
case token.EQ, token.NEQ:
    if left != right {   // TVoid == TVoid → koşul false → hata yok
        return nil, fmt.Errorf("cannot compare %s with %s", ...)
    }
```

Sonuç:

```
fun notify() -> void { run build }
var eq: bool = notify() == notify()   // checker OK — anlamsız
```

`void` bir değer tipi değil, "değer yok" belirtecidir. Karşılaştırma operatörüyle kullanılması semantik olarak anlamsız ve her zaman yanlış bir soruyu temsil eder.

### Düzeltme

`EQ` ve `NEQ` case'ine explicit void guard eklendi. Tip eşitliği kontrolü (`left != right`) korundu — `==`/`!=` design-spec §4.5.3 gereği aynı tipi gerektirir, `int == float` geçersizdir:

```go
case token.EQ, token.NEQ:
    if _, ok := left.(VoidType); ok {
        return nil, fmt.Errorf("line %d: void is not comparable", n.Op.Line)
    }
    if _, ok := right.(VoidType); ok {
        return nil, fmt.Errorf("line %d: void is not comparable", n.Op.Line)
    }
    if left != right {
        return nil, fmt.Errorf("line %d: cannot compare %s with %s",
            n.Op.Line, left.typeString(), right.typeString())
    }
    return TBool, nil
```

### Sonuç

```
fun notify() -> void { run build }
var eq: bool = notify() == notify()   // HATA: void is not comparable

var a: int   = 5
var b: float = 1.5
var eq2: bool = a == b   // HATA: cannot compare int with float (değişmedi)
```

---

## 3. `void` Tipinde Değişken Bildirimi

### Problem

`void` bir dönüş tipi belirtecidir; bir değerin tipi olamaz. Ancak `typeFromLexeme("void")` geçerli bir `TVoid` döndürdüğünden, `VarDecl` kontrolü `var x: void = notify()` ifadesini kabul ediyordu:

```
task build { timeout: 10 retries: 0 parallel: false }
fun notify() -> void { run build }
var x: void = notify()   // checker geçiriyor
print(x)                 // <nil>
```

Declared tip de `TVoid`, actual tip de `TVoid` olduğundan `typesCompatible(TVoid, TVoid)` true döndürüyordu ve hiçbir hata üretilmiyordu. C, Java, Go gibi dillerin tamamında `void` tipinde değişken bildirimi derleme hatasıdır.

### Düzeltme

`VarDecl` kontrolünde, `declared` tipi hesaplandıktan hemen sonra void kontrolü eklendi:

```go
case *ast.VarDecl:
    declared, err := typeFromLexeme(n.Type.Lexeme)
    if err != nil {
        return fmt.Errorf("line %d: %v", n.Type.Line, err)
    }
    if _, ok := declared.(VoidType); ok {
        return fmt.Errorf("line %d: cannot declare variable of type void", n.Name.Line)
    }
    // ... geri kalan kontroller
```

### Sonuç

```
var x: void = notify()   // HATA: cannot declare variable of type void
```

Mevcut `TestChecker_VoidReturnAsValue` testi etkilenmez: o testte bildirilen tip `int`'tir, `void` değil. Hata `typesCompatible(void, int)` kontrolünden gelmeye devam eder.

---

## Özet

| # | Açık | Etki | Düzeltme |
|---|------|------|---------|
| 1 | Non-void fonksiyon tüm yollarda return etmek zorunda değildi | Runtime `nil` dönüşü, sessiz hata | `blockAlwaysReturns` ile control flow analizi |
| 2 | `==`/`!=` coercion uygulamıyor, `void == void` geçerliydi | Aritmetikle tutarsız tip sistemi | `typesCompatible` + void guard |
| 3 | `var x: void` bildirimi geçerliydi | Semantik anlamsız durum, `<nil>` değer | VarDecl'de erken void reddi |


Design decisions for this record type:

- **Fixed field set.** Every task declaration must supply all three fields in the fixed order
  `timeout`, `retries`, `parallel`. There are no optional fields and no default values. This
  makes every task self-documenting and prevents accidental omissions: a missing field is a
  parse error, not a silent zero.

- **Field immutability.** Fields cannot be updated after the task is declared. The interpreter
  stores tasks as read-only structs (`TaskValue`). This guarantees reproducibility: the same
  `run build` call always executes with the same parameters, regardless of when during
  program execution it is invoked.

- **No inheritance or extension.** Tasks cannot extend other tasks or share field definitions.
  WorkflowScript does not need polymorphic dispatch; the domain calls for discrete, named units
  of work, not a hierarchy.

- **Field access via dot notation.** `t.timeout`, `t.retries`, `t.parallel` are the only way
  to read task fields. The three field names are reserved keywords to make this unambiguous in
  the lexer.

- **Static subscript check.** Accessing a non-existent field (e.g., `t.cost`) is caught at
  compile time by the type checker. There is no dynamic dispatch table, so no runtime lookup
  can fail.