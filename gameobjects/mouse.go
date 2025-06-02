package gameobjects

import (
	"math/rand"
	"time"

	"platformer-game/rendering"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// MouseState defines the various states for the mouse NPC.
type MouseState int

const (
	MouseIdle MouseState = iota
	MouseWalking
	MouseJumping
	MouseAttacking
	MouseSpecial
)

// Mouse represents an NPC with states, animations, and sounds.
type Mouse struct {
	Position        rl.Vector2
	Speed           rl.Vector2
	Width, Height   float32
	State           MouseState
	CurrentFrame    int
	FrameCounter    int
	LastStateChange time.Time
	NextStateChange time.Time // New field for fixed state duration

	IdleFrames    []rl.Texture2D
	WalkFrames    []rl.Texture2D
	JumpFrames    []rl.Texture2D
	AttackFrames  []rl.Texture2D
	SpecialFrames []rl.Texture2D

	IdleSound    rl.Sound
	WalkSound    rl.Sound
	JumpSound    rl.Sound
	AttackSound  rl.Sound
	SpecialSound rl.Sound
}

// NewMouse creates a new mouse NPC at the given position.
// It loads frames from the mouse spritesheet and sounds from assets.
func NewMouse(x, y float32) *Mouse {
	m := &Mouse{
		Position:        rl.NewVector2(x, y),
		Speed:           rl.NewVector2(0, 0),
		State:           MouseIdle,
		CurrentFrame:    0,
		FrameCounter:    0,
		LastStateChange: time.Now(),
		Width:           20, // set to the appropriate width
		Height:          12, // set to the appropriate height
	}

	// Load the spritesheet for the mouse NPC.
	spriteSheet := rendering.LoadSpriteSheet("assets/sprites/mousespritesheet1.png")

	// --- Load Idle Frames ---
	// Using updated positions and sizes from CSS:
	// Frame 1: no-repeat -104px -53px; width:20, height:12
	// Frame 2: no-repeat -171px -53px; width:20, height:12
	// Frame 3: no-repeat -238px -53px; width:20, height:12
	// Frame 4: no-repeat -304px -53px; width:20, height:12
	// Frame 5: no-repeat -370px -53px; width:20, height:12
	// Frame 6: no-repeat -437px -53px; width:19, height:12
	// Frame 7: no-repeat -503px -53px; width:20, height:12
	// Frame 8: no-repeat -570px -53px; width:20, height:12
	idleRects := []rl.Rectangle{
		{X: 104, Y: 53, Width: 20, Height: 12},
		{X: 171, Y: 53, Width: 20, Height: 12},
		{X: 238, Y: 53, Width: 20, Height: 12},
		{X: 304, Y: 53, Width: 20, Height: 12},
		{X: 370, Y: 53, Width: 20, Height: 12},
		{X: 437, Y: 53, Width: 19, Height: 12},
		{X: 503, Y: 53, Width: 20, Height: 12},
		{X: 570, Y: 53, Width: 20, Height: 12},
	}
	for _, rect := range idleRects {
		m.IdleFrames = append(m.IdleFrames, spriteSheet.ImageAt(rect, rl.Blank))
	}

	// --- Load Walking Frames ---
	// Frames from:
	// no-repeat -91px -133px; width:31, height:12
	// no-repeat -157px -133px; width:30, height:11
	// no-repeat -224px -133px; width:30, height:11
	// no-repeat -292px -133px; width:29, height:11
	// no-repeat -358px -133px; width:29, height:11
	// no-repeat -426px -133px; width:27, height:11
	walkRects := []rl.Rectangle{
		{X: 91, Y: 133, Width: 31, Height: 12},
		{X: 157, Y: 133, Width: 30, Height: 11},
		{X: 224, Y: 133, Width: 30, Height: 11},
		{X: 292, Y: 133, Width: 29, Height: 11},
		{X: 358, Y: 133, Width: 29, Height: 11},
		{X: 426, Y: 133, Width: 27, Height: 11},
	}
	for _, rect := range walkRects {
		m.WalkFrames = append(m.WalkFrames, spriteSheet.ImageAt(rect, rl.Blank))
	}

	// --- Load Jumping Frames ---
	// Frames from:
	// no-repeat -90px -210px; width:31, height:11
	// no-repeat -164px -210px; width:30, height:11
	// no-repeat -227px -208px; width:32, height:13
	// no-repeat -294px -202px; width:31, height:19
	// no-repeat -364px -198px; width:30, height:20
	// no-repeat -429px -194px; width:34, height:12
	// no-repeat -490px -190px; width:29, height:22
	// no-repeat -562px -195px; width:23, height:25
	jumpRects := []rl.Rectangle{
		{X: 90, Y: 210, Width: 31, Height: 11},
		{X: 164, Y: 210, Width: 30, Height: 11},
		{X: 227, Y: 208, Width: 32, Height: 13},
		{X: 294, Y: 202, Width: 31, Height: 19},
		{X: 364, Y: 198, Width: 30, Height: 20},
		{X: 429, Y: 194, Width: 34, Height: 12},
		{X: 490, Y: 190, Width: 29, Height: 22},
		{X: 562, Y: 195, Width: 23, Height: 25},
	}
	for _, rect := range jumpRects {
		m.JumpFrames = append(m.JumpFrames, spriteSheet.ImageAt(rect, rl.Blank))
	}

	// --- Load Attacking Frames ---
	// Frames from:
	// no-repeat -101px -280px; width:28, height:14
	// no-repeat -163px -282px; width:28, height:12
	// no-repeat -228px -276px; width:31, height:17
	// no-repeat -294px -277px; width:35, height:17
	// no-repeat -361px -281px; width:33, height:13
	attackRects := []rl.Rectangle{
		{X: 101, Y: 280, Width: 28, Height: 14},
		{X: 163, Y: 282, Width: 28, Height: 12},
		{X: 228, Y: 276, Width: 31, Height: 17},
		{X: 294, Y: 277, Width: 35, Height: 17},
		{X: 361, Y: 281, Width: 33, Height: 13},
	}
	for _, rect := range attackRects {
		m.AttackFrames = append(m.AttackFrames, spriteSheet.ImageAt(rect, rl.Blank))
	}

	// --- Load Special Frames ---
	// Frames from:
	// no-repeat -102px -355px; width:20, height:12
	// no-repeat -166px -353px; width:21, height:14
	// no-repeat -232px -349px; width:24, height:18
	// no-repeat -296px -344px; width:25, height:23
	// no-repeat -364px -340px; width:24, height:27
	// no-repeat -431px -340px; width:22, height:27
	// no-repeat -506px -344px; width:13, height:23
	// no-repeat -566px -349px; width:13, height:17
	specialRects := []rl.Rectangle{
		{X: 102, Y: 355, Width: 20, Height: 12},
		{X: 166, Y: 353, Width: 21, Height: 14},
		{X: 232, Y: 349, Width: 24, Height: 18},
		{X: 296, Y: 344, Width: 25, Height: 23},
		{X: 364, Y: 340, Width: 24, Height: 27},
		{X: 431, Y: 340, Width: 22, Height: 27},
		{X: 506, Y: 344, Width: 13, Height: 23},
		{X: 566, Y: 349, Width: 13, Height: 17},
	}
	for _, rect := range specialRects {
		m.SpecialFrames = append(m.SpecialFrames, spriteSheet.ImageAt(rect, rl.Blank))
	}

	// --- Load Sounds for each state ---
	m.IdleSound = rl.LoadSound("assets/sounds/mouse_idle.mp3")
	m.WalkSound = rl.LoadSound("assets/sounds/mouse_walk.mp3")
	m.JumpSound = rl.LoadSound("assets/sounds/mouse_jump.mp3")
	m.AttackSound = rl.LoadSound("assets/sounds/mouse_attack.mp3")
	m.SpecialSound = rl.LoadSound("assets/sounds/mouse_special.mp3")

	return m
}

// Update handles the mouse AI by switching states and updating animations.
// For demonstration, it randomly changes state every 2-4 seconds.
func (m *Mouse) Update(worldWidth, worldHeight float32) {
	// Debug print.
	//fmt.Println("Mouse State:", m.State)
	//fmt.Println("Mouse Position:", m.Position)
	//fmt.Println("Mouse Speed:", m.Speed)
	// --- Constrain Position ---
	if m.Position.X < 0 {
		m.Position.X = 0
	} else if m.Position.X > worldWidth-m.Width {
		m.Position.X = worldWidth - m.Width
	}

	if m.Position.Y < 0 {
		m.Position.Y = 0
	} else if m.Position.Y > worldHeight-m.Height {
		m.Position.Y = worldHeight - m.Height
		m.Speed.Y = 0
	}

	// --- State Switching ---
	if time.Now().After(m.NextStateChange) {
		newState := MouseState(rand.Intn(5)) // Random state from 0 to 4.
		if newState != m.State {
			m.State = newState
			m.CurrentFrame = 0
			m.FrameCounter = 0
			m.LastStateChange = time.Now()
			// Set the next state change time.
			m.NextStateChange = time.Now().Add(time.Duration(rand.Intn(2000)+2000) * time.Millisecond)

			switch m.State {
			case MouseIdle:
				rl.PlaySound(m.IdleSound)
				m.Speed = rl.NewVector2(0, 0)
			case MouseWalking:
				rl.PlaySound(m.WalkSound)
				// Use a very slow horizontal speed.
				m.Speed = rl.NewVector2(float32(rand.Intn(3)-1)*0.05, 0)
			case MouseJumping:
				rl.PlaySound(m.JumpSound)
				// Uncomment and adjust if you want an initial upward velocity:
				// m.Speed.Y = -2.0
			case MouseAttacking:
				rl.PlaySound(m.AttackSound)
				m.Speed = rl.NewVector2(0, 0)
			case MouseSpecial:
				rl.PlaySound(m.SpecialSound)
				m.Speed = rl.NewVector2(0, 0)
			}
		}
	}

	// --- State-specific Movement ---
	if m.State == MouseJumping {
		// Instead of applying gravity and modifying Y,
		// simply move forward (in the X direction).
		m.Position.X += m.Speed.X

		// Optionally, after a fixed duration in the Jumping state, switch back to Idle.
		// This prevents the mouse from remaining in the Jumping state forever.
		if time.Since(m.LastStateChange) > 1*time.Second {
			m.State = MouseIdle
			m.CurrentFrame = 0
			m.FrameCounter = 0
			m.NextStateChange = time.Now().Add(time.Duration(rand.Intn(2000)+2000) * time.Millisecond)
		}
	}

	if m.State == MouseWalking {
		m.Position.X += m.Speed.X
	}

	// --- Boundary Checking ---
	if m.Position.X < 0 {
		m.Position.X = 0
		m.Speed.X = -m.Speed.X // Optionally reverse direction.
	}
	if m.Position.X > worldWidth-m.Width {
		m.Position.X = worldWidth - m.Width
		m.Speed.X = -m.Speed.X
	}
	if m.Position.Y < 0 {
		m.Position.Y = 0
		m.Speed.Y = 0
	}
	if m.Position.Y > worldHeight-m.Height {
		m.Position.Y = worldHeight - m.Height
		m.Speed.Y = 0
	}

	// --- Update Animation Frames ---
	m.FrameCounter++
	var frameDelay int
	switch m.State {
	case MouseIdle:
		frameDelay = 300
	case MouseWalking:
		frameDelay = 150
	case MouseJumping:
		frameDelay = 200
	case MouseAttacking:
		frameDelay = 100
	case MouseSpecial:
		frameDelay = 250
	}

	if m.FrameCounter >= frameDelay/50 {
		m.CurrentFrame++
		m.FrameCounter = 0
		switch m.State {
		case MouseIdle:
			if m.CurrentFrame >= len(m.IdleFrames) {
				m.CurrentFrame = 0
			}
		case MouseWalking:
			if m.CurrentFrame >= len(m.WalkFrames) {
				m.CurrentFrame = 0
			}
		case MouseJumping:
			if m.CurrentFrame >= len(m.JumpFrames) {
				m.CurrentFrame = len(m.JumpFrames) - 1 // Hold on last frame.
			}
		case MouseAttacking:
			if m.CurrentFrame >= len(m.AttackFrames) {
				m.CurrentFrame = 0
				m.State = MouseIdle
				m.NextStateChange = time.Now().Add(time.Duration(rand.Intn(2000)+2000) * time.Millisecond)
			}
		case MouseSpecial:
			if m.CurrentFrame >= len(m.SpecialFrames) {
				m.CurrentFrame = 0
				m.State = MouseIdle
				m.NextStateChange = time.Now().Add(time.Duration(rand.Intn(2000)+2000) * time.Millisecond)
			}
		}
	}
}

// Draw renders the current frame of the mouse based on its state.
func (m *Mouse) Draw() {
	var frame rl.Texture2D
	switch m.State {
	case MouseIdle:
		frame = m.IdleFrames[m.CurrentFrame]
	case MouseWalking:
		frame = m.WalkFrames[m.CurrentFrame]
	case MouseJumping:
		frame = m.JumpFrames[m.CurrentFrame]
	case MouseAttacking:
		frame = m.AttackFrames[m.CurrentFrame]
	case MouseSpecial:
		frame = m.SpecialFrames[m.CurrentFrame]
	default:
		frame = m.IdleFrames[m.CurrentFrame]
	}

	// Draw the current frame at the mouse's position.
	rl.DrawTexture(frame, int32(m.Position.X), int32(m.Position.Y), rl.White)
}

// Unload releases all textures and sounds associated with the mouse.
func (m *Mouse) Unload() {
	for _, tex := range m.IdleFrames {
		rl.UnloadTexture(tex)
	}
	for _, tex := range m.WalkFrames {
		rl.UnloadTexture(tex)
	}
	for _, tex := range m.JumpFrames {
		rl.UnloadTexture(tex)
	}
	for _, tex := range m.AttackFrames {
		rl.UnloadTexture(tex)
	}
	for _, tex := range m.SpecialFrames {
		rl.UnloadTexture(tex)
	}
	rl.UnloadSound(m.IdleSound)
	rl.UnloadSound(m.WalkSound)
	rl.UnloadSound(m.JumpSound)
	rl.UnloadSound(m.AttackSound)
	rl.UnloadSound(m.SpecialSound)
}
