# ergomcutool

[![Ubuntu-latest](https://github.com/mcu-art/ergomcutool/actions/workflows/ubuntu-latest.yml/badge.svg)](https://github.com/mcu-art/ergomcutool/actions/workflows/ubuntu-latest.yml)
[![Ubuntu-coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/mcu-art/e32e37141973c5d7c3d84cacadaac090/raw/ergomcutool-codecov-ubuntu.json)](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/mcu-art/e32e37141973c5d7c3d84cacadaac090/raw/ergomcutool-codecov-ubuntu.json)

**ergomcutool** is a small, simple and intuitive project manager
that helps to integrate STM32 projects generated by STM32CubeMX into VSCode.
It provides a convenient way to manage STM32 projects on linux.

## Why ergomcutool?
STM32CubeMX already has an ability to generate Makefiles and CMake files. But there are still some issues with VSCode integration
that make the workflow more complicated than needed.

Benefits of `ergomcutool`:
+ less boilerplate
+ separation of machine-dependent settings and project settings
+ separation of concerns: you manage your files, STM32CubeMX manages its
+ automatic handling of VSCode intellisense


## Prerequisites
+ Linux OS (future versions may support Windows and MacOS).
  `ergomcutool` v1.1.0 was tested on Ubuntu 22.04.1.

+ STM32CubeMX
  https://www.st.com/en/development-tools/stm32cubemx.html
  `ergomcutool` v1.1.0 was tested with STM32CubeMX version 6.11.0

+ GNU ARM compiler:
```bash
sudo apt-get install gcc-arm-none-eabi
arm-none-eabi-gcc --version
# arm-none-eabi-gcc (15:10.3-2021.07-4) 10.3.1 20210621 (release)
```

+ openocd (On-Chip Debugger)
```bash
sudo apt-get install openocd
openocd --version
# Open On-Chip Debugger 0.11.0
```

+ gdb-multiarch
```bash
sudo apt-get install gdb-multiarch
gdb-multiarch --version
# GNU gdb (Ubuntu 12.1-0ubuntu1~22.04) 12.1
```
If `gdb-multiarch` can't be installed with `apt-get` due to a version conflict bug,
try installing it from a .deb file manually, e.g.
`http://archive.ubuntu.com/ubuntu/pool/universe/g/gdb/gdb-multiarch_12.1-0ubuntu1~22.04_amd64.deb`

+ ST-Link or similar device for programming and debugging the MCU that is supported by `openocd`.
  A list of the supported devices is usually located in
  `/usr/share/openocd/scripts/interface/` directory on Ubuntu.


## Installation
You need Go runtime to be installed on your computer.
The easiest way to install Go on Ubuntu is to use `update-golang` script
from https://github.com/udhos/update-golang:
```bash
git clone https://github.com/udhos/update-golang
cd update-golang
sudo ./update-golang.sh
```

Then install `ergomcutool`:
```bash
go install github.com/mcu-art/ergomcutool@v1.1.0

# Initialize ergomcutool
ergomcutool init
```
Edit `~/.ergomcutool/ergomcutool_config.yaml`
to specify your hardware debugger and other settings.


## Quick start
1. Create a new STM32 project with STM32CubeMX or use an existing one.
   Save the project.
   In the `Project Manager` tab select `Toolchain/IDE` Makefile.
   Press `Generate Code `button.

2. Create ergomcu project from the STM32CubeMX project:
   run `ergomcutool create` from the project root directory.

3. Edit the configuration files: `ergomcutool/ergomcu_project.yaml`
   and `_non_persistent/ergomcutool_config.yaml`.
   Other files like `.gitignore` and `.clang-format` may be adjusted as well.

4. Update your project: `ergomcutool update-project`.

5. Build your project `make`

6. Program your device: `make prog`


## Usage
`ergomcutool` works in tandem with STM32CubeMX, you have to create a project with it first.
Then create `ergomcu` project issuing `ergomcutool create` command.
There are two files that store your project settings:
`ergomcutool/ergomcu_project.yaml` for persistent
settings that you might want to commit, and
`_non_persistent/ergomcutool_config.yaml` for
non-persistent settings that aren't committed.

Each time you update `ergomcu_project.yaml`
or `ergomcutool_config.yaml`, run `ergomcutool update-project`
so that the changes take effect.

When changes are made by STM32CubeMX, there is no need
to manually run `ergomcutool update-project` as
it is done automatically via script.


### Setting up VSCode
The following VSCode extensions are required to be installed:
  + `C/C++` by Microsoft
  + `Cortex-Debug` by marus25

The `Makefile Tools` extension by Microsoft doesn't seem to work properly
with STM32CubeMX-generated makefiles and is recommended to be disabled.


### Adding source files
Edit `ergomcutool/ergomcu_project.yaml` to add C source files to the project.
List all your source files in the `c_src` section, e.g.
```yaml
c_src:
  - _external/example_lib/file1.c
  - "{{.EXAMPLE_LIB}}/file2.c"
```

The paths may contain Go template variables defined in the 
`external_dependencies` section, in that case use quotes
to keep yaml syntax valid, e.g.: `"{{.EXAMPLE_LIB}}/file2.c"`.

It is also possible to add all `.c` files from a directory, use `c_src_dirs` for this:
```yaml
c_src_dirs:
  - src_dir_1
  - ../path/to/src_dir_2
```


### Adding C include directories
For adding C include directories to the project use `c_include_dirs`:
```yaml
c_include_dirs:
  - _external/example_lib/include
```
The notes about template variables from section `Adding source files to your project` 
apply here as well.
The paths must not be prefixed with `-I`, the tool will
do it automatically when generating the Makefile.


### Adding C definitions
For adding C definitions into the project use `c_defs`:
```yaml
c_defs:
 - EXAMPLE_DEFINITION
```
The definitions must not be prefixed with `-D`, the tool will
do it automatically when generating the Makefile.


### SVD file
`.svd` (System View Description) files contain information about MCUs that makes
the debugging process more comfortable (register and control bit names etc.).
The debugging can be done without a `.svd` file,
in that case the `XPERIPHERIALS` tab will not be available in VSCode.

The `.svd` files are provided by ST company for each device family and can be downloaded
from their website, e.g.
`https://www.st.com/en/microcontrollers-microprocessors/stm32f103.html#cad-resources`.

If you choose to use an `.svd` file, you have two options:
1. Include it as a part of the project, then you specify
   the path to it in `ergomcutool/ergomcu_project.yaml`.
2. Use an external `.svd` file that is not a part of the project,
   in this case you specify the path to it in `_non_persistent/ergomcutool_config.yaml`.

If you don't use `.svd` file in your project, leave `svd_file_path` setting empty.


### External project dependencies
`ergomcutool` allows adding external dependencies to the project.
Permanent dependencies are supposed to be specified in 
`ergomcutool/ergomcu_project.yaml`.
Machine-specific dependencies should be specified in `ergomcutool_config.yaml`
either in the user home directory or in the project directory.

Example:
```yaml
external_dependencies:
 - var:                     EXAMPLE_LIB
   path:                    /path/to/example/lib
   create_in_project_link:  true
   link_name:               example_lib
```

In this example, an external dependency `EXAMPLE_LIB` is defined.

There are two ways of adding it into the project paths:
1. Referencing the `var` value (golang template syntax).
This way is recommended when `create_in_project_link` is false.
Note the quotes that are required for yaml syntax to be valid:
```yaml
c_src:
  - "{{.EXAMPLE_LIB}}/src/file1.c"
```
2. Referencing the `link_name` value.
This way is recommended when `create_in_project_link` is true.
It allows you to easily obtain the correct path in VSCode by right-clicking
on the chosen file in VSCode Explorer and selecting `Copy Relative Path`
from the context menu.
```yaml
c_src:
  - _external/example_lib/src/file1.c
```

The `create_in_project_link` setting instructs `ergomcutool`
to create a symlink in `your_project_root/_external/` directory
to the directory specified by the `path` value.
It is a convenient way to work with external files in VSCode,
it gives you an ability to view or edit them from inside your project.
If you don't need this functionality,
just specify `create_in_project_link: false`.
`link_name` specifies the name of the symlink to be created.
The `_external` directory is added to `.gitignore` by default.


### Intellisense
The VSCode intellisense is managed automatically by `ergomcutool`.
This is done by analyzing the Makefile in addition to `ergomcu_project.yaml`
and adding necessary entries to `.vscode/settings.json` and `.vscode/c_cpp_properties.json`.
All include directories specified in your project are always added to the intellisense.
All source file directories within your project (including ones taken from the Makefile)
are added to the intellisense as well.
This allows you to browse the files using `Go to definition` VSCode context help.
If adding source directories to the intellisense is not desired,
you can disable it by setting `intellisense.skip_adding_source_directories` to `true`
in `ergomcutool_config.yaml`.


#### Known intellisense issues
`C/C++` extension by `frannek94` uses `realpath` to obtain paths to 
the include and source directories. This behaviour leads to a possibility that
the same file can be opened in VSCode multiple times under different names, e.g.
`_external/my-lib/file1.c` and `/actual/path/to/my-lib/file1.c`.
The intellisense will still work, but one of the files will be ignored by it and a warning will be issued:
`Unable to process IntelliSense for file with same canonicalized path as existing file.`
To avoid this warning and avoid opening same file under different names,
a VSCode extension `Open file realpath` is currently being tested and will soon be published.



### Programming the MCU
To program the MCU, type `make prog` command in the terminal from the project root.
By default, `ergomcutool` adds `prog` target to the makefile based on
`~/.ergomcutool/assets/snippets/prog_task.txt.tmpl` template.

If you need to customize the `prog` target,
do not edit the Makefile directly, instead create file
`ergomcutool/snippets/prog_task.txt.tmpl` in your project directory
and customize it.

The default file contents looks as follows:
```Makefile
prog: $(BUILD_DIR)/$(TARGET).elf
	openocd -f interface/{{.OpenocdInterface}} -f target/{{.OpenocdTarget}} -c "program $(BUILD_DIR)/$(TARGET).elf verify exit reset"
```
This file is a Go template.
There are two template variables here:
  + `{{.OpenocdInterface}}` is the value of `openocd.interface`
    specified in the `ergomcutool_config.yaml`.
  + `{{.OpenocdTarget}}` is the value of `openocd.target`
    specified in the `ergomcu_project.yaml`.

Note: Makefile syntax doesn't allow usage of spaces for indentation,
always use tabs instead.
