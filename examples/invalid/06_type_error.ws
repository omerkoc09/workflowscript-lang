// Type error: mixing bool task field in arithmetic expression.
// deploy.parallel is of type bool; the + operator requires numeric operands.
// The type checker catches this before any execution begins.
task deploy { timeout: 120  retries: 1  parallel: false }

var effective_cost: int = deploy.timeout + deploy.parallel
