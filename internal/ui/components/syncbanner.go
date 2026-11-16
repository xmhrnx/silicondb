package components

import (
	"fmt"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/store"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// SyncBanner reports a running or failed sync above content that is
// already showing. It renders nothing otherwise.
func SyncBanner(st store.SyncStatus) giode.Element {
	switch st.Phase {
	case store.Syncing:
		label := tr.T("sync.checking")
		if st.Total > 0 {
			label = tr.T("sync.updating", "done", fmt.Sprint(st.Done), "total", fmt.Sprint(st.Total))
		}
		children := []giode.Element{
			giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
				giode.Spinner(giode.Styles{Width: 16, Height: 16, Color: theme.Primary}),
				giode.Text(label, theme.Caption().With(giode.Styles{Color: theme.Text})),
			),
		}
		if st.Total > 0 {
			children = append(children, giode.Progress(float32(st.Done)/float32(st.Total),
				giode.Styles{Height: 4, Color: theme.Primary, Background: theme.Faint()}))
		}
		return Card(giode.Styles{Gap: 10, Padding: giode.SymmetricInset(14, 12), BorderRadius: 14}, children...)
	case store.Failed:
		return Card(giode.Styles{Padding: giode.SymmetricInset(14, 12), BorderRadius: 14, BorderColor: theme.Tint(theme.Danger, 0.4)},
			giode.Row(giode.Styles{Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
				theme.Icon("alert", 18, theme.Danger),
				giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
					giode.Text(tr.T("sync.failed"), theme.Strong().With(giode.Styles{FontSize: 14})),
					giode.Text(tr.T("sync.offline_copy", "error", st.Err), theme.Caption().With(giode.Styles{MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis})),
				),
			))
	}
	return nil
}
