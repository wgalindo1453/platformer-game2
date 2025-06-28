package core

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"log"
	"math/rand"
	"platformer-game/database"
	"platformer-game/gameobjects"
	"time"
)

var (
	camera     rl.Camera2D
	background rl.Texture2D
	zombies    []*gameobjects.Zombie
)

// core/game.go (at top)
type SceneID int

var (
	outsideBG    rl.Texture2D
	insideBG     rl.Texture2D
	currentScene = SceneOutside
)

const (
	SceneOutside SceneID = iota
	SceneInside
)

const (
	worldWidth   = 5000
	worldHeight  = 1200
	screenWidth  = 800
	screenHeight = 450
)

const (
	miniMapWidth  = 200
	miniMapHeight = 150
	miniMapX      = screenWidth - miniMapWidth - 10
	miniMapY      = 10
	deadZoneWidth = 200
)

// We now have two world items: one Sword and one Health Pack.
var (
	testItem    gameobjects.WorldItem
	testItem2   gameobjects.WorldItem
	testItem3   gameobjects.WorldItem // For the BronzeKey
	doors       []*gameobjects.Door   // ← add this
	insideDoors []*gameobjects.Door
	fadeAlpha   float32 = 0 // 0 = fully transparent, 1 = fully black
	fading      bool    = false
	fadeDir     float32 = 0            // −1 = fade out, +1 = fade in
	targetScene SceneID = SceneOutside // where we want to go after fading
)

func InitGame(worldW, worldH int) {
	// 1) Load a background texture
	background = rl.LoadTexture("assets/background2.png")
	doorSheet := "assets/sprites/doors_spritesheet.png"

	// Load outside and inside backgrounds exactly once:
	outsideBG = rl.LoadTexture("assets/levelonebg.png")
	insideBG = rl.LoadTexture("assets/background2.png")
	src := rl.NewRectangle(0, 0, float32(insideBG.Width), float32(insideBG.Height))
	dest := rl.NewRectangle(0, 0, float32(screenWidth), float32(screenHeight))
	origin := rl.NewVector2(0, 0)
	rotation := float32(0)

	rl.DrawTexturePro(insideBG, src, dest, origin, rotation, rl.White)
	//scale insideBG to fit the screen

	currentScene = SceneOutside

	// 2) Initialize the player (sets up PlayerInstance with default health, inventory, etc.)
	gameobjects.InitPlayer(worldW, worldH)

	// 3) Open (or create) our SQLite database
	database.InitDatabase()

	// 4) Preload all item textures by name
	itemTextures := map[string]rl.Texture2D{
		"Sword":      rl.LoadTexture("assets/sword.png"),
		"HealthPack": rl.LoadTexture("assets/healthpack.png"),
		"BronzeKey":  rl.LoadTexture("assets/bronze_key.png"), // or whichever key sprite

	}

	// 5) Load whatever was saved in the "inventory" table:
	gameobjects.PlayerInstance.Inventory.LoadFromDB(itemTextures)

	// 6) Spawn two WorldItems in the scene:
	//    - Sword at (110, 1040)
	//    - HealthPack at (200, 1040)
	doorRects := []rl.Rectangle{
		{X: 19, Y: 59, Width: 78, Height: 130},  // frame 0 = closed
		{X: 118, Y: 59, Width: 78, Height: 130}, // frame 1
		{X: 218, Y: 59, Width: 77, Height: 133}, // frame 2
		{X: 317, Y: 54, Width: 77, Height: 142}, // frame 3
		{X: 416, Y: 49, Width: 78, Height: 153}, // frame 4
		{X: 515, Y: 49, Width: 78, Height: 152}, // frame 5 = fully open
	}

	doors = append(doors, gameobjects.NewAnimatedDoor(
		"BronzeKey", // that same key name from your inventory logic
		1200, float32(worldHeight-128),
		doorSheet,
		doorRects,
		100, // 100ms between frames
	))

	testItem = gameobjects.NewWorldItem(
		110, 1040,
		gameobjects.Weapon,
		"Sword",
		"assets/sword.png",
	)
	testItem2 = gameobjects.NewWorldItem(
		200, 1100,
		gameobjects.HealthPack,
		"HealthPack",
		"assets/healthpack.png",
	)
	testItem3 = gameobjects.NewWorldItem(
		300, worldHeight-100,
		gameobjects.KeyType, "BronzeKey", "assets/bronze_key.png",
	)

	// 7) Spawn some zombies
	initZombies(5)

	// 8) Set up a 2D camera that follows the player
	camera = rl.Camera2D{
		Target: gameobjects.PlayerInstance.Position,
		Offset: rl.NewVector2(float32(screenWidth)/2, float32(screenHeight)/2),
		Zoom:   1.0,
	}
}

// initZombies places `numZombies` zombies at random x-positions along the ground.
func initZombies(numZombies int) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < numZombies; i++ {
		x := float32(rand.Intn(worldWidth-100) + 50)
		y := float32(worldHeight) // Keep them at ground level
		z := gameobjects.InitZombie(x, y-50, 1)
		zombies = append(zombies, &z)
	}
}

func UpdateGame(worldH int) {
	inv := &gameobjects.PlayerInstance.Inventory

	// 1) Let the inventory handle mouse/keyboard (drag/drop, context menu, etc.)
	inv.HandleMouse()

	// 2) If the player just used a key, attempt to unlock the matching door
	if keyID := gameobjects.PlayerInstance.UsedKeyID; keyID != "" {
		if currentScene == SceneOutside {
			for _, d := range doors {
				if d.ID == keyID {
					d.TryUnlock()
					break
				}
			}
		} else /* SceneInside */ {
			for _, d := range insideDoors {
				if d.ID == keyID {
					d.TryUnlock()
					break
				}
			}
		}
		gameobjects.PlayerInstance.UsedKeyID = ""
	}

	// 3) Toggle inventory on/off with "I"
	if rl.IsKeyPressed(rl.KeyI) {
		inv.IsOpen = !inv.IsOpen
	}
	if rl.IsKeyPressed(rl.KeyK) {
		background = rl.LoadTexture("assets/background2.png")
	}

	playerPos := gameobjects.PlayerInstance.Position

	// 4) If we're outside, handle "E" to pick up world items
	if currentScene == SceneOutside && rl.IsKeyPressed(rl.KeyE) {
		// Sword pickup
		if testItem.Texture.ID != 0 && rl.Vector2Distance(playerPos, testItem.Position) < 50 {
			it := gameobjects.Item{
				Type:  testItem.Type,
				Name:  testItem.Name,
				Image: testItem.Texture,
			}
			if gameobjects.PlayerInstance.Inventory.AddItem(it) {
				log.Println("Picked up:", it.Name)
				testItem.Texture.ID = 0
				gameobjects.PlayerInstance.Inventory.SaveToDB()
			} else {
				log.Println("Inventory full!")
			}
		}
		// HealthPack pickup
		if testItem2.Texture.ID != 0 && rl.Vector2Distance(playerPos, testItem2.Position) < 50 {
			it2 := gameobjects.Item{
				Type:  testItem2.Type,
				Name:  testItem2.Name,
				Image: testItem2.Texture,
			}
			if gameobjects.PlayerInstance.Inventory.AddItem(it2) {
				log.Println("Picked up:", it2.Name)
				testItem2.Texture.ID = 0
				gameobjects.PlayerInstance.Inventory.SaveToDB()
			} else {
				log.Println("Inventory full!")
			}
		}
		// BronzeKey pickup
		if testItem3.Texture.ID != 0 && rl.Vector2Distance(playerPos, testItem3.Position) < 50 {
			it3 := gameobjects.Item{
				Type:  testItem3.Type,
				Name:  testItem3.Name,
				Image: testItem3.Texture,
			}
			if gameobjects.PlayerInstance.Inventory.AddItem(it3) {
				log.Println("Picked up:", it3.Name)
				testItem3.Texture.ID = 0
				gameobjects.PlayerInstance.Inventory.SaveToDB()
			} else {
				log.Println("Inventory full!")
			}
		}
	}

	// 5) Update player physics/movement/shooting every frame (pass zombies only if outside)
	if currentScene == SceneOutside {
		gameobjects.PlayerInstance.Update(worldH, worldWidth, zombies)
		gameobjects.PlayerInstance.Shoot()
	} else /* SceneInside */ {
		// No zombies inside; pass nil
		gameobjects.PlayerInstance.Update(worldH, worldWidth, nil)
		gameobjects.PlayerInstance.Shoot()
	}

	// 6) Advance door animations and handle scene‐switch when a door finishes opening
	// door‐opening fade logic
	if !fading {
		// advance each door
		if currentScene == SceneOutside {
			for _, d := range doors {
				d.Update()
				if d.State == gameobjects.DoorOpen {
					targetScene = SceneInside
					fading = true
					fadeDir = -1
					break
				}
			}
		} else {
			for _, d := range insideDoors {
				d.Update()
				if d.State == gameobjects.DoorOpen {
					targetScene = SceneOutside
					fading = true
					fadeDir = -1
					break
				}
			}
		}
	}
	if fading {
		dt := rl.GetFrameTime()
		fadeAlpha += fadeDir * dt
		if fadeAlpha <= 0 {
			// at black → swap scene
			fadeAlpha = 0
			fading = true // now fade in
			fadeDir = +1
			currentScene = targetScene

			// reposition & clear/spawn entities
			if currentScene == SceneInside {
				gameobjects.PlayerInstance.Position = rl.NewVector2(
					doors[0].Position.X+doors[0].Width+10,
					float32(worldHeight)-55,
				)
				zombies = nil
				// create insideDoors…
			} else {
				// back to outside
				gameobjects.PlayerInstance.Position = rl.NewVector2(
					insideDoors[0].Position.X-gameobjects.PlayerInstance.Width-10,
					float32(worldHeight)-55,
				)
				initZombies(5)
			}
		} else if fadeAlpha >= 1 {
			fadeAlpha = 1
			fading = false
		}
	}

	// 7) Update zombies—but only if we're outside
	if currentScene == SceneOutside {
		for i := len(zombies) - 1; i >= 0; i-- {
			z := zombies[i]
			z.Update(worldWidth, playerPos)
			if !z.IsAlive &&
				z.State == gameobjects.ZombieDead &&
				z.CurrentFrame == len(z.DeadFrames)-1 {
				z.UnloadSounds()
				zombies = append(zombies[:i], zombies[i+1:]...)
			}
		}
	}

	// 9) Camera follows the player with a small dead‐zone, regardless of scene
	playerX := gameobjects.PlayerInstance.Position.X
	if playerX > camera.Target.X+float32(screenWidth)/2-deadZoneWidth {
		camera.Target.X = playerX - float32(screenWidth)/2 + deadZoneWidth
	} else if playerX < camera.Target.X-float32(screenWidth)/2+deadZoneWidth {
		camera.Target.X = playerX + float32(screenWidth)/2 - deadZoneWidth
	}
	camera.Target.X = clampFloat(
		camera.Target.X,
		float32(screenWidth)/2,
		float32(worldWidth)-float32(screenWidth)/2,
	)
	camera.Target.Y = clampFloat(
		camera.Target.Y,
		float32(screenHeight)/2,
		float32(worldHeight)-float32(screenHeight)/2,
	)
}
func DrawMiniMap() {
	rl.DrawRectangle(miniMapX, miniMapY, miniMapWidth, miniMapHeight, rl.LightGray)

	scaleX := float32(miniMapWidth) / float32(worldWidth)
	scaleY := float32(miniMapHeight) / float32(worldHeight)

	rl.DrawRectangleLines(miniMapX, miniMapY, miniMapWidth, miniMapHeight, rl.DarkGray)

	viewX := miniMapX + int((camera.Target.X-float32(screenWidth)/2)*scaleX)
	viewY := miniMapY + int((camera.Target.Y-float32(screenHeight)/2)*scaleY)
	viewW := int(float32(screenWidth) * scaleX * 0.8)
	viewH := int(float32(screenHeight) * scaleY * 0.8)

	viewX = clamp(viewX, miniMapX, miniMapX+miniMapWidth-viewW)
	viewY = clamp(viewY, miniMapY, miniMapY+miniMapHeight-viewH)
	rl.DrawRectangleLines(int32(viewX), int32(viewY), int32(viewW), int32(viewH), rl.Red)
}

func DrawGame() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)
	playerPos := gameobjects.PlayerInstance.Position
	log.Printf("Player Position: (%.2f, %.2f)", playerPos.X, playerPos.Y)

	// ─── 1) Draw the correct world background (under the camera) ───
	rl.BeginMode2D(camera)
	if currentScene == SceneOutside {
		DrawWorldBG(outsideBG)

		// 2) Draw world items (only if their texture ID != 0)
		if testItem.Texture.ID != 0 {
			testItem.Draw()
		}
		if testItem2.Texture.ID != 0 {
			testItem2.Draw()
		}
		if testItem3.Texture.ID != 0 {
			testItem3.Draw()
		}

		// ─── Draw all doors (locked or open) ───
		for _, d := range doors {
			d.Draw()
		}

		// 3) Draw player (including any equipped item)
		gameobjects.PlayerInstance.Draw()

		// 4) Draw all zombies
		for _, z := range zombies {
			z.Draw()
		}
	} else {
		DrawWorldBG(insideBG)
	}
	// ─── 1) Draw the correct background in _screen_-space ───
	if currentScene == SceneOutside {
		// scale outsideBG to screen
		scaleAndDrawFullScreen(outsideBG)
	} else {
		// scale insideBG to screen
		scaleAndDrawFullScreen(insideBG)
	}

	// ─── 2) Now draw your world under the camera ───
	rl.BeginMode2D(camera)
	if currentScene == SceneOutside {
		// … draw outside items, doors, zombies …
	} else {
		for _, d := range doors {
			d.Draw()
		}
		// … draw inside doors …
	}

	// ─── 3) Draw the player (always) ───
	gameobjects.PlayerInstance.Draw()
	rl.EndMode2D()

	// ─── 4) UI & inventory ───
	if gameobjects.PlayerInstance.Inventory.IsOpen {
		gameobjects.PlayerInstance.Inventory.DrawInventory()
	}
	DrawPlayerHUD()
	DrawMiniMap()

	rl.EndDrawing()
}

// scaleAndDrawFullScreen scales `tex` to exactly 800×450 and draws it at (0,0).
func scaleAndDrawFullScreen(tex rl.Texture2D) {
	src := rl.NewRectangle(0, 0,
		float32(tex.Width), float32(tex.Height))
	dest := rl.NewRectangle(0, 0,
		float32(screenWidth), float32(screenHeight))
	origin := rl.NewVector2(0, 0)
	rl.DrawTexturePro(tex, src, dest, origin, 0, rl.White)
}

func DrawWorldBG(tex rl.Texture2D) {
	rl.DrawTexturePro(
		tex,
		rl.Rectangle{X: 0, Y: 0,
			Width:  float32(tex.Width),
			Height: float32(tex.Height)},
		rl.Rectangle{X: 0, Y: float32(worldHeight) - float32(screenHeight),
			Width:  float32(worldWidth),
			Height: float32(screenHeight)},
		rl.Vector2{}, 0, rl.White)
}

func DrawPlayerHUD() {
	player := &gameobjects.PlayerInstance

	// Health Bar
	hw := float32(200.0)
	hh := float32(20.0)
	hp := float32(player.Health / player.MaxHealth)
	rl.DrawRectangle(20, 20, int32(hw), int32(hh), rl.DarkGray)
	rl.DrawRectangle(20, 20, int32(hw*hp), int32(hh), rl.Red)
	healthText := fmt.Sprintf("Health: %.0f/%.0f", player.Health, player.MaxHealth)
	rl.DrawText(healthText, 30, 25, 10, rl.White)

	// Ammo Bar
	abY := 45
	aw := float32(200.0)
	ah := float32(15.0)
	ap := float32(player.Ammo) / float32(player.MaxAmmo)
	rl.DrawRectangle(20, int32(abY), int32(aw), int32(ah), rl.DarkGray)
	rl.DrawRectangle(20, int32(abY), int32(aw*ap), int32(ah), rl.Yellow)
	ammoText := fmt.Sprintf("Ammo: %d/%d", player.Ammo, player.MaxAmmo)
	rl.DrawText(ammoText, 30, int32(abY+3), 10, rl.White)
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func clampFloat(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
