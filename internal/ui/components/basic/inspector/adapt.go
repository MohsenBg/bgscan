package inspector

import "github.com/MohsenBg/bgscan/internal/ui/components/basic/input"

// fieldInputAdapter adapts a typed input.Input[T] to FieldInput by
// widening Value() T / SetValue(T) to Value() any / SetValue(any).
// All other methods (ID, Name, Init, Mode, OnClose, CloseCmd, Snapshot,
// AppendOnSubmit) are promoted unchanged from the embedded input.Input[T].
type fieldInputAdapter[T any] struct {
	input.Input[T]
}

func (a fieldInputAdapter[T]) Value() any { return a.Input.Value() }

// SetValue implements FieldInput, shadowing the embedded SetValue(T).
// A non-T value is ignored; in practice values always come from Value()
// on the same underlying input.
func (a fieldInputAdapter[T]) SetValue(v any) {
	tv, ok := v.(T)
	if !ok {
		return
	}
	a.Input.SetValue(tv)
}

// Adapt wraps a typed input.Input[T] so it can be used as Field.Input.
func Adapt[T any](in input.Input[T]) FieldInput {
	return fieldInputAdapter[T]{in}
}
