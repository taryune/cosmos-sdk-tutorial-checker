package keeper_test

import (
	"testing"

	"github.com/alice/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestRejectGameByRedOneMoveRemovedGame(t *testing.T) {
	msgServer, keeper, context := setupMsgServerWithOneGameForPlayMove(t)
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
		NextId: 2,
	}, systemInfo)
	_, found = keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "1")
	require.False(t, found)
}

func TestRejectGameWrongByCreator(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	rejectGameResponse, err := msgServer.RejectGame(context, &types.MsgRejectGame{
		Creator:   alice,
		GameIndex: "1",
	})
	require.Nil(t, rejectGameResponse)
	require.Equal(t, alice+": message creator is not a player", err.Error())

}
func TestRejectGameByBlackNoMove(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	rejectGameResponse, err := msgServer.RejectGame(context, &types.MsgRejectGame{
		Creator:   bob,
		GameIndex: "1",
	})
	require.Nil(t, err)
	require.Equal(t, types.MsgRejectGameResponse{}, *rejectGameResponse)
}

func TestRejectGameByRedNoMove(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	rejectGameResponse, err := msgServer.RejectGame(context, &types.MsgRejectGame{
		Creator:   carol,
		GameIndex: "1",
	})
	require.Nil(t, err)
	require.Equal(t, types.MsgRejectGameResponse{}, *rejectGameResponse)
}

func TestRejectGameWrongByRedOneMove(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
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
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
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
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
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
