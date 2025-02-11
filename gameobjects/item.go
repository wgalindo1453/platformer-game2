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
    // fmt.Println("Drawing item:", item.Name, "at position:", item.Position)
    rl.DrawTexture(item.Texture, int32(item.Position.X), int32(item.Position.Y), rl.White)
}
