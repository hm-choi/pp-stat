# Experimental Evaluation

There are four example codes and two datasets for the performance evaluation of PP-STAT.

## Usage of Experiments
To run the n-th experiment, 
(1) Change the directory into *experiment{n}*  \
(2) Run the following command:
```bash
    go run main.go
```

## Dataset
We use two real-world datasets: the UCI Adult Income dataset [1] and the Medical Cost dataset [2].
These datasets are provided in the following CSV files:
```
    examples\dataset\adult_dataset.csv
    examples\dataset\insurance.csv
```


## Experiment 1: Performance Comparison of Inverse Square Root Computation
We compare the performance of our inverse square root computation against two prior methods: Panda et al. [3] and HEaaN-STAT [4]. #BTS denotes
the number of bootstrapping operations, and Runtime (s) is measured using single-core execution. All values are averaged over 10 runs, with standard deviations shown in parentheses. 

### Table. Performance comparison of inverse square root computation over the input domain [0.001, 100]. The parameter `B` denotes the constant scaling factor. All values are averaged over ten runs; values in parentheses represent standard deviations.

| Method           | B       | #BTS | MRE                             | Runtime (s)       |
|------------------|---------|------|----------------------------------|-------------------|
| hstat            | -       | 5    | 5.36 × 10⁻³ (1.05 × 10⁻⁷)       | 225.02 (3.18)     |
| panda            | 100.0   | 6    | 3.73 × 10⁻⁴ (1.68 × 10⁻⁹)       | 273.48 (5.29)     |
| **invSqrt**      | 100.0   | 2    | 5.08 × 10⁻⁵ (1.75 × 10⁻⁹)       | 94.73 (1.05)      |

Experiment 2: Accuracy on Large-Scale Dataset
### Table. Accuracy and efficiency of Z-score normalization and other statistical measures. Z-score normalization evaluated over the domain [0, 100]; other measures over [0, 20]. All values are averaged over 10 trials; values in parentheses denote standard deviations. `B` is the constant scaling factor.

| Measure | B   | #BTS | MRE                            | Runtime (s)        |
|---------|-----|------|--------------------------------|--------------------|
| ZNorm   | 100 | 2    | 4.18 × 10⁻⁵ (6.06 × 10⁻⁶)      | 141.29 (1.55)      |
| Skew    | 20  | 2    | 8.12 × 10⁻³ (1.41 × 10⁻²)      | 154.10 (1.61)      |
| Kurt    | 20  | 2    | 3.73 × 10⁻⁴ (6.97 × 10⁻⁶)      | 154.70 (1.34)      |
| CV      | 20  | 7    | 1.25 × 10⁻⁴ (9.70 × 10⁻⁵)      | 311.08 (2.94)      |
| PCC     | 20  | 4    | 2.62 × 10⁻⁴ (3.21 × 10⁻⁵)      | 289.86 (3.47)      |

*Abbreviations: ZNorm = Z-score normalization, Skew = skewness, Kurt = kurtosis, CV = coefficient of variation, PCC = Pearson correlation coefficient.*

## Experiment 3: Evaluation on Real-world Datasets
### Expeirment 3.1 
### Table. Evaluation of statistical operations over the *Adult* dataset. The MRE is reported for Z-score normalization (ZNorm), kurtosis (Kurt), skewness (Skew), and coefficient of variation (CV) across selected features. PCC denotes the Pearson correlation coefficient between feature pairs. All values are averaged over ten trials; values in parentheses indicate standard deviations. `B` denotes the constant scaling factor.

| Operation | Feature(s)     | B   | Output   | MRE                            | Runtime (s)       |
|-----------|----------------|-----|----------|--------------------------------|-------------------|
| ZNorm     | AGE            | 50  | -        | 2.47 × 10⁻⁵ (3.39 × 10⁻²¹)     | 110.18 (2.42)     |
|           | EDU            | 50  | -        | 1.02 × 10⁻⁴ (1.36 × 10⁻²⁰)     | 110.03 (1.40)     |
|           | HPW            | 50  | -        | 7.62 × 10⁻⁵ (1.36 × 10⁻²⁰)     | 109.73 (1.71)     |
| Skew      | AGE            | 50  | 0.5576   | 7.92 × 10⁻⁵ (1.36 × 10⁻²⁰)     | 113.92 (1.83)     |
|           | EDU            | 50  | -0.3165  | 2.50 × 10⁻⁴ (0.00)             | 112.32 (2.02)     |
|           | HPW            | 50  | 0.2387   | 1.38 × 10⁻⁴ (0.00)             | 111.81 (1.68)     |
| Kurt      | AGE            | 50  | -0.1844  | 4.81 × 10⁻³ (0.00)             | 113.20 (2.07)     |
|           | EDU            | 50  | 0.6256   | 2.14 × 10⁻³ (0.00)             | 113.02 (1.58)     |
|           | HPW            | 50  | 2.9506   | 7.76 × 10⁻⁴ (0.00)             | 112.36 (1.52)     |
| CV        | AGE            | 50  | 0.3548   | 4.39 × 10⁻⁵ (6.78 × 10⁻²¹)     | 297.60 (3.84)     |
|           | EDU            | 50  | 0.2551   | 1.10 × 10⁻³ (2.17 × 10⁻¹⁹)     | 295.72 (3.92)     |
|           | HPW            | 50  | 0.3065   | 4.82 × 10⁻⁵ (6.78 × 10⁻²¹)     | 295.68 (3.08)     |
| PCC       | AGE vs HPW     | 50  | 0.0716   | 1.40 × 10⁻⁴ (2.71 × 10⁻²⁰)     | 222.39 (3.05)     |
|           | AGE vs EDU     | 50  | 0.0309   | 1.01 × 10⁻⁴ (0.00)             | 222.38 (3.35)     |

*Abbreviations: AGE = age, EDU = education-num, HPW = hours-per-week*

### Experiment 3.2
### Table. Evaluation of statistical metrics using \sysname over the *insurance* dataset. The Mean relative error (MRE) is reported for each statistical function compared to the plaintext result. PCC denotes the Pearson correlation coefficient between the target feature (`charges`) and selected predictors. Kurtosis is reported as excess kurtosis (i.e., normal kurtosis minus 3). All values are averaged over 10 trials; values in parentheses indicate standard deviations. `B` denotes the constant scaling factor.

| Operation | Feature(s)         | B   | Output   | MRE                            | Runtime (s)       |
|-----------|--------------------|-----|----------|--------------------------------|-------------------|
| ZNorm     | charges            | 100 | -        | 3.81 × 10⁻⁵ (0.00)             | 108.26 (3.34)     |
| Skew      | charges            | 20  | 1.5143   | 8.67 × 10⁻⁵ (1.36 × 10⁻²⁰)     | 105.52 (2.36)     |
| Kurt      | charges            | 20  | 1.5966   | 6.08 × 10⁻⁴ (1.08 × 10⁻¹⁹)     | 104.94 (2.45)     |
| CV        | charges            | 20  | 0.9123   | 5.54 × 10⁻⁵ (6.78 × 10⁻²¹)     | 279.02 (5.11)     |
| PCC       | AGE vs charges     | 20  | 0.2990   | 2.22 × 10⁻⁴ (0.00)             | 207.87 (4.73)     |
|           | BMI vs charges     | 20  | 0.1983   | 7.33 × 10⁻⁵ (0.00)             | 207.85 (3.64)     |
|           | SMOKER vs charges  | 20  | 0.7873   | 2.86 × 10⁻⁴ (5.42 × 10⁻²⁰)     | 209.98 (2.07)     |

[1] Barry Becker and Ronny Kohavi. 1996. Adult. UCI Machine Learning Repository.
DOI: https://doi.org/10.24432/C5XW20. \
[2] Nahida Akter and Ashadun Nobi. 2018. Investigation of the financial stability of S&P 500 using realized volatility and stock returns distribution. Journal of Risk and Financial Management 11, 2 (2018), 22. \
[3] Samanvaya Panda. 2022. Polynomial approximation of inverse sqrt function for fhe. In International Symposium on Cyber Security, Cryptology, and Machine Learning. Springer, 366–376. \
[4] Younho Lee, Jinyeong Seo, Yujin Name, Jiseok Chae, and Jung Hee Cheon. 2023. HEaaN-STAT: a privacy-preserving statistical analysis toolkit for large-scale numerical, ordinal, and categorical data. IEEE Transactions on Dependable and Secure Computing (2023).
