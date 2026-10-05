package keys

// Default returns the default key bindings.
func Default() *Keys {
	return &Keys{
		Global: GlobalKeys{
			// Normal mode: plain letters are safe to use.
			Quit: Binding{
				Action:        ActionQuitNormal,
				Keys:          []string{"q"},
				HelpText:      "quit",
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

			BackInputMode: Binding{
				Action:        ActionQuitInput,
				Keys:          []string{"esc"},
				HelpText:      "quit",
				CaseSensitive: true,
			},
		},

		Nav: NavKeys{
			Up: Binding{
				Action:        ActionUp,
				Keys:          []string{"up", "k"},
				HelpText:      "move up",
				CaseSensitive: true,
			},
			Down: Binding{
				Action:        ActionDown,
				Keys:          []string{"down", "j"},
				HelpText:      "move down",
				CaseSensitive: true,
			},
			Left: Binding{
				Action:        ActionLeft,
				Keys:          []string{"left", "h"},
				HelpText:      "move left",
				CaseSensitive: true,
			},
			Right: Binding{
				Action:        ActionRight,
				Keys:          []string{"right", "l"},
				HelpText:      "move right",
				CaseSensitive: true,
			},

			PageUp: Binding{
				Action:   ActionPageUp,
				Keys:     []string{"pgup", "ctrl+u"},
				HelpText: "page up",
				Hidden:   true,
			},
			PageDown: Binding{
				Action:   ActionPageDown,
				Keys:     []string{"pgdown", "ctrl+d"},
				HelpText: "page down",
				Hidden:   true,
			},
			Top: Binding{
				Action:        ActionTop,
				Keys:          []string{"g"},
				HelpText:      "go to top",
				Hidden:        true,
				CaseSensitive: false,
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
		},

		Confirm: ConfirmKeys{
			Accept: Binding{
				Action:        ActionAccept,
				Keys:          []string{"enter", "y"},
				HelpText:      "confirm",
				CaseSensitive: true,
			},
		},

		Notice: NoticeKeys{
			Continue: Binding{
				Action:   ActionAccept,
				Keys:     []string{"enter"},
				HelpText: "continue",
			},
		},

		Dialog: DialogKeys{
			Cancel: Binding{
				Action:    ActionDialogCancel,
				Keys:      []string{"esc"},
				HelpText:  "close dialog",
				ShortText: "close",
			},
		},

		Menu: MenuKeys{
			Select: Binding{
				Action:    ActionMenuSelect,
				Keys:      []string{"enter"},
				HelpText:  "select item",
				ShortText: "select",
			},
		},

		Form: FormKeys{
			Save: Binding{
				Action:   ActionFormSave,
				Keys:     []string{"ctrl+s"},
				HelpText: "save",
			},
		},

		Scan: ScanKeys{
			Pause: Binding{
				Action:        ActionScanPause,
				Keys:          []string{"p", "space"},
				HelpText:      "pause / resume scan",
				ShortText:     "pause",
				CaseSensitive: true,
			},
			Logs: Binding{
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
