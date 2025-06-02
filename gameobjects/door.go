package gameobjects

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"platformer-game/rendering"
)

type Door struct {
	Position rl.Vector2
	Texture  rl.Texture2D
	Width    float32
	Height   float32
}

// NewDoor loads the door texture and returns a new Door object.
func NewDoor(x, y float32, texturePath string) *Door {
	// Load the spritesheet from the texture path.
	doorSpriteSheet := rendering.LoadSpriteSheet(texturePath)
	// Extract the door texture using the desired rectangle.
	doorTexture := doorSpriteSheet.ImageAt(rl.Rectangle{X: 30, Y: 19, Width: 41, Height: 58}, rl.White)
	return &Door{
		Position: rl.NewVector2(x, y),
		Texture:  doorTexture,
		Width:    float32(doorTexture.Width),
		Height:   float32(doorTexture.Height),
	}
}

func (d *Door) Draw() {
	rl.DrawTexture(d.Texture, int32(d.Position.X), int32(d.Position.Y), rl.White)
}
