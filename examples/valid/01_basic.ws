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
