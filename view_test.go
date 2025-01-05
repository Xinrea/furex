package furex

import (
	"image"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddChildUpdateRemove(t *testing.T) {
	view := &View{
		Attrs: ViewAttrs{
			Width:      100,
			Height:     100,
			Direction:  Row,
			Justify:    JustifyStart,
			AlignItems: AlignItemStart,
		},
	}
	mock := NewMockHandler()
	child := &View{
		Attrs: ViewAttrs{
			Width:  10,
			Height: 10,
		},
		Handler: mock.ViewHandler,
	}
	require.Equal(t, view, view.AddChild(child))
	require.True(t, view.isDirty)

	view.Update()
	require.True(t, mock.IsUpdated)

	view.Draw(nil)
	require.Equal(t, image.Rect(0, 0, 10, 10), mock.Frame)

	require.True(t, view.RemoveChild(child))
	require.Equal(t, 0, len(view.children))
}

func TestUpdateWithSize(t *testing.T) {
	view := &View{
		Attrs: ViewAttrs{
			Width:      100,
			Height:     100,
			Direction:  Row,
			Justify:    JustifyCenter,
			AlignItems: AlignItemCenter,
		},
	}

	mock := NewMockHandler()
	child := &View{
		Attrs: ViewAttrs{
			Width:  10,
			Height: 10,
		},
		Handler: mock.ViewHandler,
	}
	require.Equal(t, view, view.AddChild(child))

	view.UpdateWithSize(200, 200)
	require.True(t, mock.IsUpdated)

	view.Draw(nil)
	require.Equal(t, image.Rect(95, 95, 105, 105), mock.Frame)

}

func TestAddToParent(t *testing.T) {
	root := &View{
		Attrs: ViewAttrs{
			Width:      100,
			Height:     100,
			Direction:  Row,
			Justify:    JustifyStart,
			AlignItems: AlignItemStart,
		},
	}

	mock := NewMockHandler()

	child := (&View{
		Attrs: ViewAttrs{
			Width:  10,
			Height: 10,
		},
		Handler: mock.ViewHandler,
	})

	require.Equal(t, child, child.AddTo(root))

	root.Update()
	require.True(t, mock.IsUpdated)

}

func TestAddChild(t *testing.T) {
	view := &View{
		Attrs: ViewAttrs{
			Width:      100,
			Height:     100,
			Direction:  Row,
			Justify:    JustifyStart,
			AlignItems: AlignItemStart,
		},
	}

	mocks := [2]*mockHandler{NewMockHandler(), NewMockHandler()}
	require.Equal(t, view, view.AddChild(
		&View{
			Attrs: ViewAttrs{
				Width:  10,
				Height: 10,
			},
			Handler: mocks[0].ViewHandler,
		},
		&View{
			Attrs: ViewAttrs{
				Width:  10,
				Height: 10,
			},
			Handler: mocks[1].ViewHandler,
		},
	))

	view.Update()
	require.True(t, mocks[0].IsUpdated)
	require.True(t, mocks[1].IsUpdated)

	view.Draw(nil)
	require.Equal(t, image.Rect(0, 0, 10, 10), mocks[0].Frame)
	require.Equal(t, image.Rect(10, 0, 20, 10), mocks[1].Frame)

	view.RemoveAll()
	require.Equal(t, 0, len(view.children))
}

type CountingHandler struct {
	Times int
	ViewHandler
}

func newCountingHandler() *CountingHandler {
	h := &CountingHandler{}
	h.ViewHandler.Update = h.Update
	return h
}

func (t *CountingHandler) Update(v *View) {
	t.Times++
}

func TestUpdateOnlyOnce(t *testing.T) {
	rootHandler := newCountingHandler()
	nestedHandler := newCountingHandler()

	// given
	rootView := &View{
		Handler: rootHandler.ViewHandler,
	}
	nestedView := &View{
		Handler: nestedHandler.ViewHandler,
	}
	rootView.addChild(nestedView)

	// when
	rootView.Update()

	// then
	require.True(t, rootHandler.Times == 1)
	require.True(t, nestedHandler.Times == 1)
}
