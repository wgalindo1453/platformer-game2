// /rendering/spritesheet.go taken from 
package rendering

import rl "github.com/gen2brain/raylib-go/raylib"

type SpriteSheet struct {
    Texture rl.Texture2D
    Image   *rl.Image // Keep the image in memory for cropping
}

// LoadSpriteSheet loads the texture and image for the spritesheet
func LoadSpriteSheet(filename string) SpriteSheet {
    texture := rl.LoadTexture(filename)
    image := rl.LoadImage(filename) // Load the image to crop from
    return SpriteSheet{Texture: texture, Image: image}
}

// ImageAt extracts a sub-rectangle from the spritesheet and returns a texture
func (s *SpriteSheet) ImageAt(rect rl.Rectangle, colorkey rl.Color) rl.Texture2D {
    croppedImg := rl.ImageCopy(s.Image) // Create a copy to preserve the original image
    rl.ImageCrop(croppedImg, rect)      // Crop the copied image based on the rect

    // Optional: Apply colorkey if needed to remove specific color backgrounds
    if colorkey.A > 0 {
        rl.ImageColorReplace(croppedImg, rl.GetImageColor(*croppedImg, 0, 0), colorkey)
    }

    texture := rl.LoadTextureFromImage(croppedImg)
    rl.UnloadImage(croppedImg) // Clean up the cropped image to avoid memory leaks
    return texture
}

// LoadStrip loads a strip of images from the spritesheet and returns an array of textures
// Useful for loading animation frames from a single row of sprites but needs to be updated as it depends on set sizes
func (s *SpriteSheet) LoadStrip(rect rl.Rectangle, count int, colorkey rl.Color) []rl.Texture2D {
    textures := make([]rl.Texture2D, count)
    for i := 0; i < count; i++ {
        newRect := rl.Rectangle{
            X:      rect.X + float32(i)*rect.Width,
            Y:      rect.Y,
            Width:  rect.Width,
            Height: rect.Height,
        }
        textures[i] = s.ImageAt(newRect, colorkey)
    }
    return textures
}

// Unload the sprite sheet resources
func (s *SpriteSheet) Unload() {
    rl.UnloadTexture(s.Texture)
    rl.UnloadImage(s.Image)
}
