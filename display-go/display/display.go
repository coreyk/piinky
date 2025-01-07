package display

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/devices/v3/inky"
	"periph.io/x/host/v3"
)

// BCM GPIO pin numbers for Inky Impression
const (
	DC_PIN    = 27 // GPIO27 (physical pin 13)
	RESET_PIN = 17 // GPIO17 (physical pin 11)
	BUSY_PIN  = 24 // GPIO24 (physical pin 18)
)

type DisplayService struct {
	width   int
	height  int
	isPi    bool
	display *inky.DevImpression
}

type gpioPin struct {
	number   int
	isOutput bool
	pull     gpio.Pull
	chip     string
	line     string
}

func (p *gpioPin) String() string { return fmt.Sprintf("GPIO%d", p.number) }
func (p *gpioPin) Halt() error {
	fmt.Printf("Halting GPIO%d\n", p.number)
	return nil
}
func (p *gpioPin) Name() string                           { return p.String() }
func (p *gpioPin) Number() int                            { return p.number }
func (p *gpioPin) Function() string                       { return "GPIO" }
func (p *gpioPin) DefaultPull() gpio.Pull                 { return gpio.Float }
func (p *gpioPin) Pull() gpio.Pull                        { return p.pull }
func (p *gpioPin) WaitForEdge(timeout time.Duration) bool { return false }
func (p *gpioPin) In(pull gpio.Pull, edge gpio.Edge) error {
	if p.isOutput {
		return fmt.Errorf("pin is configured as output")
	}
	p.pull = pull
	return nil
}
func (p *gpioPin) Read() gpio.Level {
	cmd := exec.Command("gpioget", p.chip, p.line)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error reading GPIO%d: %v\n", p.number, err)
		return gpio.Low
	}
	value := strings.TrimSpace(string(output))
	level := gpio.Level(value == "1")
	fmt.Printf("Read GPIO%d: %v\n", p.number, level)
	return level
}
func (p *gpioPin) Out(l gpio.Level) error {
	if !p.isOutput {
		return fmt.Errorf("pin is configured as input")
	}

	value := "0"
	if l == gpio.High {
		value = "1"
	}
	fmt.Printf("Writing GPIO%d: %v (value: %s)\n", p.number, l, value)

	cmd := exec.Command("gpioset", p.chip, fmt.Sprintf("%s=%s", p.line, value))
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error writing GPIO%d: %v\n", p.number, err)
		return fmt.Errorf("failed to write GPIO: %v", err)
	}
	return nil
}
func (p *gpioPin) PWM(duty gpio.Duty, f physic.Frequency) error {
	return fmt.Errorf("PWM not supported")
}

func openGPIO(pinNumber int, isOutput bool) (gpio.PinIO, error) {
	fmt.Printf("Opening GPIO%d (output: %v)\n", pinNumber, isOutput)

	// Find the GPIO chip and line for this pin
	cmd := exec.Command("gpioinfo")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get GPIO info: %v", err)
	}

	// Parse the output to find the chip and line
	lines := strings.Split(string(output), "\n")
	var chip, line string
	lineRegex := regexp.MustCompile(fmt.Sprintf(`\s+line\s+(\d+):\s+"GPIO%d"`, pinNumber))

	for i, l := range lines {
		if match := lineRegex.FindStringSubmatch(l); match != nil {
			// Found the GPIO line, get the chip name from the previous lines
			line = match[1]
			for j := i; j >= 0; j-- {
				if strings.HasPrefix(lines[j], "gpiochip") {
					chip = strings.Split(lines[j], " ")[0]
					break
				}
			}
			break
		}
	}

	if chip == "" || line == "" {
		return nil, fmt.Errorf("failed to find GPIO%d in gpioinfo output", pinNumber)
	}

	fmt.Printf("Found GPIO%d on %s line %s\n", pinNumber, chip, line)

	pin := &gpioPin{
		number:   pinNumber,
		isOutput: isOutput,
		pull:     gpio.Float,
		chip:     chip,
		line:     line,
	}

	// Set initial state
	if isOutput {
		if err := pin.Out(gpio.Low); err != nil {
			return nil, fmt.Errorf("failed to set initial state: %v", err)
		}
	}

	return pin, nil
}

func NewDisplayService() (*DisplayService, error) {
	ds := &DisplayService{
		width:  800,
		height: 480,
		isPi:   runtime.GOOS == "linux" && (runtime.GOARCH == "arm" || runtime.GOARCH == "arm64"),
	}

	fmt.Println("isPi: ", ds.isPi)

	if ds.isPi {
		if _, err := host.Init(); err != nil {
			return nil, fmt.Errorf("failed to initialize periph: %v", err)
		}

		spiPort, err := spireg.Open("SPI0.0")
		if err != nil {
			return nil, fmt.Errorf("failed to open SPI port: %v", err)
		}

		fmt.Println("Opening DC pin...")
		dc, err := openGPIO(DC_PIN, true)
		if err != nil {
			return nil, fmt.Errorf("failed to open DC pin: %v", err)
		}
		fmt.Println("DC pin opened successfully")

		fmt.Println("Opening Reset pin...")
		reset, err := openGPIO(RESET_PIN, true)
		if err != nil {
			return nil, fmt.Errorf("failed to open Reset pin: %v", err)
		}
		fmt.Println("Reset pin opened successfully")

		fmt.Println("Opening Busy pin...")
		busy, err := openGPIO(BUSY_PIN, false)
		if err != nil {
			return nil, fmt.Errorf("failed to open Busy pin: %v", err)
		}
		fmt.Println("Busy pin opened successfully")

		// Create new Inky Impression display with explicit options for 7.3" model
		opts := &inky.Opts{
			Model:       inky.IMPRESSION73,
			ModelColor:  inky.Multi,
			BorderColor: inky.White,
			Height:      ds.height,
			Width:       ds.width,
		}

		fmt.Println("Initializing display...")
		// Initialize the display
		display, err := inky.NewImpression(spiPort, dc, reset, busy, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize display: %v", err)
		}

		ds.display = display
		fmt.Println("Display initialized successfully")
	} else {
		fmt.Println("Not running on Raspberry Pi - display updates will be simulated")
	}

	return ds, nil
}

func (ds *DisplayService) UpdateDisplay(imagePath string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	bounds := img.Bounds()
	fmt.Printf("Image size: %dx%d\n", bounds.Dx(), bounds.Dy())

	if bounds.Dx() != ds.width || bounds.Dy() != ds.height {
		fmt.Printf("Image size: %dx%d\n", bounds.Dx(), bounds.Dy())
		fmt.Printf("Expected size: %dx%d\n", ds.width, ds.height)
		return fmt.Errorf("image dimensions do not match display dimensions")
	}

	if ds.isPi && ds.display != nil {
		fmt.Println("Starting display update...")
		// Create a new rectangle for the full display area
		fullBounds := image.Rect(0, 0, ds.width, ds.height)

		// Draw using the full display bounds
		if err := ds.display.Draw(fullBounds, img, image.Point{}); err != nil {
			return fmt.Errorf("failed to update display: %v", err)
		}
		fmt.Println("Display update completed")
	} else {
		fmt.Printf("Image would be displayed at %s\n", imagePath)
	}

	return nil
}
