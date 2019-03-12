# spvbin

Simple embedded SPIR-V for use with [vulkan-go](https://github.com/vulkan-go/vulkan) (and maybe [go-gl](https://github.com/go-gl/gl)). Similar to go-bindata, but keeps data in uint32 slices.

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

Avoid name conflicts.