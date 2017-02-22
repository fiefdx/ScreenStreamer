package flv

import (
    "fmt"
)

type Error interface {
    Error() string
    IsRecoverable() bool
}

type ReadError struct {
    incomplete *CFrame
    position int64
}

func (e *ReadError) Error() string {
    if e.incomplete == nil {
        return fmt.Sprintf("Invalid tag type @%d", e.position)
    } else {
        return fmt.Sprintf("Incomplete frame[dts=%d,stream=%d]@%d", e.incomplete.Dts, e.incomplete.Stream, e.position)
    }
}

func (e *ReadError) IsRecoverable() bool {
    return true
}

func IncompleteFrameError(incomplete *CFrame) Error {
    return &ReadError{incomplete, incomplete.Position+11}
}

func InvalidTagStart(position int64) Error {
    return &ReadError{nil, position}
}

type UnrecoverableError struct {
    message string
    position int64
}

func (e *UnrecoverableError) Error() string {
    return fmt.Sprintf("unrecoverable@%d: %s", e.position, e.message)
}

func (*UnrecoverableError) IsRecoverable () bool {
    return false
}

func Unrecoverable(message string, position int64) Error {
    return &UnrecoverableError{message, position}
}