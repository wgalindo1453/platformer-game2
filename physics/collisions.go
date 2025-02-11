package physics

import rl "github.com/gen2brain/raylib-go/raylib"

// Checks if the player is on the ground
func IsOnGround(position rl.Vector2, height float32) bool {
	groundY := float32(450) // Defining ground level based on screen size
	return position.Y+height >= groundY
}
