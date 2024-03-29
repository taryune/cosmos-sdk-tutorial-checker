package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/alice/checkers/testutil/keeper"
	"github.com/alice/checkers/x/checkers"
	"github.com/alice/checkers/x/checkers/keeper"
	"github.com/alice/checkers/x/checkers/testutil"
	"github.com/alice/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func setupMsgServerWithOneGameForRejectGame(t testing.TB) (types.MsgServer, keeper.Keeper, context.Context,
	*gomock.Controller, *testutil.MockBankEscrowKeeper) {
	ctrl := gomock.NewController(t)
	bankMock := testutil.NewMockBankEscrowKeeper(ctrl)
	k, ctx := keepertest.CheckersKeeperWithMocks(t, bankMock)
	checkers.InitGenesis(ctx, *k, *types.DefaultGenesis())
	server := keeper.NewMsgServerImpl(*k)
	context := sdk.WrapSDKContext(ctx)
	server.CreateGame(context, &types.MsgCreateGame{
		Creator: alice,
		Black:   bob,
		Red:     carol,
		Wager:   45,
		Denom:   "stake",
	})
	return server, *k, context, ctrl, bankMock
}

func TestRejectGameByRedOneMoveRemovedGame(t *testing.T) {
	msgServer, keeper, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     1,
		ToX:       2,
		ToY:       3,
	})
	msgServer.RejectGame(context, &types.MsgRejectGame{
		Creator:   carol,
		GameIndex: "1",
	})
	systemInfo, found := keeper.GetSystemInfo(sdk.UnwrapSDKContext(context))
	require.True(t, found)
	require.EqualValues(t, types.SystemInfo{
		NextId:        2,
		FifoHeadIndex: "-1",
		FifoTailIndex: "-1",
	}, systemInfo)
	_, found = keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "1")
	require.False(t, found)
}

func TestRejectGameWrongByCreator(t *testing.T) {
	msgServer, _, context, ctrl, _ := setupMsgServerWithOneGameForRejectGame(t)
	defer ctrl.Finish()
	rejectGameResponse, err := msgServer.RejectGame(context, &types.MsgRejectGame{
		Creator:   alice,
		GameIndex: "1",
	})
	require.Nil(t, rejectGameResponse)
	require.Equal(t, alice+": message creator is not a player", err.Error())

}
func TestRejectGameByBlackNoMove(t *testing.T) {
	msgServer, _, context, ctrl, _ := setupMsgServerWithOneGameForRejectGame(t)
	defer ctrl.Finish()
	rejectGameResponse, err := msgServer.RejectGame(context, &types.MsgRejectGame{
		Creator:   bob,
		GameIndex: "1",
	})
	require.Nil(t, err)
	require.Equal(t, types.MsgRejectGameResponse{}, *rejectGameResponse)
}

func TestRejectGameByRedNoMove(t *testing.T) {
	msgServer, _, context, ctrl, _ := setupMsgServerWithOneGameForRejectGame(t)
	defer ctrl.Finish()
	rejectGameResponse, err := msgServer.RejectGame(context, &types.MsgRejectGame{
		Creator:   carol,
		GameIndex: "1",
	})
	require.Nil(t, err)
	require.Equal(t, types.MsgRejectGameResponse{}, *rejectGameResponse)
}

func TestRejectGameWrongByRedOneMove(t *testing.T) {
	msgServer, _, context, ctrl, escrow := setupMsgServerWithOneGameForRejectGame(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)
	_, errMove := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	require.Nil(t, errMove)
	rejectGameResponse, err := msgServer.RejectGame(context, &types.MsgRejectGame{
		Creator:   bob,
		GameIndex: "1",
	})
	require.Nil(t, rejectGameResponse)
	require.Equal(t, "black player has already played", err.Error())
}

func TestRejectGameWrongByBlackOneMove(t *testing.T) {
	msgServer, _, context, ctrl, escrow := setupMsgServerWithOneGameForRejectGame(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)
	_, errMove := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	require.Nil(t, errMove)
	_, errMove1 := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})
	require.Nil(t, errMove1)
	rejectGameResponse, err := msgServer.RejectGame(context, &types.MsgRejectGame{
		Creator:   carol,
		GameIndex: "1",
	})
	require.Nil(t, rejectGameResponse)
	require.Equal(t, "red player has already played", err.Error())

}

func TestRejectGameEmitted(t *testing.T) {
	msgServer, _, context, ctrl, escrow := setupMsgServerWithOneGameForRejectGame(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)

	_, err := msgServer.RejectGame(context, &types.MsgRejectGame{
		Creator:   bob,
		GameIndex: "1",
	})
	require.Nil(t, err)
	ctx := sdk.UnwrapSDKContext(context)
	require.NotNil(t, ctx)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 2)
	require.EqualValues(t, sdk.StringEvent{
		Type: "game-rejected",
		Attributes: []sdk.Attribute{
			{Key: "creator", Value: bob},
			{Key: "game-index", Value: "1"},
		},
	}, events[0])

}

func TestRejectGameByBlackRefundedGas(t *testing.T) {
	msgServer, _, context, ctrl, _ := setupMsgServerWithOneGameForRejectGame(t)
	ctx := sdk.UnwrapSDKContext(context)
	defer ctrl.Finish()
	before := ctx.GasMeter().GasConsumed()
	msgServer.RejectGame(context, &types.MsgRejectGame{
		Creator:   bob,
		GameIndex: "1",
	})
	after := ctx.GasMeter().GasConsumed()
	require.LessOrEqual(t, after, before-5_000)
}
