---
name: WorkflowScript Language Design Decisions
description: CSE341 project language design - all confirmed decisions for WorkflowScript DSL
type: project
originSessionId: 1fec7301-635f-422d-a7bd-68095b7ac614
---
WorkflowScript — Görev/iş akışı tanımlama ve simülasyon DSL'i. Go ile interpreter yazılacak.

**Why:** YAML/Makefile untyped ve programlanamaz. WorkflowScript typed, döngü/koşul destekli, parse anında hata yakalar.

**Deadlines:**
- Part 1: 8 May 2026 (lexer + parser + design spec)
- Part 2: 22 May 2026 (type checker + interpreter)

## Confirmed Design Decisions

1. **Primitive types:** `int`, `float`, `bool`
2. **Structured type:** `task` — built-in record (timeout: int, retries: int, parallel: bool). Name equivalence.
3. **Control structures:** `if/else`, `while`
4. **Domain-specific constructs:** `task { }`, `pipeline { stage ... }`, `run`, `on_failure { }`
5. **Scoping:** Static (lexical)
6. **Precedence** (low→high): `||` → `&&` → `== !=` → `< <= > >=` → `+ -` → `* /` → unary `- !`
7. **Associativity:** All binary operators left-associative
8. **Coercion:** `int` → `float` implicitly in arithmetic only (var decl, fun arg, binary arithmetic). Float does not narrow to int. `==`/`!=` strict: aynı tip gerektirir, coercion uygulamaz. `void` hiçbir operatörle karşılaştırılamaz.
9. **Assignment:** Statement, not expression
10. **User functions:** `fun name(params) -> type { body }`

## Part 2 Confirmed Design Decisions (added 2026-05-17)

11. **Short-circuit evaluation:** Evet — `&&` ve `||` short-circuit. Sol false/true ise sağ eval edilmez.
12. **`run pipeline`:** Pipeline stage'lerini sırayla execute eder (task adlarını sırayla çalıştırır).
13. **Domain-specific runtime error:** `run foo` ama `foo` tanımlı değilse → `runtime error: undefined task or pipeline 'foo'`
14. **`print` built-in:** Evet — `print(expr)` stdout'a yazar; örnek programlarda hesaplama çıktısı göstermek için kullanılır.
15. **Type equivalence:** Name equivalence (Part 1'deki kararla tutarlı).
16. **Operand evaluation sırası:** Sol-dan sağa (defined, short-circuit ile tutarlı).
17. **Assignment:** Statement only (zaten mevcut parser ile tutarlı, D1 §4.7'de belgelenir).
18. **Return type checking:** Evet — checker, ReturnStmt'te ifade tipini fonksiyonun bildirdiği dönüş tipiyle karşılaştırır.
19. **RunStmt static check:** Evet — checker compile-time'da `run` hedefinin tanımlı task/pipeline olduğunu doğrular.

## Part 2 Implementation Checklist

**Code (D2): TAMAMLANDI**
- [x] Return type checking: checker'da `returnType` field ile FunDecl'de beklenen tip saklanıyor, ReturnStmt'te `typesCompatible` ile karşılaştırılıyor
- [x] Short-circuit AND/OR: `evalLogical` metodu ile AND için false ise / OR için true ise sağ taraf eval edilmiyor
- [x] Domain-specific runtime error: RunStmt tasks/pipelines map'inde yoksa `line N: undefined task or pipeline "X"` hatası
- [x] Pipeline execution: `run pipelineName` → stage'leri sırayla `fmt.Printf("[run] %s")` ile çalıştırıyor; `runLog` ile izlenebilir
- [x] `print` built-in: checker'da TVoid döner; eval'da `fmt.Println(val)` basar; her iki yerde özel case
- [x] Checker'da RunStmt static check: `c.pipelines` map'i first-pass'te dolduruldu, `run` hedefi compile-time'da doğrulanıyor
- [x] AST: `ReturnStmt`'e `Keyword token.Token` eklendi (satır numarası raporlaması için)

**Docs (D1): TAMAMLANDI — `design-spec-part2.md`**
- [x] §4.4 Semantics: `while` (While-False/While-True rules) ve `run` (Run-Task/Run-Pipeline/Run-Error rules) için big-step operational semantics
- [x] §4.5 Type System: 5 soru yanıtlandı (int/float/bool value sets, strong typing, int→float coercion only, name equivalence, task record design)
- [x] §4.7 Expressions & Assignment: precedence tablosu, short-circuit ✓, L→R eval ✓, assignment=statement ✓
- [x] §4.8 Design Rationale: 5 karar için ayrı paragraf + trade-off (strong typing, coercion, name equiv., short-circuit, assignment)

**Test (D3):**
- [ ] 3 programın actual output'unu belgele
- [ ] Type error yakalayan 1 program ekle

## Exam Defense Notes

- Static scope: "Değişkene hangi değerin bağlandığını kodu okuyarak anlarsın, çalıştırmadan"
- Name equivalence: Tip seviyesinde tüm task'lar type-compatible (`fun f(t: task)` her task'ı kabul eder). Ayrım değer seviyesinde: `build` ve `deploy` farklı `*TaskValue` pointer'larıdır, aynı field değerleri olsa bile. "geçemez" demek yanlış — tip sistemi ayırt etmez, programcı ayırt eder.
- Coercion: "Veri kaybına yol açan dönüşüm implicit olmaz"
- pipeline vs function: "Fonksiyon hesaplar, pipeline iş akışını bildirimsel tanımlar"
- on_failure: "Hata yönetimi iş akışının ayrılmaz parçası, pipeline zaten hangi stage'in başarısız olduğunu biliyor"

## How to apply
Use these decisions when writing EBNF grammar, implementing lexer/parser/interpreter, and answering design questions.
