package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/cmplx"
	"os"
	"strconv"
	"strings"
	"sync"
)

func parsePair(s string, separator rune) (int, int) {
	parts := strings.Split(s, string(separator))
	if len(parts) != 2 {
		panic("invalid format")
	}
	x, err1 := strconv.Atoi(parts[0])
	y, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		panic("invalid numbers")
	}
	return x, y
}

func parseComplex(s string) complex128 {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		panic("invalid format")
	}
	re, err1 := strconv.ParseFloat(parts[0], 64)
	im, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		panic("invalid numbers")
	}
	return complex(re, im)
}

func mandelbrot(z complex128) color.Color {
	const iterations = 255
	const contrast = 15

	var v complex128
	for n := uint8(0); n < iterations; n++ {
		v = v*v + z
		if cmplx.Abs(v) > 2 {
			return color.Gray{255 - contrast*n}
		}
	}
	return color.Black
}

func render(img *image.Gray, bounds [2]int, upperLeft, lowerRight complex128, startRow, endRow int) {
	for y := startRow; y < endRow; y++ {
		for x := 0; x < bounds[0]; x++ {
			re := real(upperLeft) + (real(lowerRight)-real(upperLeft))*float64(x)/float64(bounds[0])
			im := imag(upperLeft) + (imag(lowerRight)-imag(upperLeft))*float64(y)/float64(bounds[1])
			color := mandelbrot(complex(re, im))
			img.Set(x, y, color)
		}
	}
}

func main() {
	if len(os.Args) != 5 {
		fmt.Println("Usage: mandelbrot <output.png> <width>x<height> <upperLeft> <lowerRight>")
		os.Exit(1)
	}

	output, boundsStr, upperLeftStr, lowerRightStr := os.Args[1], os.Args[2], os.Args[3], os.Args[4]
	width, height := parsePair(boundsStr, 'x')
	upperLeft := parseComplex(upperLeftStr)
	lowerRight := parseComplex(lowerRightStr)

	img := image.NewGray(image.Rect(0, 0, width, height))
	bounds := [2]int{width, height}

	var wg sync.WaitGroup
	rowsPerBand := height / 8
	for i := 0; i < 8; i++ {
		wg.Add(1)
		startRow := i * rowsPerBand
		endRow := (i + 1) * rowsPerBand
		if i == 7 {
			endRow = height // 最後のバンドで残りのすべての行をカバー
		}
		go func(startRow, endRow int) {
			defer wg.Done()
			render(img, bounds, upperLeft, lowerRight, startRow, endRow)
		}(startRow, endRow)
	}
	wg.Wait()

	file, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, img)
}
