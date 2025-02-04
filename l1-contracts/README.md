# NilChain Verifier contract

- Contract `NilChain` is to be deployed on L1 Chain
- SyncCommittee refers to this L1 contract when submitting proof to get verified
- StateRoot updates are also done on this contract

## Build

1. install node dependencies

```sh
npm i
```


2. copy `.env.example` to `.env`

3. set all pre-requisite variables in .env
    - WALLET_ADDRESS
    - PRIVATE_KEY
  - This address is same as the address used for deployment and acts as the owner of the NilChain contract
  - The address is to be used when running SyncCommitee node


## Local Run

- For build pipeline or local testing, the contract is to be deployed on local Nil Node

### Please follow the steps mentioned below:

1. copy `.env.example` to `.env`
2. set all pre-requisite variables in .env
  - WALLET_ADDRESS
  - PRIVATE_KEY
3. This address is same as the address used for deployment and acts as the owner of the NilChain contract
4. The address is to be used when running SyncCommitee node

Try running some of the following tasks:

```shell
npx hardhat compile
npx hardhat deploy --network local
```
