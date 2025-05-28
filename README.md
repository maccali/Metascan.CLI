# MetaScan

**MetaScan** is a cross-platform command-line tool that recursively scans files in a directory, extracts metadata (including EXIF for images), computes cryptographic hashes (MD5, SHA1, SHA256), and exports everything to CSV or JSON. If files contain GPS coordinates, it automatically generates a Google Maps link.

---

## 🚀 Features

✅ Recursive or non-recursive directory scanning  
✅ Metadata extraction (EXIF: camera info, date, dimensions, etc.)  
✅ Cryptographic hashes: **MD5**, **SHA1**, **SHA256**  
✅ Automatic **Google Maps** link if GPS data exists  
✅ Output to **CSV** or **JSON**  
✅ Generates a **manifest** with output file hashes and process summary  
✅ Pre-built binaries for **Windows**, **Linux**, and **macOS**

---

## ⚙️ Installation

Download the appropriate binary from `dist/`:

| Platform | Path                        |
| -------- | --------------------------- |
| Windows  | `dist/windows/metascan.exe` |
| Linux    | `dist/linux/metascan`       |
| macOS    | `dist/macos/metascan`       |

### ✅ Linux / macOS

1. Give execution permission:

```bash
chmod +x metascan
```

2. Run directly:

```bash
./metascan --help
```

or move to `/usr/local/bin` for global access:

```bash
sudo mv metascan /usr/local/bin/
```

Now you can run:

```bash
metascan --help
```

---

### ✅ Windows

1. Run from the directory:

```cmd
metascan.exe --help
```

or add the `dist\windows\` folder to the **System PATH**:

- Open: **Control Panel** → **System** → **Advanced system settings** → **Environment Variables**.
- Edit the **PATH** variable → Add full path to `dist\windows\`.

Then you can run from **any terminal**:

```cmd
metascan --help
```

---

## 🛠️ Usage

```bash
metascan --dir <directory> [options]
```

### Options:

| Option     | Description                                   | Default                 |
| ---------- | --------------------------------------------- | ----------------------- |
| `--dir`    | Path to directory to process                  | `.` (current directory) |
| `--output` | Base name for output file (without extension) | `file_metadata_report`  |
| `--r`      | Recursively process subdirectories            | `false`                 |
| `--ext`    | Filter files by extension (e.g., `.jpg`)      | _(no filter)_           |
| `--format` | Output format: `csv` or `json`                | `csv`                   |

---

### Example 1: Basic CSV output

```bash
metascan --dir ./photos --output metadata_photos
```

Output:

- `metadata_photos.csv`
- `metadata_photos-manifest.csv`

---

### Example 2: Recursive scan with JSON output, filtering `.jpg` files

```bash
metascan --dir ./images --r --ext=".jpg" --output images_report --format json
```

Output:

- `images_report.json`
- `images_report-manifest.json`

---

## 📦 What it generates

- **Report** with metadata, hashes, and optional Google Maps link.
- **Manifest** with:
  - Output file name and format
  - Process summary (attempted, processed, errors)
  - Cryptographic hashes of the output file
  - Timestamp of generation

---

## 🎯 Use cases

## 🎯 Use Cases

✅ **Digital Forensics**  
Analyze file integrity, verify authenticity, and track metadata for investigations, audits, or security assessments.

✅ **Compliance & Auditing**  
Ensure files meet regulatory and legal standards by verifying hashes, metadata, and provenance for audits or compliance reports.

✅ **Media Management**  
Organize and catalog large collections of images, documents, or videos by automatically extracting metadata and generating structured reports.

✅ **Backup & Deduplication**  
Identify duplicate files using cryptographic hashes, optimize storage, and maintain consistent backups across systems.

✅ **Geolocation Analysis**  
Automatically extract GPS data from files (e.g., photos) and generate direct Google Maps links to visualize spatial information.

✅ **Data Pipelines**  
Integrate as a preprocessing step to enrich datasets with metadata, validate file integrity, and prepare structured outputs for downstream processing.

✅ **Chain of Custody Documentation**  
Provide verifiable cryptographic fingerprints (MD5, SHA1, SHA256) for each file to maintain a robust chain of custody in legal or compliance scenarios.

✅ **Archival & Digital Preservation**  
Document metadata and generate integrity hashes to support long-term digital preservation and ensure future accessibility.

✅ **Content Verification in Media Workflows**  
Validate that digital assets (e.g., photos, videos) are unaltered by comparing cryptographic hashes at each stage of content production and distribution.

✅ **Automated Reporting for IT Asset Management**  
Generate detailed CSV/JSON reports of file inventories across systems for IT audits, asset management, and security compliance.

✅ **Research Data Integrity**  
Facilitate reproducible research by cataloging datasets with metadata and cryptographic hashes, ensuring consistency and integrity.

✅ **Quality Assurance in File Transfers**  
Verify the completeness and integrity of files after transfer or synchronization between systems, storage media, or cloud platforms.

---

## 🔒 Hash Functions

All three are computed for each file:

- **MD5**
- **SHA1**
- **SHA256**

---

## 🗺️ Geolocation Feature

If a file contains GPS data, MetaScan automatically includes a **Google Maps** link:

```
https://www.google.com/maps?q=<lat>,<long>
```

---

## 🧑‍💻 Dependencies

The binaries are **pre-built**, so no need to install dependencies.

For developers compiling from source:

- Requires **Go 1.16+**
- Uses: [goexif](https://github.com/rwcarlsen/goexif)

Install dependencies:

```bash
go mod tidy
```

---

## ✅ License

MIT License

---

## 💡 Contributions

Pull requests are welcome!  
Feel free to open issues for bugs or feature requests.

---

## 📌 Example of Manifest (JSON)

```json
{
  "output_file": "file_metadata_report.json",
  "output_format": "json",
  "total_attempted": 10,
  "total_processed": 9,
  "total_with_errors": 1,
  "output_file_hashes": {
    "md5": "abc123...",
    "sha1": "def456...",
    "sha256": "ghi789..."
  },
  "generated_at": "2025-05-28T12:34:56Z"
}
```

---

## 🔗 Example of Google Maps Link

```
https://www.google.com/maps?q=37.774929,-122.419416
```

---

## ⚡ Note on command name

The **command name** depends on the **binary file name**.

| File name                | How to run                              |
| ------------------------ | --------------------------------------- |
| `metascan.exe`           | `metascan` or `metascan.exe`            |
| `metascan` (Linux/macOS) | `./metascan` or `metascan` (if in PATH) |

If you rename the file to `scanner`, then the command will be `scanner`.
