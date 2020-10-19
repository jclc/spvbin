I made https://github.com/jclc/spv as a better alternative. Don't use this.

# spvbin

Simple embedded SPIR-V for use with [vulkan-go](https://github.com/vulkan-go/vulkan) (and maybe [go-gl](https://github.com/go-gl/gl)). Similar to go-bindata, but keeps data in uint32 slices.

This program detects the byte order in the .spv file and works correctly for little-endian and big-endian files.

## Installing

`go get -u github.com/jclc/spvbin`

## Usage

`spvbin -package <package> [-out <output file>] [-export] <file.spv or dir/ [file2.spv or dir2/ ...]>`

Generate a file with `spvbin`. The generated file provides a getter for the binary data and a numeric constant for each .spv file.

Example:

`spvbin -package main shaders/frag.spv shaders/vert.spv -out spvbin.go`
or just `spvbin -package main shaders`

In your renderer code:
```
var vert, frag []uint32

vert = getSPV(spv_vert)
frag = getSPV(spv_frag)
```

Avoid name conflicts. eg. `dir1/frag.spv` and `dir2/frag.spv` will have the same index constant.

This program assumes your .spv files are valid SPIR-V; eg. they must be longer than 4 bytes, start with 0x07230203 and have a size divisible by 4 bytes.

## License

BSD 3-clause. Generated code is public domain or yours to choose.
