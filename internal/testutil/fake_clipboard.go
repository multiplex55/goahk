package testutil

import "context"

type FakeClipboardService struct {
	Text string
	Err  error
}

func (f *FakeClipboardService) ReadText(context.Context) (string, error) {
	if f.Err != nil {
		return "", f.Err
	}
	return f.Text, nil
}

func (f *FakeClipboardService) WriteText(_ context.Context, text string) error {
	if f.Err != nil {
		return f.Err
	}
	f.Text = text
	return nil
}

func (f *FakeClipboardService) AppendText(ctx context.Context, suffix string) error {
	text, err := f.ReadText(ctx)
	if err != nil {
		return err
	}
	return f.WriteText(ctx, text+suffix)
}

func (f *FakeClipboardService) PrependText(ctx context.Context, prefix string) error {
	text, err := f.ReadText(ctx)
	if err != nil {
		return err
	}
	return f.WriteText(ctx, prefix+text)
}
