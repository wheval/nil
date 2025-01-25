
# 🌐 Working with Pair

---

## Overview

A **Pair** is a contract that facilitates the swapping and liquidity management of two tokens. The `UniswapV2Pair` contract is responsible for handling the liquidity operations, including minting, burning, and swapping.

---

## 💡 How to Use

### 1. Retrieve Reserves

To fetch the reserves of the pair, use the following command:

```bash
npx hardhat get-reserves --pair <Pair Address>
```

Replace `<Pair Address>` with the actual pair address.

### 2. Mint Liquidity

To mint liquidity and add it to the pair, use the following command:

```bash
npx hardhat mint --pair <Pair Address> --smart-account <User Smart Account Address> --amount0 <Amount of Token0> --amount1 <Amount of Token1>
```

Replace `<Pair Address>`, `<User Smart Account Address>`, `<Amount of Token0>`, and `<Amount of Token1>` with the actual values.

### 3. Swap Tokens

To swap token0 for token1, use the following command:

```bash
npx hardhat swap --pair <Pair Address> --smart-account <User Smart Account Address> --amount <Amount of Token0>
```

Replace `<Pair Address>`, `<User Smart Account Address>`, and `<Amount of Token0>` with the actual values.

### 4. Burn Liquidity

To burn liquidity and withdraw your share of tokens from the pair, use the following command:

```bash
npx hardhat burn --pair <Pair Address> --smart-account <User Smart Account Address>
```

Replace `<Pair Address>` and `<User Smart Account Address>` with the actual values.

---
