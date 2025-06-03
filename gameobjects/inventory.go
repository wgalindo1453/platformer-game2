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
	KeyType // <-- new
	Other
)

type Item struct {
	Type  ItemType
	Name  string
	Image rl.Texture2D
}

type Inventory struct {
	Slots    []Item
	MaxSlots int
	IsOpen   bool
	// ( no longer need SelectedSlot if everything is mouse‐driven;
	// but going to keep it if you still want keyboard fallback)
	SelectedSlot int

	// --- new fields for drag/drop ---
	Dragging     bool // true while the player is holding an item
	DraggedItem  Item // a copy of the item being dragged
	DraggedIndex int  // where the item came from (so we can restore if needed)

	// ─── New: context menu fields ───
	MenuOpen     bool       // true if the context menu is visible
	MenuSlot     int        // which slot index the menu belongs to
	MenuPosition rl.Vector2 // where to draw the menu (usually at mouse pos)

}

// Helper: for slot index i, return its on‐screen x, y, width and height.
func (inv *Inventory) slotRect(i int) (x, y, w, h int32) {
	invX, invY := 100, 100 // same origin as your DrawInventory
	slotSize := 50
	padding := 10
	cols := 5

	row := i / cols
	col := i % cols

	x = int32(invX + col*(slotSize+padding))
	y = int32(invY + row*(slotSize+padding))
	w = int32(slotSize)
	h = int32(slotSize)
	return
}

func (inv *Inventory) HandleMouse() {
	mousePos := rl.GetMousePosition()
	mx, my := mousePos.X, mousePos.Y

	// ─── 1) If the context menu is open, handle clicks on it first ───
	if inv.MenuOpen && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		inv.handleMenuClick(mx, my)
		return
	}

	// ─── 2) If right‐click on a non‐empty slot and no menu/drag in progress, open context menu ───
	if rl.IsMouseButtonPressed(rl.MouseRightButton) && !inv.MenuOpen && !inv.Dragging {
		for i := 0; i < inv.MaxSlots; i++ {
			x, y, w, h := inv.slotRect(i)
			if mx >= float32(x) && mx <= float32(x+w) &&
				my >= float32(y) && my <= float32(y+h) {

				if inv.Slots[i].Type != Other {
					inv.MenuOpen = true
					inv.MenuSlot = i
					inv.MenuPosition = rl.NewVector2(mx, my)
				}
				break
			}
		}
	}

	// ─── 3) If left‐click to pick up and no drag/menu active, begin dragging ───
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !inv.Dragging && !inv.MenuOpen {
		for i := 0; i < inv.MaxSlots; i++ {
			x, y, w, h := inv.slotRect(i)
			if mx >= float32(x) && mx <= float32(x+w) &&
				my >= float32(y) && my <= float32(y+h) {

				if inv.Slots[i].Type != Other {
					inv.Dragging = true
					inv.DraggedIndex = i
					inv.DraggedItem = inv.Slots[i]
					inv.Slots[i] = Item{Type: Other}
				}
				break
			}
		}
	}

	// ─── 4) If left‐button released while dragging, attempt to drop ‒ then clear drag state ───
	if rl.IsMouseButtonReleased(rl.MouseLeftButton) && inv.Dragging {
		dropped := false

		for j := 0; j < inv.MaxSlots; j++ {
			x, y, w, h := inv.slotRect(j)
			if mx >= float32(x) && mx <= float32(x+w) &&
				my >= float32(y) && my <= float32(y+h) {

				if inv.Slots[j].Type == Other {
					inv.Slots[j] = inv.DraggedItem
					dropped = true
				} else {
					inv.Slots[inv.DraggedIndex], inv.Slots[j] = inv.Slots[j], inv.DraggedItem
					dropped = true
				}
				break
			}
		}

		if !dropped {
			inv.Slots[inv.DraggedIndex] = inv.DraggedItem
		}

		inv.Dragging = false
		inv.DraggedItem = Item{Type: Other}
		inv.DraggedIndex = -1

		// Auto-save entire inventory now that we've moved items:
		inv.SaveToDB()
	}
}

// Invoked when the player left-clicks while the context menu is open.
// mx,my is the mouse release point.
func (inv *Inventory) handleMenuClick(mx, my float32) {
	const (
		menuItemWidth  = 80
		menuItemHeight = 20
		padding        = 4
	)
	bx := int32(inv.MenuPosition.X)
	by := int32(inv.MenuPosition.Y)

	// 1) “Equip” or “Use” (only valid if it’s not Other)
	equipRect := rl.Rectangle{
		X:      float32(bx),
		Y:      float32(by),
		Width:  menuItemWidth,
		Height: menuItemHeight,
	}

	// 2) “Drop” (below “Equip”)
	dropRect := rl.Rectangle{
		X:      float32(bx),
		Y:      float32(by + menuItemHeight + padding),
		Width:  menuItemWidth,
		Height: menuItemHeight,
	}

	// 3) “Delete” (below “Drop”)
	delRect := rl.Rectangle{
		X:      float32(bx),
		Y:      float32(by + 2*(menuItemHeight+padding)),
		Width:  menuItemWidth,
		Height: menuItemHeight,
	}

	clicked := false
	slotIndex := inv.MenuSlot
	slotItem := inv.Slots[slotIndex]

	// If click inside “Equip/Use”
	if mx >= equipRect.X && mx <= equipRect.X+equipRect.Width &&
		my >= equipRect.Y && my <= equipRect.Y+equipRect.Height {

		// Only run EquipItem if it’s a valid type
		if slotItem.Type == Weapon || slotItem.Type == HealthPack || slotItem.Type == KeyType {
			PlayerInstance.EquipItem(slotIndex)
		}
		clicked = true
	}

	// If click inside “Drop”
	if !clicked &&
		mx >= dropRect.X && mx <= dropRect.X+dropRect.Width &&
		my >= dropRect.Y && my <= dropRect.Y+dropRect.Height {

		droppedItem := inv.Slots[slotIndex]
		inv.Slots[slotIndex] = Item{Type: Other}
		// Log it or spawn a world item here:
		log.Printf("Dropped item %q from slot %d\n", droppedItem.Name, slotIndex)
		inv.deleteSlotFromDB(slotIndex)
		inv.SaveToDB()
		clicked = true
	}

	// If click inside “Delete”
	if !clicked &&
		mx >= delRect.X && mx <= delRect.X+delRect.Width &&
		my >= delRect.Y && my <= delRect.Y+delRect.Height {

		inv.Slots[slotIndex] = Item{Type: Other}
		inv.deleteSlotFromDB(slotIndex)
		inv.SaveToDB()
		clicked = true
	}

	// Close menu regardless of where clicked
	inv.MenuOpen = false
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
func (inv *Inventory) deleteSlotFromDB(slotIndex int) {
	// Adjust the table/column names to match your schema
	_, err := database.DB.Exec(`DELETE FROM inventory WHERE slot = ?`, slotIndex)
	if err != nil {
		log.Printf("Error deleting slot %d from DB: %v\n", slotIndex, err)
	}
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
	slotSize := 50

	// 1) Draw each slot and its item (skip drawing the “taken” slot if dragging)
	for i, item := range inv.Slots {
		x, y, w, h := inv.slotRect(i)

		// Base slot background
		rl.DrawRectangle(x, y, w, h, rl.Gray)

		// (Optional) Draw selection highlight if you still want some visual feedback:
		if i == inv.SelectedSlot {
			rl.DrawRectangle(x, y, w, h, rl.Yellow)
		}

		// If we are dragging AND this is the original slot, skip drawing its item
		if inv.Dragging && i == inv.DraggedIndex {
			continue
		}

		// Otherwise, draw the item if present
		if item.Type != Other && item.Image.ID != 0 {
			textureWidth := float32(item.Image.Width)
			textureHeight := float32(item.Image.Height)
			scale := float32(slotSize) / max(textureWidth, textureHeight)

			drawW := int32(textureWidth * scale)
			drawH := int32(textureHeight * scale)
			drawX := x + (w-drawW)/2
			drawY := y + (h-drawH)/2

			rl.DrawTextureEx(item.Image,
				rl.Vector2{X: float32(drawX), Y: float32(drawY)},
				0, scale, rl.White)
		}
	}

	// 2) If dragging, draw the dragged item at the mouse (centered)
	if inv.Dragging && inv.DraggedItem.Type != Other && inv.DraggedItem.Image.ID != 0 {
		mpos := rl.GetMousePosition()
		tex := inv.DraggedItem.Image

		textureWidth := float32(tex.Width)
		textureHeight := float32(tex.Height)
		scale := float32(slotSize) / max(textureWidth, textureHeight)

		drawW := int32(textureWidth * scale)
		drawH := int32(textureHeight * scale)

		// Offset so the texture is centered under the cursor
		drawX := int32(mpos.X) - drawW/2
		drawY := int32(mpos.Y) - drawH/2

		rl.DrawTextureEx(tex,
			rl.Vector2{X: float32(drawX), Y: float32(drawY)},
			0, scale, rl.White)
	}

	// 3) If the context menu is open, draw it at MenuPosition
	if inv.MenuOpen {
		const (
			menuItemWidth  = 80
			menuItemHeight = 20
			padding        = 4
		)
		bx := int32(inv.MenuPosition.X)
		by := int32(inv.MenuPosition.Y)

		slotItem := inv.Slots[inv.MenuSlot]
		var topLabel string

		switch slotItem.Type {
		case Weapon:
			topLabel = "Equip"
		case HealthPack:
			topLabel = "Use"
		case KeyType:
			topLabel = "Use" // using a door key
		default:
			topLabel = ""
		}

		// Draw “Equip” / “Use”
		if topLabel != "" {
			rl.DrawRectangle(bx, by, menuItemWidth, menuItemHeight, rl.DarkGray)
			rl.DrawText(topLabel, bx+8, by+4, 12, rl.White)
		}

		// Draw “Drop”
		rl.DrawRectangle(bx, by+menuItemHeight+padding, menuItemWidth, menuItemHeight, rl.DarkGray)
		rl.DrawText("Drop", bx+16, by+menuItemHeight+padding+4, 12, rl.White)

		// Draw “Delete”
		rl.DrawRectangle(bx, by+2*(menuItemHeight+padding), menuItemWidth, menuItemHeight, rl.DarkGray)
		rl.DrawText("Delete", bx+8, by+2*(menuItemHeight+padding)+4, 12, rl.White)
	}

}

// Helper function to get the max of two float32 values
func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
