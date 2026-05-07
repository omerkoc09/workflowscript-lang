// User-defined functions and task field access
task deploy { timeout: 120 retries: 1 parallel: false }

fun max_time(t: int, r: int) -> int {
    return t * (r + 1)
}

fun is_long(t: task) -> bool {
    return t.timeout > 60
}

var m: int = max_time(120, 1)
var slow: bool = is_long(deploy)

if slow {
    run deploy
}
