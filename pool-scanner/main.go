package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"pool-scanner/scanner"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.NewWithID("com.pooltool.scanner")
	myWindow := myApp.NewWindow("Pool Scanner - COSMC")

	totalBind := binding.NewString()
	totalBind.Set("Total Blocks: Loading...")
	epochBind := binding.NewString()
	epochBind.Set("Blocks in Epoch: Loading...")
	sinceBind := binding.NewString()
	sinceBind.Set("Time since last: Loading...")
	statusBind := binding.NewString()
	statusBind.Set("Status: Initializing...")

	totalLabel := widget.NewLabelWithData(totalBind)
	epochLabel := widget.NewLabelWithData(epochBind)
	sinceLabel := widget.NewLabelWithData(sinceBind)
	statusLabel := widget.NewLabelWithData(statusBind)

	content := container.NewVBox(
		widget.NewCard("COSMC Pool Stats", "Monitoring Cardano Blocks", container.NewVBox(
			totalLabel,
			epochLabel,
			sinceLabel,
		)),
		statusLabel,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(300, 200))

	client := &http.Client{Timeout: 10 * time.Second}
	var lastKnownCount int64
	var lastBlockTime time.Time

	updateUI := func() {
		data, err := scanner.FetchPoolData(client)
		if err != nil {
			log.Printf("Fetch error: %v", err)
			statusBind.Set(fmt.Sprintf("Status: Error (%s)", time.Now().Format("15:04:05")))
			return
		}

		if lastKnownCount > 0 && data.TotalBlocks > lastKnownCount {
			myApp.SendNotification(&fyne.Notification{
				Title:   "New Block Minted!",
				Content: fmt.Sprintf("Pool COSMC just minted a new block! Total: %d", data.TotalBlocks),
			})
		}

		lastKnownCount = data.TotalBlocks
		lastBlockTime = data.LastBlockTime

		totalBind.Set(fmt.Sprintf("Total Blocks: %d", data.TotalBlocks))
		epochBind.Set(fmt.Sprintf("Blocks in Epoch: %d", data.EpochBlocks))
		statusBind.Set(fmt.Sprintf("Status: Updated (%s)", time.Now().Format("15:04:05")))
	}

	// Periodic update for "Time since last"
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			if !lastBlockTime.IsZero() {
				since := time.Since(lastBlockTime).Round(time.Second)
				sinceBind.Set(fmt.Sprintf("Time since last: %s", since))
			}
		}
	}()

	// Periodic update for scanner
	go func() {
		updateUI() // Initial update
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			updateUI()
		}
	}()

	myWindow.ShowAndRun()
}
