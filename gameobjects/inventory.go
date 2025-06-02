package gameobjects

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"log"
	"platformer-game/database"
)

type ItemType int

const (
	Weapon ItemType = iota
	HealthPack
	Other
)

type Item struct {
	Type  ItemType
	Name  string
	Image rl.Texture2D
}

type Inventory struct {
	Slots        []Item
	MaxSlots     int
	IsOpen       bool
	SelectedSlot int
}

func NewInventory(maxSlots int) Inventory {
	slots := make([]Item, maxSlots)
	for i := range slots {
		slots[i] = Item{Type: Other} // Initialize each slot as empty
	}
	return Inventory{
		Slots:    slots,
		MaxSlots: maxSlots,
	}
}

// Method to add an item to the inventory
func (inv *Inventory) AddItem(item Item) bool {
	for i := 0; i < inv.MaxSlots; i++ {
		if inv.Slots[i].Type == Other {
			inv.Slots[i] = item // Place item in the empty slot
			return true
		}
	}
	return false // Return false if inventory is full
}

func (inv *Inventory) UpdateSelection() {
	slotsPerRow := 5 // Number of slots per row
	if rl.IsKeyPressed(rl.KeyRight) {
		inv.SelectedSlot = (inv.SelectedSlot + 1) % inv.MaxSlots
	}
	if rl.IsKeyPressed(rl.KeyLeft) {
		inv.SelectedSlot = (inv.SelectedSlot - 1 + inv.MaxSlots) % inv.MaxSlots
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		inv.SelectedSlot = (inv.SelectedSlot + slotsPerRow) % inv.MaxSlots
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		inv.SelectedSlot = (inv.SelectedSlot - slotsPerRow + inv.MaxSlots) % inv.MaxSlots
	}
}

func (inv *Inventory) LoadFromDB(itemTextures map[string]rl.Texture2D) {
	db := database.DB
	rows, err := db.Query(`SELECT slot, type, name FROM inventory`)
	if err != nil {
		log.Println("Failed to load inventory from database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var slot int
		var itemType ItemType
		var name string

		if err := rows.Scan(&slot, &itemType, &name); err != nil {
			log.Println("Error scanning inventory row:", err)
			continue
		}

		if slot < 0 || slot >= inv.MaxSlots {
			log.Println("Invalid slot index from database:", slot)
			continue
		}

		// ✅ Check if texture exists
		texture, ok := itemTextures[name]
		if !ok || texture.ID == 0 {
			log.Printf("Missing or invalid texture for item name '%s', skipping slot %d\n", name, slot)
			continue
		}

		// ✅ Assign item to slot only if texture is valid
		inv.Slots[slot] = Item{
			Type:  itemType,
			Name:  name,
			Image: texture,
		}
	}
}

func (inv *Inventory) DrawInventory() {
	invX, invY := 100, 100 // Position of the inventory on the screen
	slotSize := 50
	padding := 10

	for i, item := range inv.Slots {
		x := invX + (i%5)*(slotSize+padding) // Arrange items in a grid
		y := invY + (i/5)*(slotSize+padding)
		rl.DrawRectangle(int32(x), int32(y), int32(slotSize), int32(slotSize), rl.Gray)

		// Draw the slot background with a highlight if it's selected
		if i == inv.SelectedSlot {
			rl.DrawRectangle(int32(x), int32(y), int32(slotSize), int32(slotSize), rl.Yellow) // Highlighted color
		} else {
			rl.DrawRectangle(int32(x), int32(y), int32(slotSize), int32(slotSize), rl.Gray) // Normal color
		}

		if item.Type != Other && item.Image.ID != 0 {
			// Calculate the scale factor to fit the texture within the slot
			textureWidth := float32(item.Image.Width)
			textureHeight := float32(item.Image.Height)
			scale := float32(slotSize) / max(textureWidth, textureHeight) // Scale based on the largest dimension

			// Calculate new width and height to maintain aspect ratio
			drawWidth := int32(textureWidth * scale)
			drawHeight := int32(textureHeight * scale)

			// Center the texture within the slot
			drawX := int32(x) + (int32(slotSize)-drawWidth)/2
			drawY := int32(y) + (int32(slotSize)-drawHeight)/2

			rl.DrawTextureEx(item.Image, rl.Vector2{X: float32(drawX), Y: float32(drawY)}, 0, scale, rl.White)
		}
	}
}

// Helper function to get the max of two float32 values
func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
