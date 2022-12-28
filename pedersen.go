package pedersen

import (
	"fmt"
	"github.com/dchest/blake256"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"math/big"
)

const GenpointPrefix = "PedersenGenerator"

func Hash(message []byte) *babyjub.Point {

	msg := bits(message)
	size := len(msg) / 200

	if len(msg)%200 > 0 {
		size += 1
	}

	H := babyjub.NewPoint()
	M := make([]*big.Int, size)

	for i := range M {
		M[i] = big.NewInt(0)
		ki := 200
		if i == len(M)-1 {
			ki = len(msg) % 200
		}
		var c uint = 0
		for j := 0; j < ki; j += 4 {
			encM := enc(msg[j], msg[j+1], msg[j+2], msg[j+3])
			if msg[j+3] == 1 {
				encM = encM.Neg(encM)
			}
			exp := big.NewInt(1)
			exp = exp.Lsh(exp, 5*c)
			M[i] = M[i].Add(M[i], encM.Mul(encM, exp))
			c += 1
		}

		if M[i].Sign() < 0 {
			M[i] = M[i].Add(M[i], babyjub.SubOrder)
		}

		basePoint := generateBasePoint(i)
		H = eccAdd(H, basePoint.Mul(M[i], basePoint))
	}
	return H
}

func enc(b0, b1, b2, b3 int64) *big.Int {
	ret := big.NewInt((2*b3 - 1) * (1 + b0 + 2*b1 + 4*b2))
	return ret.Abs(ret)
}

func bits(bs []byte) []int64 {
	r := make([]int64, len(bs)*8)
	for i, b := range bs {
		for j := 0; j < 8; j++ {
			r[i*8+j] = int64(b >> uint(j) & 0x01)
		}
	}
	return r
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
