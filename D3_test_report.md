# D3 — Example Programs & Test Report (Part 2)

**CSE341 — Concepts of Programming Languages**
**Gebze Technical University**
**Part 2 Submission — 22 May 2026**
**Language: WorkflowScript**

---

## Section 1: Valid Programs — Execution with Outputs

Each program is run with the command:
```
wsc <program.ws>
```

---

### Program 1 — `examples/valid/01_basic.ws`: Conditional Task Dispatch

**Source:**
```
// Conditional dispatch: choose the cheaper task based on computed cost
task fast {
    timeout:  20
    retries:  0
    parallel: true
}

task thorough {
    timeout:  90
    retries:  3
    parallel: false
}

fun total_cost(t: task) -> int {
    return t.timeout * (t.retries + 1)
}

var fast_cost: int     = total_cost(fast)
var thorough_cost: int = total_cost(thorough)

if fast_cost < thorough_cost {
    run fast
} else {
    run thorough
}
```

**Actual output:**
```
[run] fast
```

**Discussion:**
This program demonstrates user-defined functions that accept a `task` argument and access its fields (`t.timeout`, `t.retries`) to compute a derived integer value. The two tasks, `fast` and `thorough`, are passed to `total_cost`, which returns `20 * 1 = 20` and `90 * 4 = 360` respectively. Because `fast_cost < thorough_cost` is true, the `if` branch executes and `run fast` is triggered, printing `[run] fast`. The program exercises WorkflowScript's structured type (`task`), field access via dot notation, function calls with task parameters, arithmetic expressions, and the `if-else` control structure in a single coherent workflow scenario.

---

### Program 2 — `examples/valid/02_functions.ws`: Function Definitions and Task Field Access

**Source:**
```
// User-defined functions and task field access
task deploy { timeout: 120 retries: 1 parallel: false }

fun max_time(t: int, r: int) -> int {
    return t * (r + 1)
}

fun is_long(t: task) -> bool {
    return t.timeout > 60
}

var m: int   = max_time(120, 1)
var slow: bool = is_long(deploy)

if slow {
    run deploy
}
```

**Actual output:**
```
[run] deploy
```

**Discussion:**
This program showcases two function definitions with different return types (`int` and `bool`) to demonstrate that WorkflowScript's type system correctly tracks return types across function calls. `max_time` receives two `int` arguments and returns their product, computing the maximum possible wall-clock duration for a task with retries. `is_long` accepts a `task` argument and returns a boolean by comparing `t.timeout > 60`. Since `deploy.timeout` is 120, `is_long` returns `true`, the `if` body executes, and `[run] deploy` is printed. The program illustrates how functions that access structured type fields and produce boolean results can drive control flow.

---

### Program 3 — `examples/valid/03_pipeline.ws`: Pipeline Execution with While Loop

**Source:**
```
// Full pipeline with on_failure, functions, while loop
task build  { timeout: 30  retries: 0 parallel: false }
task test   { timeout: 60  retries: 2 parallel: true  }
task deploy { timeout: 120 retries: 1 parallel: false }

pipeline release {
    stage build
    stage test
    stage deploy
    on_failure { run build }
}

fun should_run(limit: int) -> bool {
    return limit > 0
}

var attempt: int = 0
var limit: int   = 3

while attempt < limit {
    attempt = attempt + 1
}

if should_run(limit) {
    run release
}
```

**Actual output:**
```
[run] build
[run] test
[run] deploy
```

**Discussion:**
This program is the most comprehensive of the three: it declares three tasks, a named pipeline with three ordered stages and an `on_failure` handler, a `while` loop that counts from 0 to 3, and a guard function that decides whether the pipeline should run. The `while` loop exits when `attempt == limit == 3`; then `should_run(3)` returns `true`, so `run release` is executed. Running a pipeline sequentially prints each stage name in declaration order — `[run] build`, `[run] test`, `[run] deploy` — which matches the [Run-Pipeline] semantic rule in the design specification. The `on_failure` block and `parallel: true` field on `test` are parsed and type-checked but are not triggered during this simulated execution, demonstrating their presence as first-class language constructs even without a real process scheduler.

---

## Section 2: Type Error Program — Type Checker in Action

### Program — `examples/invalid/06_type_error.ws`: Bool Field Used in Arithmetic

**Source:**
```
// Type error: mixing bool task field in arithmetic expression.
// deploy.parallel is of type bool; the + operator requires numeric operands.
// The type checker catches this before any execution begins.
task deploy { timeout: 120  retries: 1  parallel: false }

var effective_cost: int = deploy.timeout + deploy.parallel
```

**Type error message:**
```
type error: line 6: + requires numeric operands
```

**Discussion:**
This program contains a common domain mistake: a programmer attempts to include the `parallel` flag in a cost formula alongside `timeout`. In WorkflowScript, `deploy.parallel` has type `bool`, while `deploy.timeout` has type `int`. The `+` operator requires both operands to be numeric (`int` or `float`); applying it to a `bool` operand is forbidden by the type rules in §4.5.2 of the design specification. The type checker detects this before any execution begins and halts with a precise error on line 6. No output is produced because interpretation never starts — this demonstrates WorkflowScript's *fail-fast* design principle: type safety is enforced at compile time to prevent silent failures during long-running pipeline jobs.

---

## Section 3: Malformed Programs — Parser Error Messages

Each program below is rejected by the **parser** (before type checking or execution).

---

### Malformed Program 1 — `examples/invalid/01_missing_brace.ws`

**Source:**
```
task build { timeout: 30 retries: 0 parallel: false
```

**Error message:**
```
line 2: expected RBRACE, got ""
```

The closing `}` of the `task` body is missing. The parser reaches end-of-file while still inside the task block and reports the line on which it expected `RBRACE`.

---

### Malformed Program 2 — `examples/invalid/02_bad_field_order.ws`

**Source:**
```
task build { retries: 0 timeout: 30 parallel: false }
```

**Error message:**
```
line 1: expected KW_TIMEOUT, got "retries"
```

WorkflowScript requires task fields in the fixed order `timeout`, `retries`, `parallel`. Swapping `retries` and `timeout` violates the grammar; the parser expects `timeout` first and rejects the program immediately.

---

### Malformed Program 3 — `examples/invalid/03_empty_pipeline.ws`

**Source:**
```
pipeline ci { }
```

**Error message:**
```
line 1: pipeline must have at least one stage
```

A pipeline with no `stage` declarations is semantically empty and violates the parser's structural constraint. WorkflowScript enforces this at parse time so that a `run ci` statement cannot refer to a pipeline that would do nothing.

---

### Malformed Program 4 — `examples/invalid/04_missing_arrow.ws`

**Source:**
```
fun double(x: int) int { return x }
```

**Error message:**
```
line 1: expected ARROW, got "int"
```

The return-type annotation requires the `->` token between the parameter list and the return type. Omitting `->` causes the parser to see `int` where it expects `->` and rejects the function declaration.

---

### Malformed Program 5 — `examples/invalid/05_bad_parallel.ws`

**Source:**
```
task build { timeout: 30 retries: 0 parallel: 1 }
```

**Error message:**
```
line 1: expected BOOL_LIT, got "1"
```

The `parallel` field must be a boolean literal (`true` or `false`). Using the integer `1` is rejected by the parser, which enforces the field's type at the syntactic level. This reflects WorkflowScript's design principle of making incorrect task configurations a parse error rather than a silent type coercion.

---

*End of D3 — Example Programs & Test Report*
