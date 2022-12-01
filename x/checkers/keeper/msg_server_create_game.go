package keeper

import (
	"context"

	"github.com/alice/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CreateGame(goCtx context.Context, msg *types.MsgCreateGame) (*types.MsgCreateGameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx
	red, err := sdk.AccAddressFromBech32(msg.Red)
	if err != nil {
		return nil, error.New('invalid address for red')
	}

	return &types.MsgCreateGameResponse{}, nil
}
