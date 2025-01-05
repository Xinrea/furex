package furex

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Drawer represents a component that can be added to a container.
type Drawer interface {
	// Draw function draws the content of the component inside the frame.
	HandleDraw(screen *ebiten.Image, frame image.Rectangle, v *View)
}

// Updater represents a component that updates by one tick.
type Updater interface {
	// Update updates the state of the component by one tick.
	HandleUpdate(v *View)
}

// ButtonHandler represents a button component.
type ButtonHandler interface {
	// HandlePress handle the event when user just started pressing the button
	// The parameter (x, y) is the location relative to the window (0,0).
	// touchID is the unique ID of the touch.
	// If the button is pressed by a mouse, touchID is -1.
	HandlePress(x, y int, t ebiten.TouchID)

	// HandleRelease handle the event when user just released the button.
	// The parameter (x, y) is the location relative to the window (0,0).
	// The parameter isCancel is true when the touch/left click is released outside of the button.
	HandleRelease(x, y int, isCancel bool)
}

// TouchHandler represents a component that handle touches.
type TouchHandler interface {
	// HandleJustPressedTouchID handles the touchID just pressed and returns true if it handles the TouchID
	HandleJustPressedTouchID(touch ebiten.TouchID, x, y int) bool
	// HandleJustReleasedTouchID handles the touchID just released
	// Should be called only when it handled the TouchID when pressed
	HandleJustReleasedTouchID(touch ebiten.TouchID, x, y int)
}

// MouseHandler represents a component that handle mouse move.
type MouseHandler interface {
	// HandleMouse handles the mouch move and returns true if it handle the mouse move.
	// The parameter (x, y) is the location relative to the window (0,0).
	HandleMouse(x, y int) bool
}

// MouseLeftButtonHandler represents a component that handle mouse button left click.
type MouseLeftButtonHandler interface {
	// HandleJustPressedMouseButtonLeft handle left mouse button click just pressed.
	// The parameter (x, y) is the location relative to the window (0,0).
	// It returns true if it handles the mouse move.
	HandleJustPressedMouseButtonLeft(frame image.Rectangle, x, y int) bool
	// HandleJustReleasedTouchID handles the touchID just released.
	// The parameter (x, y) is the location relative to the window (0,0).
	HandleJustReleasedMouseButtonLeft(frame image.Rectangle, x, y int)
}

// MouseEnterHandler represets a component that handle mouse enter.
type MouseEnterLeaveHandler interface {
	// HandleMouseEnter handles the mouse enter.
	HandleMouseEnter(x, y int) bool
	// HandleMouseLeave handles the mouse leave.
	HandleMouseLeave()
}

// SwipeHandler represents different swipe directions.
type SwipeDirection int

const (
	SwipeDirectionLeft SwipeDirection = iota
	SwipeDirectionRight
	SwipeDirectionUp
	SwipeDirectionDown
)

// SwipeHandler represents a component that handle swipe.
type SwipeHandler interface {
	// HandleSwipe handles swipes.
	HandleSwipe(dir SwipeDirection)
}

// ViewHandler represents a component that can be added to a container.
// Do not directly use func in this struct, using Handle* instead.
type ViewHandler struct {
	// you can put any extra data here
	Extra                       interface{}
	Draw                        func(screen *ebiten.Image, frame image.Rectangle, v *View)
	Update                      func(v *View)
	Press                       func(x, y int, t ebiten.TouchID)
	Release                     func(x, y int, isCancel bool)
	JustPressedTouchID          func(touch ebiten.TouchID, x, y int) bool
	JustReleasedTouchID         func(touch ebiten.TouchID, x, y int)
	Mouse                       func(x, y int) bool
	JustPressedMouseButtonLeft  func(frame image.Rectangle, x, y int) bool
	JustReleasedMouseButtonLeft func(frame image.Rectangle, x, y int)
	MouseEnter                  func(x, y int) bool
	MouseLeave                  func()
	Swipe                       func(dir SwipeDirection)
}

// IsTouchHandler returns true if the handler is a touch handler. Otherwise, returns false.
//
// The resons why it requires two functions to handle touch is that some state should be kept between the touch start and end.
// If one of them is nil, nothing can be done properly.
func (h *ViewHandler) IsTouchHandler() bool {
	return h.JustPressedTouchID != nil || h.JustReleasedTouchID != nil
}

func (h *ViewHandler) IsButtonHandler() bool {
	return h.Press != nil && h.Release != nil
}

// HandleSwipe implements SwipeHandler.
func (h *ViewHandler) HandleSwipe(dir SwipeDirection) {
	if h.Swipe != nil {
		h.Swipe(dir)
	}
}

// HandleMouseEnter implements MouseEnterLeaveHandler.
func (h *ViewHandler) HandleMouseEnter(x int, y int) bool {
	if h.MouseEnter != nil {
		return h.MouseEnter(x, y)
	}
	return false
}

// HandleMouseLeave implements MouseEnterLeaveHandler.
func (h *ViewHandler) HandleMouseLeave() {
	if h.MouseLeave != nil {
		h.MouseLeave()
	}
}

// HandleJustPressedMouseButtonLeft implements MouseLeftButtonHandler.
func (h *ViewHandler) HandleJustPressedMouseButtonLeft(frame image.Rectangle, x int, y int) bool {
	if h.JustPressedMouseButtonLeft != nil {
		return h.JustPressedMouseButtonLeft(frame, x, y)
	}
	return false
}

// HandleJustReleasedMouseButtonLeft implements MouseLeftButtonHandler.
func (h *ViewHandler) HandleJustReleasedMouseButtonLeft(frame image.Rectangle, x int, y int) {
	if h.JustReleasedMouseButtonLeft != nil {
		h.JustReleasedMouseButtonLeft(frame, x, y)
	}
}

// HandleMouse implements MouseHandler.
func (h *ViewHandler) HandleMouse(x int, y int) bool {
	if h.Mouse != nil {
		return h.Mouse(x, y)
	}
	return false
}

// HandleJustPressedTouchID implements TouchHandler.
func (h *ViewHandler) HandleJustPressedTouchID(touch ebiten.TouchID, x int, y int) bool {
	if h.JustPressedTouchID != nil {
		return h.JustPressedTouchID(touch, x, y)
	}
	return false
}

// HandleJustReleasedTouchID implements TouchHandler.
func (h *ViewHandler) HandleJustReleasedTouchID(touch ebiten.TouchID, x int, y int) {
	if h.JustReleasedTouchID != nil {
		h.JustReleasedTouchID(touch, x, y)
	}
}

// HandleUpdate implements Updater.
func (h *ViewHandler) HandleUpdate(v *View) {
	if h.Update != nil {
		h.Update(v)
	}
}

// HandleDraw implements Drawer.
func (h *ViewHandler) HandleDraw(screen *ebiten.Image, frame image.Rectangle, v *View) {
	if h.Draw != nil {
		h.Draw(screen, frame, v)
	}
}

// HandlePress implements ButtonHandler.
func (h *ViewHandler) HandlePress(x, y int, t ebiten.TouchID) {
	if h.Press != nil {
		h.Press(x, y, t)
	}
}

// HandleRelease implements ButtonHandler.
func (h *ViewHandler) HandleRelease(x, y int, isCancel bool) {
	if h.Release != nil {
		h.Release(x, y, isCancel)
	}
}

var _ ButtonHandler = (*ViewHandler)(nil)
var _ Drawer = (*ViewHandler)(nil)
var _ Updater = (*ViewHandler)(nil)
var _ TouchHandler = (*ViewHandler)(nil)
var _ MouseHandler = (*ViewHandler)(nil)
var _ MouseLeftButtonHandler = (*ViewHandler)(nil)
var _ MouseEnterLeaveHandler = (*ViewHandler)(nil)
var _ SwipeHandler = (*ViewHandler)(nil)
