package pedersen

import (
	"fmt"
	"github.com/dchest/blake256"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"math/big"
)

const (
	windowSize         = 4
	nWindowsPerSegment = 50
	GenpointPrefix     = "PedersenGenerator"
)

func Hash(message []byte) *babyjub.Point {
	bitsPerSegment := windowSize * nWindowsPerSegment

	bitLen := 8 * len(message) //bits(message)
	nSegments := (bitLen + bitsPerSegment - 1) / bitsPerSegment

	accP := babyjub.NewPoint()

	for s := 0; s < nSegments; s++ {
		nWindows := 0
		if s == nSegments-1 {
			nWindows = (bitLen - (nSegments-1)*bitsPerSegment + windowSize - 1) / windowSize
		} else {
			nWindows = nWindowsPerSegment
		}
		escalar := big.NewInt(0)
		exp := big.NewInt(1)

		for w := 0; w < nWindows; w++ {
			o := s*bitsPerSegment + w*windowSize
			acc := big.NewInt(1)
			for b := 0; b < windowSize-1 && o < bitLen; b++ {
				if bits(message, o) == 1 {
					acc = new(big.Int).Add(acc, new(big.Int).Lsh(big.NewInt(1), uint(b)))
				}
				o++
			}
			if o < bitLen {
				if bits(message, o) == 1 {
					acc = new(big.Int).Neg(acc)
				}
				o++
			}
			escalar = new(big.Int).Add(escalar, new(big.Int).Mul(acc, exp))
			exp = new(big.Int).Lsh(exp, windowSize+1)
		}
		if escalar.Sign() < 0 {
			escalar = new(big.Int).Add(escalar, babyjub.SubOrder)
		}

		basePoint := generateBasePoint(s)
		accP = eccAdd(accP, basePoint.Mul(escalar, basePoint))
	}
	return accP
}

func bits(bs []byte, pos int) byte {
	return (bs[pos/8] >> (pos % 8)) & 1
}

func eccAdd(p1, p2 *babyjub.Point) *babyjub.Point {
	p1Proj := p1.Projective()
	p1Proj = p1Proj.Add(p1Proj, p2.Projective())
	return p1Proj.Affine()
}

func Blake256(m []byte) []byte {
	h := blake256.New()
	_, err := h.Write(m[:])
	if err != nil {
		panic(err)
	}
	return h.Sum(nil)
}

func generateBasePoint(pointIdx int) *babyjub.Point {
	tryIdx := 0
	point := babyjub.NewPoint()

	for {
		s := GenpointPrefix + "_" + padLeftZeros(pointIdx) + "_" + padLeftZeros(tryIdx)
		hSlice := Blake256([]byte(s))
		var h [32]byte
		copy(h[:], hSlice[:32])

		h[31] = h[31] & 0xBF
		point, err := point.Decompress(h)
		if err == nil {
			point = point.Mul(big.NewInt(8), point)

			if !point.InCurve() {
				panic("not on curve!")
			}

			return point
		}
		tryIdx += 1
	}
}

func padLeftZeros(i int) string {
	return fmt.Sprintf("%032d", i)
}
