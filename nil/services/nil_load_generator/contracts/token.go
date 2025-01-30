package contracts

import (
	"context"
	"crypto/ecdsa"

	"github.com/NilFoundation/nil/nil/client"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/cliservice"
	"github.com/ethereum/go-ethereum/crypto"
)

type Token struct {
	Contract
	Name              string
	OwnerKey          []byte
	OwnerSmartAccount SmartAccount
	Id                types.TokenId
}

func NewToken(contract Contract, name string, ownerSmartAccount SmartAccount) *Token {
	return &Token{
		Contract:          contract,
		Name:              name,
		OwnerKey:          crypto.CompressPubkey(&ownerSmartAccount.PrivateKey.PublicKey),
		OwnerSmartAccount: ownerSmartAccount,
		Id:                types.TokenId{0},
	}
}

func (c *Token) Deploy(service *cliservice.Service, deploySmartAccount SmartAccount) error {
	argsPacked, err := c.Abi.Pack("", c.Name, c.OwnerKey)
	if err != nil {
		return err
	}
	code := append(c.Contract.Code.Clone(), argsPacked...)
	c.Addr, err = DeployContract(service, deploySmartAccount.Addr, code)
	if err != nil {
		return err
	}
	c.Id = types.TokenId(c.Addr)
	return nil
}

func (c *Token) MintAndSend(ctx context.Context, client client.Client, service *cliservice.Service, smartAccountTo types.Address, mintAmount uint64) error {
	calldata, err := c.Abi.Pack("mintToken", types.NewValueFromUint64(mintAmount))
	if err != nil {
		return err
	}
	if err := sendExternalTransaction(ctx, client, service, calldata, c.Addr, c.OwnerSmartAccount.PrivateKey); err != nil {
		return err
	}
	calldata, err = c.Abi.Pack("sendToken", smartAccountTo, c.Id, types.NewValueFromUint64(mintAmount))
	if err != nil {
		return err
	}
	if err := sendExternalTransaction(ctx, client, service, calldata, c.Addr, c.OwnerSmartAccount.PrivateKey); err != nil {
		return err
	}
	return nil
}

func sendExternalTransaction(ctx context.Context, client client.Client, service *cliservice.Service, calldata types.Code, contractAddr types.Address, pk *ecdsa.PrivateKey) error {
	hash, err := client.SendExternalTransaction(ctx, calldata, contractAddr, pk, types.FeePack{})
	if err != nil {
		return err
	}
	_, err = service.WaitForReceiptCommitted(hash)
	if err != nil {
		return err
	}
	return nil
}
