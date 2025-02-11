package main

import (
	"platformer-game/core"
	"platformer-game/gameobjects"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 800
	screenHeight = 450
	worldWidth   = 5000
	worldHeight  = 1200
)

var gameOver bool

func main() {
	rl.InitWindow(screenWidth, screenHeight, "Platformer Game")
	defer rl.CloseWindow()

	core.InitGame(worldWidth, worldHeight)

	for !rl.WindowShouldClose() && !gameOver {
		core.UpdateGame(worldHeight)                       //need to pass worldHeight to update zombies
		gameOver = gameobjects.PlayerInstance.IsGameOver() // Check game-over condition
		core.DrawGame()
	}
	//check players health

	// Display "Game Over" message if game has ended
	if gameOver {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		rl.DrawText("Game Over", screenWidth/2-50, screenHeight/2-20, 40, rl.Red)
		rl.EndDrawing()
		time.Sleep(3 * time.Second) // Delay to show message before closing
	}
}
