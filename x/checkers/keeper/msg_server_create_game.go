package keeper

import (
	"context"
	"errors"
	"strconv"

	"github.com/alice/checkers/x/checkers/rules"
	"github.com/alice/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CreateGame(goCtx context.Context, msg *types.MsgCreateGame) (*types.MsgCreateGameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	red, err := sdk.AccAddressFromBech32(msg.Red)
	if err != nil {
		return nil, errors.New("invalid address for red")
	}
	black, err := sdk.AccAddressFromBech32(msg.Black)
	if err != nil {
		return nil, errors.New("invalid address for Black")
	}

	systemInfo, found := k.Keeper.GetSystemInfo(ctx)
	if !found {
		panic("systemInfo not found")
	}
	newIndex := strconv.FormatUint(systemInfo.NextId, 10)

	storedGame := types.StoredGame{
		Index: newIndex,
		Board: rules.New().String(),
		Turn:  red.String(),
		Black: black.String(),
		Red:   red.String(),
	}

	k.Keeper.SetStoredGame(ctx, storedGame)
	systemInfo.NextId++
	k.Keeper.SetSystemInfo(ctx, systemInfo)

	return &types.MsgCreateGameResponse{
		GameIndex: newIndex,
	}, nil
}
