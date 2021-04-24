package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/bit101/go-cairo"
)

func main() {
	encodeCmd := flag.NewFlagSet("enable", flag.ExitOnError)
	inputImage := encodeCmd.String("i", "", "input image")
	outputImage := encodeCmd.String("o", "", "output image")
	dataFile := encodeCmd.String("d", "", "data file")
	dataText := encodeCmd.String("t", "", "text to encode")

	decodeCmd := flag.NewFlagSet("decode", flag.ExitOnError)
	encodedImage := decodeCmd.String("i", "", "input image")
	outputFile := decodeCmd.String("o", "", "output text file")

	help := flag.Bool("h", false, "help")
	flag.Parse()

	if *help {
		fmt.Println("Usage:")
		fmt.Println("  steggo encode -i input.png -o encoded.png -d datafile ")
		fmt.Println("  steggo encode -i input.png -o encoded.png -t \"some text\"")
		fmt.Println("  steggo decode -i encoded.png")
		fmt.Println("  steggo decode -i encoded.png > output.txt")
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		fmt.Println("Expected 'enable' or 'decode' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "encode":
		encodeCmd.Parse(os.Args[2:])
		if *dataText != "" {
			encodeText(*inputImage, *outputImage, []byte(*dataText))
		} else if *dataFile != "" {
			encodeFile(*inputImage, *outputImage, *dataFile)
		} else {
			fmt.Println("encode requires '-d data file' or '-t text'")
			os.Exit(1)
		}

	case "decode":
		decodeCmd.Parse(os.Args[2:])
		decode(*encodedImage, *outputFile)

	default:
		fmt.Println("Expected 'enable' or 'decode' subcommands")
		os.Exit(1)
	}

}

func decode(image, outputFile string) {
	surface, status := cairo.NewSurfaceFromPNG(image)
	if status != cairo.StatusSuccess {
		log.Fatal("Unable to load image: ", status)
	}

	data := surface.GetData()

	message := parseData(data)
	if outputFile != "" {
		ioutil.WriteFile(outputFile, []byte(message), 0777)
	} else {
		fmt.Println(message)
	}
}

func parseData(data []byte) string {
	var message []byte
	var char byte = 0

	j := 0
	for i := 0; i < len(data); i++ {
		// skip alpha channels
		if !isAlphaChannel(i) {
			b := data[i]
			// add 1 if the bit is odd
			if b%2 == 1 {
				char += 1
			}
			j++
			if j < 8 {
				// shift left
				char = char << 1
			} else {
				// we have 8 bits now. add the char to the message
				message = append(message, char)
				j = 0
				char = 0
				i++
				// check the signal bit. if odd, return.
				if data[i]%2 == 1 {
					return string(message)
				}
			}
		}
	}
	return string(message)
}

func parseChar(c string) (string, bool) {
	// convert byte into 8 bit string, e.g. "01101101" => 109 'm'
	char, err := strconv.ParseInt(c[0:8], 2, 64)
	if err != nil {
		log.Fatal("unable to parse data")
	}
	// signal bit. 0 = continue. 1 = stop.
	return string(char), c[8] == '1'
}

func encodeFile(inputImage, outputImage, inputFile string) {
	message, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatal("couldn't read data file")
	}
	encodeText(inputImage, outputImage, message)
}

func encodeText(inputImage, outputImage string, message []byte) {

	surface, status := cairo.NewSurfaceFromPNG(inputImage)
	if status != cairo.StatusSuccess {
		log.Fatal("Unable to load image: ", status)
	}
	data := surface.GetData()

	index := 0
	for _, c := range message {
		// grab each bit left to right
		for i := 7; i >= 0; i-- {
			// don't mess with alpha channel
			if isAlphaChannel(index) {
				data[index] = 255
				index++
			}
			// isolate the bit
			b := c & byte(math.Pow(2, float64(i)))
			b = b >> i
			if b == 0 {
				setEven(data, index)
			} else if b == 1 {
				setOdd(data, index)
			}
			index++
		}
		// 9th bit is signal to continue (even) or stop (odd)
		// don't mess with alpha channel
		if isAlphaChannel(index) {
			data[index] = 255
			index++
		}
		// default to even
		setEven(data, index)
		index++
	}
	// now we are done, so set the signal bit odd
	// index has been incremented, so we need to go back one to overwrite the even byte.
	setOdd(data, index-1)

	surface.SetData(data)
	surface.WriteToPNG(outputImage)
}

func setEven(data []byte, index int) {
	if data[index]%2 == 1 {
		data[index]--
	}
}

func setOdd(data []byte, index int) {
	if data[index]%2 == 0 {
		data[index]++
	}
}

func isAlphaChannel(index int) bool {
	return index%4 == 3
}
