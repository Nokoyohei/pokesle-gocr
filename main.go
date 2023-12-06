package main

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"strings"

	vision "cloud.google.com/go/vision/apiv1"
)

func cropImage(img image.Image, rect image.Rectangle) image.Image {
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	return img.(subImager).SubImage(rect)
}

func saveImage(img image.Image, filename string) {
	out, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	png.Encode(out, img)
}

func readImage(filename string) image.Image {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	return img
}

func prepareCropImages(filename string) {
	// Read image
	im := readImage(filename)

	cropped := cropImage(im, image.Rect(180, 210, 180+500, 210+90))
	saveImage(cropped, "tmp/date.png")

	cropped = cropImage(im, image.Rect(285, 320, 285+548, 320+100))
	saveImage(cropped, "tmp/type.png")

	cropped = cropImage(im, image.Rect(270, 630, 280+100, 630+65))
	saveImage(cropped, "tmp/u_percent.png")

	cropped = cropImage(im, image.Rect(610, 630, 610+100, 630+65))
	saveImage(cropped, "tmp/s_percent.png")

	cropped = cropImage(im, image.Rect(940, 630, 940+100, 630+65))
	saveImage(cropped, "tmp/g_percent.png")

	cropped = cropImage(im, image.Rect(734, 1781, 734+326, 1781+100))
	saveImage(cropped, "tmp/u_time.png")

	cropped = cropImage(im, image.Rect(734, 1920, 734+326, 1920+100))
	saveImage(cropped, "tmp/s_time.png")

	cropped = cropImage(im, image.Rect(734, 2067, 734+326, 2067+100))
	saveImage(cropped, "tmp/g_time.png")
}

func analizeImageText(ctx context.Context, client *vision.ImageAnnotatorClient) string {
	  filelist := []string{
		"tmp/date.png",
		"tmp/type.png",
		"tmp/u_percent.png",
		"tmp/s_percent.png",
		"tmp/g_percent.png",
		"tmp/u_time.png",
		"tmp/s_time.png",
		"tmp/g_time.png",
	  }

	str := ""
	for _, file := range filelist {
		file, err := os.Open(file)

		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		defer file.Close()
		image, err := vision.NewImageFromReader(file)
		if err != nil {
			log.Fatalf("Failed to create image: %v", err)
		}

		labels, err := client.DetectDocumentText(ctx, image, nil)
		if err != nil {
			log.Fatalf("Failed to detect labels: %v", err)
		}

		for _, label := range labels.Pages {
			for _, block := range label.Blocks {
				for _, paragraph := range block.Paragraphs {
					for _, word := range paragraph.Words {
						s := []string{}
						for _, symbol := range word.Symbols {
							s = append(s, symbol.Text)
						}
						str += strings.Join(s, "")
					}
					str += ","
				}
			}			
		}
	}
	return str
}

func main() {
	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
	log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// fetch file list from img folder
	files, err := os.ReadDir("img")
	if err != nil {
		log.Fatal(err)
	}

	res, err := os.OpenFile("result.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Close()

	for _, file := range files {
		prepareCropImages("img/" + file.Name())
		str := analizeImageText(ctx, client)
		fmt.Fprintln(res, str)
	}
}