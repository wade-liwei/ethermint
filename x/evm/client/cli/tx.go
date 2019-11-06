package cli

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	emintkeys "github.com/cosmos/ethermint/keys"
	emintTypes "github.com/cosmos/ethermint/types"
	emintUtils "github.com/cosmos/ethermint/x/evm/client/utils"
	"github.com/cosmos/ethermint/x/evm/types"
)

// GetTxCmd defines the CLI commands regarding evm module transactions
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	evmTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "EVM transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	evmTxCmd.AddCommand(client.PostCommands(
		// TODO: Add back generating cosmos tx for Ethereum tx message

		GetCmdGenTx(cdc),
	)...)

	return evmTxCmd
}

// GetCmdGenTx generates an ethereum transaction wrapped in a Cosmos standard transaction
func GetCmdGenTx(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "generate-tx [amount] [gasprice] [fromAddress] [toAddr] [gasLimit] [payload]",
		Short: "generate eth tx wrapped in a Cosmos Standard tx",
		Args:  cobra.RangeArgs(3, 6),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: remove inputs and infer based on StdTx
			cliCtx := emintUtils.NewETHCLIContext().WithCodec(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			kb, err := emintkeys.NewKeyBaseFromHomeFlag()
			if err != nil {
				panic(err)
			}

			coins, err := sdk.ParseCoins(args[0])
			if err != nil {
				return err
			}

			gasLimit, err := strconv.ParseUint(args[4], 0, 64)
			if err != nil {
				return err
			}

			gasPrice, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}
			payload := args[5]

			addr, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				fmt.Println("err", err)
			}

			// TODO: Remove explicit photon check and check variables
			// Need to do conditional checks on optional variables prior to this call
			//
			msg := types.NewEmintMsg(0, nil, coins.AmountOf(emintTypes.DenomDefault), gasLimit, sdk.NewIntFromBigInt(new(big.Int).SetUint64(gasPrice)), []byte(payload), addr)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			// TODO: possibly overwrite gas values in txBldr
			return emintUtils.GenerateOrBroadcastETHMsgs(cliCtx, txBldr.WithKeybase(kb), []sdk.Msg{msg})
		},
	}
}