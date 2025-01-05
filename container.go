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
	children []*child
	isDirty  bool
	frame    image.Rectangle
	touchIDs []ebiten.TouchID

	calculatedWidth  int
	calculatedHeight int
}

func (ct *containerEmbed) processEvent() {
	ct.handleTouchEvents()
	ct.handleMouseEvents()
}

// Draw draws it's children
func (ct *containerEmbed) Draw(screen *ebiten.Image) {
	for _, c := range ct.children {
		ct.drawChild(screen, c)
	}
}

func (ct *containerEmbed) drawChild(screen *ebiten.Image, child *child) {
	b := ct.computeBounds(child)
	if ct.shouldDrawChild(child) {
		ct.handleDraw(screen, b, child)
	}
	child.item.Draw(screen)
	ct.debugDraw(screen, b, child)
}

func (ct *containerEmbed) computeBounds(child *child) image.Rectangle {
	if child.absolute {
		return scaleFrame(child.bounds)
	}
	return scaleFrame(child.bounds.Add(ct.frame.Min))
}

func (ct *containerEmbed) handleDraw(screen *ebiten.Image, b image.Rectangle, child *child) {
	child.item.Handler.Draw(screen, b, child.item)
}

func (ct *containerEmbed) shouldDrawChild(child *child) bool {
	return !child.item.Attrs.Hidden && child.item.Attrs.Display != DisplayNone && child.item.Handler.Draw != nil
}

func (ct *containerEmbed) debugDraw(screen *ebiten.Image, b image.Rectangle, child *child) {
	if Debug {
		pos := fmt.Sprintf("(%d, %d)-(%d, %d):%s:%s", b.Min.X, b.Min.Y, b.Max.X, b.Max.Y, child.item.Attrs.TagName, child.item.Attrs.ID)
		graphic.FillRect(screen, &graphic.FillRectOpts{
			Color: color.RGBA{0, 0, 0, 200},
			Rect:  image.Rect(b.Min.X, b.Min.Y, b.Min.X+len(pos)*6, b.Min.Y+12),
		})
		ebitenutil.DebugPrintAt(screen, pos, b.Min.X, b.Min.Y)
	}
}

func (ct *containerEmbed) HandleJustPressedTouchID(touchID ebiten.TouchID, x, y int) bool {
	for c := len(ct.children) - 1; c >= 0; c-- {
		child := ct.children[c]
		childFrame := ct.childFrame(child)
		if child.item.Attrs.Display == DisplayNone {
			continue
		}
		if child.HandleJustPressedTouchID(childFrame, touchID, x, y) {
			return true
		}
		if child.item.HandleJustPressedTouchID(touchID, x, y) {
			return true
		}
	}
	return false
}

func (ct *containerEmbed) HandleJustReleasedTouchID(touchID ebiten.TouchID, x, y int) {
	for c := len(ct.children) - 1; c >= 0; c-- {
		child := ct.children[c]
		childFrame := ct.childFrame(child)
		child.HandleJustReleasedTouchID(childFrame, touchID, x, y)
		child.item.HandleJustReleasedTouchID(touchID, x, y)
	}
}

func (ct *containerEmbed) handleMouse(x, y int) bool {
	for c := len(ct.children) - 1; c >= 0; c-- {
		child := ct.children[c]
		childFrame := ct.childFrame(child)
		if child.item.Attrs.Display == DisplayNone {
			continue
		}
		if isInside(childFrame, x, y) {
			if child.item.Handler.HandleMouse(x, y) {
				return true
			}
		}
		if child.item.handleMouse(x, y) {
			return true
		}
	}
	return false
}

func (ct *containerEmbed) handleMouseEnterLeave(x, y int) bool {
	result := false
	for c := len(ct.children) - 1; c >= 0; c-- {
		child := ct.children[c]
		childFrame := ct.childFrame(child)
		if child.item.Attrs.Display == DisplayNone {
			continue
		}
		if !result && !child.isMouseEntered && isInside(childFrame, x, y) {
			if child.item.Handler.HandleMouseEnter(x, y) {
				result = true
				child.isMouseEntered = true
			}
		}

		if child.isMouseEntered && !isInside(childFrame, x, y) {
			child.isMouseEntered = false
			child.item.Handler.HandleMouseLeave()
		}

		if child.item.handleMouseEnterLeave(x, y) {
			result = true
		}
	}
	return result
}

func (ct *containerEmbed) handleMouseButtonLeftPressed(x, y int) bool {
	result := false

	for c := len(ct.children) - 1; c >= 0; c-- {
		child := ct.children[c]
		childFrame := ct.childFrame(child)
		if child.item.Attrs.Display == DisplayNone {
			continue
		}
		if !result && isInside(childFrame, x, y) {
			if child.item.Handler.HandleJustPressedMouseButtonLeft(*childFrame, x, y) {
				result = true
				child.isMouseLeftButtonHandler = true
			}
		}

		if !result && isInside(childFrame, x, y) {
			if !child.isButtonPressed {
				child.isButtonPressed = true
				child.isMouseLeftButtonHandler = true
				result = true
				child.item.Handler.HandlePress(x, y, -1)
			}
		}

		if !result && child.item.handleMouseButtonLeftPressed(x, y) {
			result = true
		}
	}
	return result
}

func (ct *containerEmbed) handleMouseButtonLeftReleased(x, y int) {
	for c := len(ct.children) - 1; c >= 0; c-- {
		child := ct.children[c]
		if child.isMouseLeftButtonHandler {
			child.isMouseLeftButtonHandler = false
			child.item.Handler.HandleJustReleasedMouseButtonLeft(*ct.childFrame(child), x, y)
		}

		if child.isButtonPressed && child.isMouseLeftButtonHandler {
			child.isButtonPressed = false
			child.isMouseLeftButtonHandler = false
			if x == 0 && y == 0 {
				child.item.Handler.HandleRelease(x, y, true)
			} else {
				child.item.Handler.HandleRelease(x, y, !isInside(ct.childFrame(child), x, y))
			}
		}

		child.item.handleMouseButtonLeftReleased(x, y)
	}
}

func isInside(r *image.Rectangle, x, y int) bool {
	return r.Min.X <= x && x <= r.Max.X && r.Min.Y <= y && y <= r.Max.Y
}

func (ct *containerEmbed) handleTouchEvents() {
	justPressedTouchIds := inpututil.AppendJustPressedTouchIDs(nil)

	if justPressedTouchIds != nil {
		for i := 0; i < len(justPressedTouchIds); i++ {
			touchID := justPressedTouchIds[i]
			x, y := ebiten.TouchPosition(touchID)
			recordTouchPosition(touchID, x, y)

			ct.HandleJustPressedTouchID(touchID, x, y)
			ct.touchIDs = append(ct.touchIDs, touchID)
		}
	}

	touchIDs := ct.touchIDs
	for t := range touchIDs {
		if inpututil.IsTouchJustReleased(touchIDs[t]) {
			pos := lastTouchPosition(touchIDs[t])
			ct.HandleJustReleasedTouchID(touchIDs[t], pos.X, pos.Y)
		} else {
			x, y := ebiten.TouchPosition(touchIDs[t])
			recordTouchPosition(touchIDs[t], x, y)
		}
	}
}

func (ct *containerEmbed) handleMouseEvents() {
	x, y := ebiten.CursorPosition()
	ct.handleMouse(x, y)
	ct.handleMouseEnterLeave(x, y)
	if inpututil.IsMouseButtonJustPressed((ebiten.MouseButtonLeft)) {
		ct.handleMouseButtonLeftPressed(x, y)
	}
	if inpututil.IsMouseButtonJustReleased((ebiten.MouseButtonLeft)) {
		ct.handleMouseButtonLeftReleased(x, y)
	}
}

func (ct *containerEmbed) setFrame(frame image.Rectangle) {
	ct.frame = frame
	ct.isDirty = true
}

func (ct *containerEmbed) childFrame(c *child) *image.Rectangle {
	if c.absolute {
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
