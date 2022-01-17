package main

import (
	"log"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

var (
	clickWaitMs        = 5
	clicksLogThreshold = 400

	autoclickRunnig = false
	pressedKeyboard = make(map[uint16]bool, 256)
	pressedMouse    = make(map[uint16]bool, 256)
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("Hello there!")

	eventHook := robotgo.Start()
	var ev hook.Event

	hooks := []func(){
		startClickingHook([]string{"alt"}, []string{"mleft"}),
		stopClickingHook([]string{"ctrl", "q"}, []string{}),
	}

	for ev = range eventHook {
		shouldCheckHooks := false

		switch ev.Kind {
		case hook.KeyDown, hook.KeyHold:
			pressedKeyboard[ev.Keycode] = true
			shouldCheckHooks = true
		case hook.KeyUp:
			pressedKeyboard[ev.Keycode] = false
			shouldCheckHooks = true
		case hook.MouseHold: // hold is the first event after a click
			pressedMouse[ev.Keycode] = true
			shouldCheckHooks = true
		case hook.MouseUp, hook.MouseDown: // when the mouse moves the mouse up event is not triggered
			pressedMouse[ev.Keycode] = false
			shouldCheckHooks = true
		}

		if shouldCheckHooks {
			for _, hook := range hooks {
				hook()
			}
		}
	}
}

func startClickingHook(keyboard []string, mouse []string) func() {
	requiredKeyboardPressed := keyboardKeyNamesToCodes(keyboard)
	requiredMousePressed := mouseKeyNamesToCodes(mouse)

	return func() {
		handleEvent(requiredKeyboardPressed, requiredMousePressed, func() {
			log.Println("Start clicking - Entered!")

			if !autoclickRunnig {
				autoclickRunnig = true

				log.Printf("Starting to click every: %d ms\n", clickWaitMs)
				log.Printf("It is aprox %g clicks per second\n", (1000 / float64(clickWaitMs)))

				go clicker()
			}
		})
	}
}

func clicker() {
	mousePosX, mousePosY := robotgo.GetMousePos()
	var clicks uint64 = 0
	coutnerToLog := clicksLogThreshold

	log.Println("Started clicking!")

	clicksToLogStartTime := time.Now()

	clickerLoopLap := func() {
		robotgo.MoveMouse(mousePosX, mousePosY)
		robotgo.MouseClick()

		clicks++
		coutnerToLog--

		if coutnerToLog == 0 {
			end := time.Now()
			elapsed := end.Sub(clicksToLogStartTime)

			msPerClick := float64(elapsed.Milliseconds()) / float64(clicksLogThreshold)

			log.Printf(
				"Clicked: %d. Elapsed: %s for %d clicks which is %.3fms per click and %.3f clicks per second\n",
				clicks,
				elapsed,
				clicksLogThreshold,
				msPerClick,
				1000/msPerClick,
			)

			coutnerToLog = clicksLogThreshold
			clicksToLogStartTime = end
		}
	}

	// ticker := time.NewTicker(time.Duration(clickWaitMs) * time.Millisecond)

	// for range ticker.C {
	// 	clickerLoopLap()

	// 	if !autoclickRunnig {
	// 		ticker.Stop()
	// 		log.Println("Stopped clicking, clicked:", clicks)
	// 	}
	// }

	for autoclickRunnig {
		clickerLoopLap()
		// time.Sleep(time.Duration(clickWaitMs) * time.Millisecond)
	}

	log.Println("Stopped clicking, clicked:", clicks)
}

func stopClickingHook(keyboard []string, mouse []string) func() {
	requiredKeyboardPressed := keyboardKeyNamesToCodes(keyboard)
	requiredMousePressed := mouseKeyNamesToCodes(mouse)

	return func() {
		handleEvent(requiredKeyboardPressed, requiredMousePressed, func() {
			log.Println("Trying to stop clicking")
			autoclickRunnig = false
		})
	}
}

func keyboardKeyNamesToCodes(keyNames []string) []uint16 {
	return keyNamesToCodes(hook.Keycode, keyNames)
}

func mouseKeyNamesToCodes(mouseNames []string) []uint16 {
	return keyNamesToCodes(hook.MouseMap, mouseNames)
}

func keyNamesToCodes(nameToCode map[string]uint16, names []string) []uint16 {
	codes := []uint16{}
	for _, keyName := range names {
		codes = append(codes, nameToCode[keyName])
	}

	return codes
}

func handleEvent(requiredKeyboardPressed []uint16, requiredMousePressed []uint16, callback func()) {
	if allPressed(pressedKeyboard, requiredKeyboardPressed) && allPressed(pressedMouse, requiredMousePressed) {
		callback()
	}
}

func allPressed(pressed map[uint16]bool, keys []uint16) bool {
	for _, key := range keys {
		if !pressed[key] {
			return false
		}
	}

	return true
}
