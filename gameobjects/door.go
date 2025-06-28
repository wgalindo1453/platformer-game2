// gameobjects/door.go
package gameobjects

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"platformer-game/rendering"
	"time"
)

// DoorState is the current state of the door animation.
type DoorState int

const (
	DoorClosed DoorState = iota
	DoorOpening
	DoorOpen
)

type Door struct {
	ID            string
	Position      rl.Vector2     // top‐left corner in world coordinates
	Frames        []rl.Texture2D // door frames (closed → open)
	State         DoorState
	CurrentFrame  int           // index into Frames
	lastFrameTime time.Time     // to throttle frame updates
	FrameDelay    time.Duration // e.g. 100ms between frames

	Width  float32
	Height float32

	// Context menu fields
	MenuOpen     bool       // true if the context menu is visible
	MenuPosition rl.Vector2 // where to draw the menu (usually at mouse pos)

	// Callback for when "Leave" is clicked
	OnLeaveClicked func()
}

// NewAnimatedDoor lets you specify exactly which sub‐rectangles to pull from the spritesheet.
//   - id: unique door ID (e.g. "BronzeKey")
//   - x, y: world position to draw the door
//   - sheetPath: path to the PNG containing all door frames scattered anywhere
//   - frameRects: a slice of rl.Rectangle, one per animation frame (closed→open). Each rect is in pixels.
//   - delayMs: milliseconds between each animation frame when opening.
//
// It will load each frame by calling sheet.ImageAt(rect, rl.Blank) for every rect in frameRects.
func NewAnimatedDoor(
	id string,
	x, y float32,
	sheetPath string,
	frameRects []rl.Rectangle,
	delayMs int,
) *Door {
	// 1) Load the full spritesheet as both Texture and Image
	sheet := rendering.LoadSpriteSheet(sheetPath)

	// 2) Manually extract each frame using the provided rectangles
	allFrames := make([]rl.Texture2D, 0, len(frameRects))
	for _, rect := range frameRects {
		tex := sheet.ImageAt(rect, rl.Blank)
		allFrames = append(allFrames, tex)
	}

	// 3) Assume all frames share the same dimensions; grab from first frame
	w := float32(allFrames[0].Width)
	h := float32(allFrames[0].Height)

	return &Door{
		ID:            id,
		Position:      rl.NewVector2(x, y),
		Frames:        allFrames,
		State:         DoorClosed,
		CurrentFrame:  0,
		lastFrameTime: time.Now(),
		FrameDelay:    time.Millisecond * time.Duration(delayMs),
		Width:         w,
		Height:        h,
	}
}

// TryUnlock transitions a closed door into the "opening" animation.
func (d *Door) TryUnlock() {
	if d.State == DoorClosed {
		d.State = DoorOpening
		d.CurrentFrame = 0
		d.lastFrameTime = time.Now()
	}
}

// Update advances the opening animation when enough time has passed.
// Once the final frame is reached, State switches to DoorOpen.
func (d *Door) Update() {
	if d.State != DoorOpening {
		return
	}

	// Only move to the next frame if FrameDelay has elapsed
	if time.Since(d.lastFrameTime) < d.FrameDelay {
		return
	}

	d.lastFrameTime = time.Now()
	d.CurrentFrame++
	if d.CurrentFrame >= len(d.Frames) {
		// Reached last frame ⇒ fully open
		d.CurrentFrame = len(d.Frames) - 1
		d.State = DoorOpen
	}
}

// Draw renders whichever frame is appropriate:
//   - closed: always draw Frames[0]
//   - opening: draw Frames[CurrentFrame]
//   - open: draw Frames[last index]
func (d *Door) Draw() {
	var tex rl.Texture2D
	switch d.State {
	case DoorClosed:
		tex = d.Frames[0]
	case DoorOpening, DoorOpen:
		tex = d.Frames[d.CurrentFrame]
	}

	rl.DrawTexture(
		tex,
		int32(d.Position.X),
		int32(d.Position.Y),
		rl.White,
	)
}

// DrawContextMenu draws the context menu if it's open
func (d *Door) DrawContextMenu() {
	if !d.MenuOpen {
		return
	}

	const (
		menuItemWidth  = 80
		menuItemHeight = 20
		padding        = 4
	)
	bx := int32(d.MenuPosition.X)
	by := int32(d.MenuPosition.Y)

	// Draw "Leave" option
	rl.DrawRectangle(bx, by, menuItemWidth, menuItemHeight, rl.DarkGray)
	rl.DrawText("Leave", bx+16, by+4, 12, rl.White)
}

// CheckCollision returns true if the player's rectangle overlaps the door's rectangle.
// You can use this in core.UpdateGame to block movement when the door isn't yet open.
func (d *Door) CheckCollision(px, py, pw, ph float32) bool {
	doorRect := rl.Rectangle{
		X:      d.Position.X,
		Y:      d.Position.Y,
		Width:  d.Width,
		Height: d.Height,
	}
	playerRect := rl.Rectangle{
		X:      px,
		Y:      py,
		Width:  pw,
		Height: ph,
	}
	return rl.CheckCollisionRecs(doorRect, playerRect)
}

// HandleMouse handles mouse interactions with the door
func (d *Door) HandleMouse(playerPos rl.Vector2, playerWidth, playerHeight float32, camera rl.Camera2D) {
	mousePos := rl.GetMousePosition()
	mx, my := mousePos.X, mousePos.Y

	// Convert mouse position from screen coordinates to world coordinates
	worldMousePos := rl.GetScreenToWorld2D(mousePos, camera)
	worldMouseX := worldMousePos.X
	worldMouseY := worldMousePos.Y

	// Check if mouse is over the door
	mouseOverDoor := worldMouseX >= d.Position.X && worldMouseX <= d.Position.X+d.Width &&
		worldMouseY >= d.Position.Y && worldMouseY <= d.Position.Y+d.Height

	// Debug output
	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		fmt.Printf("Door %s: State=%d, MouseOver=%v, PlayerNear=%v, MousePos=(%.1f,%.1f), WorldPos=(%.1f,%.1f), DoorPos=(%.1f,%.1f)\n",
			d.ID, d.State, mouseOverDoor, d.CheckCollision(playerPos.X, playerPos.Y, playerWidth, playerHeight),
			mx, my, worldMouseX, worldMouseY, d.Position.X, d.Position.Y)
	}

	// If the context menu is open, handle clicks on it first
	if d.MenuOpen && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		d.handleMenuClick(mx, my)
		return
	}

	// If right-click on door and no menu open, open context menu
	if rl.IsMouseButtonPressed(rl.MouseRightButton) && !d.MenuOpen && mouseOverDoor {
		// Only show menu if door is open (temporarily removed player proximity check for testing)
		if d.State == DoorOpen {
			d.MenuOpen = true
			d.MenuPosition = rl.NewVector2(mx, my)
			fmt.Printf("Opening context menu for door %s\n", d.ID)
		}
	}

	// Close menu if clicking elsewhere
	if d.MenuOpen && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Check if click is outside the menu
		const menuItemWidth = 80
		const menuItemHeight = 20
		const padding = 4
		const menuHeight = menuItemHeight + padding + menuItemHeight

		menuRect := rl.Rectangle{
			X:      d.MenuPosition.X,
			Y:      d.MenuPosition.Y,
			Width:  menuItemWidth,
			Height: menuHeight,
		}

		if mx < menuRect.X || mx > menuRect.X+menuRect.Width ||
			my < menuRect.Y || my > menuRect.Y+menuRect.Height {
			d.MenuOpen = false
		}
	}
}

// handleMenuClick processes clicks on the context menu
func (d *Door) handleMenuClick(mx, my float32) {
	const (
		menuItemWidth  = 80
		menuItemHeight = 20
		padding        = 4
	)
	bx := int32(d.MenuPosition.X)
	by := int32(d.MenuPosition.Y)

	// "Leave" option
	leaveRect := rl.Rectangle{
		X:      float32(bx),
		Y:      float32(by),
		Width:  menuItemWidth,
		Height: menuItemHeight,
	}

	// Check if "Leave" was clicked
	if mx >= leaveRect.X && mx <= leaveRect.X+leaveRect.Width &&
		my >= leaveRect.Y && my <= leaveRect.Y+leaveRect.Height {
		// Signal that the player wants to leave through this door
		// This will be handled by the core game logic
		d.MenuOpen = false
		// You can add a callback or signal here to notify the core game
		if d.OnLeaveClicked != nil {
			d.OnLeaveClicked()
		}
	}
}
