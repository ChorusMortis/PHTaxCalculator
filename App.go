// ********************
// Names:
//    Marcellana, Patrick James
//    Mider, Brett Harley
//    Pinpin, Lord John Benedict
// Language: Go
// Paradigm: Multi-paradigm (imperative/procedural, object-oriented, concurrent)
// ********************

// The program requires the decimal module by @shopspring. It is required
// for representing money values that float64 cannot accurately represent
// due to implementation limitations.
// Documentation: https://pkg.go.dev/github.com/shopspring/decimal
// Repository: https://github.com/shopspring/decimal

// The program is tailored to the rates set and values used in 2022.

package main

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func main() {
	var input string

	fmt.Print("Enter total monthly salary: PHP ")
	fmt.Scanln(&input)
	fmt.Println()
	monthlySalary, err := decimal.NewFromString(input)
	if err != nil {
		fmt.Println(err)
		return
	}
	compute(monthlySalary)
}

func compute(monthlySalary decimal.Decimal) {
	// Step 1: Get taxable income.
	SSSContri := getSSSContri(monthlySalary)
	philHealthContri := getPhilHealthContri(monthlySalary)
	pagIBIGContri := getPagIBIGContri(monthlySalary)
	totalContri := SSSContri.Add(philHealthContri).Add(pagIBIGContri).Round(2)
	taxableIncome := monthlySalary.Sub(totalContri)

	// Step 2/3: Determine salary column and compute income tax and net pay
	// after tax.
	fixedTax, rate, compensationLevel := getWithholdingTax(taxableIncome)
	rateOverCL := taxableIncome.Sub(compensationLevel).Mul(rate)
	incomeTax := fixedTax.Add(rateOverCL).Round(2)
	netPayAfterTax := monthlySalary.Sub(incomeTax).Round(2)

	// Step 4: Compute net take home pay
	totalDeductions := totalContri.Add(incomeTax).Round(2)
	netSalary := monthlySalary.Sub(totalDeductions).Round(2)

	// Step 5: Print all necessary data
	fmt.Println("Monthly Contributions")
	fmt.Printf("%-26s PHP %s\n", "SSS", SSSContri.StringFixed(2))
	fmt.Printf("%-26s PHP %s\n", "PhilHealth", philHealthContri.StringFixed(2))
	fmt.Printf("%-26s PHP %s\n", "Pag-IBIG", pagIBIGContri.StringFixed(2))
	fmt.Printf("%-26s PHP %s\n", "Total Contributions", totalContri.StringFixed(2))
	fmt.Println()

	fmt.Println("Tax Computation")
	fmt.Printf("%-26s PHP %s\n", "Income Tax", incomeTax.StringFixed(2))
	fmt.Printf("%-26s PHP %s\n", "Net Pay After Tax", netPayAfterTax.StringFixed(2))
	fmt.Println()

	fmt.Printf("%-26s PHP %s\n", "Total Deductions", totalDeductions.StringFixed(2))
	fmt.Printf("%-26s PHP %s\n", "Net Pay After Deductions", netSalary.StringFixed(2))
}

// Computes the monthly social credit given the monthly salary. It is to be
// used in getting the SSS contributions. It has a limit of PHP 25,000.00.
// See https://moneygment.ph/wp-content/uploads/2020/12/SSS-CIR-New-Contribution-Schedule_ER-and-EE_22Dec20_521pm-1.pdf
func getSSSMSC(monthlySalary decimal.Decimal) decimal.Decimal {
	// Formula: MSC = (salary + 250) // 500 * 500 where // is integer division
	result := monthlySalary.Add(decimal.RequireFromString("250"))
	result = result.Div(decimal.RequireFromString("500")).Floor()
	result = result.Mul(decimal.RequireFromString("500"))
	return decimal.Min(decimal.RequireFromString("25000.0"), result)
}

// The monthly SSS contributions can simply be obtained by multiplying the
// monthly salary credit (MSC) by 4.5%. If the MSC is less than PHP 3,250.00,
// however, the fixed amount is PHP 135.00.
// See https://moneygment.ph/wp-content/uploads/2020/12/SSS-CIR-New-Contribution-Schedule_ER-and-EE_22Dec20_521pm-1.pdf
func getSSSContri(monthlySalary decimal.Decimal) decimal.Decimal {
	msc := getSSSMSC(monthlySalary)
	if msc.LessThan(decimal.RequireFromString("3250.0")) {
		return decimal.RequireFromString("135.0")
	}
	return msc.Mul(decimal.RequireFromString(".045"))
}

// See https://www.philhealth.gov.ph/news/2019/new_contri.php#gsc.tab=0
func getPhilHealthContri(monthlySalary decimal.Decimal) decimal.Decimal {
	var result decimal.Decimal

	switch {
	case monthlySalary.LessThanOrEqual(decimal.RequireFromString("10000.0")):
		result = decimal.RequireFromString("400.0")
	case monthlySalary.GreaterThanOrEqual(decimal.RequireFromString("80000.0")):
		result = decimal.RequireFromString("3200.0")
	default:
		result = monthlySalary.Mul(decimal.RequireFromString(".04"))
	}

	// result is halved because the employee and employer share in paying
	return result.Div(decimal.RequireFromString("2"))
}

// See https://taxcalculatorphilippines.com/pag-ibig-contribution-table
func getPagIBIGContri(monthlySalary decimal.Decimal) decimal.Decimal {
	// PHP 5,000.00 is the upper limit
	monthlySalary = decimal.Min(monthlySalary, decimal.RequireFromString("5000.0"))

	contributionRate := decimal.RequireFromString(".02")
	if monthlySalary.LessThanOrEqual(decimal.RequireFromString("1500.0")) {
		contributionRate = decimal.RequireFromString(".01")
	}

	return monthlySalary.Mul(contributionRate)
}

// See https://www.bir.gov.ph/index.php/tax-information/withholding-tax.html
func getWithholdingTax(taxableIncome decimal.Decimal) (decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	var fixedTax decimal.Decimal
	var rate decimal.Decimal
	var compensationLevel decimal.Decimal

	// Format: upper bound, fixed tax, rate, compensation level
	// the table is ordered by each column's upper bound in increasing order
	// the last column has inf as its upper bound
	monthlyTaxTable := [][]string{
		{"20833.0", "0.0", "0.0", "0.0"},
		{"33332.0", "0.0", ".2", "20833.0"},
		{"66666.0", "2500.0", ".25", "33333.0"},
		{"166666.0", "10833.33", ".3", "66667.0"},
		{"666666.0", "40833.33", ".32", "166667.0"},
		{"inf", "200833.33", ".35", "666667.0"},
	}

	// the selected salary column depends on the first upper bound the taxable
	// income fits into
	for _, vals := range monthlyTaxTable {
		upperBound, err := decimal.NewFromString(vals[0])

		// err != nil if the upper bound is inf since NewFromString() can't
		// convert +/-inf to decimal
		if err != nil || taxableIncome.LessThanOrEqual(upperBound) {
			fixedTax = decimal.RequireFromString(vals[1])
			rate = decimal.RequireFromString(vals[2])
			compensationLevel = decimal.RequireFromString(vals[3])
			break
		}
	}

	return fixedTax, rate, compensationLevel
}
