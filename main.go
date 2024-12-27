package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	inputPath := flag.String("input", ".", "The path where the tool will look for files.")
	layoutPath := flag.String("layout", "./_layout.html", "The path to the file which be used as the common layout.")
	outputPath := flag.String("output", ".", "The output path")
	flag.Parse()

	layout := readLayout(*layoutPath)

	inputFiles := getFilePaths(*inputPath)
	if len(inputFiles) <= 0 {
		fmt.Println("Found no files to process.")
		return
	}

	remain(inputFiles, layout, *outputPath)
}

func readLayout(layoutPath string) string {
	layoutBytes, err := os.ReadFile(layoutPath)
	if err != nil {
		fmt.Printf("Failed to read layout file: %v\n", layoutPath)
		return "<html><body><main></main></body></html>"
	}

	return string(layoutBytes)
}

func readContent(path string) string {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Failed to read file: %v", path)
		return ""
	}

	return string(contentBytes)
}

func getFilePaths(inputPath string) []string {
	var files []string

	err := filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		fmt.Printf("Error while gathering input files: %v", err)
	}

	return files
}

func remain(inputFiles []string, layout string, outputDir string) {
	mainReg := regexp.MustCompile(`<main[^>]*>(.*?)</main>`)
	for _, contentFile := range inputFiles {
		fullContent := readContent(contentFile)
		innerContent := string(mainReg.Find([]byte(fullContent)))
		if len(innerContent) == 0 {
			innerContent = fullContent
		}
		innerContent = "<main>" + innerContent + "</main>"
		combinedContent := mainReg.ReplaceAllString(layout, innerContent)

		outputPath := filepath.Join(outputDir, contentFile)
		dir := filepath.Dir(outputPath)

		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			println("Failed to create directory.", dir)
		}

		file, err := os.Create(outputPath)
		if err != nil {
			println("Failed to create file.", file, err)
		}
		defer file.Close()

		_, err = file.WriteString(combinedContent)
		if err != nil {
			println("Failed to write content.", err)
		}
	}
}

func printHelp() {
	fmt.Print(
		`RE<MAIN>
Version (TODO)

Usage: remain [options]

Options:
 --layout <layout-file>    The path the to file which will be used as the common layout (default: _layout.html).
 --output <output-path>    The output path (default: overwritten in place).
 --input <input-path>      The path where the tool will look for files (default: current directory).
 --help                    Display this help information.

Description:
 This tool processes HTML files in the target directory by wrapping their contents in a common layout.
 The tool rewrites the contents of the targetted HTML files to be the specified layout file content with its <main> tag contents replaced by the contents of the original file.

Examples:
 remain                                     Process .html files in the current directory and its subdirectories.
 remain --layout my-layout.html             Use my-layout.html as the layout file instead of the default _layout.html.
 remain --output publish --input my-site    Process files within the my-site directory and write the results to the publish directory.
 remain --help                              Display this help information.
`)
}
