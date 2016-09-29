package main

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/math/fixed"
)

func main() {
	var (
		flagSet    = flag.NewFlagSet("", flag.ContinueOnError)
		srcImgPath = flagSet.String("src", "", "Path to the source image file")
		dstImgPath = flagSet.String("dst", "", "Path to the destination image file")
		fontPath   = flagSet.String("font", "", "Path to the ttf/ttc font file")
		textMsg    = flagSet.String(
			"msg", "Default message", "The message to draw on the image",
		)
	)

	flagSet.Parse(os.Args[1:])

	if *srcImgPath == "" || *dstImgPath == "" || *fontPath == "" {
		exitWithInvalidCmdParamValue(
			flagSet, "Font file path, Source and destination image paths are require",
		)
	}

	var dstImgFormat string
	if idx := strings.LastIndex(*dstImgPath, "."); idx > 0 {
		switch ext := strings.ToLower((*dstImgPath)[idx+1:]); ext {
		case "png", "gif", "jpg":
			dstImgFormat = ext
		case "jpeg":
			dstImgFormat = "jpg"
		default:
			exitWithInvalidCmdParamValue(
				flagSet, "Destination image format is unkown. Extension: %s\n", ext,
			)
		}
	} else {
		exitWithInvalidCmdParamValue(
			flagSet, "Destination image file doensn't have any extension\n",
		)
	}

	var (
		err     error
		srcFile *os.File
	)

	if srcFile, err = os.Open(*srcImgPath); err != nil {
		fmt.Printf("Error reading source image file. error= %+v\n", err)
		os.Exit(1)
	}

	defer srcFile.Close()

	var srcImg image.Image
	if srcImg, _, err = image.Decode(srcFile); err != nil {
		fmt.Printf("Error decoding image file. error= %+v\n", err)
		os.Exit(1)
	}

	var font *truetype.Font
	if font, err = fontFromFile(*fontPath); err != nil {
		fmt.Printf("Error with font file. %s\n", err.Error())
		os.Exit(1)
	}

	var (
		width  int = srcImg.Bounds().Dx()
		height int = srcImg.Bounds().Dy()
		dstImg     = image.NewRGBA(
			image.Rect(0, 0, width, height),
		)
		msgStartPoint = fixed.Point26_6{
			X: fixed.Int26_6(width / 10),
			Y: fixed.Int26_6(height - height/4),
		}
		ctx = freetype.NewContext()
	)

	ctx.SetSrc(srcImg)
	ctx.SetDst(dstImg)
	ctx.SetFont(font)
	ctx.SetFontSize(20)
	if _, err := ctx.DrawString(*textMsg, msgStartPoint); err != nil {
		fmt.Printf(
			"Error drawing the message into the image. %s. error= %+v\n",
			err.Error(), err,
		)
	}

	if err = saveImage(*dstImgPath, dstImgFormat, dstImg); err != nil {
		fmt.Printf("Error when saving image. %s\n", err.Error())
		os.Exit(1)
	}
}

func exitWithInvalidCmdParamValue(fs *flag.FlagSet, msg string, fArgs ...interface{}) {
	if fArgs != nil {
		fmt.Printf(msg, fArgs...)
	} else {
		fmt.Println(msg)
	}

	fs.PrintDefaults()
	os.Exit(1)
}

// saveImage save the image to the specified filenamePath and format
// format must be "jpg", "gif" or "png" all lowercase.
func saveImage(filenamePath, format string, img image.Image) error {
	var file, err = os.Create(filenamePath)
	if err != nil {
		return fmt.Errorf(
			"Create file error. %s. error= %+v",
			err.Error(), err,
		)
	}

	switch format {
	case "png":
		err = (&png.Encoder{}).Encode(file, img)
	case "gif":
		err = gif.Encode(file, img, nil)
	case "jpg":
		err = jpeg.Encode(file, img, nil)
	}

	if err != nil {
		return fmt.Errorf("Encode image error. %s. error= %+v", err.Error(), err)
	}

	return nil
}

// fontFromFile
func fontFromFile(filenamePath string) (*truetype.Font, error) {
	var (
		err      error
		fontFile *os.File
	)
	if fontFile, err = os.Open(filenamePath); err != nil {
		return nil, fmt.Errorf(
			"Error reading font file. %s. error= %+v", err.Error(), err,
		)
	}

	defer fontFile.Close()

	var fontData []byte
	if fontData, err = ioutil.ReadAll(fontFile); err != nil {
		return nil, fmt.Errorf(
			"Error reading the font file content. %s. error= %+v", err.Error(), err,
		)
	}

	var font *truetype.Font
	if font, err = truetype.Parse(fontData); err != nil {
		return nil, fmt.Errorf(
			"Error parsing font data. %s. error= %+v", err.Error(), err,
		)
	}

	return font, nil
}
