package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type App struct {
	ctx        context.Context
	downloader *Downloader
	mu         sync.Mutex
	isRunning  bool
}

type DownloadProgress struct {
	Percentage    float64 `json:"percentage"`
	Downloaded    float64 `json:"downloaded"`
	Total         float64 `json:"total"`
	Speed         float64 `json:"speed"`
	Status        string  `json:"status"`
	Error         string  `json:"error"`
	RemainingTime int64   `json:"remainingTime"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) StartDownload(url, outputPath string, workers int) error {
	a.mu.Lock()
	if a.isRunning {
		a.mu.Unlock()
		return fmt.Errorf("download already in progress")
	}
	a.isRunning = true
	a.mu.Unlock()

	// Create downloader
	downloader, err := NewDownloader(url, outputPath, workers)
	if err != nil {
		a.mu.Lock()
		a.isRunning = false
		a.mu.Unlock()
		return err
	}

	a.downloader = downloader

	// Start download in goroutine
	go func() {
		defer func() {
			a.mu.Lock()
			a.isRunning = false
			a.mu.Unlock()
		}()

		if err := downloader.Download(); err != nil {
			fmt.Printf("Download error: %v\n", err)
		}
	}()

	return nil
}

func (a *App) GetProgress() DownloadProgress {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.downloader == nil || !a.isRunning {
		return DownloadProgress{
			Status: "idle",
		}
	}

	a.downloader.ProgressMux.Lock()
	totalDownloaded := int64(0)
	for _, progress := range a.downloader.Progress {
		totalDownloaded += progress
	}
	a.downloader.ProgressMux.Unlock()

	percentage := float64(totalDownloaded) / float64(a.downloader.TotalSize) * 100
	speed := 0.0

	// Calculate speed (simple approximation)
	if percentage > 0 && percentage < 100 {
		speed = float64(totalDownloaded) / (1024 * 1024) // MB/s approximation
	}

	status := "downloading"
	if percentage >= 100 {
		status = "completed"
	}

	remainingBytes := a.downloader.TotalSize - totalDownloaded
	remainingTime := int64(0)
	if speed > 0 {
		remainingTime = int64(float64(remainingBytes) / (speed * 1024 * 1024))
	}

	return DownloadProgress{
		Percentage:    percentage,
		Downloaded:    float64(totalDownloaded) / (1024 * 1024),
		Total:         float64(a.downloader.TotalSize) / (1024 * 1024),
		Speed:         speed,
		Status:        status,
		RemainingTime: remainingTime,
	}
}

func (a *App) ValidateURL(url string) error {
	resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}

	return nil
}

func (a *App) SelectSaveLocation() (string, error) {
	// This would typically use a file dialog, but for simplicity we'll use a default location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	downloadDir := filepath.Join(homeDir, "Downloads")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", err
	}

	return downloadDir, nil
}

func (a *App) CancelDownload() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRunning {
		return fmt.Errorf("no download in progress")
	}

	a.isRunning = false
	a.downloader = nil

	return nil
}
