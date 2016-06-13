package main

import (
	"math/big"
)

type (
	Number interface {
		Object
		Int() Int
		Double() Double
		BigInt() *big.Int
		BigFloat() *big.Float
		Ratio() *big.Rat
	}
	Ops interface {
		Combine(ops Ops) Ops
		Add(Number, Number) Number
		Subtract(Number, Number) Number
		Multiply(Number, Number) Number
		Divide(Number, Number) Number
		IsZero(Number) bool
		Lt(Number, Number) bool
		Lte(Number, Number) bool
		Gt(Number, Number) bool
	}
	IntOps      struct{}
	DoubleOps   struct{}
	BigIntOps   struct{}
	BigFloatOps struct{}
	RatioOps    struct{}
)

var (
	INT_OPS      = IntOps{}
	DOUBLE_OPS   = DoubleOps{}
	BIGINT_OPS   = BigIntOps{}
	BIGFLOAT_OPS = BigFloatOps{}
	RATIO_OPS    = RatioOps{}
)

func (ops IntOps) Combine(other Ops) Ops {
	return other
}

func (ops DoubleOps) Combine(other Ops) Ops {
	switch other.(type) {
	case BigFloatOps:
		return other
	default:
		return ops
	}
}

func (ops BigIntOps) Combine(other Ops) Ops {
	switch other.(type) {
	case IntOps:
		return ops
	default:
		return other
	}
}

func (ops BigFloatOps) Combine(other Ops) Ops {
	return ops
}

func (ops RatioOps) Combine(other Ops) Ops {
	switch other.(type) {
	case DoubleOps, BigFloatOps:
		return other
	default:
		return ops
	}
}

func GetOps(obj Object) Ops {
	switch obj.(type) {
	case Double:
		return DOUBLE_OPS
	case *BigInt:
		return BIGINT_OPS
	case *BigFloat:
		return BIGFLOAT_OPS
	case *Ratio:
		return RATIO_OPS
	default:
		return INT_OPS
	}
}

// Int conversions

func (i Int) Int() Int {
	return i
}

func (i Int) Double() Double {
	return Double{d: float64(i.i)}
}

func (i Int) BigInt() *big.Int {
	return big.NewInt(int64(i.i))
}

func (i Int) BigFloat() *big.Float {
	return big.NewFloat(float64(i.i))
}

func (i Int) Ratio() *big.Rat {
	return big.NewRat(int64(i.i), 1)
}

// Double conversions

func (d Double) Int() Int {
	return Int{i: int(d.d)}
}

func (d Double) BigInt() *big.Int {
	return big.NewInt(int64(d.d))
}

func (d Double) Double() Double {
	return d
}

func (d Double) BigFloat() *big.Float {
	return big.NewFloat(float64(d.d))
}

func (d Double) Ratio() *big.Rat {
	res := big.Rat{}
	return res.SetFloat64(float64(d.d))
}

// BigInt conversions

func (b *BigInt) Int() Int {
	return Int{i: int(b.BigInt().Int64())}
}

func (b *BigInt) BigInt() *big.Int {
	return &b.b
}

func (b *BigInt) Double() Double {
	return Double{d: float64(b.BigInt().Int64())}
}

func (b *BigInt) BigFloat() *big.Float {
	res := big.Float{}
	return res.SetInt(b.BigInt())
}

func (b *BigInt) Ratio() *big.Rat {
	res := big.Rat{}
	return res.SetInt(b.BigInt())
}

// BigFloat conversions

func (b *BigFloat) Int() Int {
	i, _ := b.BigFloat().Int64()
	return Int{i: int(i)}
}

func (b *BigFloat) BigInt() *big.Int {
	bi, _ := b.BigFloat().Int(nil)
	return bi
}

func (b *BigFloat) Double() Double {
	f, _ := b.BigFloat().Float64()
	return Double{d: f}
}

func (b *BigFloat) BigFloat() *big.Float {
	return &b.b
}

func (b *BigFloat) Ratio() *big.Rat {
	res := big.Rat{}
	return res.SetFloat64(float64(b.Double().d))
}

// Ratio conversions

func (r *Ratio) Int() Int {
	f, _ := r.Ratio().Float64()
	return Int{i: int(f)}
}

func (r *Ratio) BigInt() *big.Int {
	f, _ := r.Ratio().Float64()
	return big.NewInt(int64(f))
}

func (r *Ratio) Double() Double {
	f, _ := r.Ratio().Float64()
	return Double{d: f}
}

func (r *Ratio) BigFloat() *big.Float {
	f, _ := r.Ratio().Float64()
	return big.NewFloat(f)
}

func (r *Ratio) Ratio() *big.Rat {
	return &r.r
}

// Ops

// Add

func (ops IntOps) Add(x, y Number) Number {
	return Int{i: x.Int().i + y.Int().i}
}

func (ops DoubleOps) Add(x, y Number) Number {
	return Double{d: x.Double().d + y.Double().d}
}

func (ops BigIntOps) Add(x, y Number) Number {
	b := big.Int{}
	b.Add(x.BigInt(), y.BigInt())
	res := BigInt{b: b}
	return &res
}

func (ops BigFloatOps) Add(x, y Number) Number {
	b := big.Float{}
	b.Add(x.BigFloat(), y.BigFloat())
	res := BigFloat{b: b}
	return &res
}

func (ops RatioOps) Add(x, y Number) Number {
	r := big.Rat{}
	r.Add(x.Ratio(), y.Ratio())
	res := Ratio{r: r}
	return &res
}

// Subtract

func (ops IntOps) Subtract(x, y Number) Number {
	return Int{i: x.Int().i - y.Int().i}
}

func (ops DoubleOps) Subtract(x, y Number) Number {
	return Double{d: x.Double().d - y.Double().d}
}

func (ops BigIntOps) Subtract(x, y Number) Number {
	b := big.Int{}
	b.Sub(x.BigInt(), y.BigInt())
	res := BigInt{b: b}
	return &res
}

func (ops BigFloatOps) Subtract(x, y Number) Number {
	b := big.Float{}
	b.Sub(x.BigFloat(), y.BigFloat())
	res := BigFloat{b: b}
	return &res
}

func (ops RatioOps) Subtract(x, y Number) Number {
	r := big.Rat{}
	r.Sub(x.Ratio(), y.Ratio())
	res := Ratio{r: r}
	return &res
}

// Multiply

func (ops IntOps) Multiply(x, y Number) Number {
	return Int{i: x.Int().i * y.Int().i}
}

func (ops DoubleOps) Multiply(x, y Number) Number {
	return Double{d: x.Double().d * y.Double().d}
}

func (ops BigIntOps) Multiply(x, y Number) Number {
	b := big.Int{}
	b.Mul(x.BigInt(), y.BigInt())
	res := BigInt{b: b}
	return &res
}

func (ops BigFloatOps) Multiply(x, y Number) Number {
	b := big.Float{}
	b.Mul(x.BigFloat(), y.BigFloat())
	res := BigFloat{b: b}
	return &res
}

func (ops RatioOps) Multiply(x, y Number) Number {
	r := big.Rat{}
	r.Mul(x.Ratio(), y.Ratio())
	res := Ratio{r: r}
	return &res
}

// Divide

func (ops IntOps) Divide(x, y Number) Number {
	b := big.NewRat(int64(x.Int().i), int64(y.Int().i))
	if b.IsInt() {
		return Int{i: int(b.Num().Int64())}
	}
	res := Ratio{r: *b}
	return &res
}

func (ops DoubleOps) Divide(x, y Number) Number {
	return Double{d: x.Double().d / y.Double().d}
}

func (ops BigIntOps) Divide(x, y Number) Number {
	b := big.Rat{}
	b.Quo(x.Ratio(), y.Ratio())
	if b.IsInt() {
		res := BigInt{b: *b.Num()}
		return &res
	}
	res := Ratio{r: b}
	return &res
}

func (ops BigFloatOps) Divide(x, y Number) Number {
	b := big.Float{}
	b.Quo(x.BigFloat(), y.BigFloat())
	res := BigFloat{b: b}
	return &res
}

func (ops RatioOps) Divide(x, y Number) Number {
	r := big.Rat{}
	r.Quo(x.Ratio(), y.Ratio())
	res := Ratio{r: r}
	return &res
}

// IsZero

func (ops IntOps) IsZero(x Number) bool {
	return x.Int().i == 0
}

func (ops DoubleOps) IsZero(x Number) bool {
	return x.Double().d == 0
}

func (ops BigIntOps) IsZero(x Number) bool {
	return x.BigInt().Sign() == 0
}

func (ops BigFloatOps) IsZero(x Number) bool {
	return x.BigFloat().Sign() == 0
}

func (ops RatioOps) IsZero(x Number) bool {
	return x.Ratio().Sign() == 0
}

// Lt

func (ops IntOps) Lt(x Number, y Number) bool {
	return x.Int().i < y.Int().i
}

func (ops DoubleOps) Lt(x Number, y Number) bool {
	return x.Double().d < y.Double().d
}

func (ops BigIntOps) Lt(x Number, y Number) bool {
	return x.BigInt().Cmp(y.BigInt()) < 0
}

func (ops BigFloatOps) Lt(x Number, y Number) bool {
	return x.BigFloat().Cmp(y.BigFloat()) < 0
}

func (ops RatioOps) Lt(x Number, y Number) bool {
	return x.Ratio().Cmp(y.Ratio()) < 0
}

// Lte

func (ops IntOps) Lte(x Number, y Number) bool {
	return x.Int().i <= y.Int().i
}

func (ops DoubleOps) Lte(x Number, y Number) bool {
	return x.Double().d <= y.Double().d
}

func (ops BigIntOps) Lte(x Number, y Number) bool {
	return x.BigInt().Cmp(y.BigInt()) <= 0
}

func (ops BigFloatOps) Lte(x Number, y Number) bool {
	return x.BigFloat().Cmp(y.BigFloat()) <= 0
}

func (ops RatioOps) Lte(x Number, y Number) bool {
	return x.Ratio().Cmp(y.Ratio()) <= 0
}

// Gt

func (ops IntOps) Gt(x Number, y Number) bool {
	return x.Int().i > y.Int().i
}

func (ops DoubleOps) Gt(x Number, y Number) bool {
	return x.Double().d > y.Double().d
}

func (ops BigIntOps) Gt(x Number, y Number) bool {
	return x.BigInt().Cmp(y.BigInt()) > 0
}

func (ops BigFloatOps) Gt(x Number, y Number) bool {
	return x.BigFloat().Cmp(y.BigFloat()) > 0
}

func (ops RatioOps) Gt(x Number, y Number) bool {
	return x.Ratio().Cmp(y.Ratio()) > 0
}

func CompareNumbers(x Number, y Number) int {
	ops := GetOps(x).Combine(GetOps(y))
	if ops.Lt(x, y) {
		return -1
	}
	if ops.Lt(y, x) {
		return 1
	}
	return 0
}
