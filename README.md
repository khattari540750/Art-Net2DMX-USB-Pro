# Art-Net to DMX USB Pro

A Go application that receives Art-Net packets over UDP and converts them to DMX512 data sent via Enttec DMX USB Pro interface.

## Features

- **Art-Net Reception**: Listens for Art-Net packets on UDP port 6455
- **Universe Filtering**: Configurable universe selection (0-32767)
- **DMX Output**: Sends DMX512 data to Enttec DMX USB Pro device
- **GUI Interface**: Simple and intuitive graphical user interface
- **Real-time Status**: Live connection status and Art-Net reception monitoring
- **Cross-platform**: Built with Go and Fyne for multi-platform support

## Prerequisites

### Hardware
- Enttec DMX USB Pro device
- USB cable to connect DMX USB Pro to computer
- DMX fixtures/devices for testing

### Software
- Go 1.19 or later
- macOS, Windows, or Linux

## Installation

1. Clone the repository:
```bash
git clone https://github.com/khattari540750/Art-Net2DMX-USB-Pro.git
cd Art-Net2DMX-USB-Pro
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o artnet2dmx main.go
```

## Usage

### Running the Application

1. Connect your Enttec DMX USB Pro device to your computer
2. Run the application:
```bash
./artnet2dmx
```
Or directly with Go:
```bash
go run main.go
```

### GUI Interface

The application provides a simple GUI with the following elements:

- **Universe Input**: Enter the Art-Net universe number (0-32767) you want to monitor
- **Device Status**: Shows whether the DMX USB Pro device is connected
- **Art-Net Status**: Displays Art-Net packet reception status

### Configuration

#### Universe Selection
- Enter the desired universe number in the "Universe" field
- Valid range: 0-32767
- The application will only process Art-Net packets matching the specified universe

#### Serial Port Configuration
By default, the application looks for the DMX USB Pro device at `/dev/tty.usbserial`. If your device uses a different port, modify the port name in the source code:

```go
port := "/dev/tty.usbserial"  // Change this to your device's port
```

Common port names:
- **macOS**: `/dev/tty.usbserial-XXXXXXXX` or `/dev/cu.usbserial-XXXXXXXX`
- **Windows**: `COM1`, `COM2`, etc.
- **Linux**: `/dev/ttyUSB0`, `/dev/ttyACM0`, etc.

## Technical Details

### Art-Net Protocol
- Listens on UDP port 6455 (standard Art-Net port)
- Supports Art-Net DMX packets
- Universe calculation: `Universe = Net * 256 + SubUni`

### DMX Protocol
- Supports standard DMX512 protocol
- 512 channels per universe
- Start code: 0x00 (standard DMX)

### Enttec DMX USB Pro Protocol
The application uses the Enttec DMX USB Pro packet structure:
```
[0x7E] [Label] [Length Low] [Length High] [Start Code] [DMX Data...] [0xE7]
```

## Logging

The application provides detailed logging for:
- Art-Net packet reception with source IP and universe information
- DMX transmission status
- Device connection status
- Error messages

Log format example:
```
Art-Net received: 192.168.1.100:6455, Universe: 0, Channels: 512
```

## Troubleshooting

### Device Not Connected
- Verify the Enttec DMX USB Pro is properly connected
- Check the serial port name in the code matches your device
- Ensure the device drivers are installed

### No Art-Net Reception
- Verify Art-Net source is broadcasting to the correct IP address
- Check that the universe number matches between source and application
- Ensure UDP port 6455 is not blocked by firewall

### DMX Output Issues
- Check DMX cable connections
- Verify DMX fixtures are configured for the correct addresses
- Ensure the universe contains valid DMX data

## Development

### Dependencies
- [Fyne](https://fyne.io/): Cross-platform GUI framework
- [go-artnet](https://github.com/jsimonetti/go-artnet): Art-Net protocol implementation
- [serial](https://github.com/tarm/serial): Serial port communication

### Building for Different Platforms

**Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -o artnet2dmx.exe main.go
```

**Linux:**
```bash
GOOS=linux GOARCH=amd64 go build -o artnet2dmx main.go
```

**macOS:**
```bash
GOOS=darwin GOARCH=amd64 go build -o artnet2dmx main.go
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Art-Net Protocol](https://art-net.org.uk/) by Artistic Licence
- [Enttec](https://www.enttec.com/) for DMX USB Pro specifications
- [Fyne](https://fyne.io/) for the excellent GUI framework

## Support

If you encounter any issues or have questions, please [open an issue](https://github.com/khattari540750/Art-Net2DMX-USB-Pro/issues) on GitHub.
