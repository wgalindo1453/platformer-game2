package gameobjects

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Bullet struct {
	Position rl.Vector2
	Speed    float32
	Direction rl.Vector2 // Vector indicating direction
	IsActive bool        // Track if the bullet is active
}

// Initialize a new bullet based on the player’s position and facing direction
func NewBullet(x, y, speed float32, facingRight bool) *Bullet {
	direction := rl.NewVector2(1, 0)
	if !facingRight {
		direction.X = -1
	}
	return &Bullet{
		Position:  rl.Vector2{X: x, Y: y},
		Speed:     speed,
		Direction: direction,
		IsActive:  true,
	}
}

// Update bullet position based on its speed and direction
func (b *Bullet) Update() {
	b.Position.X += b.Direction.X * b.Speed
}

func (b *Bullet) Draw() {
	if b.IsActive {
		rl.DrawCircleV(b.Position, 5, rl.Red) 
	}
}
