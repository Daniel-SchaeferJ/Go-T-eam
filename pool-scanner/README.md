# Cardano Pool Scanner (COSMC)

This is a mobile application built with Go and Fyne (which uses gomobile) that monitors the COSMC stake pool.

## Features
- **Real-time UI**: Shows total blocks minted, **blocks in the current epoch**, and time since the last block.
- **Push Notifications**: Sends a notification whenever a new block is detected.
- **Auto-polling**: Scans the blockchain every 30 seconds using the Koios API.

## Requirements
- Go 1.25 or later
- Fyne v2
- Fyne helper tool (Install with `go install fyne.io/fyne/v2/cmd/fyne@latest`)
- Gomobile (installed via `fyne setup` or `go install golang.org/x/mobile/cmd/gomobile@latest`)

### Linux Desktop Dependencies
To build or run on Linux desktop, you need the following development libraries:
```bash
sudo apt-get install libgl1-mesa-dev xorg-dev libxxf86vm-dev
```

## How to Build

### Desktop (for testing)
```bash
go run .
```

### Docker (for consistent build environment)
If you don't want to install dependencies locally, you can use Docker:
```bash
docker build -t pool-scanner .
```

### Android
To build an APK with the Cardano icon:
```bash
fyne package -os android -appID com.pooltool.scanner -icon icon.png
```

### iOS
To build an IPA (requires macOS and Xcode):
```bash
fyne package -os ios -appID com.pooltool.scanner -icon icon.png
```

## How to Install on Your Phone (Android)

1. **Build the APK**: Run the Android build command above. This will generate a file named `pool-scanner.apk`.
2. **Transfer to Phone**:
   - Connect your phone to your computer via USB and copy the file.
   - OR upload the file to a cloud service (Google Drive, Dropbox) and download it on your phone.
   - OR send it to yourself via email or a messaging app.
3. **Enable Unknown Sources**: On your phone, go to **Settings > Security** (or **Apps**) and enable **"Install unknown apps"** for your file manager or browser.
4. **Install**: Open the `pool-scanner.apk` file on your phone and tap **Install**.
5. **Launch**: Find the "Pool Scanner" icon (with the Cardano logo) on your home screen and open it!

## Technical Details
- **API**: Uses [Koios API](https://api.koios.rest/) to fetch pool information.
- **Pool**: Monitors COSMC (`pool1c3ecg6h4m73a2mftmdm9mz3gxklxmnzuwzl9e38lzd4y26qcm8a`).
- **Framework**: Built with [Fyne](https://fyne.io/), which provides a cross-platform UI and leverages `gomobile` for mobile integration.
