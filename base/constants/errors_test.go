package constants

import "testing"

func Test_applicationErrors_Error(t *testing.T) {
	tests := []struct {
		name string
		e    applicationErrors
		want string
	}{
		{
			name: "success",
			e:    ErrInterconnectionNotFound,
			want: "not found interconnection",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("applicationErrors.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
