package main

import (
	"fmt"
	"os"
	"os/exec"
	"platformer-game/core"
	"platformer-game/gameobjects"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 800
	screenHeight = 450
	worldWidth   = 5000
	worldHeight  = 1200
)

var gameOver bool
var restarting bool // Prevents multiple restarts

// Play GameOver.mp4 INSIDE the game window (not fullscreen)
func PlayGameOverVideo() {
	fmt.Println("Playing GameOver.mp4 in window...")

	// Launch ffplay with window position and size
	cmd := exec.Command("ffplay",
		"-autoexit",                  // Closes when done
		"-noborder",                  // No window decorations
		"-window_title", "Game Over", // Title to match game
		"-x", fmt.Sprintf("%d", screenWidth), // Set width
		"-y", fmt.Sprintf("%d", screenHeight), // Set height
		"assets/video/GameOver.mp4",
	)

	// Set window position near the game window
	cmd.Env = append(os.Environ(), "SDL_VIDEO_WINDOW_POS=100,100")

	// Run the command
	err := cmd.Start()
	if err != nil {
		fmt.Println("Error playing video:", err)
		return
	}
	cmd.Wait() // Wait until video finishes
}

// Show "Try Again" button after the Game Over video
func ShowTryAgainButton() {
	buttonX := screenWidth/2 - 100
	buttonY := screenHeight/2 + 50
	buttonWidth := 200
	buttonHeight := 50

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// Draw "Game Over" text
		rl.DrawText("Game Over", screenWidth/2-100, screenHeight/2-100, 40, rl.Red)

		// Draw "Try Again" button
		rl.DrawRectangle(int32(buttonX), int32(buttonY), int32(buttonWidth), int32(buttonHeight), rl.DarkGray)
		rl.DrawText("Try Again", int32(buttonX+50), int32(buttonY+15), 20, rl.White)

		// Check for mouse click on the button
		mousePos := rl.GetMousePosition()
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			if mousePos.X > float32(buttonX) && mousePos.X < float32(buttonX+buttonWidth) &&
				mousePos.Y > float32(buttonY) && mousePos.Y < float32(buttonY+buttonHeight) {
				RestartGame() // Restart game when button is clicked
				return
			}
		}

		rl.EndDrawing()
	}
}

// Restart the game correctly without crashing
func RestartGame() {
	if restarting {
		return
	}
	restarting = true

	fmt.Println("Restarting Game...")

	// Close the window properly
	rl.CloseWindow()

	// Restart the game by executing a new process
	execPath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}

	cmd := exec.Command(execPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error restarting game:", err)
		return
	}

	os.Exit(0) // Kill the current process to allow restart
}

func main() {
	restarting = false // Reset restart flag
	rl.InitWindow(screenWidth, screenHeight, "Platformer Game")

	// Initialize the game
	core.InitGame(worldWidth, worldHeight)

	for !rl.WindowShouldClose() && !gameOver {
		//fmt.Println("Game loop running...")

		core.UpdateGame(worldHeight)
		gameOver = gameobjects.PlayerInstance.IsGameOver()
		core.DrawGame()
	}

	// If game is over, play video, then show "Try Again" button
	if gameOver {
		PlayGameOverVideo()  // Play Game Over video inside window
		ShowTryAgainButton() // Show retry button after video
	}

	rl.CloseWindow() // 🔧 Always close the window properly
}
