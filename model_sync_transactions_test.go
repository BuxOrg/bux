package bux

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSyncTransaction_GetModelName will test the method GetModelName()
func TestSyncTransaction_GetModelName(t *testing.T) {
	t.Parallel()

	t.Run("valid name", func(t *testing.T) {
		syncTx := newSyncTransaction(testTxID, &SyncConfig{SyncOnChain: true, Broadcast: true}, New())
		require.NotNil(t, syncTx)
		assert.Equal(t, ModelSyncTransaction.String(), syncTx.GetModelName())
	})

	t.Run("missing config", func(t *testing.T) {
		syncTx := newSyncTransaction(testTxID, nil, New())
		require.Nil(t, syncTx)
	})
}

func Test_areParentsBroadcast(t *testing.T) {
	ctx, client, deferMe := CreateTestSQLiteClient(t, false, true, WithCustomTaskManager(&taskManagerMockBase{}))
	defer deferMe()

	opts := []ModelOps{WithClient(client)}

	tx := newTransaction(testTxHex, append(opts, New())...)
	txErr := tx.Save(ctx)
	require.NoError(t, txErr)

	tx = newTransaction(testTx2Hex, append(opts, New())...)
	txErr = tx.Save(ctx)
	require.NoError(t, txErr)

	tx = newTransaction(testTx3Hex, append(opts, New())...)
	txErr = tx.Save(ctx)
	require.NoError(t, txErr)

	// input of testTxID
	syncTx := newSyncTransaction("65bb8d2733298b2d3b441a871868d6323c5392facf0d3eced3a6c6a17dc84c10", &SyncConfig{SyncOnChain: false, Broadcast: false}, append(opts, New())...)
	syncTx.BroadcastStatus = SyncStatusComplete
	txErr = syncTx.Save(ctx)
	require.NoError(t, txErr)

	// input of testTxInID
	syncTx = newSyncTransaction("89fbccca3a5e2bfc8a161bf7f54e8cb5898e296ae8c23b620b89ed570711f931", &SyncConfig{SyncOnChain: false, Broadcast: false}, append(opts, New())...)
	txErr = syncTx.Save(ctx)
	require.NoError(t, txErr)

	type args struct {
		tx   *SyncTransaction
		opts []ModelOps
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "no parents",
			args: args{
				tx:   newSyncTransaction(testTxID3, &SyncConfig{SyncOnChain: true, Broadcast: true}, New()),
				opts: opts,
			},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "parent not broadcast",
			args: args{
				tx:   newSyncTransaction(testTxID2, &SyncConfig{SyncOnChain: true, Broadcast: true}, New()),
				opts: opts,
			},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "parent broadcast",
			args: args{
				tx:   newSyncTransaction(testTxID, &SyncConfig{SyncOnChain: true, Broadcast: true}, New()),
				opts: opts,
			},
			want:    true,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := areParentsBroadcast(ctx, tt.args.tx, tt.args.opts...)
			if !tt.wantErr(t, err, fmt.Sprintf("areParentsBroadcast(%v, %v, %v)", ctx, tt.args.tx, tt.args.opts)) {
				return
			}
			assert.Equalf(t, tt.want, got, "areParentsBroadcast(%v, %v, %v)", ctx, tt.args.tx, tt.args.opts)
		})
	}
}
