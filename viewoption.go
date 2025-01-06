package furex

type ViewOption func(v *View)

func Height(h int) ViewOption {
	return func(v *View) {
		v.Attrs.Height = h
	}
}

func Width(w int) ViewOption {
	return func(v *View) {
		v.Attrs.Width = w
	}
}

func Left(l int) ViewOption {
	return func(v *View) {
		v.Attrs.Left = l
	}
}

func Top(t int) ViewOption {
	return func(v *View) {
		v.Attrs.Top = t
	}
}

func Right(r int) ViewOption {
	return func(v *View) {
		v.Attrs.Right = &r
	}
}

func Bottom(b int) ViewOption {
	return func(v *View) {
		v.Attrs.Bottom = &b
	}
}

func MarginLeft(ml int) ViewOption {
	return func(v *View) {
		v.Attrs.MarginLeft = ml
	}
}

func MarginTop(mt int) ViewOption {
	return func(v *View) {
		v.Attrs.MarginTop = mt
	}
}

func MarginRight(mr int) ViewOption {
	return func(v *View) {
		v.Attrs.MarginRight = mr
	}
}

func MarginBottom(mb int) ViewOption {
	return func(v *View) {
		v.Attrs.MarginBottom = mb
	}
}

func Position(p FlexPosition) ViewOption {
	return func(v *View) {
		v.Attrs.Position = p
	}
}

func Direction(d FlexDirection) ViewOption {
	return func(v *View) {
		v.Attrs.Direction = d
	}
}

func Wrap(w FlexWrap) ViewOption {
	return func(v *View) {
		v.Attrs.Wrap = w
	}
}

func Justify(j FlexJustify) ViewOption {
	return func(v *View) {
		v.Attrs.Justify = j
	}
}

func AlignItems(ai FlexAlignItem) ViewOption {
	return func(v *View) {
		v.Attrs.AlignItems = ai
	}
}

func AlignContent(ac FlexAlignContent) ViewOption {
	return func(v *View) {
		v.Attrs.AlignContent = ac
	}
}

func Grow(g float64) ViewOption {
	return func(v *View) {
		v.Attrs.Grow = g
	}
}

func Shrink(s float64) ViewOption {
	return func(v *View) {
		v.Attrs.Shrink = s
	}
}

func Display(d FlexDisplay) ViewOption {
	return func(v *View) {
		v.Attrs.Display = d
	}
}

func Hidden(h bool) ViewOption {
	return func(v *View) {
		v.Attrs.Hidden = h
	}
}

func ID(id string) ViewOption {
	return func(v *View) {
		v.Attrs.ID = id
	}
}

func TagName(tagName string) ViewOption {
	return func(v *View) {
		v.Attrs.TagName = tagName
	}
}

type HandlerProvider interface {
	Handler() ViewHandler
}

// Handler sets the handler for the view, it accepts a ViewHandler or a HandlerProvider.
// If a HandlerProvider is provided, it will call Handler() to get the ViewHandler.
// If a ViewHandler is provided, it will set the handler directly.
// If neither is provided, it will panic.
//
// Interface constraints cannot set in Go (cannot union types with behavioral interface, https://github.com/golang/go/issues/49054), so it is up to the caller to provide the correct type.
func Handler(h any) ViewOption {
	return func(v *View) {
		switch handler := h.(type) {
		case ViewHandler:
			v.Handler = handler
		case HandlerProvider:
			v.Handler = handler.Handler()
		default:
			panic("invalid handler type")
		}
	}
}
