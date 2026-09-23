package main

import (
	"debug/dwarf"
	"debug/elf"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func section(file *elf.File, name string) []byte {
	value := file.Section(name)
	if value == nil {
		return nil
	}
	data, err := value.Data()
	if err != nil {
		panic(err)
	}
	return data
}

func sample(version int) string {
	directory, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	target, err := os.CreateTemp("", "dwarf-fixture-*.elf")
	if err != nil {
		panic(err)
	}
	path := target.Name()
	target.Close()
	defer os.Remove(path)
	command := exec.Command("gcc", fmt.Sprintf("-gdwarf-%d", version), "-gstrict-dwarf", "-O0", "-no-pie", "-fdebug-prefix-map="+directory+"=/fixture", "source.c", "-o", path)
	if output, err := command.CombinedOutput(); err != nil {
		panic(fmt.Sprintf("gcc: %v: %s", err, output))
	}
	file, err := elf.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	for _, part := range []struct{ section, suffix string }{
		{".debug_info", "info"},
		{".debug_abbrev", "abbrev"},
		{".debug_line", "line"},
		{".debug_str", "str"},
		{".debug_line_str", "line-str"},
	} {
		data := section(file, part.section)
		if err := os.WriteFile(fmt.Sprintf("v%d-%s.bin", version, part.suffix), data, 0644); err != nil {
			panic(err)
		}
	}
	data, err := file.DWARF()
	if err != nil {
		panic(err)
	}
	reader := data.Reader()
	entries := 0
	names := []string{}
	var unit *dwarf.Entry
	for {
		entry, err := reader.Next()
		if err == io.EOF || entry == nil {
			break
		}
		if err != nil {
			panic(err)
		}
		if entry.Tag == 0 {
			continue
		}
		entries++
		if entry.Tag == dwarf.TagCompileUnit {
			unit = entry
		}
		if name, ok := entry.Val(dwarf.AttrName).(string); ok {
			names = append(names, name)
		}
	}
	if unit == nil {
		panic("missing compilation unit")
	}
	lines, err := data.LineReader(unit)
	if err != nil {
		panic(err)
	}
	rows := []string{}
	files := []string{}
	for _, file := range lines.Files() {
		if file != nil {
			files = append(files, filepath.Base(file.Name))
		}
	}
	for {
		var row dwarf.LineEntry
		err := lines.Next(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		name := ""
		if row.File != nil {
			name = filepath.Base(row.File.Name)
		}
		rows = append(rows, fmt.Sprintf("%d:%s:%d:%t", row.Address, name, row.Line, row.EndSequence))
	}
	return fmt.Sprintf("%d\t%d\t%s\t%s\t%s", version, entries, strings.Join(names, ","), strings.Join(files, ","), strings.Join(rows, ","))
}

func main() {
	for _, version := range []int{2, 4, 5} {
		fmt.Println(sample(version))
	}
}
