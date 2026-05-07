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
var limit: int = 3

while attempt < limit {
    attempt = attempt + 1
}

if should_run(limit) {
    run release
}
