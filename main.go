package main

import (
	"archive/zip"
	"bufio"
	"crypto/md5" // #nosec G501
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const masterFilelistURL = "http://data.gdeltproject.org/gdeltv2/masterfilelist.txt"
const downloadedFilesLog = "downloaded_files.log"
const downloadDir = "gdelt_data"

type GDELTFile struct {
	Size     int64
	MD5Sum   string
	URL      string
	Filename string
}

func main() {
	unzipFiles := flag.Bool("unzip", false, "Unzip all downloaded GDELT files")
	checkForNew := flag.Bool("check-new", false, "Check for new files and download them")
	flag.Parse()

	fmt.Println("GDELT Downloader started.")

	if *unzipFiles {
		fmt.Println("Unzipping all downloaded files...")
		if err := unzipAllFiles(); err != nil {
			fmt.Printf("Error unzipping files: %v\n", err)
		}
		fmt.Println("Unzipping complete.")
		return 
	}

	if err := os.MkdirAll(downloadDir, 0700); err != nil { // #nosec G301
		fmt.Printf("Error creating download directory: %v\n", err)
		return
	}

	downloaded := loadDownloadedFiles()
	fmt.Printf("Loaded %d previously downloaded files.\n", len(downloaded))

	masterList, err := fetchMasterFilelist(masterFilelistURL) //#nosec G107
	if err != nil {
		fmt.Printf("Error fetching master filelist: %v\n", err)
		return
	}
	fmt.Printf("Fetched %d files from master filelist.\n", len(masterList))

	if *checkForNew {
		fmt.Println("Checking for new files...")
		newFilesCount := 0
		for _, file := range masterList {
			if _, ok := downloaded[file.URL]; !ok {
				fmt.Printf("New file available: %s\n", file.Filename)
				newFilesCount++
			}
		}
		if newFilesCount == 0 {
			fmt.Println("No new files found.")
		} else {
			fmt.Printf("%d new files found.\n", newFilesCount)
		}
		return 
	}

	var wg sync.WaitGroup
	for _, file := range masterList {
		if _, ok := downloaded[file.URL]; ok {
			fmt.Printf("Skipping %s (already downloaded).\n", file.Filename)
			continue
		}

		wg.Add(1)
		go func(f GDELTFile) {
			defer wg.Done()
			if err := downloadFile(f); err != nil {
				fmt.Printf("Error downloading %s: %v\n", f.Filename, err)
			} else {
				markFileAsDownloaded(f.URL)
				fmt.Printf("Successfully downloaded %s\n", f.Filename)
			}
		}(file)
	}

	wg.Wait()
	fmt.Println("GDELT Downloader finished.")
}

func unzipAllFiles() error {
	files, err := os.ReadDir(downloadDir)
	if err != nil {
		return fmt.Errorf("failed to read download directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".zip") {
			zipFilePath := filepath.Join(downloadDir, file.Name())
			destDir := filepath.Join(downloadDir, strings.TrimSuffix(file.Name(), ".zip"))
			fmt.Printf("Unzipping %s to %s...\n", file.Name(), destDir)
			if err := unzipFile(zipFilePath, destDir); err != nil {
				fmt.Printf("Error unzipping %s: %v\n", file.Name(), err)
			} else {
				fmt.Printf("Successfully unzipped %s\n", file.Name())
			}
		}
	}
	return nil
}

func unzipFile(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := r.Close(); cerr != nil && err == nil { // #nosec G104
			err = cerr
		}
	}()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name) // #nosec G305

		// Check for ZipSlip (directory traversal)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, 0700); err != nil { // #nosec G301
				return err
			}
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), 0700); err != nil { // #nosec G301
				return err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode()) // #nosec G302, G304
			if err != nil {
				return err
			}

			rc, err := f.Open()
			if err != nil {
				outFile.Close() // #nosec G104
				return err
			}

			_, err = io.Copy(outFile, rc) // #nosec G110

			if cerr := outFile.Close(); cerr != nil && err == nil { // #nosec G104
				err = cerr
			}
			if cerr := rc.Close(); cerr != nil && err == nil { // #nosec G104
				err = cerr
			}

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func fetchMasterFilelist(url string) ([]GDELTFile, error) {
	resp, err := http.Get(url) // #nosec G107
	if err != nil {
		return nil, fmt.Errorf("failed to fetch master filelist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch master filelist: status code %d", resp.StatusCode)
	}

	var files []GDELTFile
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}

		size, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}

		fileURL := parts[2]
		filename := filepath.Base(fileURL)

		files = append(files, GDELTFile{
			Size:     size,
			MD5Sum:   parts[1],
			URL:      fileURL,
			Filename: filename,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading master filelist: %w", err)
	}

	return files, nil
}

func downloadFile(file GDELTFile) error {
	if strings.Contains(file.Filename, "..") || strings.Contains(file.Filename, "/") || strings.Contains(file.Filename, "\\") {
		return fmt.Errorf("invalid filename detected: %s", file.Filename)
	}

	filePath := filepath.Join(downloadDir, file.Filename)
	tempFilePath := filePath + ".tmp"

	req, err := http.NewRequest("GET", file.URL, nil) // #nosec G107
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %w", file.Filename, err)
	}

	var startByte int64
	if fileInfo, err := os.Stat(tempFilePath); err == nil {
		startByte = fileInfo.Size()
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startByte))
		fmt.Printf("Resuming download for %s from byte %d\n", file.Filename, startByte)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", file.Filename, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("unexpected status code for %s: %d", file.Filename, resp.StatusCode)
	}

	out, err := os.OpenFile(tempFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to open temporary file for %s: %w", file.Filename, err)
	}
	defer out.Close()

	hash := md5.New() // #nosec G401
	writer := io.MultiWriter(out, hash)

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write data for %s: %w", file.Filename, err)
	}

	calculatedMD5 := hex.EncodeToString(hash.Sum(nil))
	if calculatedMD5 != file.MD5Sum {
		return fmt.Errorf("MD5 sum mismatch for %s: expected %s, got %s", file.Filename, file.MD5Sum, calculatedMD5)
	}

	if err := os.Rename(tempFilePath, filePath); err != nil {
		return fmt.Errorf("failed to rename temporary file for %s: %w", file.Filename, err)
	}

	return nil
}

func loadDownloadedFiles() map[string]struct{} {
	downloaded := make(map[string]struct{})
	file, err := os.Open(downloadedFilesLog)
	if err != nil {
		if os.IsNotExist(err) {
			return downloaded
		}
		fmt.Printf("Error opening downloaded files log: %v\n", err)
		return downloaded
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		downloaded[scanner.Text()] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading downloaded files log: %v\n", err)
	}
	return downloaded
}

func markFileAsDownloaded(url string) {
	file, err := os.OpenFile(downloadedFilesLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("Error opening downloaded files log for writing: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "%s\n", url); err != nil {
		fmt.Printf("Error writing to downloaded files log: %v\n", err)
	}
}
