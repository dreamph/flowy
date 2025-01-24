package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

func Benchmark_GoHttpRequests(b *testing.B) {
	b.Run("UploadTest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			UploadTest()
		}
	})
}

func UploadTest() {
	// Define the URL
	url := "http://localhost:3000/upload"

	// Create a buffer to hold the multipart form data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add the "name" field
	_ = writer.WriteField("name", "test")

	// Add the first file
	if err := addFile(writer, "file", "../api.go"); err != nil {
		fmt.Println("Error adding file:", err)
		return
	}

	// Add the second file
	if err := addFile(writer, "file2", "../utils.go"); err != nil {
		fmt.Println("Error adding file2:", err)
		return
	}

	// Close the writer to finalize the form data
	if err := writer.Close(); err != nil {
		fmt.Println("Error closing writer:", err)
		return
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set the content type to multipart/form-data
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Print the response
	_, _ = io.ReadAll(resp.Body)
	//fmt.Printf("Response: %s\n", respBody)
}

// Helper function to add a file to the multipart writer
func addFile(writer *multipart.Writer, fieldName, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := writer.CreateFormFile(fieldName, file.Name())
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	return err
}
