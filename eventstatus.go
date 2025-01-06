package furex

import (
	"image"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type EventStatus struct {
	isMousePressed bool
	isMouseEntered bool
	handledTouchID ebiten.TouchID
	swipe
}

type swipe struct {
	downX, downY int
	upX, upY     int
	downTime     time.Time
	upTime       time.Time
	swipeDir     SwipeDirection
	swipeTouchID ebiten.TouchID
}

func (c *View) checkSwipeHandlerStart(frame *image.Rectangle, touchID ebiten.TouchID, x, y int) bool {
	if c.Handler.Swipe != nil {
		if isInside(frame, x, y) {
			c.Status.swipeTouchID = touchID
			c.Status.swipe.downTime = time.Now()
			c.Status.swipe.downX, c.Status.swipe.downY = x, y
			return true
		}
	}
	return false
}

func (c *View) checkSwipeHandlerEnd(touchID ebiten.TouchID, x, y int) bool {
	if c.Handler.Swipe != nil {
		if c.Status.swipeTouchID != touchID {
			return false
		}
		c.Status.swipeTouchID = -1
		c.Status.upTime = time.Now()
		c.Status.upX, c.Status.upY = x, y
		if c.checkSwipe() {
			c.Handler.HandleSwipe(c.Status.swipeDir)
			return true
		}
	}
	return false
}

const swipeThresholdDist = 50.
const swipeThresholdTime = time.Millisecond * 300

func (c *View) checkSwipe() bool {
	dur := c.Status.upTime.Sub(c.Status.downTime)
	if dur > swipeThresholdTime {
		return false
	}

	deltaX := float64(c.Status.downX - c.Status.upX)
	if math.Abs(deltaX) >= swipeThresholdDist {
		if deltaX > 0 {
			c.Status.swipeDir = SwipeDirectionLeft
		} else {
			c.Status.swipeDir = SwipeDirectionRight
		}
		return true
	}

	deltaY := float64(c.Status.downY - c.Status.upY)
	if math.Abs(deltaY) >= swipeThresholdDist {
		if deltaY > 0 {
			c.Status.swipeDir = SwipeDirectionUp
		} else {
			c.Status.swipeDir = SwipeDirectionDown
		}
		return true
	}

	return false
}
