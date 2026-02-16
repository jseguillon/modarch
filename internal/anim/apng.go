package anim

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
)

func WriteAPNG(output string, framePaths []string, delayNumerator uint16) error {
	if len(framePaths) == 0 {
		return nil
	}
	frames := make([]parsedPNG, 0, len(framePaths))
	for _, p := range framePaths {
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		f, err := parsePNG(b)
		if err != nil {
			return err
		}
		frames = append(frames, f)
	}
	out := &bytes.Buffer{}
	out.Write(pngSignature)
	out.Write(writeChunk("IHDR", frames[0].ihdr))

	actl := make([]byte, 8)
	binary.BigEndian.PutUint32(actl[0:4], uint32(len(frames)))
	binary.BigEndian.PutUint32(actl[4:8], 0)
	out.Write(writeChunk("acTL", actl))

	seq := uint32(0)
	for i, f := range frames {
		fctl := make([]byte, 26)
		binary.BigEndian.PutUint32(fctl[0:4], seq)
		seq++
		binary.BigEndian.PutUint32(fctl[4:8], f.width)
		binary.BigEndian.PutUint32(fctl[8:12], f.height)
		binary.BigEndian.PutUint32(fctl[12:16], 0)
		binary.BigEndian.PutUint32(fctl[16:20], 0)
		binary.BigEndian.PutUint16(fctl[20:22], delayNumerator)
		binary.BigEndian.PutUint16(fctl[22:24], 100)
		fctl[24] = 0
		fctl[25] = 0
		out.Write(writeChunk("fcTL", fctl))

		for _, id := range f.idats {
			if i == 0 {
				out.Write(writeChunk("IDAT", id))
			} else {
				fd := make([]byte, 4+len(id))
				binary.BigEndian.PutUint32(fd[0:4], seq)
				seq++
				copy(fd[4:], id)
				out.Write(writeChunk("fdAT", fd))
			}
		}
	}
	out.Write(writeChunk("IEND", nil))
	return os.WriteFile(output, out.Bytes(), 0o644)
}

var pngSignature = []byte{137, 80, 78, 71, 13, 10, 26, 10}

type parsedPNG struct {
	ihdr          []byte
	idats         [][]byte
	width, height uint32
}

func parsePNG(b []byte) (parsedPNG, error) {
	if len(b) < 8 || !bytes.Equal(b[:8], pngSignature) {
		return parsedPNG{}, fmt.Errorf("invalid png")
	}
	o := 8
	var out parsedPNG
	for o+8 <= len(b) {
		ln := int(binary.BigEndian.Uint32(b[o : o+4]))
		typ := string(b[o+4 : o+8])
		o += 8
		if o+ln+4 > len(b) {
			return parsedPNG{}, fmt.Errorf("invalid chunk")
		}
		data := append([]byte{}, b[o:o+ln]...)
		o += ln + 4
		switch typ {
		case "IHDR":
			out.ihdr = data
			out.width = binary.BigEndian.Uint32(data[0:4])
			out.height = binary.BigEndian.Uint32(data[4:8])
		case "IDAT":
			out.idats = append(out.idats, data)
		case "IEND":
			return out, nil
		}
	}
	return out, nil
}

func writeChunk(typ string, data []byte) []byte {
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, binary.BigEndian, uint32(len(data)))
	buf.WriteString(typ)
	buf.Write(data)
	crc := crc32.NewIEEE()
	crc.Write([]byte(typ))
	crc.Write(data)
	_ = binary.Write(buf, binary.BigEndian, crc.Sum32())
	return buf.Bytes()
}
