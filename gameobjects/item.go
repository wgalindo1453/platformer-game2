package gameobjects

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type WorldItem struct {
	Position rl.Vector2
	Texture  rl.Texture2D
	Type     ItemType // Referencing ItemType from Inventory
	Name     string
}

func NewWorldItem(x, y float32, itemType ItemType, name string, texturePath string) WorldItem {
	return WorldItem{
		Position: rl.NewVector2(x, y),
		Texture:  rl.LoadTexture(texturePath),
		Type:     itemType,
		Name:     name,
	}
}

func (item *WorldItem) Draw() {
	scale := float32(1.0)
	switch item.Type {
	case HealthPack:
		scale = 0.5
	case KeyType:
		scale = 0.4
		// you can add more custom scales here (e.g. Weapon = 0.8, etc.)
	}
	rl.DrawTextureEx(item.Texture,
		rl.Vector2{X: item.Position.X, Y: item.Position.Y},
		0, scale, rl.White)
}
