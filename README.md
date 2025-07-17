# GDELT Downloader

Download GDELT data with file integrity checks, and options to unzip files or check for new ones.

## Features

-   **Download GDELT Data**: Fetches the master file list and downloads zip files.
-   **Remember Downloaded Files**: Skips already downloaded files using `downloaded_files.log`.
-   **Unzip Functionality**: Unzips all downloaded `.zip` files in the `gdelt_data` directory when `--unzip` flag passed.
-   **Check for New Files**: Identifies new files available in the master list without downloading with `--check-new`.

## Usage

### Download Binaries from Releases

You can download pre-compiled binaries for your operating system directly from the [GitHub Releases page](https://github.com/Sudo-Ivan/gdelt-downloader/releases). This is the quickest way to get started without needing to install Go or Docker.

### Go Installation and Usage

If you have Go installed, you can build and run the application directly.

1.  **Install the executable (recommended)**:
    ```bash
    go install github.com/Sudo-Ivan/gdelt-downloader@latest
    ```
    This will install the `gdelt-downloader` executable in your `$GOPATH/bin` directory (or `$HOME/go/bin` if `GOPATH` is not set). Make sure this directory is in your system's `PATH`.

2.  **Manual Build and Run**:
    Alternatively, you can build and run the application from the source code:
    ```bash
    go build -o gdelt-downloader main.go
    ./gdelt-downloader
    ```
    Or run directly:
    ```bash
    go run main.go
    ```

    **Command-line Flags**:
    -   **Default (download files)**:
        ```bash
        ./gdelt-downloader
        ```
        or
        ```bash
        go run main.go
        ```
    -   **Unzip all downloaded files**:
        ```bash
        ./gdelt-downloader --unzip
        ```
        or
        ```bash
        go run main.go --unzip
        ```
    -   **Check for new files (without downloading)**:
        ```bash
        ./gdelt-downloader --check-new
        ```
        or
        ```bash
        go run main.go --check-new
        ```

### Docker Usage

You can either build the Docker image yourself or pull it from GitHub Container Registry (GHCR).

#### Pulling from GitHub Container Registry (GHCR)

```bash
docker pull ghcr.io/sudo-ivan/gdelt-downloader:latest
```

#### Building the Docker Image (Optional)

If you prefer to build the image locally:

```bash
docker build -t gdelt-downloader .
```

#### Running the Docker Container

Before running, ensure your `gdelt_data` directory has the correct permissions for the container's user (UID/GID 65532, common for Chainguard images):

```bash
mkdir -p gdelt_data
sudo chown -R 65532:65532 gdelt_data
```

Now, run the container. Replace `gdelt-downloader` with `ghcr.io/sudo-ivan/gdelt-downloader:latest` if you pulled the image from GHCR.

-   **Default (download files)**:
    ```bash
    docker run --rm -v "./gdelt_data:/gdelt_data" ghcr.io/sudo-ivan/gdelt-downloader:latest
    ```
-   **Unzip all downloaded files**:
    ```bash
    docker run --rm -v "./gdelt_data:/gdelt_data" ghcr.io/sudo-ivan/gdelt-downloader:latest --unzip
    ```
-   **Check for new files (without downloading)**:
    ```bash
    docker run --rm -v "./gdelt_data:/gdelt_data" ghcr.io/sudo-ivan/gdelt-downloader:latest --check-new
    ```
The `-v "./gdelt_data:/gdelt_data"` flag mounts the local `gdelt_data` directory into the container, allowing downloaded files to persist on your host machine.

After downloading change permissions to `$USER:$USER` or your user to access the files. 
