# D4 — AI Usage Journal — Part 2 Entries (11–18)

**CSE 341 — Concepts of Programming Languages**  
**Gebze Technical University**  
**Language:** WorkflowScript | **Student:** Ömer Faruk Koç | **Student #:** 210104004061  
**Part:** 2 | **Entries:** 8 (Entry 11–18) | **Period:** 2026-05-17 — 2026-05-22

---

## Journal Overview

| Entry | Date | Phase | Topic |
|-------|------|-------|-------|
| 11 | 2026-05-18 | Planning | Part-2 gap analysis — requirements mapping from handout |
| 12 | 2026-05-18 | Design | Part-2 design decisions (short-circuit, print, pipeline, type equivalence) |
| 13 | 2026-05-18 | Design / Review | Review of existing type checker and interpreter design decisions |
| 14 | 2026-05-18 | Architecture | Evaluation of external language implementor's advice |
| 15 | 2026-05-18 | Testing / Implementation | TDD implementation of 6 Part-2 features |
| 16 | 2026-05-17 | Testing / Bug Fix | Type checker weakness discovery and three-fix resolution |
| 17 (E2) | 2026-05-22 | Design | Name vs structural equivalence for `task` — Sebesta §6.14 check |
| 18 (E3) | 2026-05-22 | Testing | Type checker implementation — 5 known-answer inputs, bugs caught |

---

## Entry 11 — Part-2 Gap Analysis

| Field | Content |
|-------|---------|
| **Entry #** | 11 of 18 |
| **Date** | 2026-05-18 |
| **Phase** | Planning |
| **AI tool** | Claude Sonnet 4.6 |
| **Goal** | Identify what is missing for Part 2 by having the AI read the handout PDF against the current implementation, without making assumptions. |
| **Prompt** | *"@CSE341_Project_Handout.pdf projenin part-1 ini teslim ettik şimdi part-2 yi yapmamız gerekiyor. Ne eksik tespit et. Design kararlarını bana sor. Varsayım yapma."* |
| **Response (key part)** | The AI structured the gap analysis across four deliverables. D1 (documentation): §4.4 Semantics (operational/denotational semantics for 2 constructs), §4.5 Type System (5 questions: strong/weak, coercion, equivalence, structured type), §4.7 Expressions & Assignment (precedence table, short-circuit, eval order, assignment), §4.8 Design Rationale (1 paragraph + trade-off per decision). D2 (implementation): `run` statement has no runtime validation — `run undefined_task` only prints `[run] undefined_task` without error; `&&`/`\|\|` perform eager evaluation of both operands; pipeline `run` only prints the pipeline name, does not execute stages. D3 (example programs): actual output missing from all three valid programs, no program demonstrating a type error. D4/D5: minimum 6 journal entries for the 9–22 May period, retrospective update. |
| **Accepted** | The full gap structure. Categorizing by deliverable (D1/D2/D3) made it easy to prioritize: code gaps (D2) had to be fixed before documentation (D1) could be finalized, because the semantics section describes runtime behavior. |
| **Rejected / modified** | Nothing rejected. The AI correctly declined to assume design decisions and listed them as open questions to be answered separately. |
| **Errors I caught** | None in this exchange. The categorization was accurate and matched a manual re-reading of the handout. |
| **Reflection** | *Asking the AI to read the handout directly against the codebase produced a structured gap analysis faster than doing it manually. The key constraint — "don't assume, ask me" — prevented the AI from filling in design choices that were mine to make. This saved time and kept design ownership with me.* |

---

## Entry 12 — Part-2 Design Decisions

| Field | Content |
|-------|---------|
| **Entry #** | 12 of 18 |
| **Date** | 2026-05-18 |
| **Phase** | Design |
| **AI tool** | Claude Sonnet 4.6 |
| **Goal** | Confirm the 7 open design decisions identified in Entry 11, covering evaluation semantics, built-in functions, and expression properties. |
| **Prompt** | The AI asked 7 questions sequentially; my verbatim answers were: *"Evet, short-circuit"* · *"Stage'leri sırayla execute et"* · *"Undefined task/pipeline hatası"* · *"Evet, print built-in ekle"* · *"Name equivalence"* · *"Sol-dan sağa (defined)"* · *"Statement (sadece)"* |
| **Response (key part)** | The AI recorded all 7 decisions, linked each to a specific implementation location, and explained why they were consistent with the existing design: short-circuit is coherent with left-to-right evaluation order; `print` as a built-in avoids adding string-literal parsing (consistent with the Part 1 decision to exclude string types); name equivalence was already the rule for `task`, so extending it uniformly avoids a two-rule type system; statement-only assignment avoids expression-side-effects which would complicate the scoping model. The AI then produced an updated implementation checklist with 6 items and their target files. |
| **Accepted** | All 7 decisions. The AI's argument that `print` as a special-case built-in (not a declared function) avoids reopening the string type decision was the most useful reasoning — it confirmed the design stayed consistent without requiring backtracking. |
| **Rejected / modified** | None. |
| **Errors I caught** | None. The AI's linkage of each decision to a file and line number was verified against the actual codebase before proceeding to implementation. |
| **Reflection** | *Having the AI ask me the questions rather than me listing requirements produced cleaner decisions. Each question came with a brief justification so I could see the implications before answering. This is more reliable than stating requirements in isolation, because I could catch conflicts immediately — for example, the short-circuit/left-to-right pair was confirmed as consistent in the same exchange.* |

---

## Entry 13 — Review of Existing Type Checker and Interpreter Design

| Field | Content |
|-------|---------|
| **Entry #** | 13 of 18 |
| **Date** | 2026-05-18 |
| **Phase** | Design / Review |
| **AI tool** | Claude Sonnet 4.6 |
| **Goal** | Cross-check the existing Part 1 implementation against the Part 2 rubric requirements ("type checker consistent with declared type system", "interpreter respecting scoping, precedence, associativity, parameter-passing"). |
| **Prompt** | *"daha önce interpreter ve type checker ı yazmıştık onun design kararlarını gözden geçirelim."* Followed by: *"A type checker consistent with the type system you declared in your final D1 [...] An interpreter that executes programs respecting your declared scoping rule [...] Informative runtime errors [...] bunlardan eksik var mı?"* |
| **Response (key part)** | The AI produced two tables. Type checker — correct: strong typing (`typesCompatible`), int→float coercion (one direction), name equivalence (separate struct per type), forward declaration first-pass, static scoping (scope stack). Type checker — broken: `ReturnStmt` checks only that the expression is well-typed, not that it matches the declared return type (`fun f() -> int { return true }` passes); `RunStmt` returns `nil` unconditionally; void return not blocked. Interpreter — correct: static scoping (`NewEnv(interp.env)`, not caller env), call-by-value, int/int=int truncated, mixed int+float=float, division-by-zero error, stack-dynamic block lifetime. Interpreter — broken: `&&`/`\|\|` both operands evaluated before the operator case runs; `run pipelineName` only prints the name, stages not iterated. |
| **Accepted** | The full analysis. The two-column "correct / broken" framing was more useful than a pass/fail list because it confirmed which parts did not need to change, avoiding unnecessary refactoring. |
| **Rejected / modified** | Nothing rejected. |
| **Errors I caught** | The AI correctly identified that `fun f() -> int { return true }` passes the checker — I verified this by running the file and confirming no error was produced before proceeding to fix it. |
| **Reflection** | *Reading the implementation against the rubric requirements line by line surfaced three checker gaps that would have gone undetected until the grader ran test cases. The most dangerous was the missing return-type match — the checker was validating type correctness of the expression but not checking it against the function's declared return type, which is the central claim of a "strongly typed" language.* |

---

## Entry 14 — External Language Implementor Advice Evaluation

| Field | Content |
|-------|---------|
| **Entry #** | 14 of 18 |
| **Date** | 2026-05-18 |
| **Phase** | Architecture |
| **AI tool** | Claude Sonnet 4.6 |
| **Goal** | Evaluate three architectural recommendations received from an external language implementor against the current codebase, without automatically accepting them. |
| **Prompt** | *"bir programa dili geliştiren kişiden şu tavsiyeleri aldım. Bunları bizim projeye uyarlarsak nasıl olur sence? incele: (1) Dilin AST'si, type checking'i, interpreter'ini bir kütüphane olarak yazmak. (2) AST'ye source span dahil etmek. (3) Hata mesajlarını `enum` olarak tanımlamak."* |
| **Response (key part)** | **Recommendation 1 (library architecture):** Already done — `internal/` packages are the library, `cmd/wsc/main.go` is the application. Go's `internal/` convention enforces this separation. Adding a REPL or LSP would only require a new `cmd/` directory. Action: none. **Recommendation 2 (source spans):** Partially present — `Token.Line` exists but `Col` is absent. Several AST nodes (`IfStmt`, `WhileStmt`) do not even carry a token reference. Adding full `line:col` spans requires updating `Token`, the lexer column tracker, and every affected AST node. Value for this project: the rubric says "informative runtime errors" but does not require column precision. Cost: high for the remaining 5 days. **Recommendation 3 (error enums):** Not present — all errors are `fmt.Errorf` strings. Typed error structs would enable `errors.As` in tests and cleaner error formatting, but the rubric does not test for error type identity, only error message content. Cost: high refactor; benefit: low within this scope. |
| **Accepted** | The AI's cost/benefit framing for recommendations 2 and 3. Both are correct architectural improvements for a production language, but neither improves the grade within the 5-day remaining window. |
| **Rejected / modified** | The implicit suggestion to implement recommendation 2 or 3. Rejected in favor of finishing the 6 open implementation items from Entry 12. |
| **Errors I caught** | None in the AI's analysis. The AI correctly recognized that the library architecture was already in place rather than proposing a refactor. |
| **Reflection** | *This was a case where the AI's role was to give a second opinion, not to accept advice uncritically. The external recommendations were all sound for a production compiler, but the AI correctly weighted them against the project's actual constraints (rubric requirements, deadline). Asking the AI to evaluate rather than apply advice prevented scope creep.* |

---

## Entry 15 — TDD Implementation of Part-2 Features

| Field | Content |
|-------|---------|
| **Entry #** | 15 of 18 |
| **Date** | 2026-05-18 |
| **Phase** | Testing / Implementation |
| **AI tool** | Claude Sonnet 4.6 |
| **Goal** | Implement the 6 confirmed Part-2 features (return type checking, RunStmt static check, short-circuit AND/OR, pipeline execution, `print` built-in, undefined target runtime error) using TDD: failing tests first, verified red, then implementation. |
| **Prompt** | *"tamam project_workflowscript design ı implement edelim"* (TDD skill was invoked before implementation began.) |
| **Response (key part)** | The AI wrote failing tests across two files before touching implementation code. `checker_test.go`: `TestChecker_ReturnTypeMismatch` (`fun f() -> int { return true }` → error), `TestChecker_VoidFunctionReturnValue` (`fun notify() -> void { return 42 }` → error), `TestChecker_RunUndefinedTarget` (`run ghost` → compile-time error), `TestChecker_PrintBuiltin` (`print(42)` → no error). `eval_test.go`: `TestEval_ShortCircuitAnd` (right side `1/0 == 0` must not evaluate when left is false), `TestEval_ShortCircuitOr` (right side must not evaluate when left is true), `TestEval_RunUndefinedTarget` (`run ghost` → runtime error), `TestEval_PipelineExecution` (run ci → runLog == ["build", "test"]), `TestEval_PrintBuiltin` (`print(result)` → no error). All 9 tests failed (confirmed red). Implementation: `ReturnStmt` gained a `Keyword` token field for line numbers; checker gained `returnType` field and `pipelines` map; `evalLogical` separated short-circuit logic from `evalBinary`; `RunStmt` handler dispatches to task or pipeline with `runLog` tracking; `print` handled as special-case in both checker and eval. All 9 new tests plus all previous tests passed (green). |
| **Accepted** | The `runLog` field on the `Interpreter` struct — a testing-only slice that records executed task/stage names so `TestEval_PipelineExecution` can assert execution order without parsing stdout. This is cleaner than output capture and does not affect production behavior. The separation of `evalLogical` from `evalBinary` was also accepted: it keeps short-circuit semantics explicit and the case for AND/OR in `evalBinary` can be removed entirely, avoiding a hidden bug where both operands were already evaluated before the operator branch ran. |
| **Rejected / modified** | An initial version of `TestEval_PipelineExecution` that only checked `err == nil` was rejected — it passed even before pipeline execution was implemented (because `run ci` only printed and returned nil). The test was strengthened to assert the exact contents of `runLog`. |
| **Errors I caught** | The initial pipeline test was a false green — it passed before the feature existed. Detection: I questioned whether the test actually proved anything if `run ci` already returned nil without error. The fix (asserting `runLog`) turned it red as required, then green after implementation. |
| **Reflection** | *The TDD discipline caught one false-positive test before it became a false sense of confidence. Writing a test that compiles and returns `nil` is not a failing test — a failing test must assert an observable property that the implementation does not yet satisfy. The `runLog` approach gave a concrete, binary assertion rather than absence-of-error.* |

---

## Entry 16 — Type Checker Weakness Discovery and Three-Fix Resolution

| Field | Content |
|-------|---------|
| **Entry #** | 16 of 18 |
| **Date** | 2026-05-17 |
| **Phase** | Testing / Bug Fix |
| **AI tool** | Claude Sonnet 4.6 |
| **Goal** | Identify design ambiguities or implementation gaps in the Part-2 type checker and interpreter that could expose weaknesses during exam defense, then fix the ones that are genuine bugs rather than acceptable scope limitations. |
| **Prompt** | *"şimdi part-2 design kararlarımızı ve implementasyon yöntemlerinin düşün. Bunların için de sebestanın ambiguity olarak ifade ettiği bir yer var mı? Yani bizi açıklarken zor duruma sokacak 'şu şekilde daha iyi yapılabilir aslında?' dediğin bir yer var mı?"* — followed by, after the AI listed 5 weaknesses: *"düzeltilmesi gerekenkeri düzeltelim savunmaya çalışmayalım."* |
| **Response (key part)** | The AI identified 5 weaknesses by running test programs against the binary: **(1) Exhaustive return check missing** — `fun max(a: int, b: int) -> int { if a > b { return a } }` passes the checker but returns `<nil>` at runtime when the else branch is taken. Java, Go, and Kotlin reject this at compile time. Contradicts "reliability" as a design priority. **(2) EQ/NEQ coercion inconsistency** — `a + b` (int + float) is accepted via `typesCompatible`, as is `a > b`, but `a == b` (int == float) fails with "cannot compare int with float" because `==` used direct `left != right` (struct pointer equality) instead of `typesCompatible`. Two different rules for the same coercion. **(3) `var x: void` allowed** — `var x: void = notify()` compiles and prints `<nil>`. `void` is a return-type marker, not a value type; `var x: void` should be a compile-time error. **(4) `on_failure` block is dead code** — interpreter iterates stages but tasks simulate no real process, so no failure condition ever triggers. Block is type-checked but never executed. **(5) `void == void` valid** — `notify() == notify()` passes the checker and returns `bool`. Fixes applied for (1), (2), (3): `blockAlwaysReturns`/`stmtAlwaysReturns` added; `EQ`/`NEQ` case changed to use `typesCompatible` + explicit `void` guard; `VarDecl` check added for `VoidType`. 4 new tests added: `TestChecker_ExhaustiveReturn`, `TestChecker_ExhaustiveReturnIfElse`, `TestChecker_VoidVariable`, `TestChecker_VoidComparison`. All tests passed. |
| **Accepted** | Fixes (1) and (3). Exhaustive return check is required for a language claiming strong static typing — `blockAlwaysReturns`/`stmtAlwaysReturns` added. `var x: void` is semantically meaningless and was rejected with a VarDecl guard. Both are genuine implementation errors with no ambiguity. Fix (2) — EQ/NEQ `typesCompatible` change — was initially accepted in this session but later **reverted** after reading `design-spec-part2.md` (see Entry 17/E2): §4.5.3 explicitly lists where coercion applies (var decl, fun arg, binary arithmetic) and EQ/NEQ is not in that list. The original `left != right` is correct per the spec. The void guard on EQ/NEQ was kept, since `void` is not comparable under any reading of the spec. |
| **Rejected / modified** | Weaknesses (4) and (5) were not fixed. `on_failure` dead code is a known scope limitation — tasks simulate behavior, not real processes. `void == void` is covered by the void guard added to EQ/NEQ. The EQ/NEQ `typesCompatible` change was applied then reverted — see Entry 17 for the spec-based reasoning. |
| **Errors I caught** | The AI's claim that EQ/NEQ coercion was an "inconsistency" was itself incorrect. `a > b` (int vs float) is accepted by `isNumeric` checks on both sides, which widens both to numeric without `typesCompatible`. `a == b` using `left != right` is strict same-type equality. These are two different code paths implementing two different rules — the spec intends them to be different. The AI misread the coercion rule as applying to all operators; it only applies to the three contexts listed in §4.5.3. |
| **Reflection** | *The EQ/NEQ "fix" applied in this session was wrong. It took reading the design spec to catch it — not any test, because the existing tests did not cover `int == float`. This shows a gap in the test suite (no test for EQ/NEQ with mixed numeric types) and a gap in my own review process: I accepted the AI's framing of "inconsistency" without checking whether the spec actually required consistent behavior. The spec was authoritative; the AI was not.* |

---

## Entry 17 (E2) — Name vs Structural Equivalence for `task`

**Experiment E2 task:** *"Ask the AI to explain the difference between name equivalence and structural equivalence for your record type. Does its explanation match Sebesta §6.14? Does the example it gives actually run under the equivalence rule it described?"*

| Field | Content |
|-------|---------|
| **Entry #** | 17 of 18 (Experiment E2) |
| **Date** | 2026-05-22 |
| **Phase** | Design |
| **AI tool** | Claude Sonnet 4.6 |
| **Goal** | Ask the AI to explain the difference between name and structural equivalence for the `task` type, then verify its explanation against Sebesta §6.14 by running the example programs it produces. |
| **Prompt** | *"WorkflowScript'in task tipi için name equivalence ile structural equivalence arasındaki farkı açıkla. Hangi kuralı kullanıyoruz ve neden?"* |
| **Response (key part)** | **Name equivalence (Sebesta §6.14.1):** two types are equivalent if they share the same name. In WorkflowScript every task declaration — `task build { ... }`, `task deploy { ... }` — belongs to the single type named `task`. A function `fun f(t: task)` accepts any declared task because they all carry the type name `task`. **Structural equivalence (Sebesta §6.14.2):** two types are equivalent if they have the same structure (same fields in the same order). Since WorkflowScript has only one `task` type with fixed fields (`timeout: int`, `retries: int`, `parallel: bool`), both rules give the same result for any two task values. The observable difference would arise if WorkflowScript had two *differently-named* record types with identical fields (e.g., a hypothetical `pipeline_config` type with the same three fields as `task`): under name equivalence they would be incompatible; under structural equivalence they would be compatible. WorkflowScript uses name equivalence — confirmed by the type checker: `typesCompatible(TTask, TTask)` is `a == b` (Go struct equality on `TaskType{}`), and non-task values (`int`, `bool`) are rejected even if they hold a numeric value like `timeout`. The AI also noted a subtlety: within the coercion rule (§4.5.3), EQ/NEQ comparison uses strict `left != right` (same type pointer), not `typesCompatible`. This means `int == float` is a compile-time error even though `int + float` is valid — the coercion rule applies to the three contexts listed in §4.5.3, not to equality operators. |
| **Accepted** | The two-level framing of name vs structural equivalence, and the observation that the distinction has no practical effect in the current language because there is only one record type. The explanation correctly cites Sebesta §6.14 and the Celsius/Fahrenheit motivating example (which appears in §6.15 as well). The EQ/NEQ observation was used to catch and revert a wrong code change (see Entry 16). |
| **Rejected / modified** | An initial claim that `int == float` "should" be valid by analogy with `int + float` was rejected. Reading §4.5.3 showed that coercion applies only to var decl, function arguments, and binary arithmetic — equality operators are strict. The AI's analogy was plausible but not backed by the spec. |
| **Errors I caught** | **(1)** The AI initially suggested applying `typesCompatible` to EQ/NEQ for consistency with arithmetic operators. This was incorrect: §4.5.3 enumerates exactly three contexts for coercion, and EQ/NEQ is not among them. I caught this by reading the design spec directly rather than accepting the AI's reasoning. **(2)** The AI did not initially distinguish between name equivalence's effect at the *type* level (all tasks are `task`) and identity at the *value* level (build ≠ deploy at runtime). After prompting, it clarified the separation correctly. |
| **Reflection** | *The experiment confirmed the spec is self-consistent: name equivalence with a single `task` type means all task values are type-compatible at the checker level, while remaining individually identifiable at runtime. The more valuable finding was negative — the AI's analogy-based argument for extending coercion to EQ/NEQ was technically coherent but spec-incorrect. Checking primary sources (the design spec) rather than AI reasoning is necessary whenever a proposed change affects the documented type rules.* |

**Test programs run and actual output:**

```
// Test 1: both task values accepted by same function
task build  { timeout: 30 retries: 0 parallel: false }
task deploy { timeout: 60 retries: 3 parallel: true  }
fun max_timeout(t: task) -> int { return t.timeout }
var a: int = max_timeout(build)
var b: int = max_timeout(deploy)
print(a)   // output: 30
print(b)   // output: 60
```
Result: ✓ compiles and runs. Both task values accepted — name equivalence confirmed.

```
// Test 2: int rejected where task expected (type name matters, not field shape)
fun get_timeout(t: task) -> int { return t.timeout }
var x: int = get_timeout(42)
```
Result: ✗ `type error: line 2: arg 1: expected task, got int` — correct; `42` has the same value as a `timeout` field but is not of type `task`.

```
// Test 3: int == float under strict EQ/NEQ rule
var a: int   = 5
var b: float = 5.0
var eq: bool = a == b
```
Result: ✗ `type error: line 3: cannot compare int with float` — correct per §4.5.3.

---

## Entry 18 (E3) — Type Checker Implementation: Known-Answer Tests and Bugs

**Experiment E3 task:** *"Ask the AI to implement your type checker (or a non-trivial portion of it). Run its code on at least three test inputs where you know the right answer. Report any bugs, subtle or blatant, and how you fixed them."*

| Field | Content |
|-------|---------|
| **Entry #** | 18 of 18 (Experiment E3) |
| **Date** | 2026-05-22 |
| **Phase** | Testing |
| **AI tool** | Claude Sonnet 4.6 |
| **Goal** | Have the AI implement the type checker, then run 5 known-answer programs against it and report every bug found with how it was fixed. |
| **Prompt** | *"tamam project_workflowscript design ı implement edelim"* — this was the implementation request (Entry 15) that produced the type checker. The E3 known-answer tests were run immediately after by sending each program as: *"bu programı çalıştır, beklenen çıktı X"* for each of the 5 inputs, and comparing the checker's actual output to the expected result. |
| **Response (key part)** | The AI implemented the type checker across several files: `checker.go` (main dispatch, scope stack, `returnType` field, `blockAlwaysReturns`/`stmtAlwaysReturns`), `types.go` (`typesCompatible`, `typeFromLexeme`). The non-trivial portions were: return type checking via `c.returnType` saved/restored on function entry/exit; exhaustive return analysis via `stmtAlwaysReturns` (ReturnStmt always, IfStmt only if both branches return, while never); RunStmt static check via first-pass `c.pipelines` map; `print` as a special-case built-in returning `TVoid`; void guard on VarDecl and EQ/NEQ. |
| **Accepted** | The scope stack (push/pop per block), the first-pass registration of tasks/pipelines/functions for forward references, the `blockAlwaysReturns` logic (while loop deliberately excluded as not statically guaranteed), and the `print` built-in as a special case rather than a declared function. |
| **Rejected / modified** | The EQ/NEQ `typesCompatible` change (see Entry 16 and Entry 17/E2) was rejected and reverted after reading the design spec. The checker now correctly uses `left != right` for EQ/NEQ, keeping strict same-type equality as documented in §4.5.3. |
| **Errors I caught** | **(1 — blatant)** The original checker (`ReturnStmt` case) checked only that the return expression was well-typed, not that its type matched the function's declared return type. `fun f() -> int { return true }` compiled without error. Fix: added `c.returnType` field; stored/restored on FunDecl entry/exit; checked `typesCompatible(actual, c.returnType)` in ReturnStmt. **(2 — subtle)** The EQ/NEQ `typesCompatible` "fix" applied in Entry 16 was itself a bug: it made the checker accept `int == float`, contradicting §4.5.3. The existing test suite did not cover this case — no test for EQ/NEQ with mixed numeric types existed. Detection: reading the design spec directly. Fix: revert to `left != right`. **(3 — moderate)** `blockAlwaysReturns` is an approximation: `while true { return 1 }` diverges in theory only if the condition is always true, but the function is rejected as "does not return on all paths" because `while` is conservatively excluded. This is the standard safe approximation (Sebesta §6.14 notes the halting-problem undecidability of exact termination analysis) and was accepted as intentional. |
| **Reflection** | *The most instructive finding was bug (2): a fix that was syntactically correct, passed all existing tests, and had a plausible rationale (consistency) was still wrong. The only thing that caught it was reading primary sources. This demonstrates that AI-generated code changes must be verified against the specification, not just against the test suite — especially when the test suite was also partially generated by the AI and may share its blind spots.* |

**Test programs run and actual output:**

| # | Input | Expected | Got | Match |
|---|-------|----------|-----|-------|
| 1 | `fun f() -> int { return true }` | type error: return bool from int | `type error: line 1: cannot return bool from function declared to return int` | ✓ |
| 2 | `var x: float = 3 + 1.5` then `print(x)` | OK, output `4.5` | `4.5` | ✓ |
| 3 | `run ghost` | compile error: undefined | `type error: line 1: undefined task or pipeline "ghost"` | ✓ |
| 4 | `fun max(a: int, b: int) -> int { if a > b { return a } }` | error: missing else | `type error: line 1: function "max" does not return on all paths` | ✓ |
| 5 | `fun max(a:int, b:int)->int { if a>b {return a} else {return b} }` then `print(max(3,7))` | OK, output `7` | `7` | ✓ |
