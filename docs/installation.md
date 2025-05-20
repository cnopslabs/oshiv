# Installation Guide

## Install `oshiv` using Homebrew

`oshiv` can be installed using [Homebrew](https://brew.sh) for macOS and Linux.

1. Add the `cnopslabs/oshiv` tap:
   ```bash
   brew tap cnopslabs/oshiv https://github.com/cnopslabs/oshiv
   ```

2. Install `oshiv`:
   ```bash
   brew install oshiv
   ```

3. Verify the installation:
   ```bash
   oshiv -h
   ```

## Manual Installation

You can download the latest binary from the [oshiv releases](https://github.com/cnopslabs/oshiv/releases) page.

### Place the Binary in Your `PATH`

#### macOS/Linux

After downloading the binary, move it to a directory included in your `PATH` (e.g., `/usr/local/bin` or any custom location).

**Example: Adding Binary to a Custom Path**

1. Check your `PATH`:
   ```bash
   echo $PATH
   ```
   Example output:
   ```
   /usr/local/bin:/Users/YOUR_USER/.local/bin
   ```

2. Move the binary to a directory in your `PATH` (e.g., `~/.local/bin`):
   ```bash
   mv ~/Downloads/oshiv ~/.local/bin
   ```

3. For macOS, clear quarantine (if applicable):
   ```bash
   sudo xattr -d com.apple.quarantine ~/.local/bin/oshiv
   ```

4. Make the binary executable:
   ```bash
   chmod +x ~/.local/bin/oshiv
   ```

5. Verify the installation:
   ```bash
   oshiv -h
   ```

#### Windows Setup

To use `oshiv` on Windows, you need to add its location to the `PATH` environment variable.

Steps:

1. Open **Control Panel** → **System** → **System Settings** → **Environment Variables**.
2. Scroll down in the **System Variables** section and locate the `PATH` variable.
3. Click **Edit** and add the location of your `oshiv` binary to the `PATH` variable. For example, `c:\oshiv`.

   *Note: When adding a new location, ensure that a semicolon (`;`) is included as a delimiter if appending to existing entries. Example: `c:\path;c:\oshiv`.*

4. Launch a new console session to apply the updated environment variable.

### Verify Installation

Once your `PATH` is updated, verify the installation by running:

```bash
oshiv -h
```

If the command prints the `oshiv` help information, the setup is complete.

### Troubleshooting

If oshiv gets quarantined by your OS:

```
sudo xattr -d com.apple.quarantine PATH_TO_OSHIV
```

Example:

```
sudo xattr -d com.apple.quarantine ~/.local/bin/oshiv
```