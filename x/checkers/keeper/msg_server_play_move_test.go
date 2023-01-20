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

func setupMsgServerWithOneGameForPlayMove(t testing.TB) (types.MsgServer, keeper.Keeper, context.Context,
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

func TestPlayMove(t *testing.T) {
	msgServer, _, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)
	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	require.Nil(t, err)
	require.EqualValues(t, types.MsgPlayMoveResponse{
		CapturedX: -1,
		CapturedY: -1,
		Winner:    "*",
	}, *playMoveResponse)
}

func TestPlayMoveGameNotFound(t *testing.T) {
	msgServer, _, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)
	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "2",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	require.Nil(t, playMoveResponse)
	require.Equal(t, "2: game by id not found", err.Error())
}

func TestPlayMoveCreatorNotPlayer(t *testing.T) {
	msgServer, _, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)
	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   alice,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	require.Nil(t, playMoveResponse)
	require.Equal(t, alice+": message creator is not a player", err.Error())
}

func TestPlayMoveNotPlayerTurn(t *testing.T) {
	msgServer, _, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)

	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	require.Nil(t, playMoveResponse)
	require.Equal(t, "{red}: player tried to play out of turn", err.Error())
}

func TestPlayMoveWrongMove(t *testing.T) {
	msgServer, _, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)

	playMoveResponse, err := msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       1,
		ToY:       2,
	})
	require.Nil(t, playMoveResponse)
	require.Equal(t, "Already piece at destination position: {1 2}: wrong move", err.Error())

}

func TestPlayMoveCannotParseGame(t *testing.T) {
	msgServer, k, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)

	ctx := sdk.UnwrapSDKContext(context)
	storedGame, _ := k.GetStoredGame(ctx, "1")
	storedGame.Board = "not a board"
	k.SetStoredGame(ctx, storedGame)
	defer func() {
		r := recover()
		require.NotNil(t, r, "The code did not panic")
		require.Equal(t, r, "game cannot be parsed: invalid board string: not a board")
	}()
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
}

func TestPlayMove2Emitted(t *testing.T) {
	msgServer, _, context, ctrl, escrow := setupMsgServerWithOneGameForPlayMove(t)
	defer ctrl.Finish()
	escrow.ExpectAny(context)

	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   bob,
		GameIndex: "1",
		FromX:     1,
		FromY:     2,
		ToX:       2,
		ToY:       3,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator:   carol,
		GameIndex: "1",
		FromX:     0,
		FromY:     5,
		ToX:       1,
		ToY:       4,
	})
	ctx := sdk.UnwrapSDKContext(context)
	require.NotNil(t, ctx)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 2)
	require.Equal(t, "move-played", events[0].Type)
	require.EqualValues(t, []sdk.Attribute{
		{Key: "creator", Value: carol},
		{Key: "game-index", Value: "1"},
		{Key: "captured-x", Value: "-1"},
		{Key: "captured-y", Value: "-1"},
		{Key: "winner", Value: "*"},
		{Key: "board", Value: "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|*r******|**r*r*r*|*r*r*r*r|r*r*r*r*"},
	}, events[0].Attributes[6:])

}
