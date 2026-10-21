package inspector

import "github.com/MohsenBg/bgscan/internal/ui/components/basic/input"

// fieldInputAdapter widens input.Input[T] to the FieldInput interface.
type fieldInputAdapter[T any] struct {
	input.Input[T]
}

func (a fieldInputAdapter[T]) Value() any { return a.Input.Value() }

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
