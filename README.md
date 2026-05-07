# WorkflowScript

A typed task/workflow orchestration DSL. 

## Build

    go build -o wsc ./cmd/wsc/

## Run

    ./wsc <file.ws>              # parse and report OK or error
    ./wsc --dump-ast <file.ws>   # print AST

## Examples

    ./wsc --dump-ast examples/valid/03_pipeline.ws
    ./wsc examples/invalid/01_missing_brace.ws
