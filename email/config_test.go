package email

import "testing"

func TestWithFormValidation(t *testing.T) {
	type args struct {
		enabled bool
	}
	tests := []struct {
		name           string
		initial        Config
		args           args
		expectedResult Config
	}{
		{
			name: "Enable validation",
			initial: Config{
				ValidateForm: false,
			},
			args: args{
				enabled: true,
			},
			expectedResult: Config{
				ValidateForm: true,
			},
		},
		{
			name: "Disable validation",
			initial: Config{
				ValidateForm: true,
			},
			args: args{
				enabled: false,
			},
			expectedResult: Config{
				ValidateForm: false,
			},
		},
		{
			name: "No change when already set",
			initial: Config{
				ValidateForm: true,
			},
			args: args{
				enabled: true,
			},
			expectedResult: Config{
				ValidateForm: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithFormValidation(tt.args.enabled)
			got := opt(tt.initial)
			if got != tt.expectedResult {
				t.Errorf("WithFormValidation() = %v, want %v", got, tt.expectedResult)
			}
		})
	}
}
