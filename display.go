package main

import (
	"context"
	"image/color"
	"time"

	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freesans"
)

func FlashMsg(ctx context.Context, display *ssd1306.Device, text string) {
	// 4. Clear the buffer and write text
	display.ClearBuffer()
	white := color.RGBA{255, 255, 255, 255}
	// tinyfont.WriteLine(adapter, font, x, y, text, color)
	tinyfont.WriteLine(display, &freesans.Regular12pt7b, 10, 25, "CAR BACK!", white)

	// 5. Push the buffer to the screen
	println(display.Display())

	// Keep the chip awake
	// Logic flow inside your goroutine
	for {
		select {
		case <-ctx.Done():
			display.ClearDisplay()
			return
		default:
			tinyfont.WriteLine(display, &freesans.Regular12pt7b, 10, 25, text, white)

			// 5. Push the buffer to the screen
			println(display.Display())

			time.Sleep(time.Second)
			display.ClearDisplay()
			time.Sleep(time.Second)
		}
	}
}
