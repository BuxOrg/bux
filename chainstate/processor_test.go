package chainstate

import (
	"fmt"
	"github.com/centrifugal/centrifuge-go"
	"github.com/libsv/go-bt/v2"
	"github.com/stretchr/testify/assert"
	boom "github.com/tylertreat/BoomFilters"
	"testing"
)

func Test(t *testing.T) {
	type fields struct {
		filter *boom.StableBloomFilter
	}
	type args struct {
		item string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BloomProcessor{
				filter: tt.fields.filter,
			}
			assert.Equalf(t, tt.want, m.Test(tt.args.item), "Test(%v)", tt.args.item)
		})
	}
}

func TestBloomProcessor_Add(t *testing.T) {
	type fields struct {
		filter *boom.StableBloomFilter
	}
	type args struct {
		item string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "valid add",
			fields: fields{
				filter: boom.NewDefaultStableBloomFilter(uint(1000), float64(0.001)),
			},
			args: args{
				item: "006a",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BloomProcessor{
				filter: tt.fields.filter,
			}
			m.Add(tt.args.item)
			passes := m.Test("006a")
			assert.Truef(t, passes, "%v - test of filter failed", tt.name)
		})
	}
}

func TestBloomProcessor_AddFilter(t *testing.T) {
	type fields struct {
		filter *boom.StableBloomFilter
	}
	type args struct {
		filterType TransactionType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BloomProcessor{
				filter: tt.fields.filter,
			}
			m.AddFilter(tt.args.filterType)
		})
	}
}

func TestBloomProcessor_FilterMempoolPublishEvent(t *testing.T) {
	type fields struct {
		filter *boom.StableBloomFilter
	}
	type args struct {
		e centrifuge.ServerPublishEvent
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *bt.Tx
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &BloomProcessor{
				filter: tt.fields.filter,
			}
			got, err := p.FilterMempoolPublishEvent(tt.args.e)
			if !tt.wantErr(t, err, fmt.Sprintf("FilterMempoolPublishEvent(%v)", tt.args.e)) {
				return
			}
			assert.Equalf(t, tt.want, got, "FilterMempoolPublishEvent(%v)", tt.args.e)
		})
	}
}

func TestBloomProcessor_Reload(t *testing.T) {
	type fields struct {
		filter *boom.StableBloomFilter
	}
	type args struct {
		items []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BloomProcessor{
				filter: tt.fields.filter,
			}
			tt.wantErr(t, m.Reload(tt.args.items), fmt.Sprintf("Reload(%v)", tt.args.items))
		})
	}
}

func TestNewBloomProcessor(t *testing.T) {
	type args struct {
		maxCells          uint
		falsePositiveRate float64
	}
	tests := []struct {
		name string
		args args
		want *BloomProcessor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewBloomProcessor(tt.args.maxCells, tt.args.falsePositiveRate), "NewBloomProcessor(%v, %v)", tt.args.maxCells, tt.args.falsePositiveRate)
		})
	}
}
