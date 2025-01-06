// Referenced code: https://github.com/golang/exp/blob/master/shiny/widget/flex/flex.go
package furex

import (
	"fmt"
	"image"
	"math"
)

// FlexDirection is the direction in which flex items are laid out
type FlexDirection uint8

const (
	Row FlexDirection = iota
	Column
)

func (d FlexDirection) String() string {
	switch d {
	case Row:
		return "row"
	case Column:
		return "column"
	default:
		return fmt.Sprintf("unknown direction: %d", d)
	}
}

// FlexJustify aligns items along the main axis.
type FlexJustify uint8

const (
	JustifyStart        FlexJustify = iota // pack to start of line
	JustifyEnd                             // pack to end of line
	JustifyCenter                          // pack to center of line
	JustifySpaceBetween                    // even spacing
	JustifySpaceAround                     // even spacing, half-size on each end
)

func (f FlexJustify) String() string {
	switch f {
	case JustifyStart:
		return "flex-start"
	case JustifyEnd:
		return "flex-end"
	case JustifyCenter:
		return "center"
	case JustifySpaceBetween:
		return "space-between"
	case JustifySpaceAround:
		return "space-around"
	default:
		return fmt.Sprintf("unknown justify: %d", f)
	}
}

// FlexAlign represents align of flex children
type FlexAlign int

const (
	FlexCenter FlexAlign = iota
	FlexStart
	FlexEnd
	FlexSpaceBetween
)

func (f FlexAlign) String() string {
	switch f {
	case FlexCenter:
		return "center"
	case FlexStart:
		return "flex-start"
	case FlexEnd:
		return "flex-end"
	case FlexSpaceBetween:
		return "space-between"
	default:
		return fmt.Sprintf("unknown flex-align: %d", f)
	}
}

// FlexAlignItem aligns items along the cross axis.
type FlexAlignItem uint8

const (
	AlignItemStretch FlexAlignItem = iota
	AlignItemStart
	AlignItemEnd
	AlignItemCenter
)

func (f FlexAlignItem) String() string {
	switch f {
	case AlignItemStretch:
		return "stretch"
	case AlignItemStart:
		return "flex-start"
	case AlignItemEnd:
		return "flex-end"
	case AlignItemCenter:
		return "center"
	default:
		return fmt.Sprintf("unknown align-item: %d", f)
	}
}

// FlexWrap controls whether the container is single- or multi-line,
// and the direction in which the lines are laid out.
type FlexWrap uint8

const (
	NoWrap FlexWrap = iota
	WrapNormal
	WrapReverse
)

func (f FlexWrap) String() string {
	switch f {
	case NoWrap:
		return "nowrap"
	case WrapNormal:
		return "wrap"
	case WrapReverse:
		return "wrap-reverse"
	default:
		return fmt.Sprintf("unknown flex-wrap: %d", f)
	}
}

// FlexAlignContent is the 'align-content' property.
// It aligns container lines when there is extra space on the cross-axis.
type FlexAlignContent uint8

const (
	AlignContentStart FlexAlignContent = iota
	AlignContentEnd
	AlignContentCenter
	AlignContentSpaceBetween
	AlignContentSpaceAround
	AlignContentStretch
)

func (f FlexAlignContent) String() string {
	switch f {
	case AlignContentStart:
		return "start"
	case AlignContentEnd:
		return "end"
	case AlignContentCenter:
		return "center"
	case AlignContentSpaceBetween:
		return "space-between"
	case AlignContentSpaceAround:
		return "space-around"
	case AlignContentStretch:
		return "stretch"
	}
	return fmt.Sprintf("unknown align-content: %d", f)
}

// FlexPosition is the 'position' property
type FlexPosition uint8

const (
	PositionStatic FlexPosition = iota
	PositionAbsolute
)

func (p FlexPosition) String() string {
	switch p {
	case PositionStatic:
		return "static"
	case PositionAbsolute:
		return "absolute"
	}
	return fmt.Sprintf("unknown position: %d", p)
}

// FlexDisplay is the 'display' property
type FlexDisplay uint8

const (
	DisplayFlex FlexDisplay = iota
	DisplayNone
)

func (d FlexDisplay) String() string {
	switch d {
	case DisplayFlex:
		return "flex"
	case DisplayNone:
		return "none"
	}
	return fmt.Sprintf("unknown display: %d", d)
}

// layout is the main routine that implements a subset of flexbox layout
// https://www.w3.org/TR/css-flexbox-1/#layout-algorithm
func (f *View) layout(width, height int, container *containerEmbed) {
	// 9.2. Line Length Determination
	// Determine the available main and cross space for the flex items.
	containerMainSize := float64(f.mainSize(width, height))
	containerCrossSize := float64(f.crossSize(width, height))

	// Determine the flex base size and hypothetical main size of each item:
	var children []element
	for _, c := range container.children {
		if c.Attrs.Display == DisplayNone {
			continue
		}
		if c.Attrs.Position == PositionAbsolute {
			x := container.frame.Min.X
			if c.Attrs.Left != 0 {
				x = container.frame.Min.X + c.Attrs.Left
			} else if c.Attrs.Right != nil {
				x = container.frame.Max.X - *c.Attrs.Right - c.Attrs.Width
			}
			y := container.frame.Min.Y
			if c.Attrs.Top != 0 {
				y = container.frame.Min.Y + c.Attrs.Top
			} else if c.Attrs.Bottom != nil {
				y = container.frame.Max.Y - *c.Attrs.Bottom - c.Attrs.Height
			}
			c.frame = image.Rect(x, y, x+c.Attrs.Width, y+c.Attrs.Height)
			c.bounds = image.Rect(0, 0, c.Attrs.Width, c.Attrs.Height)
			continue
		}
		children = append(children, element{
			widthInPct:   c.Attrs.WidthInPct,
			heightInPct:  c.Attrs.HeightInPct,
			flexBaseSize: float64(f.flexBaseSize(c)),
			node:         c,
		})
	}

	// Depending on the flex container direction, apply calculation for width and height in percent.
	switch f.Attrs.Direction {
	case Row:
		// Calculate the remaining width after taking out the fixed width items.
		remFree := width
		for _, c := range children {
			remFree -= (c.node.Attrs.Width + c.node.Attrs.MarginLeft + c.node.Attrs.MarginRight)
		}
		// If there is remaining space, distribute it among the flexible items.
		if remFree > 0 {
			for i, c := range children {
				if c.widthInPct > 0 {
					v := float64(width) * c.widthInPct / 100.
					children[i].node.calculatedWidth = int(math.Min(v, float64(remFree)))
					children[i].flexBaseSize = float64(f.flexBaseSize(children[i].node))
				}
			}
		}
		// If the container is a row, calculate the height of each item.
		for _, c := range children {
			if c.heightInPct > 0 {
				// Calculate the new width based on the item's width percentage.
				c.node.calculatedHeight = int(float64(height) * c.node.Attrs.HeightInPct / 100)
			}
		}
	case Column:
		// Calculate the remaining height after taking out the fixed width items.
		remFree := height
		for _, c := range children {
			remFree -= (c.node.Attrs.Height + c.node.Attrs.MarginTop + c.node.Attrs.MarginBottom)
		}
		// If there is remaining space, distribute it among the flexible items.
		if remFree > 0 {
			for i, c := range children {
				if c.heightInPct > 0 {
					v := float64(height) * c.heightInPct / 100.
					children[i].node.calculatedHeight = int(math.Min(v, float64(remFree)))
					children[i].flexBaseSize = float64(f.flexBaseSize(children[i].node))
				}
			}
		}
		// If the container is a column, calculate the width of each item.
		for _, c := range children {
			if c.widthInPct > 0 {
				// Calculate the new width based on the item's width percentage.
				c.node.calculatedWidth = int(float64(width) * c.node.Attrs.WidthInPct / 100)
			}
		}
	default:
		panic(fmt.Sprint("flex: bad direction ", f.Attrs.Direction))
	}

	// §9.3. Main Size Determination
	// Collect flex items into flex lines
	var lines []flexLine
	if f.Attrs.Wrap == NoWrap {
		// Single line
		line := flexLine{child: make([]*element, len(children))}
		for i := range children {
			child := &children[i]
			child.mainMargin = f.mainMargin(child.node)
			line.child[i] = child
			line.mainSize += child.flexBaseSize +
				(child.mainMargin[0] + child.mainMargin[1])
		}
		lines = []flexLine{line}
	} else {
		// Multi line
		var line flexLine
		for i := range children {
			child := &children[i]
			child.mainMargin = f.mainMargin(child.node)

			// hypotheticalMainSize = flexBaseSize + main margin
			hypotheticalMainSize := child.flexBaseSize +
				(child.mainMargin[0] + child.mainMargin[1])

			if line.mainSize > 0 && line.mainSize+hypotheticalMainSize > containerMainSize {
				lines = append(lines, line)
				line = flexLine{}
			}
			line.child = append(line.child, child)
			line.mainSize += hypotheticalMainSize
		}

		if len(line.child) > 0 || len(children) == 0 {
			lines = append(lines, line)
		}
	}

	// §9.3.6 resolve flexible lengths (details in section §9.7)
	for l := range lines {
		line := &lines[l]

		grow := line.mainSize < containerMainSize // §9.7.1

		// §9.7.2 freeze inflexible children.
		for _, child := range line.child {
			mainSize := float64(f.mainSize(child.node.width(), child.node.height()))
			if grow {
				if child.node.Attrs.Grow == 0 {
					child.frozen = true
					child.mainSize = mainSize
				}
			} else {
				if child.node.Attrs.Shrink == 0 {
					child.frozen = true
					child.mainSize = mainSize
				}
			}
		}

		// §9.7.3 calculate initial free space
		freeSpace := float64(f.mainSize(width, height))
		for _, child := range line.child {
			freeSpace -= (float64(f.flexBaseSize(child.node)) +
				(child.mainMargin[0] + child.mainMargin[1]))
		}

		// §9.7.4 flex loop
		for {
			// Check for flexible items.
			allFrozen := true
			for _, child := range line.child {
				if !child.frozen {
					allFrozen = false
					break
				}
			}
			if allFrozen {
				break
			}

			// Calculate remaining free space.
			remFreeSpace := float64(f.mainSize(width, height))
			unfrozenFlexFactor := 0.0
			for _, child := range line.child {
				mainMargin := child.mainMargin[0] + child.mainMargin[1]
				if child.frozen {
					remFreeSpace -= (child.mainSize + mainMargin)
				} else {
					remFreeSpace -= (float64(f.mainSize(child.node.width(), child.node.height())) + mainMargin)
					if grow {
						unfrozenFlexFactor += child.node.Attrs.Grow
					} else {
						unfrozenFlexFactor += child.node.Attrs.Shrink
					}
				}
			}

			if unfrozenFlexFactor < 1 {
				p := freeSpace * unfrozenFlexFactor
				if math.Abs(p) < math.Abs(remFreeSpace) {
					remFreeSpace = p
				}
			}

			// Distribute free space proportional to flex factors.
			if grow {
				for _, child := range line.child {
					if child.frozen {
						continue
					}
					r := child.node.Attrs.Grow / unfrozenFlexFactor
					child.mainSize = float64(f.mainSize(
						child.node.width(), child.node.height(),
					)) + r*remFreeSpace
				}
			} else {
				sumScaledShrinkFactor := 0.0
				for _, child := range line.child {
					if child.frozen {
						continue
					}
					scaledShrinkFactor := float64(f.mainSize(
						child.node.width(), child.node.height(),
					)) * child.node.Attrs.Shrink
					sumScaledShrinkFactor += scaledShrinkFactor
				}
				for _, child := range line.child {
					if child.frozen {
						continue
					}
					scaledShrinkFactor := float64(f.mainSize(
						child.node.width(), child.node.height(),
					)) * child.node.Attrs.Shrink
					r := float64(scaledShrinkFactor) / sumScaledShrinkFactor
					child.mainSize = float64(f.mainSize(
						child.node.width(), child.node.height(),
					)) - r*math.Abs(float64(remFreeSpace))
				}
			}

			for _, child := range line.child {
				child.frozen = true
			}

		}
	}

	// §9.4. Cross Size Determination
	// Determine the hypothetical cross size of each item
	for l := range lines {
		for _, c := range lines[l].child {
			c.crossMargin = f.crossMargin(c.node)
			c.crossSize = float64(
				f.crossSize(c.node.width(), c.node.height()),
			)
		}
	}

	// §9.4.8 Calculate the cross size of each flex line.
	if len(lines) == 1 {
		// Single line
		lines[0].crossSize = containerCrossSize
	} else {
		// Multi line
		for l := range lines {
			line := &lines[l]
			max := 0.0
			for _, child := range line.child {
				if child.crossSize > max {
					max = child.crossSize +
						(child.crossMargin[0] + child.crossMargin[1])
				}
			}
			line.crossSize = max
		}
	}

	off := 0.0
	for l := range lines {
		line := &lines[l]
		line.crossOffset = off
		off += line.crossSize
	}

	// §9.4.9 align-content: stretch
	remCrossSize := containerCrossSize - off
	if f.Attrs.AlignContent == AlignContentStretch && remCrossSize > 0 {
		add := remCrossSize / float64(len(lines))
		for l := range lines {
			line := &lines[l]
			line.crossOffset += float64(l) * add
			line.crossSize += add
		}
	}

	// §9.4.11 align-item: stretch
	for l := range lines {
		line := &lines[l]
		for _, child := range line.child {
			if f.Attrs.AlignItems == AlignItemStretch &&
				!f.isCrossSizeFixed(child.node) &&
				child.crossSize < line.crossSize {
				crossMargin := child.crossMargin[0] + child.crossMargin[1]
				child.crossSize = line.crossSize - crossMargin
			}
		}
	}

	// §9.5. Main-Axis Alignment
	for l := range lines {
		line := &lines[l]
		total := 0.0
		for _, child := range line.child {
			total += child.mainSize +
				(child.mainMargin[0] + child.mainMargin[1])
		}
		remFree := containerMainSize - total
		off, spacing := 0.0, 0.0
		switch f.Attrs.Justify {
		case JustifyStart:
		case JustifyEnd:
			off = remFree
		case JustifyCenter:
			off = remFree / 2
		case JustifySpaceBetween:
			spacing = remFree / float64(len(line.child)-1)
		case JustifySpaceAround:
			spacing = remFree / float64(len(line.child))
			off = spacing / 2
		}
		for _, child := range line.child {
			child.mainOffset = off + (child.mainMargin[0])
			off += spacing + child.mainSize +
				(child.mainMargin[0] + child.mainMargin[1])
		}
	}

	// §9.6. Cross axis alignment
	for l := range lines {
		line := &lines[l]
		for _, child := range line.child {
			child.crossOffset = line.crossOffset + (child.crossMargin[0])
			if child.crossSize == line.crossSize {
				continue
			}
			diff := line.crossSize - child.crossSize -
				(child.crossMargin[0] + child.crossMargin[1])
			switch f.Attrs.AlignItems {
			case AlignItemStart:
				// already laid out correctly
			case AlignItemEnd:
				child.crossOffset = line.crossOffset + diff +
					(child.crossMargin[0])
			case AlignItemCenter:
				child.crossOffset = line.crossOffset + diff/2 +
					(child.crossMargin[0])
			}
		}
	}

	// §9.6.15 determine container cross size used
	crossSize := lines[len(lines)-1].crossOffset + lines[len(lines)-1].crossSize
	remFree := containerCrossSize - crossSize

	// §9.6.16 align flex lines, 'align-content'.
	if remFree > 0 {
		spacing, off := 0.0, 0.0
		switch f.Attrs.AlignContent {
		case AlignContentStart:
			// already laid out correctly
		case AlignContentEnd:
			off = remFree
		case AlignContentCenter:
			off = remFree / 2
		case AlignContentSpaceBetween:
			spacing = remFree / float64(len(lines)-1)
		case AlignContentSpaceAround:
			spacing = remFree / float64(len(lines))
			off = spacing / 2
		}
		if f.Attrs.AlignContent != AlignContentStart {
			for l := range lines {
				line := &lines[l]
				line.crossOffset += off
				for _, child := range line.child {
					child.crossOffset += off
				}
				off += spacing
			}
		}
	}

	// §9.9.1. Flex Container Intrinsic Main Sizes
	intrinsicMainSize := 0.0
	for _, line := range lines {

		largestMaxContentFlexFraction := -math.MaxFloat64
		for _, child := range line.child {
			// 1. Calculate the max-content flex fraction for each item.
			flexBaseSize := child.flexBaseSize
			maxContentSize := child.mainSize
			var maxContentFlexFraction float64
			if maxContentSize > flexBaseSize {
				// Positive free space, divide by flex grow factor (floored at 1).
				flexGrowFactor := math.Max(1, child.node.Attrs.Grow)
				maxContentFlexFraction = (maxContentSize - flexBaseSize) / flexGrowFactor
			} else {
				// Negative free space, divide by scaled flex shrink factor (floored at 1).
				flexShrinkFactor := math.Max(1, child.node.Attrs.Shrink)
				maxContentFlexFraction = (flexBaseSize - maxContentSize) / (flexBaseSize * flexShrinkFactor)
			}
			child.maxContentFlexFraction = maxContentFlexFraction
			if maxContentFlexFraction > largestMaxContentFlexFraction {
				largestMaxContentFlexFraction = maxContentFlexFraction
			}
		}

		// 2. Add each item’s flex base size to the product of its flex grow/shrink factor and the largest max-content flex fraction.
		for _, child := range line.child {
			var newMainSize float64
			if largestMaxContentFlexFraction > 0 {
				newMainSize = child.flexBaseSize + (child.node.Attrs.Grow * largestMaxContentFlexFraction)
			} else {
				newMainSize = child.flexBaseSize - (child.node.Attrs.Shrink * child.flexBaseSize * largestMaxContentFlexFraction)
			}
			child.mainSize = newMainSize
		}

		// 3. Determine line size and update intrinsicMainSize.
		lineSize := 0.0
		for _, child := range line.child {
			lineSize += child.mainSize
		}
		if lineSize > intrinsicMainSize {
			intrinsicMainSize = lineSize
		}
	}
	f.setMainSize(int(intrinsicMainSize))

	// §9.9.2. Flex Container Intrinsic Cross Sizes
	// The min-content/max-content cross size of a single-line flex container
	// is the largest min-content contribution/max-content contribution (respectively)
	// of its flex items.
	intrinsicCrossSize := 0.0
	for _, line := range lines {
		if intrinsicCrossSize < line.crossOffset+line.crossSize {
			intrinsicCrossSize = line.crossOffset + line.crossSize
		}

		min := math.Inf(1)
		max := -1.
		for _, child := range line.child {
			if child.crossOffset < min {
				min = child.crossOffset
			}
			if max == -1 || child.crossOffset+child.crossSize > max {
				max = child.crossOffset + child.crossSize
			}
		}
		if intrinsicCrossSize < max-min {
			intrinsicCrossSize = max - min
		}
	}
	f.setCrossSize(int(intrinsicCrossSize))

	// TODO: Calculate min-content/max-content cross size for multi-line flex container.
	// For a multi-line flex container, the min-content/max-content cross size is
	// the sum of the flex line cross sizes resulting from sizing the flex container
	// under a cross-axis min-content constraint/max-content constraint (respectively).
	// However, if the flex container is flex-flow: column wrap;, then it’s sized
	// by first finding the largest min-content/max-content cross-size contribution
	// among the flex items (respectively), then using that size as the available
	// space in the cross axis for each of the flex items during layout.

	// Layout complete. Update children position
	for l := range lines {
		line := &lines[l]
		for _, child := range line.child {
			switch f.Attrs.Direction {
			case Row:
				child.node.bounds = image.Rect(
					round(child.mainOffset),
					round(child.crossOffset),
					round(child.mainOffset+child.mainSize),
					round(child.crossOffset+child.crossSize))
				child.node.setFrame(child.node.bounds.Add(f.frame.Min))
			case Column:
				child.node.bounds = image.Rect(
					round(child.crossOffset),
					round(child.mainOffset),
					round(child.crossOffset+child.crossSize),
					round(child.mainOffset+child.mainSize))
				child.node.setFrame(child.node.bounds.Add(f.frame.Min))
			default:
				panic(fmt.Sprint("flex: bad direction ", f.Attrs.Direction))
			}
		}
	}
}

type element struct {
	node                   *View
	flexBaseSize           float64
	mainSize               float64
	mainOffset             float64
	mainMargin             []float64
	crossSize              float64
	crossOffset            float64
	crossMargin            []float64
	frozen                 bool
	maxContentFlexFraction float64
	widthInPct             float64
	heightInPct            float64
}

type flexLine struct {
	mainSize    float64
	crossSize   float64
	crossOffset float64
	child       []*element
}

func (f *View) mainSize(x, y int) int {
	switch f.Attrs.Direction {
	case Row:
		return x
	case Column:
		return y
	default:
		panic(fmt.Sprint("flex: bad direction ", f.Attrs.Direction))
	}
}

func (f *View) setCrossSize(v int) {
	switch f.Attrs.Direction {
	case Row:
		f.calculatedHeight = v
	case Column:
		f.calculatedWidth = v
	default:
		panic(fmt.Sprint("flex: bad direction ", f.Attrs.Direction))
	}
}

func (f *View) setMainSize(v int) {
	switch f.Attrs.Direction {
	case Row:
		f.calculatedWidth = v
	case Column:
		f.calculatedHeight = v
	default:
		panic(fmt.Sprint("flex: bad direction ", f.Attrs.Direction))
	}
}

func (f *View) isCrossSizeFixed(v *View) bool {
	switch f.Attrs.Direction {
	case Row:
		return v.isHeightFixed()
	case Column:
		return v.isWidthFixed()
	default:
		panic(fmt.Sprint("flex: bad direction ", f.Attrs.Direction))
	}
}

func (f *View) crossSize(x, y int) int {
	switch f.Attrs.Direction {
	case Row:
		return y
	case Column:
		return x
	default:
		panic(fmt.Sprint("flex: bad direction ", f.Attrs.Direction))
	}
}

func (f *View) mainMargin(c *View) []float64 {
	switch f.Attrs.Direction {
	case Row:
		return []float64{
			float64(c.Attrs.MarginLeft),
			float64(c.Attrs.MarginRight)}
	case Column:
		return []float64{
			float64(c.Attrs.MarginTop),
			float64(c.Attrs.MarginBottom)}
	default:
		panic("unreachable")
	}
}

func (f *View) crossMargin(c *View) []float64 {
	switch f.Attrs.Direction {
	case Row:
		return []float64{
			float64(c.Attrs.MarginTop),
			float64(c.Attrs.MarginBottom)}
	case Column:
		return []float64{
			float64(c.Attrs.MarginLeft),
			float64(c.Attrs.MarginRight)}
	default:
		panic("unreachable")
	}
}

func (f *View) flexBaseSize(c *View) int {
	w := c.Attrs.Width
	if w == 0 {
		w = c.calculatedWidth
	}
	h := c.Attrs.Height
	if h == 0 {
		h = c.calculatedHeight
	}
	return f.mainSize(w, h)
}

func (f *View) clampSize(size, width, height int) int {
	minSize := f.mainSize(width, height)
	if minSize > size {
		size = minSize
	}
	if size < 0 {
		return 0
	}
	return size
}

func round(f float64) int {
	return int(math.Floor(f + .5))
}
