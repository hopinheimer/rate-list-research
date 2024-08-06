package util

import (
	"fmt"
	"io"
	"os"
)

func LogHello() {
	file, err := os.Open("./util/hello.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	fmt.Println(string(content))
}
