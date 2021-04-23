package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/bit101/go-cairo"
)

func main() {

	if os.Args[1] == "encode" {
		encode(os.Args[2], os.Args[3], os.Args[4])
	} else if os.Args[1] == "decode" {
		decode(os.Args[2])
	} else if os.Args[1] == "-h" {
		fmt.Println("Usage:")
		fmt.Println("  steggo encode input.png datafile encoded.png")
		fmt.Println("  steggo decode encoded.png")
		fmt.Println("  steggo decode encoded.png > output.txt")
	}

}

func decode(image string) {
	surface, status := cairo.NewSurfaceFromPNG(image)
	if status != cairo.StatusSuccess {
		log.Fatal("Unable to load image: ", status)
	}

	data := surface.GetData()
	message := parseData(data)
	// ioutil.WriteFile(outputData, []byte(message), 0777)
	fmt.Println(message)
}

func parseData(data []byte) string {
	message := make([]byte, 1000)
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

func encode(inputImage, inputData, outputImage string) {
	surface, status := cairo.NewSurfaceFromPNG(inputImage)
	if status != cairo.StatusSuccess {
		log.Fatal("Unable to load image: ", status)
	}

	data := surface.GetData()
	message, err := ioutil.ReadFile(inputData)
	if err != nil {
		log.Fatal("couldn't read file")
	}

	index := 0
	for _, c := range message {
		// convert byte into 8 bit string. e.g. 109 'm' => "01101101"
		b := fmt.Sprintf("%08b", byte(c))
		// loop through the 8 bits
		for _, d := range b {
			// don't mess with alpha channel
			if isAlphaChannel(index) {
				data[index] = 255
				index++
			}
			if d == '0' {
				setEven(data, index)
			} else if d == '1' {
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
