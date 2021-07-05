package net_io

import (
	"strings"
)

const (
    EXT = 3
)

// NormalizeWriteData builds the write data with EXT as the last byte
func NormalizeWriteData(writeData string) []byte {
    // fmt.Println(writeData) // DEBUG

    // Removes '\n'
    res := strings.ReplaceAll(writeData, "\n", "")

    // Init a slice with len + 1 for EXT
    slice := make([]byte, 0, len(res) + 1)

    // Append line into slice
    slice = append(slice, res...)

    // Add EXT at the end of slice
    slice = append(slice, EXT)

    // fmt.Println(slice) // DEBUG
    return slice
}

// NormalizeReadData builds the read data without EXT as the last byte
func NormalizeReadData(readData string) string {

    // Get the lenght of the read data
    lenData := len(readData)

    // Init a slice with juste enough space for the actual data whithout any '\n' or EXT
    slice := make([]byte, 0, lenData)

    // Append the data with EXT
    slice = append(slice, readData...)

    // Remove the last byte (EXT)
    slice = append(slice[:lenData - 1], slice[lenData:]...)

    return string(slice)
}
