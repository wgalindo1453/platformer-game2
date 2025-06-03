package gameobjects

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"log"
	"platformer-game/database" // Add this line
	"platformer-game/rendering"
	"time"
)

type PlayerState int

const (
	Idle PlayerState = iota
	Walking
	Running
	Shooting
	Sitting
	SittingShooting
	Jumping
	Resting
	Sleeping
	Dying
	ThrowingGrenade
	Reloading
)
const (
	jumpVelocity = -5.0 // Initial upward velocity for jumping
	gravity      = 800  // Gravity value, pulling player down each frame
	groundYPos   = 0    // The ground level, adjust to your world height
)

type Player struct {
	Position              rl.Vector2
	Speed                 rl.Vector2
	Acceleration          rl.Vector2
	Width, Height         float32
	Color                 rl.Color
	FacingRight           bool           // Direction the player is facing
	CurrentFrame          int            // Current frame index for animation
	FrameCounter          int            // Counter to control frame switch timing
	State                 PlayerState    // Current animation state
	IdleTimer             time.Time      // Timer for idle state
	RestTimer             time.Time      // Timer for resting state
	WalkFrames            []rl.Texture2D // Frames for walking animation
	RunFrames             []rl.Texture2D // Frames for running animation
	IdleFrames            []rl.Texture2D // Frames for idle animation
	ShootFrames           []rl.Texture2D // Frames for shooting animation
	ReloadingFrames       []rl.Texture2D // Frames for reloading animation
	SittingFrames         []rl.Texture2D // Frames for sitting animation
	SittingShootingFrames []rl.Texture2D // Frames for sitting shooting animation
	JumpFrames            []rl.Texture2D // Frames for jumping animation
	RestingFrames         []rl.Texture2D // Frames for resting animation
	SleepingFrames        []rl.Texture2D // Frames for sleeping animation
	DyingFrames           []rl.Texture2D // Frames for dying animation
	GrenadeFrames         []rl.Texture2D // Frames for throwing grenade animation
	Bullets               []*Bullet      // Add bullets slice
	Explosions            []*Explosion   // Slice to hold active explosions
	ExplosionTex          rl.Texture2D   // Texture for the explosion
	switchDown            bool           // Indicates when to start descending
	throwingFinishedTime  time.Time      // Track when grenade throw animation finishes
	threwGrenade          bool           // Track if grenade was thrown
	Ammo                  int            // Current ammo count
	MaxAmmo               int            // Maximum ammo capacity
	IsReloading           bool           // Flag to check if reloading

	// Sounds
	WalkSound      rl.Sound
	RunSound       rl.Sound
	ShootSound     rl.Sound
	ReloadSound    rl.Sound
	EmptyClipSound rl.Sound
	GrenadeExplode rl.Sound

	// New attributes
	Health    float64 // Player health
	MaxHealth float64 // Maximum health to keep track for the health bar
	Inventory Inventory
	HeldItem  Item // The currently held item

	UsedKeyID string // if non‐empty, means “player just used this key”
}

func (p *Player) UpdateHeldItem() {
	if p.Inventory.Slots[p.Inventory.SelectedSlot].Type != Other {
		p.HeldItem = p.Inventory.Slots[p.Inventory.SelectedSlot]
	} else {
		p.HeldItem = Item{} // No item held if slot is empty
	}
}
func (inv *Inventory) SaveToDB() {
	db := database.DB
	for i, item := range inv.Slots {
		_, err := db.Exec(`
			INSERT OR REPLACE INTO inventory (slot, type, name)
			VALUES (?, ?, ?);`,
			i, item.Type, item.Name)
		if err != nil {
			log.Println("Failed to save inventory item:", err)
		}
	}
}

// EquipItem is called when the user right-clicks “Equip” or “Use” on slotIndex.
func (p *Player) EquipItem(slotIndex int) {
	if slotIndex < 0 || slotIndex >= p.Inventory.MaxSlots {
		return
	}
	it := p.Inventory.Slots[slotIndex]

	switch it.Type {
	case Weapon:
		p.HeldItem = it
		fmt.Printf("Equipped weapon: %s\n", it.Name)
		p.Inventory.Slots[slotIndex] = Item{Type: Other}
		p.Inventory.SaveToDB()

	case HealthPack:
		healAmount := 25.0
		p.Health += healAmount
		if p.Health > p.MaxHealth {
			p.Health = p.MaxHealth
		}
		fmt.Printf("Used health pack: healed %.0f, now at %.0f/%.0f\n",
			healAmount, p.Health, p.MaxHealth)
		p.Inventory.Slots[slotIndex] = Item{Type: Other}
		p.Inventory.SaveToDB()

	case KeyType:
		// Instead of calling core.UnlockDoor here, just record “I used key X”:
		p.UsedKeyID = it.Name // e.g. “BronzeKey”
		fmt.Printf("Used key %q (will notify core to unlock)\n", it.Name)
		p.Inventory.Slots[slotIndex] = Item{Type: Other}
		p.Inventory.SaveToDB()

	default:
		// Other/nothing
	}
}

func (p *Player) Shoot() {
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if p.Inventory.IsOpen || p.Inventory.MenuOpen {
			return
		}
		//if p.State != ThrowingGrenade {
		//	p.setState(ThrowingGrenade)
		//	if !rl.IsSoundPlaying(p.GrenadeExplode) {
		//		rl.PlaySound(p.GrenadeExplode)
		//	}
		//	p.threwGrenade = false // Reset for next throw
		//}
		if p.Ammo > 0 && !p.IsReloading {
			bulletPosition := p.Position
			bulletPosition.Y += p.Height / 2 // Adjust to shoot from the middle
			newBullet := NewBullet(bulletPosition.X, bulletPosition.Y, 10, p.FacingRight)
			p.Bullets = append(p.Bullets, newBullet)

			p.Ammo-- // Reduce ammo when shooting

			if !rl.IsSoundPlaying(p.ShootSound) {
				rl.PlaySound(p.ShootSound)
			}
		} else if p.Ammo == 0 && !p.IsReloading {
			// Play empty clip sound if out of ammo
			if !rl.IsSoundPlaying(p.EmptyClipSound) {
				rl.PlaySound(p.EmptyClipSound)
			}

			// Set shooting state but only display the first frame
			p.setState(Shooting)
			p.CurrentFrame = 0 // Always show the first frame when out of ammo
		}
	}
}
func (p *Player) IsGameOver() bool {

	return p.Health <= 0
}

func (p *Player) Unload() {
	for _, frame := range p.WalkFrames {
		rl.UnloadTexture(frame)
	}
	for _, frame := range p.RunFrames {
		rl.UnloadTexture(frame)
	}
	for _, frame := range p.IdleFrames {
		rl.UnloadTexture(frame)
	}
	for _, frame := range p.ShootFrames {
		rl.UnloadTexture(frame)
	}
	// Unload sounds
	rl.UnloadSound(p.WalkSound)
	rl.UnloadSound(p.RunSound)
	rl.UnloadSound(p.ShootSound)
}

var PlayerInstance Player

func InitPlayer(worldWidth, worldHeight int) {

	rl.InitAudioDevice() // Initialize audio device
	PlayerInstance = Player{
		Position:     rl.NewVector2(100, float32(worldHeight)-55), // Start at the bottom of the world
		Speed:        rl.NewVector2(0, 0),
		Acceleration: rl.NewVector2(0, 0.5),
		Width:        113,
		Height:       113,
		Color:        rl.White,
		CurrentFrame: 0,
		FrameCounter: 0,
		State:        Idle,
		FacingRight:  true,
		Health:       100,              // Initialize with full health
		MaxHealth:    100,              // Set maximum health
		Inventory:    NewInventory(10), // Initialize with 10 slots
		Ammo:         30,               // Set starting ammo
		MaxAmmo:      30,               // Max ammo capacity
		IsReloading:  false,            // Initialize reloading state
	}
	// Load sounds
	PlayerInstance.WalkSound = rl.LoadSound("assets/sounds/walking.mp3")
	PlayerInstance.RunSound = rl.LoadSound("assets/sounds/running.mp3")
	PlayerInstance.ShootSound = rl.LoadSound("assets/sounds/machineguneffect.wav")
	PlayerInstance.ReloadSound = rl.LoadSound("assets/sounds/reload.mp3")
	PlayerInstance.EmptyClipSound = rl.LoadSound("assets/sounds/emptyclip.mp3")
	PlayerInstance.GrenadeExplode = rl.LoadSound("assets/sounds/grenade_explosion.mp3")

	/***********************************LOAD SPRITES*********************************************** */

	// Load sprite sheet
	spriteSheet := rendering.LoadSpriteSheet("assets/sprites/shooterspritesheet.png")
	spriteSheet2 := rendering.LoadSpriteSheet("assets/sprites/shooterspritesheet2.png")
	spriteSheet3 := rendering.LoadSpriteSheet("assets/sprites/shooterspritesheet3.png")
	spriteSheet4 := rendering.LoadSpriteSheet("assets/sprites/shooterspritesheet4.png")

	// Load explosion frame from spritesheet4
	explosionRect := rl.Rectangle{X: 1600, Y: 346, Width: 157, Height: 93}
	PlayerInstance.ExplosionTex = spriteSheet4.ImageAt(explosionRect, rl.Blank) // Load the explosion texture	// Load walking frames
	walkingFrames := []rl.Rectangle{
		{X: 309, Y: 301, Width: 63, Height: 136},  // Frame 1
		{X: 500, Y: 301, Width: 66, Height: 136},  // Frame 2
		{X: 690, Y: 303, Width: 72, Height: 134},  // Frame 3
		{X: 878, Y: 302, Width: 72, Height: 136},  // Frame 4
		{X: 1075, Y: 299, Width: 70, Height: 138}, // Frame 5
	}
	for _, frame := range walkingFrames {
		PlayerInstance.WalkFrames = append(PlayerInstance.WalkFrames, spriteSheet.ImageAt(frame, rl.Blank))
	}

	// Load running frames
	runningFrames := []rl.Rectangle{
		{X: 267, Y: 525, Width: 76, Height: 122},  // Frame 1
		{X: 456, Y: 535, Width: 78, Height: 122},  // Frame 2
		{X: 644, Y: 535, Width: 84, Height: 122},  // Frame 3
		{X: 840, Y: 525, Width: 80, Height: 122},  // Frame 4
		{X: 1042, Y: 533, Width: 68, Height: 124}, // Frame 5
	}
	for _, frame := range runningFrames {
		PlayerInstance.RunFrames = append(PlayerInstance.RunFrames, spriteSheet.ImageAt(frame, rl.Blank))
	}

	// Load idle frames
	idleFrames := []rl.Rectangle{
		{X: 296, Y: 71, Width: 94, Height: 134},  // Frame 1
		{X: 488, Y: 71, Width: 94, Height: 134},  // Frame 2
		{X: 681, Y: 69, Width: 94, Height: 136},  // Frame 3
		{X: 873, Y: 69, Width: 94, Height: 136},  // Frame 4
		{X: 1063, Y: 69, Width: 94, Height: 136}, // Frame 5
		{X: 1256, Y: 71, Width: 93, Height: 134}, // Frame 6
	}
	for _, frame := range idleFrames {
		PlayerInstance.IdleFrames = append(PlayerInstance.IdleFrames, spriteSheet.ImageAt(frame, rl.Blank))
	}

	// Load shooting frames
	shooting1Frames := []rl.Rectangle{
		{X: 294, Y: 739, Width: 95, Height: 130},  // Frame 1
		{X: 487, Y: 739, Width: 108, Height: 130}, // Frame 2
		{X: 677, Y: 739, Width: 125, Height: 130}, // Frame 3
		{X: 869, Y: 739, Width: 102, Height: 131}, // Frame 4
		{X: 300, Y: 951, Width: 102, Height: 130}, // Frame 5
		{X: 492, Y: 951, Width: 111, Height: 130}, // Frame 6
		{X: 683, Y: 951, Width: 130, Height: 130}, // Frame 7
		{X: 877, Y: 951, Width: 106, Height: 130}, // Frame 8
	}
	for _, frame := range shooting1Frames {
		PlayerInstance.ShootFrames = append(PlayerInstance.ShootFrames, spriteSheet.ImageAt(frame, rl.Blank))
	}

	// Load Reloading frames
	reloadingFrames := []rl.Rectangle{
		{X: 306, Y: 73, Width: 80, Height: 135},  // Frame 1
		{X: 499, Y: 47, Width: 54, Height: 161},  // Frame 2
		{X: 691, Y: 47, Width: 55, Height: 162},  // Frame 3
		{X: 884, Y: 47, Width: 75, Height: 161},  // Frame 4
		{X: 1074, Y: 47, Width: 54, Height: 160}, // Frame 5
		{X: 1265, Y: 47, Width: 53, Height: 159}, // Frame 6 (New)
		{X: 1457, Y: 47, Width: 57, Height: 159}, // Frame 7 (New)
	}

	for _, frame := range reloadingFrames {
		PlayerInstance.ReloadingFrames = append(PlayerInstance.ReloadingFrames, spriteSheet3.ImageAt(frame, rl.Blank))
	}

	// Load Sitting frames
	sittingFrames := []rl.Rectangle{
		{X: 234, Y: 82, Width: 75, Height: 88}, // Frame 1
		{X: 394, Y: 83, Width: 75, Height: 87}, // Frame 2
		{X: 555, Y: 85, Width: 75, Height: 86}, // Frame 3
	}
	for _, frame := range sittingFrames {
		PlayerInstance.SittingFrames = append(PlayerInstance.SittingFrames, spriteSheet2.ImageAt(frame, rl.Blank))
	}

	//Load Sitting Shooting frames
	sittingShootingFrames := []rl.Rectangle{
		{X: 242, Y: 275, Width: 85, Height: 89},  // Frame 1
		{X: 399, Y: 275, Width: 84, Height: 89},  // Frame 2
		{X: 560, Y: 275, Width: 110, Height: 89}, // Frame 3
	}
	for _, frame := range sittingShootingFrames {
		PlayerInstance.SittingShootingFrames = append(PlayerInstance.SittingShootingFrames, spriteSheet2.ImageAt(frame, rl.Blank))
	}

	//Jumping frames
	jumpingFrames := []rl.Rectangle{
		{X: 240, Y: 444, Width: 78, Height: 103}, // Frame 1
		{X: 401, Y: 450, Width: 80, Height: 96},  // Frame 2
		{X: 561, Y: 434, Width: 79, Height: 113}, // Frame 3
		{X: 722, Y: 444, Width: 78, Height: 98},  // Frame 4
		{X: 1043, Y: 457, Width: 68, Height: 89}, // Frame 5

	}
	for _, frame := range jumpingFrames {
		PlayerInstance.JumpFrames = append(PlayerInstance.JumpFrames, spriteSheet2.ImageAt(frame, rl.Blank))
	}

	//Resting frames
	restingFrames := []rl.Rectangle{
		{X: 240, Y: 621, Width: 78, Height: 102}, // Frame 1
		{X: 400, Y: 626, Width: 78, Height: 97},  // Frame 2
		{X: 559, Y: 644, Width: 71, Height: 79},  // Frame 3
		{X: 686, Y: 651, Width: 87, Height: 72},  // Frame 4
	}

	for _, frame := range restingFrames {
		PlayerInstance.RestingFrames = append(PlayerInstance.RestingFrames, spriteSheet2.ImageAt(frame, rl.Blank))
	}

	//Sleeping frames
	sleepingFrames := []rl.Rectangle{
		{X: 231, Y: 864, Width: 113, Height: 32}, // Frame 1
		{X: 390, Y: 847, Width: 115, Height: 49}, // Frame 2
		{X: 541, Y: 825, Width: 124, Height: 71}, // Frame 3
		{X: 711, Y: 864, Width: 114, Height: 32}, // Frame 4
		{X: 869, Y: 863, Width: 114, Height: 33}, // Frame 5

	}
	for _, frame := range sleepingFrames {
		PlayerInstance.SleepingFrames = append(PlayerInstance.SleepingFrames, spriteSheet2.ImageAt(frame, rl.Blank))
	}

	//Dying frames
	dyingFrames := []rl.Rectangle{
		{X: 315, Y: 952, Width: 92, Height: 128},
		{X: 504, Y: 943, Width: 94, Height: 137},
		{X: 651, Y: 984, Width: 128, Height: 96},
		{X: 814, Y: 1041, Width: 160, Height: 39},
	}

	for _, frame := range dyingFrames {
		PlayerInstance.DyingFrames = append(PlayerInstance.DyingFrames, spriteSheet3.ImageAt(frame, rl.Blank))
	}

	//Grenade frames
	grenadeFrames := []rl.Rectangle{
		{X: 294, Y: 300, Width: 72, Height: 141},
		{X: 477, Y: 301, Width: 82, Height: 140},
		{X: 686, Y: 299, Width: 67, Height: 140},
		{X: 874, Y: 300, Width: 71, Height: 139},
		{X: 1040, Y: 299, Width: 94, Height: 140},
		{X: 1251, Y: 307, Width: 64, Height: 133},
		{X: 1444, Y: 312, Width: 117, Height: 127},
	}

	for _, frame := range grenadeFrames {
		PlayerInstance.GrenadeFrames = append(PlayerInstance.GrenadeFrames, spriteSheet4.ImageAt(frame, rl.Blank))
	}
	//{X: 1600, Y: 346, Width: 157, Height: 93}, //this is start of explosion frames

}

// Method to throw a grenade and create an explosion
func (p *Player) ThrowGrenade() {
	// Check if enough time has passed since the last grenade throw
	timeSinceThrow := time.Since(p.throwingFinishedTime)
	//check if end of frames

	// Only allow throwing a grenade if 5 seconds have passed
	if timeSinceThrow.Seconds() > 5 && !p.threwGrenade {
		fmt.Println("Throwing grenade...")
		if !rl.IsSoundPlaying(p.GrenadeExplode) {
			rl.PlaySound(p.GrenadeExplode)
		}
		p.threwGrenade = false // Reset threwGrenade for the next throw

		// Calculate explosion position in front of the player
		explosionX := p.Position.X + 50 // Adjust distance as desired
		if !p.FacingRight {
			explosionX = p.Position.X - 50
		}
		explosionY := p.Position.Y + p.Height/2 // Adjust height for ground level

		// Create the explosion object and add it to the player's explosions
		explosion := NewExplosion(explosionX, explosionY, p.ExplosionTex)
		p.Explosions = append(p.Explosions, &explosion)

		// Reset the throwingFinishedTime to the current time for cooldown
		p.throwingFinishedTime = time.Now()

		fmt.Println("Grenade thrown! Cooldown started.")
	} else {
		// If cooldown is still active, notify player or prevent action
		fmt.Println("Grenade on cooldown. Time remaining:", 5-timeSinceThrow.Seconds(), "seconds")
	}
}

/***********************************STATES*********************************************** */

func (p *Player) setState(state PlayerState) {
	if p.State != state {
		p.State = state
		p.CurrentFrame = 0
		p.FrameCounter = 0
	}

	// Reset timers when changing to idle, resting, or sleeping states
	if state == Idle {
		//set reloading to false
		p.IsReloading = false
		p.IdleTimer = time.Now()
		p.RestTimer = time.Time{}
	} else if state == Resting {
		p.RestTimer = time.Now()
	} else {
		p.IdleTimer = time.Time{}
		p.RestTimer = time.Time{}
	}
}

/***********************************UPDATE*********************************************** */

func (p *Player) Update(worldHeight int, worldWidth int, zombies []*Zombie) {
	// fmt.Println("players starting out y position: ", p.Position.Y)
	if p.State == Reloading {
		fmt.Println("currents state is reloading")
		if p.CurrentFrame >= len(p.ReloadingFrames)-1 {
			fmt.Println("current frame is greater than or equal to the length of reloading frames")

			p.Ammo = p.MaxAmmo // Refill ammo
			p.IsReloading = false
			p.setState(Idle)
			rl.StopSound(p.ReloadSound)
			p.FrameCounter = 0
		} else {
			p.FrameCounter++
		}
	}

	// Update bullets
	for _, bullet := range p.Bullets {
		if bullet.IsActive {
			bullet.Update()

			// Here we are checking if bullet hits any zombie
			for _, zombie := range zombies {
				if zombie.IsAlive && rl.CheckCollisionPointCircle(bullet.Position, zombie.Position, zombie.Width/2) {
					zombie.TakeDamage(20) // Adjust damage as needed
					bullet.IsActive = false
					break
				}
			}

			// Deactivate bullet if it goes out of bounds
			if bullet.Position.X < 0 || bullet.Position.X > float32(worldWidth) {
				bullet.IsActive = false
			}
		}
	}
	// Update explosions
	for i := len(p.Explosions) - 1; i >= 0; i-- {
		p.Explosions[i].Update()
		if !p.Explosions[i].IsActive {
			p.Explosions = append(p.Explosions[:i], p.Explosions[i+1:]...)
		}
	}

	//print inventory
	//if inventory is not empty then print the item inside
	if len(p.Inventory.Slots) != 0 {
		// fmt.Println("Inventory:", p.Inventory)

	}

	// Filter out inactive bullets
	activeBullets := p.Bullets[:0]
	for _, bullet := range p.Bullets {
		if bullet.IsActive {
			activeBullets = append(activeBullets, bullet)
		}
	}
	p.Bullets = activeBullets
	// Check if player is on the ground
	//onGround := p.Position.Y >= float32(worldHeight)-p.Height
	//
	//// Apply gravity and handle jumping
	//if !onGround || p.State == Jumping {
	//	if p.Speed.Y < 0 && !p.switchDown { // Ascending phase
	//		if p.Speed.Y >= -1.0 { // Nearing the peak of jump
	//			p.switchDown = true
	//		}
	//		p.Speed.Y += gravity * 0.0005 // Gradual deceleration while ascending
	//	} else { // Descending phase
	//		p.Speed.Y += gravity * 0.02 // Faster descent for natural gravity
	//	}
	//
	//	// Update the player's vertical position
	//	p.Position.Y += p.Speed.Y
	//}

	// Handle jump input
	//if rl.IsKeyPressed(rl.KeySpace) && onGround {
	//	p.setState(Jumping)
	//	p.Speed.Y = -6.5 // Set initial upward velocity
	//	p.switchDown = false
	//}

	//// If player is grounded, reset jump state
	//if p.Position.Y >= float32(worldHeight)-p.Height {
	//	p.Position.Y = float32(worldHeight) - p.Height
	//	p.Speed.Y = 0
	//	p.switchDown = false // Reset for next jump
	//	if p.State == Jumping {
	//		p.setState(Idle) // Reset to Idle after landing
	//	}
	//}

	// Player state logic based on key inputs, prioritizing crouching
	switch {
	case rl.IsKeyDown(rl.KeyR):
		fmt.Println("still have ammo: ", p.Ammo)

		if p.State != Reloading && p.Ammo < p.MaxAmmo {
			fmt.Println("Reloading...")
			rl.PlaySound(p.ReloadSound)
			p.setState(Reloading)
			p.IsReloading = true
			p.Speed.X = 0
			rl.StopSound(p.WalkSound)
			rl.StopSound(p.RunSound)
			rl.StopSound(p.ShootSound)
		}

	case rl.IsMouseButtonDown(rl.MouseRightButton) && p.State != Sitting && p.State != SittingShooting:
		// Trigger grenade throw when holding down the left mouse button
		p.setState(ThrowingGrenade)
		//set reloading to false
		p.IsReloading = false
		p.FacingRight = true // Adjust if needed based on player orientation
		p.Speed.X = 0
		rl.StopSound(p.WalkSound)
	case rl.IsKeyDown(rl.KeyLeftControl):
		// Crouching has priority, halts forward movement
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			if p.Inventory.IsOpen || p.Inventory.MenuOpen {
				return
			}

			//check if out of ammo
			if p.Ammo == 0 {
				fmt.Println("Out of ammo")
				p.CurrentFrame = 0 // Lock shooting animation to first frame
				rl.StopSound(p.ShootSound)
				return
			}

			fmt.Println("Sitting and shooting")
			p.setState(SittingShooting)
			p.Shoot() // Call shoot when sitting and shooting
			//call shoot method simul
			p.Speed.X = 0 // Halt horizontal movement
			rl.StopSound(p.WalkSound)

			if !rl.IsSoundPlaying(p.ShootSound) {
				rl.PlaySound(p.ShootSound)
			}
			//stop walking sound
			rl.StopSound(p.WalkSound)

		} else {
			p.setState(Sitting)
			p.Speed.X = 0 // Halt horizontal movement
			rl.StopSound(p.WalkSound)
		}

		// When initiating the jump, set a lower initial speed
	//case rl.IsKeyPressed(rl.KeySpace) && onGround:
	//	// Jump initiation
	//	p.setState(Jumping)
	//	p.Speed.Y = -2.0 // Lower initial jump speed for a shorter jump
	//	fmt.Println("Jumping")
	//
	//	// Apply gravity and handle jumping
	//	if !onGround || p.State == Jumping {
	//		// Print player's position for debugging
	//		fmt.Println(p.Position.Y)
	//
	//		if p.Speed.Y < 0 && !p.switchDown { // Ascending
	//			fmt.Println("Ascending")
	//			if p.Speed.Y > -0.1 {
	//				fmt.Println("Switching down")
	//				p.switchDown = true
	//			}
	//			p.Speed.Y += p.Acceleration.Y * 0.001 // Maintain slow upward deceleration
	//		} else { // Descending
	//			p.Speed.Y += p.Acceleration.Y * 0.1 // Slightly faster but controlled descent
	//		}
	//		p.Position.Y += p.Speed.Y
	//	}
	//
	//	// If player lands on the ground, reset to Idle and reset switchDown
	//	if p.Position.Y >= float32(worldHeight)-p.Height {
	//		p.Position.Y = float32(worldHeight) - p.Height
	//		p.Speed.Y = 0
	//		p.switchDown = false // Reset switchDown for the next jump
	//		if p.State == Jumping {
	//			p.setState(Idle) // Reset to Idle after landing
	//		}
	//	}
	//	// Check for reloading
	case rl.IsMouseButtonDown(rl.MouseLeftButton) && p.State != Sitting && p.State != SittingShooting:
		if p.Inventory.IsOpen || p.Inventory.MenuOpen {
			break
		}
		if p.State == Shooting && p.Ammo == 0 {
			fmt.Println("Out of ammo")
			p.CurrentFrame = 0 // Lock shooting animation to first frame
			rl.StopSound(p.ShootSound)
		} else {
			// Shooting (no horizontal movement)
			p.setState(Shooting)
			p.Speed.X = 0
			if !rl.IsSoundPlaying(p.ShootSound) {
				rl.PlaySound(p.ShootSound)
			}
			//stop walking sound
			rl.StopSound(p.WalkSound)
			//stop running sound
			rl.StopSound(p.RunSound)
		}

	case rl.IsKeyDown(rl.KeyD) && rl.IsKeyDown(rl.KeyLeftShift) && p.State != Shooting && p.State != Sitting:
		// Running (right) if not shooting or crouching
		p.setState(Running)
		p.FacingRight = true
		p.Speed.X = 0.2
		if !rl.IsSoundPlaying(p.RunSound) {
			rl.PlaySound(p.RunSound)
		}
		rl.StopSound(p.WalkSound)

	case rl.IsKeyDown(rl.KeyD) && p.State != Shooting && p.State != Sitting && p.State != SittingShooting:
		// Walking (right) if not shooting or crouching
		p.setState(Walking)
		p.FacingRight = true
		p.Speed.X = 0.05
		if !rl.IsSoundPlaying(p.WalkSound) {
			rl.PlaySound(p.WalkSound)
		}
		rl.StopSound(p.RunSound)

	case rl.IsKeyDown(rl.KeyA) && rl.IsKeyDown(rl.KeyLeftShift) && p.State != Shooting && p.State != Sitting:
		// Running (left) if not shooting or crouching
		p.setState(Running)
		p.FacingRight = false
		p.Speed.X = -0.2
		if !rl.IsSoundPlaying(p.RunSound) {
			rl.PlaySound(p.RunSound)
		}
		rl.StopSound(p.WalkSound)

	case rl.IsKeyDown(rl.KeyA) && p.State != Shooting && p.State != Sitting && p.State != SittingShooting:
		// Walking (left) if not shooting or crouching
		p.setState(Walking)
		p.FacingRight = false
		p.Speed.X = -0.05
		if !rl.IsSoundPlaying(p.WalkSound) {
			rl.PlaySound(p.WalkSound)
		}
		rl.StopSound(p.RunSound)

	case p.State != Resting && p.State != Sleeping:
		// Idle if no movement
		p.setState(Idle)
		rl.StopSound(p.ReloadSound)
		p.Speed.X = 0
		rl.StopSound(p.WalkSound)
		rl.StopSound(p.RunSound)
		rl.StopSound(p.ShootSound)
	}

	if !rl.IsMouseButtonDown(rl.MouseLeftButton) {
		rl.StopSound(p.ShootSound)
	}

	// Update horizontal position
	p.Position.X += p.Speed.X

	// this is to constrain player within screen bounds (X-axis)
	if p.Position.X < 0 {
		p.Position.X = 0
	} else if p.Position.X > float32(worldWidth)-p.Width {
		p.Position.X = float32(worldWidth) - p.Width
	}

	// this is to Ensure player doesn't sink below ground level (Y-axis)
	//if p.Position.Y >= float32(worldHeight-128) {
	//	p.Position.Y = float32(worldHeight - 128)
	//	p.Speed.Y = 0
	//}
	// Updating animation frames based on state of the player
	p.FrameCounter++
	var frames []rl.Texture2D
	frameDelay := 300
	switch p.State {
	case Walking:
		frames = p.WalkFrames
	case Running:
		frames = p.RunFrames
	case Shooting:
		frames = p.ShootFrames
	case Reloading:
		frames = p.ReloadingFrames
		frameDelay = 1000
	case Sitting:
		frames = p.SittingFrames
	case SittingShooting:
		frames = p.SittingShootingFrames
	case Jumping:
		frames = p.JumpFrames
		frameDelay = 500
	case Resting:
		frames = p.RestingFrames
		frameDelay = 5000
	case Sleeping:
		frames = p.SleepingFrames
		frameDelay = 5000
	case Dying:
		frames = p.DyingFrames
		frameDelay = 10000
	case ThrowingGrenade:
		frames = p.GrenadeFrames
		frameDelay = 800
		//only set to true if on last frame
		if p.CurrentFrame == len(p.GrenadeFrames)-1 {
			fmt.Println("Grenade thrown")
			p.ThrowGrenade()
			p.threwGrenade = true
			p.setState(Idle) // Ensure transition happens only after the throw
		}

	default:
		frames = p.IdleFrames
	}

	if p.threwGrenade && p.FrameCounter >= frameDelay {
		p.setState(Idle)
	}

	// Only update frame based on delay
	if len(frames) > 0 && p.FrameCounter >= frameDelay {
		p.CurrentFrame = (p.CurrentFrame + 1) % len(frames)
		p.FrameCounter = 0
	}
}

/***********************************DRAW*********************************************** */

func (p *Player) Draw() {

	var frame rl.Texture2D
	switch p.State {
	case Walking:
		frame = p.WalkFrames[p.CurrentFrame]
	case Running:

		frame = p.RunFrames[p.CurrentFrame]
	case Shooting:

		frame = p.ShootFrames[p.CurrentFrame]
	case Sitting:

		frame = p.SittingFrames[p.CurrentFrame]
	case SittingShooting:

		frame = p.SittingShootingFrames[p.CurrentFrame]
	case Reloading:
		fmt.Println("Reloading frame: ", p.CurrentFrame)
		frame = p.ReloadingFrames[p.CurrentFrame]
	case Jumping:

		frame = p.JumpFrames[p.CurrentFrame]
	case Resting:

		frame = p.RestingFrames[p.CurrentFrame]
	case Sleeping:

		frame = p.SleepingFrames[p.CurrentFrame]
	case Dying:

		frame = p.DyingFrames[p.CurrentFrame]
	case ThrowingGrenade:

		frame = p.GrenadeFrames[p.CurrentFrame]
	default:

		frame = p.IdleFrames[p.CurrentFrame]
	}

	if p.HeldItem.Type != Other && p.HeldItem.Image.ID != 0 {
		heldX := p.Position.X - 10                                                           // Adjust for desired position relative to player
		heldY := p.Position.Y - 10                                                           // Adjust for desired position relative to player
		rl.DrawTextureEx(p.HeldItem.Image, rl.Vector2{X: heldX, Y: heldY}, 0, 0.5, rl.White) // Scale to desired size
	}

	// Source rectangle starts normally
	sourceRect := rl.Rectangle{X: 0, Y: 0, Width: float32(frame.Width), Height: float32(frame.Height)}
	// Flip the source width to achieve a horizontal flip
	if !p.FacingRight {
		sourceRect.Width = -sourceRect.Width
	}

	// Destination rectangle keeps player position and scale
	destinationRect := rl.Rectangle{
		X:      p.Position.X,
		Y:      p.Position.Y,
		Width:  p.Width,
		Height: p.Height,
	}

	// Draw the current frame with adjusted sourceRect for flipping
	if frame.ID != 0 { // Ensure the frame texture is loaded
		rl.DrawTexturePro(
			frame,
			sourceRect,      // Flipped if FacingRight is false
			destinationRect, // Destination position and size on the screen
			rl.Vector2{X: p.Width / 2, Y: p.Height / 2}, // Origin remains centered
			0,       // No rotation
			p.Color, // Tint color
		)
	}
	// Draw active explosions
	for _, explosion := range p.Explosions {
		explosion.Draw()
	}

	// Drawing bullets
	for _, bullet := range p.Bullets {
		bullet.Draw()
	}
}
