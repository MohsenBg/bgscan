package inspector

import "bgscan/internal/ui/components/basic/input"

// fieldInputAdapter adapts a typed input.Input[T] to FiledInput by
// widening Value() T / SetValue(T) to Value() any / SetValue(any).
// All other methods (ID, Name, Init, Mode, OnClose, CloseCmd, Snapshot,
// AppendOnSubmit) are promoted unchanged from the embedded input.Input[T].
type fieldInputAdapter[T any] struct {
	input.Input[T]
}

func (a fieldInputAdapter[T]) Value() any { return a.Input.Value() }

// SetValue implements FiledInput, shadowing the embedded SetValue(T).
// It panics if v is not assertable to T, which should never happen in
// practice since v always originates from a call to Value() on the same
// underlying input.
func (a fieldInputAdapter[T]) SetValue(v any) {
	tv, ok := v.(T)
	if !ok {
		return
	}
	a.Input.SetValue(tv)
}

// Adapt wraps a typed input.Input[T] so it can be used as Field.Input.
func Adapt[T any](in input.Input[T]) FiledInput {
	return fieldInputAdapter[T]{in}
}
