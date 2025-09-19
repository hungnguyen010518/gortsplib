# librtsp_client Build Instructions

This Makefile simplifies the building of the `librtsp_client.so` shared library for different target architectures and projects.

## Targets

The Makefile provides the following targets:

- **`all` (default):** Builds the library for the `oiuc-dev` project (x86_64 architecture).
- **`build0`:** Builds the library for the `oiuc-dev` project (x86_64 architecture).
- **`build1`:** Builds the library for the `nvdicom` project (ARMv8 architecture).
- **`build2`:** Builds the library for the `riuc-v2` project (ARMv8 architecture, performs an additional `go mod tidy` in the `RTSPClient` subdirectory).
- **`clean`:** Removes the generated library files and the header file.

## Usage

1. **Prerequisites:**
   - Make sure you have Go installed and configured on your system.
   - For ARMv8 targets (`build1` and `build2`), you'll need the `aarch64-linux-gnu-gcc` cross-compiler installed.

2. **Building:**
   - **Default (oiuc-dev, x86_64):** Simply run `make` in your terminal.
   - **Specific target:** Run `make <target>`, where `<target>` is one of `build0`, `build1`, or `build2`.

3. **Cleaning:**
   - Run `make clean` to remove all generated files. 

## Example

To build the library for the `nvdicom` project:

```bash
make build1 


