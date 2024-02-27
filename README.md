# PHTaxCalculator

A Philippine tax calculator written in Go for the course CSADPRG. Given a total monthly salary in Philippine Peso, the CLI program calculates fees in terms of monthly contributions to SSS, PhilHealth, and Pag-IBIG, as well as income tax, net pay after tax, and net pay after all deductions. The program currently uses the Philippine tax table as of 2022 for computation.

The program utilizes the [`decimal`](https://github.com/shopspring/decimal) package for representing monetary values instead of the native Go `float64` data type to address floating-type implementation limitations and preserve computation accuracy.
