# PP-STAT: An Efficient Privacy-Preserving Statistical Analysis Framework using Homomorphic Encryption

PP-STAT is an experimental toolkit for privacy-preserving statistical analysis using Homomorphic Encryption. It includes efficient implementations of:

- Z-Score Normalization
- Skewness
- Kurtosis
- Coefficient of Variation (CV)
- Pearson Correlation Coefficient (PCC)



## Abstract
With the widespread adoption of cloud computing, the need for outsourcing statistical analysis to third-party platforms is growing rapidly. However, handling sensitive data such as medical records and financial information in cloud environments raises serious privacy concerns. In this paper, we present PP-STAT, a novel and efficient Homomorphic Encryption (HE)-based framework for privacy-preserving statistical analysis. HE enables computations to be performed directly on encrypted data without revealing the underlying plaintext. PP-STAT supports advanced statistical measures, including Z-score normalization, skewness, kurtosis, coefficient of variation, and Pearson correlation coefficient, all computed securely over encrypted data. To improve efficiency, PP-STAT introduces
two key optimizations: (1) a Chebyshev-based approximation strategy for initializing inverse square root operations, and (2) a prenormalization scaling technique that reduces multiplicative depth by folding constant scaling factors into mean and variance computations. These techniques significantly lower computational overhead and minimize the number of expensive bootstrapping procedures. Our evaluation on real-world datasets demonstrates that PP-STAT
achieves high numerical accuracy, with mean relative error (MRE)
below 2.4 × 10−4. Notably, the encrypted Pearson correlation between the smoker attribute and charges reaches 0.7873, with an MRE of 2.86×10−4. These results confirm the practical utility of PPSTAT for secure and precise statistical analysis in privacy-sensitive domains.

## 1. Server Setting
- In this evaluation, Intel(R) Xeon(R) Gold 6248R CPU @3.00GHz processor, 64GB RAM, and 500GB SSD with Ubuntu 24.04 LTS.
- The experiment will work well even in any Linux environment.

## 2. Requirements
- Go: go1.23 or higher version
- Lattigo V6 library (https://github.com/tuneinsight/lattigo)

## 3. Run the codes
The five examples are introduced in the following directory.
./examples
The detailed explanation and how to run the example code is in the README.md in the example directory.

## 4. License
This is available for non-commercial purposes only.

