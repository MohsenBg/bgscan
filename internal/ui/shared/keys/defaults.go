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

		Table: table,
		Crud: CrudKeys{
			Select: Binding{
				Action:    ActionCrudSelect,
				Keys:      []string{"enter"},
				HelpText:  "select item",
				ShortText: "select",
			},
			Add: Binding{
				Action:    ActionCrudAdd,
				Keys:      []string{"a"},
				HelpText:  "create item",
				ShortText: "add",
			},
			Delete: Binding{
				Action:    ActionCrudDelete,
				Keys:      []string{"x"},
				HelpText:  "delete item",
				ShortText: "delete",
			},
			Rename: Binding{
				Action:    ActionCrudRename,
				Keys:      []string{"r"},
				HelpText:  "rename item",
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
