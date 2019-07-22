# exec [![Build Status](https://travis-ci.org/k1LoW/exec.svg?branch=master)](https://travis-ci.org/k1LoW/exec) [![codecov](https://codecov.io/gh/k1LoW/exec/branch/master/graph/badge.svg)](https://codecov.io/gh/k1LoW/exec)

## Usage

``` golang
import (
    // "os/exec"
    "github.com/k1LoW/exec"
)
```

## Difference between `os/exec` and `k1LoW/exec`

- `k1LoW/exec.Command` returns `*os/exec.Cmd` with PGID set.
- When context cancelled, `k1LoW/exec.CommandContext` send signal to process group.

## References

- [Songmu/timeout](https://github.com/Songmu/timeout): Timeout invocation. Go porting of GNU timeout and able to use as Go package
    - [k1LoW/exec](https://github.com/k1LoW/exec) is porting source code to handle processes from [Songmu/timeout](https://github.com/Songmu/timeout)
