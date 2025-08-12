package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

func generateFileWithSize(size int) error {
	// Create a file with the specified size
	file, err := os.Create("upload.flv")
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Write random data to the file until it reaches the specified size
	data := make([]byte, 1024) // 1 KB buffer
	for i := range data {
		data[i] = byte(rand.Intn(256)) // Fill with random bytes
	}

	for i := 0; i < size/1024; i++ {
		// Fill the buffer with random data
		if _, err := file.Write(data); err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}

		rand.Shuffle(len(data), func(i, j int) {
			data[i], data[j] = data[j], data[i]
		})
	}

	return nil
}

func main() {
	// Check if the correct number of arguments is provided
	if len(os.Args) < 5 {
		fmt.Println("Usage: go run main.go <c_parameter> <auth_key_value>" +
			" <x_een_port_value> <size_of_file> <delay_seconds> <chunk_size> <output_file>")
		return
	}

	// Get the ~c~ parameter, ~auth_key~ value, and ~X-Een-Port~ value from the command-line arguments
	cParameter := os.Args[1]
	authKeyValue := os.Args[2]
	xEenPortValue := os.Args[3]

	sizeOfFileStr := os.Args[4]
	sizeOfFile, err := strconv.Atoi(sizeOfFileStr)
	if err != nil {
		fmt.Println("Error converting size of file to integer:", err)
		return
	}

	delaySecondsStr := os.Args[5]
	delaySeconds, err := strconv.Atoi(delaySecondsStr)
	if err != nil {
		fmt.Println("Error converting delay seconds to integer:", err)
		return
	}

	chunkSizeStr := os.Args[6]
	chunkSize, err := strconv.Atoi(chunkSizeStr)
	if err != nil {
		fmt.Println("Error converting chunk size to integer:", err)
		return
	}

	outputFile := os.Args[7]

	// convert outputFile to base64
	outputFileBase64 := base64.StdEncoding.EncodeToString([]byte(outputFile))

	// Generate a file with the specified size
	if err := generateFileWithSize(sizeOfFile); err != nil {
		fmt.Println("Error generating file:", err)
		return
	}

	// Open the file
	file, err := os.Open("upload.flv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Define the URL with the ~c~ parameter
	url := fmt.Sprintf("http://esn-smart-dispatcher/camera/command?t=download&c=%s&a=%s", cParameter,
		outputFileBase64)

	// Create a pipe
	reader, writer := io.Pipe()
	// Start a goroutine to write chunks to the pipe
	go func() {
		buffer := make([]byte, chunkSize)
		for {
			// Read a chunk from the file
			n, err := file.Read(buffer)
			if err != nil {
				if err == io.EOF {
					writer.Close() // Close the writer when done
					break
				}
				fmt.Println("Error reading file:", err)
				writer.CloseWithError(err)
				return
			}

			// Write the chunk to the pipe
			_, err = writer.Write(buffer[:n])
			if err != nil {
				fmt.Println("Error writing to pipe:", err)
				writer.CloseWithError(err)
				return
			}
			writer.Write([]byte("\r\n"))
			fmt.Println("chunk written")
			// Wait for 5 seconds before sending the next chunk
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}
	}()

	// Create a POST request with the pipe reader as the body
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Transfer-Encoding", "chunked")
	req.Header.Set("X-Een-Port", xEenPortValue) // Add the custom header

	// Add the cookie with ~auth_key~ and its value
	req.AddCookie(&http.Cookie{
		Name:  "auth_key",
		Value: authKeyValue,
	})
	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// dump response headers
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("Response header: %s: %s\n", key, value)
		}
	}

	// dump response status
	fmt.Printf("Response status: %s\n", resp.Status)

	// dump the response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println("Response body:", string(data))

	// Print status
	fmt.Println("File upload completed with status:", resp.Status)
}
