package evm

import (
	"fmt"
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/core"
	"github.com/cosmos/ethermint/x/evm/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcore "github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethvm "github.com/ethereum/go-ethereum/core/vm"
	ethparams "github.com/ethereum/go-ethereum/params"
)

// NewHandler returns a handler
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.EthereumTxMsg:
			return handleEthTxMsg(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized nameservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle a message to set name
func handleEthTxMsg(ctx sdk.Context, keeper Keeper, msg types.EthereumTxMsg) sdk.Result {
	chCfg := ethparams.MainnetChainConfig
	chCtx := core.NewChainContext() // TODO: Do we use ethermint implementation here?
	gp := new(ethcore.GasPool)
	db := keeper.csdb.WithContext(ctx)

	header := &ethtypes.Header{
		ParentHash:  ethcmn.Hash{},
		UncleHash:   ethcmn.Hash{},
		Coinbase:    ethcmn.Address{},
		Root:        ethcmn.Hash{},
		TxHash:      ethcmn.Hash{},
		ReceiptHash: ethcmn.Hash{},
		Bloom:       ethtypes.Bloom{},            // TODO: Add bloom filter to CSDB logs
		Difficulty:  ethparams.MinimumDifficulty, // TODO: Decide a constant to use, possibly introduce dynamic difficulty later
		Number:      big.NewInt(ctx.BlockHeight()),
		GasLimit:    ctx.BlockGasMeter().Limit(),
		GasUsed:     ctx.BlockGasMeter().GasConsumed(),
		Time:        uint64(time.Now().Unix()),
		Extra:       nil,
		MixDigest:   ethcmn.Hash{},
		Nonce:       ethtypes.BlockNonce{},
	}
	usedGas := new(uint64)
	vmCfg := ethvm.Config{}


	d := msg.Data
	// TODO: This should be part of Msg type
	tx := ethtypes.NewTransaction(d.AccountNonce, *d.Recipient, d.Amount, d.GasLimit, d.Price, d.Payload)

	// TODO: Setup keeper.TxIndex so it increments after each TX
	db.Prepare(tx.Hash(), header.Hash(), keeper.TxIndex(ctx))

	receipt, gas, err := keeper.applyTransaction(chCfg, chCtx, zeroAddress, gp, db, header, tx, usedGas, vmCfg)

	fmt.Printf("receipt: %v gas: %v err: %v", receipt, gas, err)

	return sdk.Result{GasUsed: gas}
}
