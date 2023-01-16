package utils

import (
	"crypto/rand"
	"io"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"golang.org/x/crypto/sha3"
)

type GroupParallel struct {
	max, current, step int64
	wg                 sync.WaitGroup
}

func (gp *GroupParallel) Next() (int64, bool) {
	base := atomic.AddInt64(&gp.current, gp.step) - gp.step

	if base >= gp.max {
		return 0, false
	} else {
		return base, true
	}
}

type Parallel struct {
	gp           *GroupParallel
	max, current int64
}

func (p *Parallel) Next() (int, bool) {
	if p.current >= p.max {
		r, ok := p.gp.Next()
		if !ok {
			return 0, false
		}
		p.current, p.max = r, r+p.gp.step
		if p.max > p.gp.max {
			p.max = p.gp.max
		}
	}

	r := p.current
	p.current += 1

	return int(r), true
}

func ParallelFor(n int, f func(p *Parallel)) {
	// TODO: this formula could probably be more clever
	step := n / runtime.NumCPU() / 100
	if step < 10 {
		step = 10
	}

	gp := &GroupParallel{
		max:     int64(n),
		current: 0,
		step:    int64(step),
	}

	gp.wg.Add(runtime.NumCPU())

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			p := &Parallel{
				gp: gp,
			}
			f(p)
			gp.wg.Done()
		}()
	}

	gp.wg.Wait()
}

type Span struct {
	Start int
	Count int
}

func Spans(total, spanSize int) []Span {
	spans := make([]Span, 0, (total+spanSize-1)/spanSize)
	var c int
	for i := 0; i < total; i += c {
		if i+spanSize <= total {
			c = spanSize
			spans = append(spans, Span{Start: i, Count: c})
		} else {
			c = total - i
			if c > spanSize/2 || len(spans) == 0 {
				spans = append(spans, Span{Start: i, Count: c})
			} else {
				spans[len(spans)-1].Count += c
			}
		}
	}
	return spans
}

func StrictSpans(total, spanSize int) []Span {
	spans := make([]Span, 0, (total+spanSize-1)/spanSize)
	var c int
	for i := 0; i < total; i += c {
		if i+spanSize <= total {
			c = spanSize
		} else {
			c = total - i
		}
		spans = append(spans, Span{Start: i, Count: c})
	}
	return spans
}

func HashToPrime(data []byte) *big.Int {
	// Unclear if this is a good hash function.
	h := sha3.NewShake256()
	h.Write(data)
	p, err := rand.Prime(h, 256)
	if err != nil {
		panic(err)
	}
	return p
}

// PrivateKey is the private key for an RSA accumulator.
// It is not needed for typical uses of an accumulator.
type PrivateKey struct {
	P, Q    *big.Int
	N       *big.Int // N = P*Q
	Totient *big.Int // Totient = (P-1)*(Q-1)
}

type PublicKey struct {
	N *big.Int
}

var base = big.NewInt(65537)
var bigOne = big.NewInt(1)
var bigTwo = big.NewInt(2)

// GenerateKey generates an RSA accumulator keypair. The private key
// is mostly used for debugging and should usually be destroyed
// as part of a trusted setup phase.
func GenerateKey(random io.Reader) (*PublicKey, *PrivateKey, error) {
	for {
		p, err := rand.Prime(random, 1024)
		if err != nil {
			return nil, nil, err
		}
		q, err := rand.Prime(random, 1024)
		if err != nil {
			return nil, nil, err
		}

		pminus1 := new(big.Int).Sub(p, bigOne)
		qminus1 := new(big.Int).Sub(q, bigOne)
		totient := new(big.Int).Mul(pminus1, qminus1)

		g := new(big.Int).GCD(nil, nil, base, totient)
		if g.Cmp(bigOne) == 0 {
			privateKey := &PrivateKey{
				P:       p,
				Q:       q,
				N:       new(big.Int).Mul(p, q),
				Totient: totient,
			}
			publicKey := &PublicKey{
				N: new(big.Int).Set(privateKey.N),
			}
			return publicKey, privateKey, nil
		}
	}
}

func (key *PrivateKey) Accumulate(items ...[]byte) (acc *big.Int, witnesses []*big.Int) {
	primes := make([]*big.Int, len(items))
	ParallelFor(len(items), func(p *Parallel) {
		for i, ok := p.Next(); ok; i, ok = p.Next() {
			primes[i] = HashToPrime(items[i])
		}
	})

	exp := big.NewInt(1)
	for i := range primes {
		exp.Mul(exp, primes[i])
		exp.Mod(exp, key.Totient)
	}
	acc = new(big.Int).Exp(base, exp, key.N)

	witnesses = make([]*big.Int, len(items))
	ParallelFor(len(items), func(p *Parallel) {
		for i, ok := p.Next(); ok; i, ok = p.Next() {
			inv := new(big.Int).ModInverse(primes[i], key.Totient)
			inv.Mul(exp, inv)
			inv.Mod(inv, key.Totient)
			witnesses[i] = new(big.Int).Exp(base, inv, key.N)
		}
	})

	return
}

func (key *PublicKey) Accumulate(items ...[]byte) (acc *big.Int, witnesses []*big.Int) {
	primes := make([]*big.Int, len(items))
	ParallelFor(len(items), func(p *Parallel) {
		for i, ok := p.Next(); ok; i, ok = p.Next() {
			primes[i] = HashToPrime(items[i])
		}
	})

	acc = new(big.Int).Set(base)
	for i := range primes {
		acc.Exp(acc, primes[i], key.N)
	}

	witnesses = make([]*big.Int, len(items))
	ParallelFor(len(items), func(p *Parallel) {
		for i, ok := p.Next(); ok; i, ok = p.Next() {
			// TODO reuse computations
			wit := new(big.Int).Set(base)
			for j := range primes {
				if j != i {
					wit.Exp(wit, primes[j], key.N)
				}
			}
			witnesses[i] = wit
		}
	})

	return
}

func (key *PublicKey) Verify(acc *big.Int, witness *big.Int, item []byte) bool {
	c := HashToPrime(item)
	v := new(big.Int).Exp(witness, c, key.N)
	return acc.Cmp(v) == 0
}
