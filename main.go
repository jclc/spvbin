package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	export = flag.Bool("export", false,
		"Set to true to export a getter function")
	pkg = flag.String("package", "",
		"Name of the package in the output file")
	outFile = flag.String("output", "spvbin.go",
		"Name of the output file")
	clearFunc = flag.Bool("clear-func", false,
		"Include a function to clears the modules from memory")
)

const (
	genComment = "// Code generated by github.com/jclc/spvbin. DO NOT EDIT."
)

func main() {
	flag.Parse()

	if *pkg == "" {
		fmt.Println("No package name given")
		os.Exit(1)
	}

	files := flag.Args()
	var dirs []int // indices of dirs to remove from 'files'

	for i, f := range files {
		info, err := os.Stat(f)
		if os.IsNotExist(err) {
			fmt.Println("File", f, "does not exist")
			os.Exit(1)
		}
		if info.IsDir() {
			// If a directory is supplied as an argument, add all .spv files in that directory.
			dirs = append(dirs, i)
			d, err := ioutil.ReadDir(f)
			if err != nil {
				panic(err)
			}

			for i := range d {
				if !d[i].IsDir() && strings.HasSuffix(d[i].Name(), ".spv") {
					files = append(files, filepath.Join(f, d[i].Name()))
				}
			}
		} else {
			if !strings.HasSuffix(f, ".spv") {
				fmt.Println("File", f, "is not an .spv file")
				os.Exit(1)
			}
		}
	}

	// Remove any directories in 'files'
	if dirs != nil {
		tmp := files
		files = make([]string, 0, len(tmp)-len(dirs))
		diri := 0
		for i := range tmp {
			if diri < len(dirs) && dirs[diri] == i {
				diri++
			} else {
				files = append(files, tmp[i])
			}
		}
	}

	if len(files) == 0 {
		fmt.Println("No output files.")
		os.Exit(1)
	}

	sort.Strings(files)

	r := strings.NewReplacer(
		".", "_",
	)
	var constPrefix string
	if *export {
		constPrefix = "SPV_"
	} else {
		constPrefix = "spv_"
	}
	constNames := make(map[string]string, len(files))
	for _, f := range files {
		fbase := filepath.Base(f)
		c := constPrefix + r.Replace(fbase[:len(fbase)-4]) // trim .spv
		constNames[f] = c
	}

	if err := os.MkdirAll(filepath.Dir(*outFile), os.ModePerm); err != nil {
		fmt.Println("Error creating output file directory:", err)
		os.Exit(1)
	}
	w, err := os.Create(*outFile)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		os.Exit(1)
	}
	defer w.Close()

	// Write header

	w.WriteString(genComment)
	w.WriteString("\n\n")
	w.WriteString("package ")
	w.WriteString(*pkg)
	w.WriteString("\n\n")

	// Write constants

	var constType string
	if *export {
		constType = "SPVModuleIndex"
	} else {
		constType = "spvModuleIndex"
	}

	var getter string
	if *export {
		getter = "GetSPV"
	} else {
		getter = "getSPV"
	}

	w.WriteString("type " + constType + " int\n\nconst (\n")
	for i, f := range files {
		w.WriteString("\t")
		w.WriteString(constNames[f])
		if i == 0 {
			w.WriteString(" " + constType + " = iota")
		}
		w.WriteString("\n")
	}
	w.WriteString(")\n\n")

	// Write getter

	w.WriteString("// " + getter + " returns a SPIR-V module for the given index.\n")

	w.WriteString("func " + getter + "(which " + constType + ") []uint32 {\n" +
		"\tif which < 0 || which > ")
	w.WriteString(fmt.Sprintf("%d {\n", len(files)-1))
	w.WriteString("\t\tpanic(\"Invalid spvbin index\")\n")
	w.WriteString("\t}\n\n")
	w.WriteString("\tif _spvBin == nil {\n")
	w.WriteString("\t\tpanic(\"spvbin data already cleared\")\n")
	w.WriteString("\t}\n\n")
	w.WriteString("\treturn _spvBin[which]\n")
	w.WriteString("}\n\n")

	// Write binary data

	h := hex.NewEncoder(w)
	e := binary.BigEndian // Literals in Go code are always big-endian

	w.WriteString("var _spvBin = [][]uint32{\n")
	for _, f := range files {
		spv, err := os.Open(f)
		if err != nil {
			fmt.Printf("Error opening file %s:", f)
			fmt.Println(err)
			os.Exit(1)
		}

		w.WriteString("\t[]uint32{")
		var b bytes.Buffer
		var fileEndianness binary.ByteOrder
		var bb [4]byte
		var ui uint32

		spv.Read(bb[:])
		if bb[0] != 0x07 { // assume spirv magic number; 0x07230203
			fileEndianness = binary.LittleEndian
		} else {
			fileEndianness = binary.BigEndian
		}
		ui = fileEndianness.Uint32(bb[:])
		w.WriteString("0x")
		binary.Write(h, e, ui)
		w.WriteString(", ")

		_, err = b.ReadFrom(spv)
		if err != nil {
			fmt.Printf("Error reading file %s:", f)
			fmt.Println(err)
			os.Exit(1)
		}

		for {
			// SPIR-V should always be 4-byte aligned, so we don't worry about
			// how many bytes we read
			_, err := b.Read(bb[:])
			if err != nil {
				break
			}
			w.WriteString("0x")
			ui = fileEndianness.Uint32(bb[:])
			binary.Write(h, e, ui)

			if b.Len() == 0 {
				break
			} else {
				w.WriteString(", ")
			}
		}

		w.WriteString("},\n")
	}
	w.WriteString("}\n\n")

	// Write clearing function

	if *clearFunc {
		var clearFunc string
		if *export {
			clearFunc = "SPVClear"
		} else {
			clearFunc = "spvClear"
		}

		w.WriteString("// " + clearFunc + " clears the embedded SPIR-V modules from memory.\n")
		w.WriteString("func " + clearFunc + "() {\n")
		w.WriteString("\t_spvBin = nil\n")
		w.WriteString("}\n\n")
	}
}
