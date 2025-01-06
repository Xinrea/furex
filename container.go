package furex

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/furex/v2/internal/graphic"
)

type containerEmbed struct {
	// frame is the frame of the container in flex layout
	frame image.Rectangle
	// bounds is the bounds of the container
	bounds   image.Rectangle
	children []*View
	isDirty  bool
	touchIDs []ebiten.TouchID

	calculatedWidth  int
	calculatedHeight int
}

// processEvent processes touch and mouse events, it can only be called by root view
func (ct *View) processEvent() {
	ct.handleTouchEvents(&ct.frame)
	ct.handleMouseEvents(&ct.frame)
}

// Draw draws it's children
func (ct *View) childrenDraw(screen *ebiten.Image) {
	for _, c := range ct.children {
		ct.drawChild(screen, c)
	}
}

func (ct *View) drawChild(screen *ebiten.Image, child *View) {
	b := ct.computeBounds(child)
	if ct.shouldDrawChild(child) {
		ct.handleDraw(screen, b, child)
	}
	child.Draw(screen)
	ct.debugDraw(screen, b, child)
}

func (ct *View) computeBounds(child *View) image.Rectangle {
	return scaleFrame(child.frame)
}

func (ct *View) handleDraw(screen *ebiten.Image, b image.Rectangle, child *View) {
	child.Handler.Draw(screen, b, child)
}

func (ct *View) shouldDrawChild(child *View) bool {
	return !child.Attrs.Hidden && child.Attrs.Display != DisplayNone && child.Handler.Draw != nil
}

func (ct *View) debugDraw(screen *ebiten.Image, b image.Rectangle, child *View) {
	if Debug {
		pos := fmt.Sprintf("(%d, %d)-(%d, %d):%s:%s", b.Min.X, b.Min.Y, b.Max.X, b.Max.Y, child.Attrs.TagName, child.Attrs.ID)
		graphic.FillRect(screen, &graphic.FillRectOpts{
			Color: color.RGBA{0, 0, 0, 200},
			Rect:  image.Rect(b.Min.X, b.Min.Y, b.Min.X+len(pos)*6, b.Min.Y+12),
		})
		ebitenutil.DebugPrintAt(screen, pos, b.Min.X, b.Min.Y)
	}
}

// HandleJustPressedTouchID handles touch event. LayoutFrame is the frame of the container in flex layout.
func (ct *View) HandleJustPressedTouchID(layoutFrame *image.Rectangle, touchID ebiten.TouchID, x, y int) bool {
	for c := len(ct.children) - 1; c >= 0; c-- {
		child := ct.children[c]
		childFrame := ct.childFrame(child)
		if child.HandleJustPressedTouchID(childFrame, touchID, x, y) {
			return true
		}
	}
	if ct.Attrs.Display == DisplayNone {
		return false
	}
	if layoutFrame == nil {
		layoutFrame = &ct.frame
	}
	ct.checkSwipeHandlerStart(layoutFrame, touchID, x, y)
	if isInside(layoutFrame, x, y) {
		if ct.Handler.HandleJustPressedTouchID(touchID, x, y) {
			ct.Status.handledTouchID = touchID
			return true
		}
	}
	return false
}

func (ct *View) HandleJustReleasedTouchID(layoutFrame *image.Rectangle, touchID ebiten.TouchID, x, y int) {
	for c := len(ct.children) - 1; c >= 0; c-- {
		child := ct.children[c]
		childFrame := ct.childFrame(child)
		child.HandleJustReleasedTouchID(childFrame, touchID, x, y)
	}
	if layoutFrame == nil {
		layoutFrame = &ct.frame
	}
	ct.checkSwipeHandlerEnd(touchID, x, y)
	if ct.Status.handledTouchID == touchID {
		ct.Handler.HandleJustReleasedTouchID(touchID, x, y, !isInside(layoutFrame, x, y))
		ct.Status.handledTouchID = -1
	}
}

func (ct *View) handleMouse(layoutFrame *image.Rectangle, x, y int) bool {
	for c := len(ct.children) - 1; c >= 0; c-- {
		child := ct.children[c]
		childFrame := ct.childFrame(child)
		if child.handleMouse(childFrame, x, y) {
			return true
		}
	}
	if ct.Attrs.Display == DisplayNone {
		return false
	}
	if layoutFrame == nil {
		layoutFrame = &ct.frame
	}
	if isInside(layoutFrame, x, y) {
		return ct.Handler.HandleMouse(x, y)
	}
	return false
}

func (ct *View) handleMouseEnterLeave(layoutFrame *image.Rectangle, x, y int) bool {
	result := false
	for c := len(ct.children) - 1; c >= 0; c-- {
		child := ct.children[c]
		childFrame := ct.childFrame(child)
		if child.handleMouseEnterLeave(childFrame, x, y) {
			result = true
		}
	}
	if ct.Attrs.Display == DisplayNone {
		return result
	}
	if !result && !ct.Status.isMouseEntered && isInside(layoutFrame, x, y) {
		if ct.Handler.HandleMouseEnter(x, y) {
			result = true
			ct.Status.isMouseEntered = true
		}
	}

	if ct.Status.isMouseEntered && !isInside(layoutFrame, x, y) {
		ct.Status.isMouseEntered = false
		ct.Handler.HandleMouseLeave()
	}
	return result
}

func (ct *View) handleMouseButtonLeftPressed(layoutFrame *image.Rectangle, x, y int) bool {
	for c := len(ct.children) - 1; c >= 0; c-- {
		child := ct.children[c]
		childFrame := ct.childFrame(child)
		if child.handleMouseButtonLeftPressed(childFrame, x, y) {
			return true
		}
	}
	if ct.Attrs.Display == DisplayNone {
		return false
	}
	if layoutFrame == nil {
		layoutFrame = &ct.frame
	}

	if isInside(layoutFrame, x, y) {
		ct.Status.isMousePressed = true
		return ct.Handler.HandleJustPressedMouseButtonLeft(*layoutFrame, x, y)
	}
	return false
}

func (ct *View) handleMouseButtonLeftReleased(layoutFrame *image.Rectangle, x, y int) {
	for c := len(ct.children) - 1; c >= 0; c-- {
		child := ct.children[c]
		childFrame := ct.childFrame(child)
		child.handleMouseButtonLeftReleased(childFrame, x, y)
	}
	if layoutFrame == nil {
		layoutFrame = &ct.frame
	}
	if ct.Status.isMousePressed {
		ct.Handler.HandleJustReleasedMouseButtonLeft(*layoutFrame, x, y)
		ct.Status.isMousePressed = false
	}
}

func isInside(r *image.Rectangle, x, y int) bool {
	return r.Min.X <= x && x <= r.Max.X && r.Min.Y <= y && y <= r.Max.Y
}

func (ct *View) handleTouchEvents(layoutFrame *image.Rectangle) {
	justPressedTouchIds := inpututil.AppendJustPressedTouchIDs(nil)

	if justPressedTouchIds != nil {
		for i := 0; i < len(justPressedTouchIds); i++ {
			touchID := justPressedTouchIds[i]
			x, y := ebiten.TouchPosition(touchID)
			recordTouchPosition(touchID, x, y)

			ct.HandleJustPressedTouchID(layoutFrame, touchID, x, y)
			ct.touchIDs = append(ct.touchIDs, touchID)
		}
	}

	touchIDs := ct.touchIDs
	for t := range touchIDs {
		if inpututil.IsTouchJustReleased(touchIDs[t]) {
			pos := lastTouchPosition(touchIDs[t])
			ct.HandleJustReleasedTouchID(layoutFrame, touchIDs[t], pos.X, pos.Y)
		} else {
			x, y := ebiten.TouchPosition(touchIDs[t])
			recordTouchPosition(touchIDs[t], x, y)
		}
	}
}

func (ct *View) handleMouseEvents(layoutFrame *image.Rectangle) {
	x, y := ebiten.CursorPosition()
	ct.handleMouse(layoutFrame, x, y)
	ct.handleMouseEnterLeave(layoutFrame, x, y)
	if inpututil.IsMouseButtonJustPressed((ebiten.MouseButtonLeft)) {
		ct.handleMouseButtonLeftPressed(layoutFrame, x, y)
	}
	if inpututil.IsMouseButtonJustReleased((ebiten.MouseButtonLeft)) {
		ct.handleMouseButtonLeftReleased(layoutFrame, x, y)
	}
}

func (ct *View) setFrame(frame image.Rectangle) {
	ct.frame = frame
	ct.isDirty = true
}

func (ct *View) childFrame(c *View) *image.Rectangle {
	if c.IsAbsolute() {
		bounds := scaleFrame(c.bounds)
		return &bounds
	}
	bounds := scaleFrame(c.bounds.Add(ct.frame.Min))
	return &bounds
}

type touchPosition struct {
	X, Y int
}

var (
	touchPositions = make(map[ebiten.TouchID]touchPosition)
)

func recordTouchPosition(t ebiten.TouchID, x, y int) {
	touchPositions[t] = touchPosition{x, y}
}

func lastTouchPosition(t ebiten.TouchID) *touchPosition {
	s, ok := touchPositions[t]
	if ok {
		return &s
	}
	return &touchPosition{0, 0}
}

func scaleFrame(frame image.Rectangle) image.Rectangle {
	return image.Rect(
		int(float64(frame.Min.X)*GlobalScale),
		int(float64(frame.Min.Y)*GlobalScale),
		int(float64(frame.Max.X)*GlobalScale),
		int(float64(frame.Max.Y)*GlobalScale),
	)
}
