package string_helper

import "testing"

func TestValidateIDList(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"valid-id-list",
			args{
				str: "12,34,56",
			},
			true,
		}, {
			"invalid-id-list",
			args{
				str: "aaa12,34,56",
			},
			false,
		}, {
			"invalid-id-list",
			args{
				str: "aaa12 oneof=1 and",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateIDList(tt.args.str); got != tt.want {
				t.Errorf("ValidateIDList() = %v, want %v", got, tt.want)
			}
		})
	}
}
