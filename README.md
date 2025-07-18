# Audio Synchronization Tool

This is a command-line tool written in Go to process and synchronize audio files based on a manifest. It provides functionalities to adjust the speed of audio clips to match specified durations and to build a single audio track from multiple clips, inserting silence where necessary to maintain correct timing.

## Prerequisites

Before using this tool, you must have the following installed on your system:

- **Go**: Version 1.18 or later.
- **ffmpeg**: A recent version of ffmpeg must be available in your system's PATH.

## Installation and Building

1.  **Clone the repository (if you haven't already):**
    ```sh
    git clone <repository-url>
    cd sync-audio-with-timestamps
    ```

2.  **Build the application:**
    Run the following command from the project root to build the `sync-audio` executable:
    ```sh
    go build -o sync-audio .
    ```

## Manifest File Format

The tool relies on a manifest file (`.txt`) to understand the structure of the audio. Each line in the manifest represents a single audio segment and must follow this format:

`[<start_time>s–<end_time>s] (<speaker_id>) /path/to/audio/file.wav`

-   **`<start_time>`/`<end_time>`**: The target start and end times for the clip in seconds (e.g., `5.7s`, `12.0s`).
-   **`<speaker_id>`**: An identifier for the speaker (e.g., `SPEAKER_00`).
-   **`/path/to/audio/file.wav`**: The absolute or relative path to the audio file for that segment.

**Example:**
```
[0.0s–5.0s] (SPEAKER_00) /path/to/audio/000.wav
[5.7s–8.4s] (SPEAKER_00) /path/to/audio/001.wav
```

## Usage

The tool has two main commands: `adjust-speed` and `build`.

### 1. Adjust Speed (`adjust-speed`)

This command reads a manifest file, compares the duration of each audio clip to the target duration specified by the timestamps, and creates new, speed-adjusted versions of the audio files. It also generates a new manifest file pointing to these new clips.

**Command:**
```sh
./sync-audio adjust-speed --manifest /path/to/your/manifest.txt
```

**Arguments:**
-   `--manifest` or `-m`: (Required) The path to the input manifest file.

**Process:**
1.  For each entry, it calculates the required speed factor (`actual_duration / manifest_duration`).
2.  The speed factor is clamped to a safe range (`0.8`–`1.5`) to avoid heavy distortion.
3.  A new audio file is created with the `_synced` suffix (e.g., `000_synced.wav`).
4.  After processing all entries, a new manifest file is created with the `_synced` suffix (e.g., `manifest_synced.txt`) containing the paths to the new audio files.

### 2. Build (`build`)

This command takes a manifest file (typically the `_synced` manifest from the `adjust-speed` step) and concatenates all the audio clips into a single audio file, respecting the timestamps.

**Command:**
```sh
./sync-audio build --manifest /path/to/manifest_synced.txt --output final_track.wav
```

**Arguments:**
-   `--manifest` or `-m`: (Required) The path to the manifest file containing the clips to be merged.
-   `--output` or `-o`: (Required) The path for the final, combined audio file.

**Process:**
1.  The command processes the manifest entries in order.
2.  If there is a time gap between the end of one clip and the start of the next, it generates and inserts a corresponding period of silence.
3.  It progressively concatenates the clips and silence into a single track.
4.  The final, complete audio track is saved to the specified output path.
