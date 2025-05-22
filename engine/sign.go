package engine

import (
	"math/big"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/minimax"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/tuneinsight/lattigo/v6/utils"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

// Evaluator is an evaluator providing an API for homomorphic comparisons.
// All fields of this struct are public, enabling custom instantiations.
type Evaluator struct {
	Parameters ckks.Parameters
	*minimax.Evaluator
	MinimaxCompositeSignPolynomial minimax.Polynomial
}

// NewEvaluator instantiates a new ComparisonEvaluator.
// The default ckks.Evaluator is compliant with the EvaluatorForMinimaxCompositePolynomial interface.
//
// Giving a MinimaxCompositePolynomial is optional, but it is highly recommended to provide one that is optimized
// for the circuit requiring the comparisons as this polynomial will define the internal precision of all computations
// performed by this evaluator.
//
// The MinimaxCompositePolynomial must be a composite minimax approximation of the sign function:
// f(x) = 1 if x > 0, -1 if x < 0, else 0, in the interval [-1, 1].
// Such composite polynomial can be obtained with the function GenMinimaxCompositePolynomialForSign.
//
// If no MinimaxCompositePolynomial is given, then it will use by default the variable DefaultMinimaxCompositePolynomialForSign.
// See the doc of DefaultMinimaxCompositePolynomialForSign for additional information about the performance of this approximation.
//
// This method is allocation free if a MinimaxCompositePolynomial is given.
func NewEvaluator(params ckks.Parameters, eval *minimax.Evaluator, signPoly ...minimax.Polynomial) *Evaluator {
	if len(signPoly) == 1 {
		return &Evaluator{
			Parameters:                     params,
			Evaluator:                      eval,
			MinimaxCompositeSignPolynomial: signPoly[0],
		}
	} else {
		return &Evaluator{
			Parameters:                     params,
			Evaluator:                      eval,
			MinimaxCompositeSignPolynomial: minimax.NewPolynomial(DefaultPolynomialForSign),
		}
	}
}

// DefaultCompositePolynomialForSign is an example of composite minimax polynomial
// for the sign function that is able to distinguish between value with a delta of up to
// 2^{-alpha=30}, tolerates a scheme error of 2^{-35} and outputs a binary value (-1, or 1)
// of up to 20x4 bits of precision.
//
// It was computed with GenMinimaxCompositePolynomialForSign(256, 30, 35, []int{15, 15, 15, 17, 31, 31, 31, 31})
// which outputs a minimax composite polynomial of precision 21.926741, which is further composed with
// CoeffsSignX4Cheby to bring it to ~80bits of precision.

// Sign evaluates f(x) = 1 if x > 0, -1 if x < 0, else 0.
// This will ensure that sign.Scale = params.DefaultScale().
func (eval Evaluator) Sign(op0 *rlwe.Ciphertext) (sign *rlwe.Ciphertext, err error) {
	return eval.Evaluate(op0, eval.MinimaxCompositeSignPolynomial)
}

// Step evaluates f(x) = 1 if x > 0, 0 if x < 0, else 0.5 (i.e. (sign+1)/2).
// This will ensure that step.Scale = params.DefaultScale().
func (eval Evaluator) Step(op0 *rlwe.Ciphertext) (step *rlwe.Ciphertext, err error) {

	n := len(eval.MinimaxCompositeSignPolynomial)

	stepPoly := make([]bignum.Polynomial, n)

	for i := 0; i < n; i++ {
		stepPoly[i] = eval.MinimaxCompositeSignPolynomial[i]
	}

	half := new(big.Float).SetFloat64(0.5)

	// (x+1)/2
	lastPoly := eval.MinimaxCompositeSignPolynomial[n-1].Clone()
	for i := range lastPoly.Coeffs {
		lastPoly.Coeffs[i][0].Mul(lastPoly.Coeffs[i][0], half)
	}
	lastPoly.Coeffs[0][0].Add(lastPoly.Coeffs[0][0], half)

	stepPoly[n-1] = lastPoly

	return eval.Evaluate(op0, stepPoly)
}

// Max returns the smooth maximum of op0 and op1, which is defined as: op0 * x + op1 * (1-x) where x = step(diff = op0-op1).
// Use must ensure that:
//   - op0 + op1 is in the interval [-1, 1].
//   - op0.Scale = op1.Scale.
//
// This method ensures that max.Scale = params.DefaultScale.
func (eval Evaluator) Max(op0, op1 *rlwe.Ciphertext) (max *rlwe.Ciphertext, err error) {

	// step * diff
	var stepdiff *rlwe.Ciphertext
	if stepdiff, err = eval.stepdiff(op0, op1); err != nil {
		return
	}

	// max = step * diff + op1
	if err = eval.Add(stepdiff, op1, stepdiff); err != nil {
		return
	}

	return stepdiff, nil
}

// Min returns the smooth min of op0 and op1, which is defined as: op0 * (1-x) + op1 * x where x = step(diff = op0-op1)
// Use must ensure that:
//   - op0 + op1 is in the interval [-1, 1].
//   - op0.Scale = op1.Scale.
//
// This method ensures that min.Scale = params.DefaultScale.
func (eval Evaluator) Min(op0, op1 *rlwe.Ciphertext) (min *rlwe.Ciphertext, err error) {

	// step * diff
	var stepdiff *rlwe.Ciphertext
	if stepdiff, err = eval.stepdiff(op0, op1); err != nil {
		return
	}

	// min = op0 - step * diff
	if err = eval.Sub(op0, stepdiff, stepdiff); err != nil {
		return
	}

	return stepdiff, nil
}

func (eval Evaluator) stepdiff(op0, op1 *rlwe.Ciphertext) (stepdiff *rlwe.Ciphertext, err error) {
	params := eval.Parameters

	// diff = op0 - op1
	var diff *rlwe.Ciphertext
	if diff, err = eval.SubNew(op0, op1); err != nil {
		return
	}

	// Required for the scale matching before the last multiplication.
	if diff.Level() < params.LevelsConsumedPerRescaling()*2 {
		if diff, err = eval.BtsEval.Bootstrap(diff); err != nil {
			return
		}
	}

	// step = 1 if diff > 0, 0 if diff < 0 else 0.5
	var step *rlwe.Ciphertext
	if step, err = eval.Step(diff); err != nil {
		return
	}

	// Required for the following multiplication
	if step.Level() < params.LevelsConsumedPerRescaling() {
		if step, err = eval.BtsEval.Bootstrap(step); err != nil {
			return
		}
	}

	// Extremum gate: op0 * step + op1 * (1 - step) = step * diff + op1
	level := utils.Min(diff.Level(), step.Level())

	ratio := rlwe.NewScale(1)
	for i := 0; i < params.LevelsConsumedPerRescaling(); i++ {
		ratio = ratio.Mul(rlwe.NewScale(params.Q()[level-i]))
	}

	ratio = ratio.Div(diff.Scale)
	if err = eval.Mul(diff, &ratio.Value, diff); err != nil {
		return
	}

	if err = eval.Rescale(diff, diff); err != nil {
		return
	}
	diff.Scale = diff.Scale.Mul(ratio)

	// max = step * diff
	if err = eval.MulRelin(diff, step, diff); err != nil {
		return
	}

	if err = eval.Rescale(diff, diff); err != nil {
		return
	}

	return diff, nil
}

// GetChebyshevPoly returns the Chebyshev polynomial approximation of f the
// in the interval [-K, K] for the given degree.
func GetChebyshevPoly(K float64, degree int, f64 func(x float64) (y float64)) bignum.Polynomial {

	FBig := func(x *big.Float) (y *big.Float) {
		xF64, _ := x.Float64()
		return new(big.Float).SetPrec(x.Prec()).SetFloat64(f64(xF64))
	}

	var prec uint = 128

	interval := bignum.Interval{
		A:     *bignum.NewFloat(-K, prec),
		B:     *bignum.NewFloat(K, prec),
		Nodes: degree,
	}
	// Returns the polynomial.
	return bignum.ChebyshevApproximation(FBig, interval)
}

// // 29 = 3 + 4 + 4 + 4 + 3 + 5 + 3 + 3
// var DefaultPolynomialForSign = [][]string{{"0", "0.6390324059720205", "0", "-0.2198072442832513", "0", "0.1414406903374961", "0", "-0.5606601964716950"},
// 	{"0", "0.6371916849818620", "0", "-0.2138182920938642", "0", "0.1300528543273972", "0", "-0.0948905401988566", "0", "0.0760465373114826", "0", "-0.0647752466148176", "0", "0.0577934985612991", "0", "-0.5275291175103518"},
// 	{"0", "0.6377206727878986", "0", "-0.2139936398068216", "0", "0.1301568728941797", "0", "-0.0949635425470285", "0", "0.0761019418378052", "0", "-0.0648191338716430", "0", "0.0578291285745938", "0", "-0.5271294130801401"},
// 	{"0", "0.6443950923061463", "0", "-0.2162055771803981", "0", "0.1314684297456403", "0", "-0.0958833600353299", "0", "0.0767993066598502", "0", "-0.0653707358190023", "0", "0.0582760544840787", "0", "-0.5220846192664677"},
// 	{"0", "0.6811845934958654", "0", "-0.2334032891661299", "0", "0.1490228759436919", "0", "-0.5303235460737469"},
// 	{"0", "1.2724046571429532", "0", "-0.4219130535252118", "0", "0.2504961101202538", "0", "-0.1761103958158040", "0", "0.1340921891162222", "0", "-0.1068120717134459", "0", "0.0874944606348786", "0", "-0.0729814769275110", "0", "0.0616046472097950", "0", "-0.0524007824752258", "0", "0.0447759297248556", "0", "-0.0383446741769455", "0", "0.0328466191095667", "0", "-0.0281000581726133", "0", "0.0239749328696791", "0", "-0.0724805452224001"},
// 	minimax.CoeffsSignX4Cheby,
// 	minimax.CoeffsSignX4Cheby,
// }

// 32 = 3 + 3 + 4 + 4 + 4 + 4 + 5 + 5
var DefaultPolynomialForSign = [][]string{
	{"0", "0.639028938435711219", "0", "-0.219806118878047092", "0", "0.141440053455946451", "0", "-0.560662696277302517"},
	{"0", "0.639029489633228222", "0", "-0.219806297800259927", "0", "0.141440154749489224", "0", "-0.560662301247238699"},
	{"0", "0.637154696201570660", "0", "-0.213806030962452086", "0", "0.130045580652836942", "0", "-0.094885435152665168", "0", "0.076042662624281770", "0", "-0.064772177102166506", "0", "0.057791006255316676", "0", "-0.527557080198263608"},
	{"0", "0.637252723778154157", "0", "-0.213838525423059134", "0", "0.130064857400618902", "0", "-0.094898964641665779", "0", "0.076052931457460596", "0", "-0.064780312107983876", "0", "0.057797611601832795", "0", "-0.527483011957071916"},
	{"0", "0.638492548163143535", "0", "-0.214249489332586510", "0", "0.130308633915153971", "0", "-0.095070037868134732", "0", "0.076182750613643440", "0", "-0.064883127935866152", "0", "0.057881063854726339", "0", "-0.526546163851403631"},
	{"0", "0.654072536322215222", "0", "-0.219411111586902031", "0", "0.133367149639755852", "0", "-0.097212758106646793", "0", "0.077804778872091079", "0", "-0.066163376778950260", "0", "0.058915286191840665", "0", "-0.514764737657234364"},
	{"0", "0.985321201923117642", "0", "-0.328119170883870823", "0", "0.196617511234660686", "0", "-0.140045613570337270", "0", "0.108634034466585055", "0", "-0.088498752002149384", "0", "0.074572306762014301", "0", "-0.064287415234822877", "0", "0.056401814845377282", "0", "-0.050183389367761095", "0", "0.045086211659977983", "0", "-0.040959900956987588", "0", "0.037395730021827179", "0", "-0.034512330854589498", "0", "0.031915520720098515", "0", "-0.241509663775553008"},
	{"0", "1.262673861720083096", "0", "-0.393697035515367173", "0", "0.206535085221700125", "0", "-0.120395410087615135", "0", "0.071169308776617400", "0", "-0.041082546981498569", "0", "0.022667637149739957", "0", "-0.011770429894589077", "0", "0.005672674575918509", "0", "-0.002500516475970389", "0", "0.000990582083790523", "0", "-0.000344556747529926", "0", "0.000101727287456808", "0", "-0.000024132899916296", "0", "0.000004146320405856", "0", "-0.000000395353444811"},
}
