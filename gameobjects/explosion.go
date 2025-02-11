package gameobjects

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"time"
)

type Explosion struct {
	Position    rl.Vector2
	Texture     rl.Texture2D
	IsActive    bool
	StartTime   time.Time
	Duration    time.Duration
}

// Initialize a new explosion
func NewExplosion(x, y float32, texture rl.Texture2D) Explosion {
	return Explosion{
		Position:  rl.NewVector2(x, y),
		Texture:   texture,
		IsActive:  true,
		StartTime: time.Now(),
		Duration:  500 * time.Millisecond, // Explosion lasts for 0.5 seconds
	}
}

// Update the explosion status based on elapsed time
func (e *Explosion) Update() {
	if time.Since(e.StartTime) > e.Duration {
		e.IsActive = false // Deactivate the explosion after the duration
	}
}

// Draw the explosion if active
func (e *Explosion) Draw() {
	if e.IsActive {
		rl.DrawTexture(e.Texture, int32(e.Position.X), int32(e.Position.Y), rl.White)
	}
}
