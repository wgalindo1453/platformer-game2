package gameobjects

import (
	"fmt"
	"platformer-game/rendering"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)
var gameOver bool // Variable to track game over state

type ZombieState int

const (
	ZombieIdle ZombieState = iota
	ZombieWalking
	ZombieAttacking
	ZombieHurt
	ZombieDead
)

const (
	stateSwitchDelay = 3 * time.Second // Delay between state switches
	frameDelay       = 5000          // Default delay for frame updates
	attackRange      = 50.0            // Range within which zombie will attack the player
	followRange      = 300.0           // Range within which zombie will follow the player
)


var lastIdleSoundTime time.Time // Global cooldown for zombie idle sound
var isIdleSoundPlaying bool     // Global flag to check if idle sound is currently playing

const idleSoundCooldown = 5 * time.Second // Cooldown duration for the idle sound
const idleSoundProximityRange = 200       // Range within which idle sound plays


type Zombie struct {
	Position        rl.Vector2
	Speed           rl.Vector2
	Width, Height   float32
	Color           rl.Color
	FacingRight     bool             // Direction the zombie is facing
	State           ZombieState      // Current animation state
	FrameCounter    int              // Counter to control frame switch timing
	CurrentFrame    int              // Current frame index for animation
	IdleFrames      []rl.Texture2D   // Frames for idle animation
	WalkFrames      []rl.Texture2D   // Frames for walking animation
	AttackingFrames []rl.Texture2D   // Frames for attacking animation
	HurtFrames 	[]rl.Texture2D   // Frames for hurt animation
	DeadFrames 	[]rl.Texture2D   // Frames for dead animation
	LastSwitch      time.Time        // Timer for switching states
	Health          int              // Health points
    IsAlive         bool             // Whether zombie is alive

	// Sounds
    ClawSound       rl.Sound
    HurtSound       rl.Sound
    DeathSound      rl.Sound
	IdleSound       rl.Sound
	IdleSoundCooldown time.Time        // Cooldown timer for idle sound

}

// Initializing  zombie with default settings and load frames for animations
func InitZombie(x, y float32, zombieType int) Zombie {
	spriteSheet := rendering.LoadSpriteSheet("assets/sprites/zombiespritesheet1girl_processed.png")
	spriteSheet2 := rendering.LoadSpriteSheet("assets/sprites/zombiespritesheet2girl_processed.png")

	// Load sounds for zombie actions
	clawSound := rl.LoadSound("assets/sounds/zombie_attack.mp3")
	hurtSound := rl.LoadSound("assets/sounds/zombie_hurt.mp3")
	deathSound := rl.LoadSound("assets/sounds/zombie_death.mp3")
	idleSound := rl.LoadSound("assets/sounds/zombie_idle.mp3")

	// animation frames
	idleFrames := []rl.Rectangle{
		{X: 233, Y: 67, Width: 55, Height: 99}, //frame 1
		{X: 385, Y: 67, Width: 55, Height: 99}, //frame 2
		{X: 540, Y: 67, Width: 56, Height: 99},//frame 3
		{X: 694, Y: 67, Width: 59, Height: 99},//frame 4
		{X: 844, Y: 67, Width: 59, Height: 99},//frame 5
		{X: 1000, Y: 67, Width: 60, Height: 99},//frame 6
		{X: 1150, Y: 67, Width: 57, Height: 99},//frame 7
	}

	walkFrames := []rl.Rectangle{
		{X: 229, Y: 243, Width: 67, Height: 108},//frame 1
		{X: 380, Y: 244, Width: 72, Height: 107},//frame 2
		{X: 536, Y: 243, Width: 70, Height: 108},//frame 3
		{X: 702, Y: 241, Width: 55, Height: 110},//frame 4
		{X: 837, Y: 241, Width: 73, Height: 110},//frame 5
		{X: 1000, Y: 241, Width: 66, Height: 110},//frame 6
		{X: 1150, Y: 241, Width: 68, Height: 110},//frame 7
		{X: 1308, Y: 241, Width: 64, Height: 110},//frame 8
	}

	//attacking frames
	attackingFrames := []rl.Rectangle{
		{X: 241, Y: 56, Width: 56, Height: 110},//frame 1
		{X: 387, Y: 54, Width: 51, Height: 112},//frame 2
		{X: 544, Y: 58, Width: 80, Height: 108},//frame 3
		{X: 698, Y: 58, Width: 72, Height: 108},//frame 4
		{X: 837, Y: 59, Width: 71, Height: 107},//frame 5
	}

	//hurt frames
	hurtFrames := []rl.Rectangle{
		{X: 229, Y: 596, Width: 61, Height: 101},
		{X: 383, Y: 598, Width: 63, Height: 99},
		{X: 537, Y: 598, Width: 58, Height: 99},
	}
	//Dead frames
	deadFrames := []rl.Rectangle{
		{X: 200, Y: 772, Width: 106, Height: 94},
		{X: 358, Y: 775, Width: 106, Height: 91},
		{X: 516, Y: 832, Width: 124, Height: 34},
		{X: 667, Y: 834, Width: 124, Height: 32},
	}





	var idleTextures, walkTextures, attackingTextures, hurtTextures, deadTextures []rl.Texture2D
	for _, frame := range idleFrames {
		idleTextures = append(idleTextures, spriteSheet.ImageAt(frame, rl.Blank))
	}
	for _, frame := range walkFrames {
		walkTextures = append(walkTextures, spriteSheet.ImageAt(frame, rl.Blank))
	}
	for _, frame := range attackingFrames {
		attackingTextures = append(attackingTextures, spriteSheet2.ImageAt(frame, rl.Blank))
	}
	for _, frame := range hurtFrames {
		hurtTextures = append(hurtTextures, spriteSheet2.ImageAt(frame, rl.Blank))
	}
	for _, frame := range deadFrames {
		deadTextures = append(deadTextures, spriteSheet2.ImageAt(frame, rl.Blank))
	}
	
	return Zombie{
		Position:        rl.Vector2{X: x, Y: y},
		Speed:           rl.Vector2{X: 0.05, Y: 0},
		Width:           113,
		Height:          113,
		Color:           rl.Green,
		FacingRight:     true,
		State:           ZombieIdle,
		IdleFrames:      idleTextures,
		WalkFrames:      walkTextures,
		AttackingFrames: attackingTextures,
		HurtFrames:      hurtTextures,
		LastSwitch:      time.Now(),
		DeadFrames:      deadTextures,
		Health:          100, // Set zombie health
        IsAlive:         true,

		// Assign loaded sounds
        ClawSound:       clawSound,
        HurtSound:       hurtSound,
        DeathSound:      deathSound,
		IdleSound:       idleSound, // Assign idle sound
	}
}

// TakeDamage reduces the zombie's health by the specified amount, sets it to hurt or dead if health reaches zero
func (z *Zombie) TakeDamage(damage int) {
    z.Health -= damage
    if z.Health <= 0 {
        z.Health = 0
        z.setState(ZombieDead)
        z.IsAlive = false
        if !rl.IsSoundPlaying(z.DeathSound) {
            rl.PlaySound(z.DeathSound)
        }
    } else {
        z.setState(ZombieHurt)
        if !rl.IsSoundPlaying(z.HurtSound) {
            rl.PlaySound(z.HurtSound)
        }
    }
}


// Updating zombie behavior to follow and attack player if within range
func (z *Zombie) Update(worldWidth int, playerPosition rl.Vector2) {
    if z.State == ZombieDead && z.CurrentFrame >= len(z.DeadFrames)-1 {
        // Hold the last death frame, marking the zombie as inactive
        z.IsAlive = false
        return
    }

	// Calculating distance to player for behavior
	distanceToPlayer := rl.Vector2Distance(z.Position, playerPosition)

	if z.State == ZombieAttacking && distanceToPlayer <= attackRange {
		if !rl.IsSoundPlaying(z.ClawSound) {
            rl.PlaySound(z.ClawSound)
        }
		//stop other sounds
		rl.StopSound(z.IdleSound)
		// Reduce player health when attacked
		if PlayerInstance.Health > 0 {
			PlayerInstance.Health -= 0.001 // Adjust damage as needed
			if PlayerInstance.Health <= 0 {
				PlayerInstance.Health = 0
				if PlayerInstance.IsGameOver() {
					fmt.Println("Game Over: Player Health is 0")
				}
			}
		}
	}

	// Checking if the zombie's health has reached zero, setting it to dead if so
	if z.Health <= 0 && z.IsAlive {
		z.setState(ZombieDead)
		z.IsAlive = false // Start death animation but zombie is marked inactive
		return
	}
    if z.IsAlive {
        switch {
        case distanceToPlayer <= attackRange:
            z.setState(ZombieAttacking)
        case distanceToPlayer <= followRange:
            z.setState(ZombieWalking)
			//print th edistance to player
			//print the idleSoundProximityRange
			if distanceToPlayer <= idleSoundProximityRange && !isIdleSoundPlaying && time.Since(lastIdleSoundTime) > idleSoundCooldown {
                rl.PlaySound(z.IdleSound)
                lastIdleSoundTime = time.Now() // Reset global cooldown timer
                isIdleSoundPlaying = true      // Set idle sound as currently playing
            }
            if playerPosition.X < z.Position.X {
                z.FacingRight = false
                z.Speed.X = -0.02 // Slower speed for zombie movement
            } else {
                z.FacingRight = true
                z.Speed.X = 0.02
            }
            z.Position.X += z.Speed.X
        default:
            // Randomly switch between idle and walking if outside follow range
            if time.Since(z.LastSwitch) > stateSwitchDelay {
                if z.State == ZombieIdle {
                    z.setState(ZombieWalking)
                } else {
					//play idle sound
					
                    z.setState(ZombieIdle)
                }
                z.LastSwitch = time.Now()
            }

            // Manages edge flipping in walking state
            if z.State == ZombieWalking {
                if z.Position.X < 0 || z.Position.X > float32(worldWidth)-z.Width {
                    z.Speed.X = -z.Speed.X
                    z.FacingRight = !z.FacingRight
                }
                z.Position.X += z.Speed.X
            }
        }
    }

	if isIdleSoundPlaying && distanceToPlayer > idleSoundProximityRange {
        rl.StopSound(z.IdleSound)
        isIdleSoundPlaying = false
    }
}


// Helper method to set zombie state and reset frame data
func (z *Zombie) setState(state ZombieState) {
    if z.State != state {
        // Stop sounds as needed
        if state == ZombieDead {
            rl.StopSound(z.ClawSound) // Stop attack sound if zombie dies
			rl.StopSound(z.IdleSound) // Stop idle sound if zombie dies
        }
		
        
        z.State = state
        z.CurrentFrame = 0
        z.FrameCounter = 0
    }
}
func (z *Zombie) UnloadSounds() {
    rl.UnloadSound(z.ClawSound)
    rl.UnloadSound(z.HurtSound)
    rl.UnloadSound(z.DeathSound)
	rl.UnloadSound(z.IdleSound) 
}

// Drawing zombie based on the current frame and state
func (z *Zombie) Draw() {
    var frames []rl.Texture2D
    switch z.State {
    case ZombieWalking:
        frames = z.WalkFrames
    case ZombieAttacking:
        frames = z.AttackingFrames
    case ZombieHurt:
        frames = z.HurtFrames
    case ZombieDead:
        frames = z.DeadFrames
    default:
        frames = z.IdleFrames
    }

    if len(frames) > 0 {
        frame := frames[z.CurrentFrame]
        z.FrameCounter++

        // Differentiate frame timing for the death state
        if z.State == ZombieDead {
            if z.FrameCounter >= frameDelay / 20 { // Slower death frame rate
                if z.CurrentFrame < len(frames)-1 {
                    z.CurrentFrame++
                }
                z.FrameCounter = 0
            }
        } else {
            // Standard frame delay for all other states
            if z.FrameCounter >= frameDelay / 10 {
                z.CurrentFrame = (z.CurrentFrame + 1) % len(frames)
                z.FrameCounter = 0
            }
        }

        // Source rectangle setup for animation and flipping
        sourceRect := rl.Rectangle{X: 0, Y: 0, Width: float32(frame.Width), Height: float32(frame.Height)}
        if !z.FacingRight {
            sourceRect.Width = -sourceRect.Width
        }

        // Drawing the current frame
        destinationRect := rl.Rectangle{
            X:      z.Position.X,
            Y:      z.Position.Y,
            Width:  z.Width,
            Height: z.Height,
        }
        if frame.ID != 0 {
            rl.DrawTexturePro(frame, sourceRect, destinationRect, rl.Vector2{X: z.Width / 2, Y: z.Height / 2}, 0, z.Color)
        }
    }
}
