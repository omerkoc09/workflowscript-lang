# WorkflowScript — Design Specification (D1) — Part 2 Additions

**CSE341 — Concepts of Programming Languages**  
**Gebze Technical University**  
**Part 2 Submission — 22 May 2026**

> This document extends `design-spec.md` (Part 1). Sections §4.4, §4.5, §4.7, and §4.8
> are new in Part 2. Sections §4.1–§4.3 and §4.6 are unchanged from Part 1.

---

## 4.4 Semantics

Operational semantics (Sebesta §3.5.1) is used to describe the meaning of WorkflowScript
constructs. The approach follows the method Sebesta describes: each language construct is
translated into an equivalent sequence of statements in a simpler **intermediate language**
whose meaning is self-evident. The human reader acts as the virtual machine that executes
these intermediate statements.

### 4.4.1 Intermediate Language

The intermediate language used below consists of the following statement forms. All statement
forms are drawn from the vocabulary Sebesta §3.5.1 introduces:

| Intermediate statement          | Meaning                                          |
|---------------------------------|--------------------------------------------------|
| `ident = expr`                  | Assign the value of `expr` to variable `ident`  |
| `goto label`                    | Unconditional jump to `label`                   |
| `if not(expr) goto label`       | Jump to `label` if `expr` is false              |
| `output(str)`                   | Print `str` to standard output (side effect)    |
| `error(str)`                    | Halt with error message `str`                   |

Labels are written as `label:` at the start of a line. T and P denote the global, immutable
task store and pipeline store respectively, both populated before any statement executes.

---

### 4.4.2 While Statement

The `while` statement is the only looping construct in WorkflowScript. Its meaning is given
by translating it into the intermediate language. The translation has two parts: a conditional
check at the top and an unconditional back-edge at the bottom, matching the pattern Sebesta
§3.5.1 uses for the C `for` construct.

```
WorkflowScript                   Intermediate Language
──────────────────────────────   ──────────────────────────────────────────────
while cond {                     loop: if not(cond) goto end
    s₁                                 s₁
    s₂                                 s₂
    …                                  …
    sₙ                                 sₙ
}                                      goto loop
                                 end:
```

**Reading the translation.** On each iteration the interpreter evaluates `cond`. If it is
false, control jumps to `end` and the loop is skipped (or exits). If it is true, the body
statements `s₁ … sₙ` execute in order and control returns to `loop` for the next test.

**Non-termination.** If `cond` never becomes false, the `goto loop` back-edge fires
indefinitely and the program diverges. WorkflowScript provides no `break` or `continue`;
termination is the programmer's responsibility. The type checker guarantees `cond` is of type
`bool` but cannot determine whether the loop terminates.

**Nested blocks.** Each statement `sᵢ` in the body may itself be a block; if so, it is
expanded recursively by the same translation rules. An empty body translates to only the
`goto loop` back-edge, producing a spin-loop that depends on a side-effecting condition.

---

### 4.4.3 Run Statement (Domain-Specific Construct)

`run` is WorkflowScript's primary domain-specific statement. It triggers either a single
named task or an entire named pipeline. There are two cases depending on whether `name`
resolves to a task or a pipeline in the global stores.

**Case 1 — Running a task** (`name ∈ dom(T)`):

```
WorkflowScript                   Intermediate Language
──────────────────────────────   ──────────────────────────────────────────────
run name                         output("[run] name")
```

A single output line is produced. The environment (variable bindings) is not modified;
`run` is a pure output action.

**Case 2 — Running a pipeline** (`name ∈ dom(P)`, stages = `[t₁, t₂, … tₙ]`):

```
WorkflowScript                   Intermediate Language
──────────────────────────────   ──────────────────────────────────────────────
run name                         output("[run] t₁")
                                 output("[run] t₂")
                                 …
                                 output("[run] tₙ")
```

The pipeline's stage list is unrolled into one `output` call per stage in declaration order.
There is no conditional branching; every stage always executes.

**Case 3 — Undefined name** (`name ∉ dom(T)` and `name ∉ dom(P)`):

```
WorkflowScript                   Intermediate Language
──────────────────────────────   ──────────────────────────────────────────────
run name                         error("undefined task or pipeline 'name'")
```

This case is unreachable in well-typed programs: the static type checker verifies that every
`run` target is declared before execution begins, so `error(...)` is included here only for
completeness.

**Key observations.**

1. `run` does **not** modify variable bindings. Task parameters (`timeout`, `retries`,
   `parallel`) are fixed at declaration time; the intermediate translation contains no
   assignment to any program variable.

2. Case 1 vs Case 2 is resolved by checking `dom(T)` first. A name cannot simultaneously
   be a task and a pipeline because both share the same global namespace and the parser
   rejects duplicate declarations.

3. The `on_failure` block of a `pipeline` declaration is statically type-checked but is
   never entered during normal execution: since `run` is simulated (output only), no failure
   signal is ever raised. It would be triggered by an extension that models real process
   failure.

---

## 4.5 Type System

### 4.5.1 Primitive Types and Value Sets

WorkflowScript has three primitive types:

| Type    | Go runtime representation | Value set                                              |
|---------|--------------------------|--------------------------------------------------------|
| `int`   | `int64`                  | 64-bit signed integers stored in two's-complement; range −2⁶³ to 2⁶³ − 1 |
| `float` | `float64`                | IEEE 754 double precision; ≈ ±1.8 × 10³⁰⁸, 15–17 sig. digits |
| `bool`  | `bool`                   | {`true`, `false`}                                      |

`int` covers all realistic timeout values (seconds) and retry counts. `float` is available for
computations that mix integer task parameters into fractional results. `bool` corresponds to
the `parallel` task field and to all comparison results.

There is no `string` type. Pipeline stage names are identifiers in the source text, not runtime
data; no string manipulation is needed in the workflow domain.

### 4.5.2 Strong or Weak Typing?

WorkflowScript is **strongly typed** (Sebesta §6.14 definition: a language is strongly typed
if type errors are always detected). Every variable carries a declared type annotation that
does not change across assignments. The type checker enforces the following rules before any
execution begins:

- A variable declared as `int` cannot be assigned a `bool` value.
- Arithmetic operators (`+`, `-`, `*`, `/`) require both operands to be numeric (`int` or
  `float`); applying them to `bool` is a compile-time error.
- Boolean operators (`&&`, `||`, `!`) require `bool` operands.
- A function declared `-> int` must return an expression of type `int` (or `float` — see
  §4.5.3); returning `bool` is a compile-time error.
- `void` cannot be used as a value: `var x: void` is rejected, and the result of a `void`
  function cannot appear in an expression.

No type information is deferred to runtime: every name's type is known statically, and every
expression's type is determined from its sub-expressions without executing the program.

**Caveat — coercion and strong typing.** Sebesta §6.14 notes that implicit coercion is the
primary way languages weaken the strong-typing guarantee: if an `int` is silently accepted
where a `float` is expected, a programmer who writes the wrong variable by mistake will not
receive a type error. WorkflowScript's single coercion rule (int→float widening, §4.5.3)
introduces this theoretical weakening. In practice the risk is minimal: widening only applies
in one direction, it only occurs between numeric types, and the domain (task configuration
values) is narrow enough that accidentally passing an integer where a float is expected is
unlikely to cause a semantic error. Nevertheless, the language is best described as
*effectively strongly typed* rather than *absolutely strongly typed* in Sebesta's strictest
sense.

### 4.5.3 Implicit Type Coercion

There is exactly one coercion rule:

> **int-to-float widening:** an expression of type `int` is implicitly accepted wherever a
> `float` is expected, in both variable declarations and function arguments.

```
typesCompatible(int, float) = true
typesCompatible(float, int) = false   -- narrowing is forbidden
typesCompatible(T, T)       = true    -- reflexive for all other types
typesCompatible(T, U)       = false   -- otherwise
```

This coercion is applied in:

- Variable declarations: `var x: float = 1 + 2` is valid; the integer result `3` widens to `3.0`.
- Function arguments: a parameter of type `float` may receive an `int` argument.
- Binary arithmetic: if one operand is `float` and the other is `int`, the `int` is promoted to
  `float` and the result is `float`.

Narrowing (`float` → `int`) is deliberately absent. Truncation loses information silently and
conflicts with WorkflowScript's reliability goal. Any narrowing conversion must be written
explicitly by the programmer (not supported in the current version, consistent with the
language's minimal scope).

### 4.5.4 Type Equivalence

WorkflowScript uses **name equivalence** for the `task` structured type (Sebesta §6.15).

There is exactly one type named `task`. Every task declaration — regardless of the values
assigned to `timeout`, `retries`, and `parallel` — belongs to this single type. Two tasks with
identical field values are therefore type-compatible; a function accepting `task` may receive
any declared task as an argument.

Name equivalence is the right choice for two reasons.

**Primary reason — forward extensibility and semantic safety.** Sebesta §6.15 illustrates the
hazard of structural equivalence with the classic Celsius/Fahrenheit example: two record types
with an identical single `float` field are structurally equivalent, yet treating a Celsius
value as Fahrenheit is a semantic error that structural equivalence silently permits. The same
hazard applies here. If a second structured type — say, `pipeline` — were later added to
WorkflowScript, it would likely carry the same fields (`timeout: int`, `retries: int`,
`parallel: bool`). Under structural equivalence a function that expected a `task` would
silently accept a `pipeline`, conflating two fundamentally different domain concepts. Name
equivalence prevents this: the type checker would reject the mismatch regardless of field
layout. The principle is that structurally similar types can be semantically incompatible, and
name equivalence is the mechanism that enforces this separation.

**Secondary reason — conceptual fit.** Tasks are domain roles identified by name
(`build`, `deploy`, `test`), not by their parameter values. Two tasks with identical timeouts
and retry counts are still different units of work. Name equivalence reflects this: the type
system recognises only the declared type name `task`, not the field values, consistent with
the domain model.

### 4.5.5 Structured Type Design Decisions (Sebesta §6.5–§6.7)

The `task` type is WorkflowScript's single structured type. It is a **record** (Sebesta §6.7)
with three fixed fields:

| Field      | Type   | Meaning                                      |
|------------|--------|----------------------------------------------|
| `timeout`  | `int`  | Maximum seconds the task may run             |
| `retries`  | `int`  | Number of retry attempts on failure          |
| `parallel` | `bool` | Whether the task may run concurrently        |

Sebesta §6.7 identifies two primary design issues for record types; a third concern (implementation)
is addressed in the same section. Each is resolved below for `task`.

**Design issue 1 — Field reference syntax.**
Sebesta §6.7 contrasts the dot notation used in most languages (`record.field`) with COBOL's
reversed `OF` syntax (`field OF record`). WorkflowScript uses dot notation exclusively:
`t.timeout`, `t.retries`, `t.parallel`. Dot notation was chosen because it is the
near-universal convention in languages that WorkflowScript programmers are likely to know (Go,
Python, Java, JavaScript). Reversing the order (COBOL style) would add cognitive friction with
no benefit in a single-record language. The three field names (`timeout`, `retries`, `parallel`)
are reserved keywords so that the lexer can tokenise `t.timeout` unambiguously even though the
dot is also used as a decimal separator in other languages (WorkflowScript has no floating-point
literals, so no ambiguity arises).

**Design issue 2 — Elliptical references.**
In languages with nested records, Sebesta §6.7 notes that some designs allow the programmer to
omit intermediate record names when the field name is unambiguous in scope (an *elliptical
reference*). WorkflowScript deliberately **disallows** elliptical references. Every field access
must be written as `variable.field` (e.g., `t.timeout`), never as a bare `timeout`. This
decision keeps name resolution simple: the parser never needs to search the environment for a
record that owns a given field name. It also prevents shadowing bugs where a local variable
named `timeout` might be confused with a task field of the same name. Because WorkflowScript
has only one structured type and three field names, the verbosity cost is negligible.

**Implementation — compile-time descriptor, no run-time descriptor.**
Sebesta §6.7 explains that record field offsets are fixed at compile time and can be stored in
a *type descriptor*. No descriptor is needed at run time because field names are never looked
up dynamically. WorkflowScript follows this model exactly: the type checker resolves
`t.timeout` to the `timeout` field of the `task` type at compile time and rejects any unknown
field name (e.g., `t.cost`) as a type error. The interpreter then accesses `TaskValue` fields
directly by name in Go struct literals — there is no hash-map lookup or dynamic dispatch at
run time. This means field access is O(1) and cannot fail after the type-checking pass.

---

## 4.7 Expressions and Assignment

### Operator Precedence and Associativity

The table below lists all binary operators from lowest to highest binding strength. All binary
operators are left-associative. Precedence is encoded directly into the recursive-descent
grammar (§4.3) — each level calls the level above it — so the table is a reading aid, not
a separate mechanism.

| Precedence | Operators              | Associativity |
|-----------|------------------------|---------------|
| 1 (lowest) | `\|\|`                | left          |
| 2          | `&&`                   | left          |
| 3          | `==`  `!=`             | left          |
| 4          | `<`  `<=`  `>`  `>=`  | left          |
| 5          | `+`  `-`               | left          |
| 6          | `*`  `/`               | left          |
| 7 (highest)| unary `-`  `!`         | right (prefix)|

The unary operators bind more tightly than any binary operator. `!a && b` parses as `(!a) && b`,
and `-a * b` parses as `(-a) * b`.

### Short-Circuit Evaluation of Boolean Operators

WorkflowScript uses **short-circuit (lazy) evaluation** for `&&` and `||` (Sebesta §7.6):

- `e₁ && e₂`: if `e₁` evaluates to `false`, `e₂` is not evaluated and the result is `false`.
- `e₁ || e₂`: if `e₁` evaluates to `true`, `e₂` is not evaluated and the result is `true`.

Sebesta §7.6 motivates this choice with the canonical array-bounds guard example: without
short-circuit evaluation, the second operand of `&&` would be evaluated even when the first
is false, causing a potential runtime error (e.g., out-of-range index). WorkflowScript adopts
short-circuit semantics to support the same guard pattern in `while` conditions.

Sebesta §7.6 also warns that short-circuit evaluation interacts with side effects: if the
unevaluated operand contains a side-effecting expression, the side effect is silently skipped.
WorkflowScript mitigates this by restricting side effects: the only side-effecting statement
is `run`, which cannot appear inside an expression. Function calls in expressions return
values but do not modify global state, so the skip-on-short-circuit hazard does not apply.

The type checker evaluates both operands unconditionally for type correctness. This is sound
because type checking is a static pass that does not execute expressions.

### Order of Operand Evaluation

Sebesta §7.2.2 identifies operand evaluation order as a design issue whenever operands can
have side effects: if a function call in an expression modifies a variable that also appears
as another operand, the result depends on which operand is evaluated first. Two solutions
exist: disallow functional side effects, or define a specific evaluation order (Sebesta
§7.2.2.1).

WorkflowScript chooses the second option. Operand evaluation order is **defined as
left-to-right**: for a binary expression `e₁ op e₂`, `e₁` is always evaluated before `e₂`.
This matches the approach Java takes (Sebesta §7.2.2.1: "The Java language definition
guarantees that operands appear to be evaluated in left-to-right order") and is consistent
with short-circuit semantics, which requires the left operand to be known before deciding
whether to evaluate the right.

### Assignment: Statement or Expression?

Assignment (`x = expr`) is a **statement**, not an expression (Sebesta §7.7.1, §7.7.5). It
cannot appear inside a larger expression; it must stand alone on its own line. The parser
enforces this by distinguishing `IDENT "="` (assignment statement) from `IDENT "("` (call
expression) using one token of lookahead.

Sebesta §7.7.5 discusses the C design where assignment is an expression, and identifies its
primary hazard: `if (x = y)` is silently accepted when the programmer intended `if (x == y)`.
WorkflowScript structurally prevents this class of bug — `=` is never valid in an expression
context, so the mistake is a parse error. Chained assignment (`a = b = 5`) is not supported,
consistent with Fortran and Ada's approach (Sebesta §7.7.1).

---

## 4.8 Design Rationale

The following paragraphs justify the five key design decisions emerging from §4.5 and §4.7.

### Decision 1 — Strong Static Typing

WorkflowScript is strongly typed because reliability is its primary language criterion. A
pipeline configuration that silently accepts the wrong type of value can fail at execution
time, potentially after a long-running job has already started. By detecting all type errors
before execution begins, WorkflowScript guarantees that a program which passes the type checker
will not crash due to a type mismatch. The trade-off accepted is reduced flexibility: users
cannot, for instance, write a function that accepts "any numeric type" without specifying
`int` or `float`. This is acceptable because the domain is narrow — task parameters are
always concrete integers or booleans — and generics would significantly complicate the type
checker without enabling any new workflow patterns.

### Decision 2 — int-to-float Widening Only

Allowing `int` to widen to `float` eliminates friction when computing derived metrics (e.g.,
`t.timeout * 1.5` for a soft deadline). Disallowing `float`-to-`int` narrowing prevents silent
truncation: `retries: 2.7` is meaningless and should be a compile-time error. The trade-off is
that converting a float result back to an integer must be done explicitly, which is not
currently supported in the language. This is intentional: if a programmer needs `int(x / 3.0)`,
they should be forced to think about what truncation means for their pipeline logic. Keeping
coercion one-directional keeps the `typesCompatible` function simple and its behavior obvious.

### Decision 3 — Name Equivalence for `task`

Name equivalence was chosen because `task` is a domain role, not a data structure.
A `build` task and a `test` task carry the same three fields, but they represent fundamentally
different units of work. If structural equivalence were used, any ad-hoc record with
`timeout: int`, `retries: int`, `parallel: bool` fields would be interchangeable with a
declared task — which would allow a programmer to accidentally pass the wrong task to a
pipeline stage. With name equivalence and a single `task` type, all tasks are
interchangeable at the type level (they all have type `task`), but their identities are
preserved at the value level (the runtime stores each task under its declared name). The
trade-off is that the language cannot distinguish `task build` from `task deploy` in a
function signature; both accept any `task`. A richer type system with per-task singleton types
would fix this but is beyond the scope of the project.

### Decision 4 — Short-Circuit Boolean Evaluation

Short-circuit evaluation is the natural choice for a language where the right-hand side of
`&&` or `||` may be an expression with a side effect or a potential runtime error (e.g.,
division by zero in a guard condition). It matches the expectations of programmers coming from
Go, Python, C, and most mainstream languages. The implementation is clean: `evalLogical` is
a separate method from `evalBinary`, called specifically when the operator is `&&` or `||`.
The trade-off is a small asymmetry: the type checker always evaluates both operands for type
correctness, while the interpreter may not evaluate the right operand. This means a type error
in the right operand of `false && <bad_expr>` is caught at compile time even though the
expression would never be evaluated at runtime. This is the correct behavior — type checking
and execution are separate passes.

### Decision 5 — Assignment as Statement, Not Expression

Treating assignment as a statement eliminates an entire class of bugs. The most common is
`if x = 5` where the programmer intended `if x == 5`: in languages where assignment is an
expression, the if-condition is silently the assigned value (`5`, which is truthy), not the
comparison result. WorkflowScript's parser rejects this pattern structurally: `=` is never
parsed inside an expression context. The trade-off is that chained assignment (`a = b = 5`)
is unsupported, but chaining is rarely needed in workflow configuration scripts and its
absence does not reduce expressiveness in the domain.
