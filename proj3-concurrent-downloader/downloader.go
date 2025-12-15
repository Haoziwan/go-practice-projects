package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type DownloadChunk struct {
	Start int64
	End   int64
	Index int
}

type Downloader struct {
	URL         string
	OutputPath  string
	NumWorkers  int
	TotalSize   int64
	ChunkSize   int64
	Progress    []int64
	ProgressMux sync.Mutex
}

func NewDownloader(url, outputPath string, numWorkers int) (*Downloader, error) {
	// Create HTTP client with better configuration
	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
	}

	// Create request with proper headers
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Concurrent-Downloader/1.0")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %s", resp.Status)
	}

	totalSize := resp.ContentLength
	if totalSize <= 0 {
		return nil, fmt.Errorf("unable to determine file size")
	}

	// Check if server supports range requests
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		fmt.Println("Warning: Server doesn't support range requests, using single-threaded download")
		numWorkers = 1
	}

	chunkSize := totalSize / int64(numWorkers)

	return &Downloader{
		URL:        url,
		OutputPath: outputPath,
		NumWorkers: numWorkers,
		TotalSize:  totalSize,
		ChunkSize:  chunkSize,
		Progress:   make([]int64, numWorkers),
	}, nil
}

func (d *Downloader) Download() error {
	fmt.Printf("Starting download: %s\n", d.URL)
	fmt.Printf("File size: %.2f MB\n", float64(d.TotalSize)/(1024*1024))
	fmt.Printf("Workers: %d\n", d.NumWorkers)

	// Create temporary directory for chunks
	tempDir := d.OutputPath + ".tmp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, d.NumWorkers)

	// Start progress monitor
	done := make(chan bool)
	go d.monitorProgress(done)

	// Start workers
	for i := 0; i < d.NumWorkers; i++ {
		start := int64(i) * d.ChunkSize
		end := start + d.ChunkSize - 1

		// Last chunk gets the remainder
		if i == d.NumWorkers-1 {
			end = d.TotalSize - 1
		}

		chunk := DownloadChunk{
			Start: start,
			End:   end,
			Index: i,
		}

		wg.Add(1)
		go d.downloadChunk(chunk, tempDir, &wg, errChan)
	}

	wg.Wait()
	done <- true
	close(errChan)

	// Check for errors
	if len(errChan) > 0 {
		err := <-errChan
		os.RemoveAll(tempDir)
		return fmt.Errorf("download failed: %w", err)
	}

	// Merge chunks
	fmt.Println("\nMerging chunks...")
	if err := d.mergeChunks(tempDir); err != nil {
		os.RemoveAll(tempDir)
		return fmt.Errorf("failed to merge chunks: %w", err)
	}

	// Clean up
	os.RemoveAll(tempDir)
	fmt.Println("\nDownload completed successfully!")

	return nil
}

func (d *Downloader) downloadChunk(chunk DownloadChunk, tempDir string, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	chunkPath := filepath.Join(tempDir, fmt.Sprintf("chunk_%d", chunk.Index))

	// Create HTTP request with Range header
	req, err := http.NewRequest("GET", d.URL, nil)
	if err != nil {
		errChan <- err
		return
	}

	rangeHeader := fmt.Sprintf("bytes=%d-%d", chunk.Start, chunk.End)
	req.Header.Set("Range", rangeHeader)
	req.Header.Set("User-Agent", "Concurrent-Downloader/1.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{
		Timeout: 30 * time.Minute,
		Transport: &http.Transport{
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			DisableKeepAlives:     false,
			DisableCompression:    true,
			ResponseHeaderTimeout: 60 * time.Second,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		errChan <- err
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("worker %d: server returned status %s", chunk.Index, resp.Status)
		return
	}

	// Create chunk file
	file, err := os.Create(chunkPath)
	if err != nil {
		errChan <- err
		return
	}
	defer file.Close()

	// Download with progress tracking
	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				errChan <- writeErr
				return
			}

			d.ProgressMux.Lock()
			d.Progress[chunk.Index] += int64(n)
			d.ProgressMux.Unlock()
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			errChan <- err
			return
		}
	}
}

func (d *Downloader) mergeChunks(tempDir string) error {
	outputFile, err := os.Create(d.OutputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	for i := 0; i < d.NumWorkers; i++ {
		chunkPath := filepath.Join(tempDir, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(outputFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Downloader) monitorProgress(done chan bool) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			d.ProgressMux.Lock()
			totalDownloaded := int64(0)
			for _, progress := range d.Progress {
				totalDownloaded += progress
			}
			d.ProgressMux.Unlock()

			percentage := float64(totalDownloaded) / float64(d.TotalSize) * 100
			elapsed := time.Since(startTime).Seconds()
			speed := float64(totalDownloaded) / elapsed / (1024 * 1024) // MB/s

			fmt.Printf("\rProgress: %.2f%% | Downloaded: %.2f MB / %.2f MB | Speed: %.2f MB/s",
				percentage,
				float64(totalDownloaded)/(1024*1024),
				float64(d.TotalSize)/(1024*1024),
				speed)
		}
	}
}
