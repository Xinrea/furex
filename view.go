package furex

import (
	"fmt"
	"image"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

func NewView(opts ...ViewOption) *View {
	v := &View{}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

type ViewAttrs struct {
	Left         int
	Right        *int
	Top          int
	Bottom       *int
	Width        int
	WidthInPct   float64
	Height       int
	HeightInPct  float64
	MarginLeft   int
	MarginTop    int
	MarginRight  int
	MarginBottom int
	Position     FlexPosition
	Direction    FlexDirection
	Wrap         FlexWrap
	Justify      FlexJustify
	AlignItems   FlexAlignItem
	AlignContent FlexAlignContent
	Grow         float64
	Shrink       float64
	Display      FlexDisplay

	ID         string
	Raw        string
	TagName    string
	Text       string
	ExtraAttrs map[string]string
	Hidden     bool
}

// View represents a UI element.
// You can set flex options, size, position and so on.
// Handlers can be set to create custom component such as button or list.
type View struct {
	Attrs   ViewAttrs
	Handler ViewHandler

	containerEmbed
	flexEmbed
	lock      sync.Mutex
	hasParent bool
	parent    *View
}

// Update updates the view
func (v *View) Update() {
	if v.isDirty {
		v.startLayout()
	}
	if !v.hasParent {
		v.processHandler()
	}
	for _, v := range v.children {
		v.item.Update()
		v.item.processHandler()
	}
	if !v.hasParent {
		v.processEvent()
	}
}

func (v *View) processHandler() {
	v.Handler.HandleUpdate(v)
}

func (v *View) startLayout() {
	v.lock.Lock()
	defer v.lock.Unlock()
	if !v.hasParent {
		v.frame = image.Rect(v.Attrs.Left, v.Attrs.Top, v.Attrs.Left+v.Attrs.Width, v.Attrs.Top+v.Attrs.Height)
	}
	v.flexEmbed.View = v

	for _, child := range v.children {
		if child.item.Attrs.Position == PositionStatic {
			child.item.startLayout()
		}
	}

	v.layout(v.frame.Dx(), v.frame.Dy(), &v.containerEmbed)
	v.isDirty = false
}

// UpdateWithSize the view with modified height and width
func (v *View) UpdateWithSize(width, height int) {
	if !v.hasParent && (v.Attrs.Width != width || v.Attrs.Height != height) {
		v.Attrs.Height = height
		v.Attrs.Width = width
		v.isDirty = true
	}
	v.Update()
}

// Layout marks the view as dirty
func (v *View) Layout() {
	v.isDirty = true
	if v.hasParent {
		v.parent.isDirty = true
	}
}

// Draw draws the view
func (v *View) Draw(screen *ebiten.Image) {
	if v.isDirty {
		v.startLayout()
	}
	if !v.hasParent {
		// scale frame with GlobalScale
		v.handleDrawRoot(screen, scaleFrame(v.frame))
	}
	if !v.Attrs.Hidden && v.Attrs.Display != DisplayNone {
		v.containerEmbed.Draw(screen)
	}
	if Debug && !v.hasParent && v.Attrs.Display != DisplayNone {
		debugBorders(screen, v.containerEmbed)
	}
}

// AddTo add itself to a parent view
func (v *View) AddTo(parent *View) *View {
	if v.hasParent {
		panic("this view has been already added to a parent")
	}
	parent.AddChild(v)
	return v
}

// AddChild adds one or multiple child views
func (v *View) AddChild(views ...*View) *View {
	for _, vv := range views {
		v.addChild(vv)
	}
	return v
}

// RemoveChild removes a specified view
func (v *View) RemoveChild(cv *View) bool {
	for i, child := range v.children {
		if child.item == cv {
			v.children = append(v.children[:i], v.children[i+1:]...)
			v.isDirty = true
			cv.hasParent = false
			cv.parent = nil
			return true
		}
	}
	return false
}

// RemoveAll removes all children view
func (v *View) RemoveAll() {
	v.isDirty = true
	for _, child := range v.children {
		child.item.hasParent = false
		child.item.parent = nil
	}
	v.children = []*child{}
}

// PopChild remove the last child view add to this view
func (v *View) PopChild() *View {
	if len(v.children) == 0 {
		return nil
	}
	c := v.children[len(v.children)-1]
	v.children = v.children[:len(v.children)-1]
	v.isDirty = true
	c.item.hasParent = false
	c.item.parent = nil
	return c.item
}

func (v *View) addChild(cv *View) *View {
	child := &child{item: cv, handledTouchID: -1}
	v.children = append(v.children, child)
	v.isDirty = true
	cv.hasParent = true
	cv.parent = v
	return v
}

func (v *View) isWidthFixed() bool {
	return v.Attrs.Width != 0 || v.Attrs.WidthInPct != 0
}

func (v *View) width() int {
	if v.Attrs.Width == 0 {
		return v.calculatedWidth
	}
	return v.Attrs.Width
}

func (v *View) isHeightFixed() bool {
	return v.Attrs.Height != 0 || v.Attrs.HeightInPct != 0
}

func (v *View) height() int {
	if v.Attrs.Height == 0 {
		return v.calculatedHeight
	}
	return v.Attrs.Height
}

func (v *View) GetChildren() []*View {
	if v == nil || v.children == nil {
		return nil
	}
	ret := make([]*View, len(v.children))
	for i, child := range v.children {
		ret[i] = child.item
	}
	return ret
}

// Len returns the number of children.
func (v *View) Len() int {
	return len(v.children)
}

func (v *View) NthChild(n int) *View {
	if n < 0 || n >= len(v.children) {
		return nil
	}
	return v.children[n].item
}

func (v *View) First() *View {
	return v.NthChild(0)
}

func (v *View) Last() *View {
	return v.NthChild(v.Len() - 1)
}

// GetByID returns the view with the specified id.
// It returns nil if not found.
func (v *View) GetByID(id string) (*View, bool) {
	if v.Attrs.ID == id {
		return v, true
	}
	for _, child := range v.children {
		if v, ok := child.item.GetByID(id); ok {
			return v, true
		}
	}
	return nil, false
}

// MustGetByID returns the view with the specified id.
// It panics if not found.
func (v *View) MustGetByID(id string) *View {
	vv, ok := v.GetByID(id)
	if !ok {
		panic("view not found")
	}
	return vv
}

// FilterByTagName returns views with the specified tag name.
func (v *View) FilterByTagName(tagName string) []*View {
	var views []*View
	if v.Attrs.TagName == tagName {
		views = append(views, v)
	}
	for _, child := range v.children {
		views = append(views, child.item.FilterByTagName(tagName)...)
	}
	return views
}

// SetLeft sets the left position of the view.
func (v *View) SetLeft(left int) {
	v.Attrs.Left = left
	v.Layout()
}

// SetRight sets the right position of the view.
func (v *View) SetRight(right int) {
	v.Attrs.Right = Int(right)
	v.Layout()
}

// SetTop sets the top position of the view.
func (v *View) SetTop(top int) {
	v.Attrs.Top = top
	v.Layout()
}

// SetBottom sets the bottom position of the view.
func (v *View) SetBottom(bottom int) {
	v.Attrs.Bottom = Int(bottom)
	v.Layout()
}

// SetWidth sets the width of the view.
func (v *View) SetWidth(width int) {
	v.Attrs.Width = width
	v.Layout()
}

// SetHeight sets the height of the view.
func (v *View) SetHeight(height int) {
	v.Attrs.Height = height
	v.Layout()
}

// SetMarginLeft sets the left margin of the view.
func (v *View) SetMarginLeft(marginLeft int) {
	v.Attrs.MarginLeft = marginLeft
	v.Layout()
}

// SetMarginTop sets the top margin of the view.
func (v *View) SetMarginTop(marginTop int) {
	v.Attrs.MarginTop = marginTop
	v.Layout()
}

// SetMarginRight sets the right margin of the view.
func (v *View) SetMarginRight(marginRight int) {
	v.Attrs.MarginRight = marginRight
	v.Layout()
}

// SetMarginBottom sets the bottom margin of the view.
func (v *View) SetMarginBottom(marginBottom int) {
	v.Attrs.MarginBottom = marginBottom
	v.Layout()
}

// SetPosition sets the position of the view.
func (v *View) SetPosition(position FlexPosition) {
	v.Attrs.Position = position
	v.Layout()
}

// SetDirection sets the direction of the view.
func (v *View) SetDirection(direction FlexDirection) {
	v.Attrs.Direction = direction
	v.Layout()
}

// SetWrap sets the wrap property of the view.
func (v *View) SetWrap(wrap FlexWrap) {
	v.Attrs.Wrap = wrap
	v.Layout()
}

// SetJustify sets the justify property of the view.
func (v *View) SetJustify(justify FlexJustify) {
	v.Attrs.Justify = justify
	v.Layout()
}

// SetAlignItems sets the align items property of the view.
func (v *View) SetAlignItems(alignItems FlexAlignItem) {
	v.Attrs.AlignItems = alignItems
	v.Layout()
}

// SetAlignContent sets the align content property of the view.
func (v *View) SetAlignContent(alignContent FlexAlignContent) {
	v.Attrs.AlignContent = alignContent
	v.Layout()
}

// SetGrow sets the grow property of the view.
func (v *View) SetGrow(grow float64) {
	v.Attrs.Grow = grow
	v.Layout()
}

// SetShrink sets the shrink property of the view.
func (v *View) SetShrink(shrink float64) {
	v.Attrs.Shrink = shrink
	v.Layout()
}

// SetDisplay sets the display property of the view.
func (v *View) SetDisplay(display FlexDisplay) {
	v.Attrs.Display = display
	v.Layout()
}

// SetHidden sets the hidden property of the view.
func (v *View) SetHidden(hidden bool) {
	v.Attrs.Hidden = hidden
	v.Layout()
}

func (v *View) Config() ViewConfig {
	cfg := ViewConfig{
		TagName:      v.Attrs.TagName,
		ID:           v.Attrs.ID,
		Left:         v.Attrs.Left,
		Right:        v.Attrs.Right,
		Top:          v.Attrs.Top,
		Bottom:       v.Attrs.Bottom,
		Width:        v.Attrs.Width,
		Height:       v.Attrs.Height,
		MarginLeft:   v.Attrs.MarginLeft,
		MarginTop:    v.Attrs.MarginTop,
		MarginRight:  v.Attrs.MarginRight,
		MarginBottom: v.Attrs.MarginBottom,
		Position:     v.Attrs.Position,
		Direction:    v.Attrs.Direction,
		Wrap:         v.Attrs.Wrap,
		Justify:      v.Attrs.Justify,
		AlignItems:   v.Attrs.AlignItems,
		AlignContent: v.Attrs.AlignContent,
		Grow:         v.Attrs.Grow,
		Shrink:       v.Attrs.Shrink,
		children:     []ViewConfig{},
	}
	for _, child := range v.GetChildren() {
		cfg.children = append(cfg.children, child.Config())
	}
	return cfg
}

func (v *View) handleDrawRoot(screen *ebiten.Image, b image.Rectangle) {
	v.Handler.HandleDraw(screen, b, v)
}

// This is for debugging and testing.
type ViewConfig struct {
	TagName      string
	ID           string
	Left         int
	Right        *int
	Top          int
	Bottom       *int
	Width        int
	Height       int
	MarginLeft   int
	MarginTop    int
	MarginRight  int
	MarginBottom int
	Position     FlexPosition
	Direction    FlexDirection
	Wrap         FlexWrap
	Justify      FlexJustify
	AlignItems   FlexAlignItem
	AlignContent FlexAlignContent
	Grow         float64
	Shrink       float64
	children     []ViewConfig
}

func (cfg ViewConfig) Tree() string {
	return cfg.tree("")
}

// TODO: This is a bit of a mess. Clean it up.
func (cfg ViewConfig) tree(indent string) string {
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s<%s ", indent, cfg.TagName))
	if cfg.ID != "" {
		sb.WriteString(fmt.Sprintf("id=\"%s\" ", cfg.ID))
	}
	sb.WriteString("style=\"")
	sb.WriteString(
		fmt.Sprintf("left: %d, right: %d, top: %d, bottom: %d, width: %d, height: %d, marginLeft: %d, marginTop: %d, marginRight: %d, marginBottom: %d, position: %s, direction: %s, wrap: %s, justify: %s, alignItems: %s, alignContent: %s, grow: %f, shrink: %f",
			cfg.Left, *cfg.Right, cfg.Top, *cfg.Bottom, cfg.Width, cfg.Height, cfg.MarginLeft, cfg.MarginTop, cfg.MarginRight, cfg.MarginBottom, cfg.Position, cfg.Direction, cfg.Wrap, cfg.Justify, cfg.AlignItems, cfg.AlignContent, cfg.Grow, cfg.Shrink))
	sb.WriteString("\">\n")
	for _, child := range cfg.children {
		sb.WriteString(child.tree(indent + "  "))
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("%s</%s>", indent, cfg.TagName))
	sb.WriteString("\n")
	return sb.String()
}
