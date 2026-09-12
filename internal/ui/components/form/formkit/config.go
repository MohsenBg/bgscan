package formkit

import (
	"strings"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// TunnelConfig is the validation contract shared by every tunnel
// configuration type stored by the dns package.
type TunnelConfig interface {
	Validate() map[string]error
}

// Service is the configuration CRUD surface shared by every tunnel service.
// The dns package services (VayDNSService, DNSTTService, ...) satisfy it
// structurally.
type Service[C TunnelConfig] interface {
	SaveConfig(config C, name string) error
	EditConfig(config C, originalName string) error
	RenameConfig(oldName, newName string) error
}

// Save runs the generic save-or-edit flow for a tunnel config form: it
// validates the name, edits/renames or creates the config through srv, and
// emits the appropriate success/error notice.
func Save[C TunnelConfig](b *Base, label string, srv Service[C], cfg *C) tea.Msg {
	name := strings.TrimSpace(b.name)
	if name == "" {
		return notice.NewNoticeCmd(
			b.Layout(),
			"Error",
			"config name is required",
			notice.NOTICE_ERROR,
		)()
	}

	if b.originalName != "" {
		if err := srv.EditConfig(*cfg, b.originalName); err != nil {
			logger.UIError("Failed to edit %s config: %v", label, err)
			return notice.NewNoticeCmd(
				b.Layout(),
				"Edit Failed",
				err.Error(),
				notice.NOTICE_ERROR,
			)()
		}

		if b.originalName != name {
			if err := srv.RenameConfig(b.originalName, name); err != nil {
				logger.UIError("Failed to rename %s config: %v", label, err)
				return notice.NewNoticeCmd(
					b.Layout(),
					"Rename Failed",
					err.Error(),
					notice.NOTICE_ERROR,
				)()
			}
		}
	} else if err := srv.SaveConfig(*cfg, name); err != nil {
		logger.UIError("Failed to save %s config: %v", label, err)
		return notice.NewNoticeCmd(
			b.Layout(),
			"Save Failed",
			err.Error(),
			notice.NOTICE_ERROR,
		)()
	}

	return tea.Sequence(
		notice.NewNoticeCmd(
			b.Layout(),
			"Saved",
			label+" config saved",
			notice.NOTICE_SUCCESS,
		),
		func() tea.Msg {
			return ui.CloseComponentMsg{ID: b.ID()}
		},
	)()
}
