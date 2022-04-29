package envs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSfcSourceFlowBot_Decode(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		sd      *SfcSourceFlowBot
		args    args
		wantErr bool
		want    *SfcSourceFlowBot
	}{
		{
			name: "success",
			sd:   &SfcSourceFlowBot{},
			args: args{
				value: `SFB001={"subject":"Estado de pedido","providers":{"whatsapp":{"button_id":"5737b000000GmhG","owner_id":"00G7b000002vwIs"},"facebook":{"button_id":"5737b000000GmhG","owner_id":"00G7b000002vwIs"}}};SFB002={"subject":"Devoluciones y cancelaciones","providers":{"whatsapp":{"button_id":"5737b000000GmhG","owner_id":"00G7b000002vwIs"},"facebook":{"button_id":"5737b000000GmhG","owner_id":"00G7b000002vwIs"}}}`,
			},
			wantErr: false,
			want: &SfcSourceFlowBot{
				"SFB001": {
					Subject: "Estado de pedido",
					Providers: map[string]Provider{
						"whatsapp": {
							ButtonID: "5737b000000GmhG",
							OwnerID:  "00G7b000002vwIs",
						},
						"facebook": {
							ButtonID: "5737b000000GmhG",
							OwnerID:  "00G7b000002vwIs",
						},
					},
				},
				"SFB002": {
					Subject: "Devoluciones y cancelaciones",
					Providers: map[string]Provider{
						"whatsapp": {
							ButtonID: "5737b000000GmhG",
							OwnerID:  "00G7b000002vwIs",
						},
						"facebook": {
							ButtonID: "5737b000000GmhG",
							OwnerID:  "00G7b000002vwIs",
						},
					},
				},
			},
		},
		{
			name: "error parse",
			args: args{
				value: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.sd.Decode(tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("SfcSourceFlowBot.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, tt.sd)
		})
	}
}
