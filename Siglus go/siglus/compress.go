package siglus

import "encoding/binary"

const (
	lookback    = 4096
	level       = 17
	acceptLevel = 3
)

// Decompress décompresse les données selon l'algorithme LZ77 de SiglusEngine.
// Le format attendu en entrée est : [complen int32][declen int32][données compressées...]
func Decompress(data []byte) []byte {
	if len(data) < 8 {
		return nil
	}
	decLen := int(binary.LittleEndian.Uint32(data[4:8]))
	out := make([]byte, decLen)
	src := data[8:]
	dst := 0

	for dst < decLen {
		if len(src) == 0 {
			break
		}
		flag := src[0]
		src = src[1:]

		for i := 0; i < 8 && dst < decLen; i++ {
			if len(src) == 0 {
				break
			}
			if flag&1 != 0 {
				// Octet littéral
				out[dst] = src[0]
				dst++
				src = src[1:]
			} else {
				// Référence arrière : 2 octets
				if len(src) < 2 {
					break
				}
				ref := int(binary.LittleEndian.Uint16(src[0:2]))
				src = src[2:]
				copyLen := (ref & 0x0F) + 2
				offset := ref >> 4
				for j := 0; j < copyLen && dst < decLen; j++ {
					out[dst] = out[dst-offset]
					dst++
				}
			}
			flag >>= 1
		}
	}
	return out
}

// Compress compresse les données selon l'algorithme LZ77 de SiglusEngine.
// Retourne un buffer avec header [complen int32][declen int32][données compressées...]
func Compress(in []byte) []byte {
	srcLen := len(in)
	// Buffer de sortie worst case
	out := make([]byte, 8+srcLen*2)
	outPos := 8
	inPos := 0

	for inPos <= srcLen {
		ctrlPos := outPos
		outPos++
		ctrlByte := byte(0)

		for i := 0; i < 8; i++ {
			if inPos >= srcLen {
				break
			}

			offset, mlen := searchData(in, inPos)

			if offset == 0 {
				// Octet littéral
				if outPos >= len(out) {
					out = append(out, 0)
				}
				out[outPos] = in[inPos]
				outPos++
				inPos++
				ctrlByte |= 1 << uint(i)
			} else {
				// Référence arrière
				ref := uint16((offset << 4) | (mlen & 0x0F))
				if outPos+1 >= len(out) {
					out = append(out, 0, 0)
				}
				binary.LittleEndian.PutUint16(out[outPos:], ref)
				outPos += 2
				inPos += mlen + 2
			}

			if inPos >= srcLen {
				break
			}
		}

		if ctrlPos < len(out) {
			out[ctrlPos] = ctrlByte
		}
	}

	// Écriture du header
	compLen := outPos
	binary.LittleEndian.PutUint32(out[0:4], uint32(compLen))
	binary.LittleEndian.PutUint32(out[4:8], uint32(srcLen))

	return out[:compLen]
}

// searchData cherche la meilleure correspondance dans la fenêtre glissante.
// Retourne (offset, longueur-2) ou (0, 0) si aucune correspondance valide.
func searchData(buf []byte, pos int) (offset, mlen int) {
	if pos < acceptLevel {
		return 0, 0
	}

	maxSearch := pos
	if maxSearch > lookback {
		maxSearch = lookback
	}

	maxMatch := level
	if pos+maxMatch > len(buf) {
		maxMatch = len(buf) - pos
	}

	if maxMatch < acceptLevel {
		return 0, 0
	}

	bestLen := 0
	bestOff := 0

	for off := 1; off <= maxSearch; off++ {
		matchLen := 0
		start := pos - off
		for matchLen < maxMatch && buf[start+matchLen] == buf[pos+matchLen] {
			matchLen++
		}
		if matchLen > bestLen {
			bestLen = matchLen
			bestOff = off
		}
		if bestLen == maxMatch {
			break
		}
	}

	if bestLen < acceptLevel {
		return 0, 0
	}
	return bestOff, bestLen - 2
}
