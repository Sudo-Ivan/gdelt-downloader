# GDELT Downloader

Download GDELT data with file integrity checks, and options to unzip files or check for new ones.

## Features

-   **Download GDELT Data**: Fetches the master file list and downloads zip files.
-   **Remember Downloaded Files**: Skips already downloaded files using `downloaded_files.log`.
-   **Unzip Functionality**: Unzips all downloaded `.zip` files in the `gdelt_data` directory when `--unzip` flag passed.
-   **Check for New Files**: Identifies new files available in the master list without downloading with `--check-new`.

## Usage

1.  **Build**:
    ```bash
    go build -o gdelt-downloader main.go
    ```

2.  **Run the downloader**:
    -   **Default (download files)**:
        ```bash
        ./gdelt-downloader
        ```
    -   **Unzip all downloaded files**:
        ```bash
        ./gdelt-downloader --unzip
        ```
    -   **Check for new files (without downloading)**:
        ```bash
        ./gdelt-downloader --check-new
