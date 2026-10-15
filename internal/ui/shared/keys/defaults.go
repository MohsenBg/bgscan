package keys

// OrDefault returns k when it carries bindings, and the defaults when it does
// not.
//
// A Keys value assembled by hand — as test Deps and one-off constructions do
// — leaves every binding empty, and a component built from one responds to no
// key at all. That failure is silent, so components normalise through here.
func OrDefault(k Keys) Keys {
	if len(k.Nav.Up.Keys) > 0 {
		return k
	}

	return Default()
}

// Default returns the default key bindings.
func Default() Keys {
	global := GlobalKeys{
		Quit: Binding{
			Action:        ActionQuit,
			Keys:          []string{"q"},
			HelpText:      "quit",
			CaseSensitive: false,
		},

		ForceQuit: Binding{
			Action:        ActionForceQuit,
			Keys:          []string{"ctrl+c"},
			HelpText:      "quit",
			CaseSensitive: true,
		},

		DumpGoroutines: Binding{
			Action:        ActionDumpGoroutines,
			Keys:          []string{"ctrl+t"},
			HelpText:      "dump Goroutines",
			CaseSensitive: true,
		},

		Help: Binding{
			Action:    ActionHelp,
			Keys:      []string{"?"},
			HelpText:  "show all keys",
			ShortText: "more",
		},
		Back: Binding{
			Action:    ActionBack,
			Keys:      []string{"esc", "b"},
			HelpText:  "go back",
			ShortText: "back",
		},
		BackFromInput: Binding{
			Action:        ActionBackFromInput,
			Keys:          []string{"esc"},
			HelpText:      "back",
			CaseSensitive: false,
		},
	}

	nav := NavKeys{
		Up: Binding{
			Action:        ActionUp,
			Keys:          []string{"up", "k"},
			HelpText:      "move up",
			CaseSensitive: false,
		},
		Down: Binding{
			Action:        ActionDown,
			Keys:          []string{"down", "j"},
			HelpText:      "move down",
			CaseSensitive: false,
		},
		Left: Binding{
			Action:        ActionLeft,
			Keys:          []string{"left", "h"},
			HelpText:      "move left",
			CaseSensitive: false,
		},
		Right: Binding{
			Action:        ActionRight,
			Keys:          []string{"right", "l"},
			HelpText:      "move right",
			CaseSensitive: false,
		},
		PageUp: Binding{
			Action:        ActionPageUp,
			Keys:          []string{"pgup", "ctrl+u"},
			HelpText:      "page up",
			Hidden:        true,
			CaseSensitive: false,
		},
		PageDown: Binding{
			Action:        ActionPageDown,
			Keys:          []string{"pgdown", "ctrl+d"},
			HelpText:      "page down",
			Hidden:        true,
			CaseSensitive: false,
		},
		Top: Binding{
			Action:        ActionTop,
			Keys:          []string{"g"},
			HelpText:      "go to top",
			Hidden:        true,
			CaseSensitive: true,
		},
		Bottom: Binding{
			Action:        ActionBottom,
			Keys:          []string{"G"},
			HelpText:      "go to bottom",
			Hidden:        true,
			CaseSensitive: true,
		},
		NextTab: Binding{
			Action:   ActionNextTab,
			Keys:     []string{"tab"},
			HelpText: "next tab",
		},
		PrevTab: Binding{
			Action:   ActionPrevTab,
			Keys:     []string{"shift+tab"},
			HelpText: "previous tab",
		},
	}

	table := TableKeys{
		Up:       nav.Up,
		Down:     nav.Down,
		PageUp:   nav.PageUp,
		PageDown: nav.PageDown,
		Top:      nav.Top,
		Bottom:   nav.Bottom,

		HalfPageUp: Binding{
			Action:   ActionTableHalfPageUp,
			Keys:     []string{"u"},
			HelpText: "½ page up",
		},
		HalfPageDown: Binding{
			Action:   ActionTableHalfPageDown,
			Keys:     []string{"d"},
			HelpText: "½ page down",
		},
		ChangeSort: Binding{
			Action:   ActionTableSort,
			Keys:     []string{"s"},
			HelpText: "sort",
		},
		ToggleOrder: Binding{
			Action:   ActionTableOrder,
			Keys:     []string{"o"},
			HelpText: "order",
		},
		Filter: Binding{
			Action:   ActionTableFilter,
			Keys:     []string{"/"},
			HelpText: "filter",
		},
	}

	textInput := TextInputKeys{
		Submit: Binding{
			Action:   ActionTextInputSubmit,
			Keys:     []string{"enter"},
			HelpText: "submit",
		},
		CharacterForward: Binding{
			Action:   ActionTextInputCharacterForward,
			Keys:     []string{"right", "ctrl+f"},
			HelpText: "cursor right",
		},
		CharacterBackward: Binding{
			Action:   ActionTextInputCharacterBackward,
			Keys:     []string{"left", "ctrl+b"},
			HelpText: "cursor left",
		},
		WordForward: Binding{
			Action:   ActionTextInputWordForward,
			Keys:     []string{"alt+right", "ctrl+right", "alt+f"},
			HelpText: "word right",
		},
		WordBackward: Binding{
			Action:   ActionTextInputWordBackward,
			Keys:     []string{"alt+left", "ctrl+left", "alt+b"},
			HelpText: "word left",
		},
		DeleteWordBackward: Binding{
			Action:   ActionTextInputDeleteWordBackward,
			Keys:     []string{"alt+backspace", "ctrl+w", "ctrl+backspace"},
			HelpText: "delete word back",
		},
		DeleteWordForward: Binding{
			Action:   ActionTextInputDeleteWordForward,
			Keys:     []string{"alt+delete", "alt+d", "ctrl+delete"},
			HelpText: "delete word fwd",
		},
		DeleteAfterCursor: Binding{
			Action:   ActionTextInputDeleteAfterCursor,
			Keys:     []string{"ctrl+k"},
			HelpText: "delete to end",
		},
		DeleteBeforeCursor: Binding{
			Action:   ActionTextInputDeleteBeforeCursor,
			Keys:     []string{"ctrl+u"},
			HelpText: "delete to start",
		},
		DeleteCharacterBackward: Binding{
			Action:   ActionTextInputDeleteCharacterBackward,
			Keys:     []string{"backspace", "ctrl+h"},
			HelpText: "backspace",
		},
		DeleteCharacterForward: Binding{
			Action:   ActionTextInputDeleteCharacterForward,
			Keys:     []string{"delete", "ctrl+d"},
			HelpText: "delete",
		},
		LineStart: Binding{
			Action:   ActionTextInputLineStart,
			Keys:     []string{"home", "ctrl+a"},
			HelpText: "line start",
		},
		LineEnd: Binding{
			Action:   ActionTextInputLineEnd,
			Keys:     []string{"end", "ctrl+e"},
			HelpText: "line end",
		},
		Paste: Binding{
			Action:   ActionTextInputPaste,
			Keys:     []string{"ctrl+v"},
			HelpText: "paste",
		},
	}

	textArea := TextAreaKeys{
		Submit: Binding{
			Action:   ActionTextAreaSubmit,
			Keys:     []string{"enter"},
			HelpText: "submit",
		},
		NewLine: Binding{
			Action:   ActionTextAreaNewLine,
			Keys:     []string{"shift+enter", "ctrl+m"},
			HelpText: "newline",
		},
		CharacterForward: Binding{
			Action:   ActionTextAreaCharacterForward,
			Keys:     []string{"right", "ctrl+f"},
			HelpText: "cursor right",
		},
		CharacterBackward: Binding{
			Action:   ActionTextAreaCharacterBackward,
			Keys:     []string{"left", "ctrl+b"},
			HelpText: "cursor left",
		},
		WordForward: Binding{
			Action:   ActionTextAreaWordForward,
			Keys:     []string{"alt+right", "ctrl+right", "alt+f"},
			HelpText: "word right",
		},
		WordBackward: Binding{
			Action:   ActionTextAreaWordBackward,
			Keys:     []string{"alt+left", "ctrl+left", "alt+b"},
			HelpText: "word left",
		},
		LineNext: Binding{
			Action:   ActionTextAreaLineNext,
			Keys:     []string{"down", "ctrl+n"},
			HelpText: "line down",
		},
		LinePrevious: Binding{
			Action:   ActionTextAreaLinePrevious,
			Keys:     []string{"up", "ctrl+p"},
			HelpText: "line up",
		},
		DeleteWordBackward: Binding{
			Action:   ActionTextAreaDeleteWordBackward,
			Keys:     []string{"alt+backspace", "ctrl+w", "ctrl+backspace"},
			HelpText: "delete word back",
		},
		DeleteWordForward: Binding{
			Action:   ActionTextAreaDeleteWordForward,
			Keys:     []string{"alt+delete", "alt+d", "ctrl+delete"},
			HelpText: "delete word fwd",
		},
		DeleteAfterCursor: Binding{
			Action:   ActionTextAreaDeleteAfterCursor,
			Keys:     []string{"ctrl+k"},
			HelpText: "delete to end",
		},
		DeleteBeforeCursor: Binding{
			Action:   ActionTextAreaDeleteBeforeCursor,
			Keys:     []string{"ctrl+u"},
			HelpText: "delete to start",
		},
		InsertNewline: Binding{
			Action:   ActionTextAreaInsertNewline,
			Keys:     []string{"shift+enter", "ctrl+m"},
			HelpText: "newline",
		},
		DeleteCharacterBackward: Binding{
			Action:   ActionTextAreaDeleteCharacterBackward,
			Keys:     []string{"backspace", "ctrl+h"},
			HelpText: "backspace",
		},
		DeleteCharacterForward: Binding{
			Action:   ActionTextAreaDeleteCharacterForward,
			Keys:     []string{"delete", "ctrl+d"},
			HelpText: "delete",
		},
		LineStart: Binding{
			Action:   ActionTextAreaLineStart,
			Keys:     []string{"home", "ctrl+a"},
			HelpText: "line start",
		},
		LineEnd: Binding{
			Action:   ActionTextAreaLineEnd,
			Keys:     []string{"end", "ctrl+e"},
			HelpText: "line end",
		},
		PageUp: Binding{
			Action:   ActionTextAreaPageUp,
			Keys:     []string{"pgup"},
			HelpText: "page up",
		},
		PageDown: Binding{
			Action:   ActionTextAreaPageDown,
			Keys:     []string{"pgdown"},
			HelpText: "page down",
		},
		Paste: Binding{
			Action:   ActionTextAreaPaste,
			Keys:     []string{"ctrl+v"},
			HelpText: "paste",
		},
		InputBegin: Binding{
			Action:   ActionTextAreaInputBegin,
			Keys:     []string{"alt+<", "ctrl+home"},
			HelpText: "input start",
		},
		InputEnd: Binding{
			Action:   ActionTextAreaInputEnd,
			Keys:     []string{"alt+>", "ctrl+end"},
			HelpText: "input end",
		},
		CapitalizeWordForward: Binding{
			Action:   ActionTextAreaCapitalizeWordForward,
			Keys:     []string{"alt+c"},
			HelpText: "capitalize",
		},
		LowercaseWordForward: Binding{
			Action:   ActionTextAreaLowercaseWordForward,
			Keys:     []string{"alt+l"},
			HelpText: "lowercase",
		},
		UppercaseWordForward: Binding{
			Action:   ActionTextAreaUppercaseWordForward,
			Keys:     []string{"alt+u"},
			HelpText: "uppercase",
		},
		TransposeCharacterBackward: Binding{
			Action:   ActionTextAreaTransposeCharacterBackward,
			Keys:     []string{"ctrl+t"},
			HelpText: "transpose",
		},
		SelectCharacterForward: Binding{
			Action:   ActionTextAreaSelectCharacterForward,
			Keys:     []string{"shift+right"},
			HelpText: "select right",
		},
		SelectCharacterBackward: Binding{
			Action:   ActionTextAreaSelectCharacterBackward,
			Keys:     []string{"shift+left"},
			HelpText: "select left",
		},
		SelectWordForward: Binding{
			Action:   ActionTextAreaSelectWordForward,
			Keys:     []string{"ctrl+shift+right", "alt+shift+right", "alt+shift+f"},
			HelpText: "select word right",
		},
		SelectWordBackward: Binding{
			Action:   ActionTextAreaSelectWordBackward,
			Keys:     []string{"ctrl+shift+left", "alt+shift+left", "alt+shift+b"},
			HelpText: "select word left",
		},
		SelectLineUp: Binding{
			Action:   ActionTextAreaSelectLineUp,
			Keys:     []string{"shift+up"},
			HelpText: "select line up",
		},
		SelectLineDown: Binding{
			Action:   ActionTextAreaSelectLineDown,
			Keys:     []string{"shift+down"},
			HelpText: "select line down",
		},
		SelectAll: Binding{
			Action:   ActionTextAreaSelectAll,
			Keys:     []string{"ctrl+g"},
			HelpText: "select all",
		},
		CopySelection: Binding{
			Action:   ActionTextAreaCopySelection,
			Keys:     []string{"ctrl+shift+c"},
			HelpText: "copy",
		},
	}

	selectInput := SelectInputKeys{
		Submit: Binding{
			Action:   ActionSelectInputSubmit,
			Keys:     []string{"enter"},
			HelpText: "submit",
		},
		Prev: Binding{
			Action:   ActionSelectInputPrev,
			Keys:     []string{"shift+tab"},
			HelpText: "back",
		},
		Next: Binding{
			Action:   ActionSelectInputNext,
			Keys:     []string{"tab"},
			HelpText: "select",
		},
		Up: Binding{
			Action:   ActionSelectInputUp,
			Keys:     []string{"up", "k", "ctrl+k", "ctrl+p"},
			HelpText: "up",
		},
		Down: Binding{
			Action:   ActionSelectInputDown,
			Keys:     []string{"down", "j", "ctrl+j", "ctrl+n"},
			HelpText: "down",
		},
		HalfPageUp: Binding{
			Action:   ActionSelectInputHalfPageUp,
			Keys:     []string{"ctrl+u"},
			HelpText: "½ up",
		},
		HalfPageDown: Binding{
			Action:   ActionSelectInputHalfPageDown,
			Keys:     []string{"ctrl+d"},
			HelpText: "½ down",
		},
		GotoTop: Binding{
			Action:   ActionSelectInputGotoTop,
			Keys:     []string{"home", "g"},
			HelpText: "top",
		},
		GotoBottom: Binding{
			Action:   ActionSelectInputGotoBottom,
			Keys:     []string{"end", "G"},
			HelpText: "bottom",
		},
		Left: Binding{
			Action:   ActionSelectInputLeft,
			Keys:     []string{"h", "left"},
			HelpText: "left",
		},
		Right: Binding{
			Action:   ActionSelectInputRight,
			Keys:     []string{"l", "right"},
			HelpText: "right",
		},
		Filter: Binding{
			Action:   ActionSelectInputFilter,
			Keys:     []string{"/"},
			HelpText: "filter",
		},
		SetFilter: Binding{
			Action:   ActionSelectInputSetFilter,
			Keys:     []string{"esc"},
			HelpText: "set filter",
		},
		ClearFilter: Binding{
			Action:   ActionSelectInputClearFilter,
			Keys:     []string{"esc"},
			HelpText: "clear filter",
		},
	}

	multiSelectInput := MultiSelectInputKeys{
		Submit: Binding{
			Action:   ActionMultiSelectSubmit,
			Keys:     []string{"enter"},
			HelpText: "submit",
		},
		Prev: Binding{
			Action:   ActionMultiSelectPrev,
			Keys:     []string{"shift+tab"},
			HelpText: "back",
		},
		Next: Binding{
			Action:   ActionMultiSelectNext,
			Keys:     []string{"enter", "tab"},
			HelpText: "confirm",
		},
		Toggle: Binding{
			Action:   ActionMultiSelectToggle,
			Keys:     []string{"space", "x"},
			HelpText: "toggle",
		},
		Up: Binding{
			Action:   ActionMultiSelectUp,
			Keys:     []string{"up", "k", "ctrl+p"},
			HelpText: "up",
		},
		Down: Binding{
			Action:   ActionMultiSelectDown,
			Keys:     []string{"down", "j", "ctrl+n"},
			HelpText: "down",
		},
		HalfPageUp: Binding{
			Action:   ActionMultiSelectHalfPageUp,
			Keys:     []string{"ctrl+u"},
			HelpText: "½ up",
		},
		HalfPageDown: Binding{
			Action:   ActionMultiSelectHalfPageDown,
			Keys:     []string{"ctrl+d"},
			HelpText: "½ down",
		},
		GotoTop: Binding{
			Action:   ActionMultiSelectGotoTop,
			Keys:     []string{"home", "g"},
			HelpText: "top",
		},
		GotoBottom: Binding{
			Action:   ActionMultiSelectGotoBottom,
			Keys:     []string{"end", "G"},
			HelpText: "bottom",
		},
		Filter: Binding{
			Action:   ActionMultiSelectFilter,
			Keys:     []string{"/"},
			HelpText: "filter",
		},
		SetFilter: Binding{
			Action:   ActionMultiSelectSetFilter,
			Keys:     []string{"enter", "esc"},
			HelpText: "set filter",
		},
		ClearFilter: Binding{
			Action:   ActionMultiSelectClearFilter,
			Keys:     []string{"esc"},
			HelpText: "clear filter",
		},
		SelectAll: Binding{
			Action:   ActionMultiSelectSelectAll,
			Keys:     []string{"ctrl+a"},
			HelpText: "all",
		},
		SelectNone: Binding{
			Action:   ActionMultiSelectSelectNone,
			Keys:     []string{"ctrl+a"},
			HelpText: "none",
		},
	}

	toggleInput := ToggleInputKeys{
		Prev: Binding{
			Action:   ActionToggleInputPrev,
			Keys:     []string{"shift+tab"},
			HelpText: "back",
		},
		Next: Binding{
			Action:   ActionToggleInputNext,
			Keys:     []string{"enter", "tab"},
			HelpText: "next",
		},
		Submit: Binding{
			Action:   ActionToggleInputSubmit,
			Keys:     []string{"enter"},
			HelpText: "submit",
		},
		Toggle: Binding{
			Action:   ActionToggleInputToggle,
			Keys:     []string{"right", "left", "h", "l"},
			HelpText: "toggle",
		},
		Accept: Binding{
			Action:   ActionToggleInputAccept,
			Keys:     []string{"y", "Y"},
			HelpText: "Yes",
		},
		Reject: Binding{
			Action:   ActionToggleInputReject,
			Keys:     []string{"n", "N"},
			HelpText: "No",
		},
	}

	inspector := InspectorKeys{
		Up:      nav.Up,
		Down:    nav.Down,
		NextTab: nav.NextTab,
		PrevTab: nav.PrevTab,
		Edit: Binding{
			Action:    ActionInspectorEdit,
			Keys:      []string{"enter"},
			HelpText:  "edit selected field",
			ShortText: "edit",
		},
		Description: Binding{
			Action:    ActionInspectorDescription,
			Keys:      []string{"d"},
			HelpText:  "show description",
			ShortText: "description",
		},
	}

	return Keys{
		Global: global,
		Nav:    nav,

		Confirm: ConfirmKeys{
			Accept: Binding{
				Action:        ActionAccept,
				Keys:          []string{"enter", "y"},
				HelpText:      "confirm",
				CaseSensitive: false,
			},
		},

		Notice: NoticeKeys{
			Continue: Binding{
				Action:   ActionAccept,
				Keys:     []string{"enter"},
				HelpText: "continue",
			},
		},

		Menu: newMenuKeys(nav, Binding{
			Action:    ActionMenuSelect,
			Keys:      []string{"enter"},
			HelpText:  "select item",
			ShortText: "select",
		}),

		Dialog: newDialogKeys(nav, Binding{
			Action:    ActionDialogCancel,
			Keys:      []string{"esc"},
			HelpText:  "close dialog",
			ShortText: "close",
		}),

		Form: newFormKeys(nav, Binding{
			Action:        ActionFormSave,
			Keys:          []string{"s"},
			HelpText:      "save",
			CaseSensitive: false,
		}),

		Inspector: inspector,

		Table: table,
		Crud: CrudKeys{
			Select: Binding{
				Action:    ActionCrudSelect,
				Keys:      []string{"enter"},
				HelpText:  "select",
				ShortText: "select",
			},
			Add: Binding{
				Action:    ActionCrudAdd,
				Keys:      []string{"a"},
				HelpText:  "create",
				ShortText: "add",
			},
			Delete: Binding{
				Action:    ActionCrudDelete,
				Keys:      []string{"x"},
				HelpText:  "delete",
				ShortText: "delete",
			},
			Rename: Binding{
				Action:    ActionCrudRename,
				Keys:      []string{"r"},
				HelpText:  "rename",
				ShortText: "rename",
			},
		},
		Filter: FilterKeys{
			Accept: Binding{
				Action:   ActionAccept,
				Keys:     []string{"enter"},
				HelpText: "apply filter",
			},
			Clear: Binding{
				Action:   ActionBack,
				Keys:     []string{"esc"},
				HelpText: "clear filter",
			},
			Backspace: Binding{
				Keys:     []string{"backspace"},
				HelpText: "delete character",
				Hidden:   true,
			},
		},
		Scan: ScanKeys{
			TogglePause: Binding{
				Action:        ActionScanPause,
				Keys:          []string{"p", "space"},
				HelpText:      "pause / resume scan",
				ShortText:     "pause",
				CaseSensitive: true,
			},
			ToggleLogs: Binding{
				Action:        ActionScanLogs,
				Keys:          []string{"l"},
				HelpText:      "show logs",
				ShortText:     "logs",
				CaseSensitive: true,
			},
			Cancel: Binding{
				Action:        ActionScanCancel,
				Keys:          []string{"esc", "b", "q"},
				HelpText:      "cancel scan",
				ShortText:     "cancel",
				CaseSensitive: true,
			},
		},

		TextInput:        textInput,
		TextArea:         textArea,
		SelectInput:      selectInput,
		MultiSelectInput: multiSelectInput,
		ToggleInput:      toggleInput,
	}
}

func newMenuKeys(nav NavKeys, sel Binding) MenuKeys {
	return MenuKeys{
		Up: nav.Up, Down: nav.Down,
		Select: sel,
	}
}

func newDialogKeys(nav NavKeys, cancel Binding) DialogKeys {
	return DialogKeys{
		Left: nav.Left, Right: nav.Right,
		Cancel: cancel,
	}
}

func newFormKeys(nav NavKeys, save Binding) FormKeys {
	return FormKeys{
		Up: nav.Up, Down: nav.Down,
		Save: save,
	}
}
