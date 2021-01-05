package zookeeper

import "testing"

func TestCompareSlice(t *testing.T) {
	type args struct {
		s1 []string
		s2 []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal",
			args: args{
				s1: []string{"1", "3", "2"},
				s2: []string{"1", "2", "3"},
			},
			want: true,
		},
		{
			name: "not equal",
			args: args{
				s1: []string{"1", "3", "2", "1"},
				s2: []string{"1", "2", "3"},
			},
			want: false,
		},
		{
			name: "nil",
			args: args{
				s1: nil,
				s2: nil,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareSlice(tt.args.s1, tt.args.s2); got != tt.want {
				t.Errorf("CompareSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
