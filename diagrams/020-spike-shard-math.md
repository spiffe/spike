![SPIKE](../assets/spike-banner-lg.png)

## Shamir Secret Sharing Mathematics

### Threshold Scheme

**Parameters:**
- `t`: Threshold (minimum shards - 1)
- `n`: Total shards
- Need `t+1` shards to reconstruct

**Example Configuration:**
```
ShamirThreshold = 3
ShamirShares = 5

t = 3 - 1 = 2
n = 5

Need any 3 of 5 shards to reconstruct root key
```

### Polynomial Construction

```
Secret s is split using a random polynomial of degree t:

f(x) = a₀ + a₁x + a₂x² + ... + aₜx^t

Where:
- a₀ = s (the secret)
- a₁, a₂, ..., aₜ are random coefficients
- Operations in finite field (P256 curve)

Shards are points on this polynomial:
Shard₁ = f(1)
Shard₂ = f(2)
...
Shardₙ = f(n)

Any t+1 shards can reconstruct the polynomial and recover s = a₀
```

### Lagrange Interpolation (Reconstruction)

```
Given t+1 points (xᵢ, yᵢ), reconstruct polynomial:

f(x) = Σ yᵢ * Lᵢ(x)

Where Lᵢ(x) is the Lagrange basis polynomial:

Lᵢ(x) = Π (x - xⱼ) / (xᵢ - xⱼ)  for j ≠ i

Evaluate at x=0 to recover secret: s = f(0)
```