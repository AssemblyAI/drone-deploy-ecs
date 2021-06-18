package main

// Need to refactor the mock functions to allow this to work
/*
func Test_determineBlueGreen(t *testing.T) {
	type args struct {
		e            types.ECSClient
		blueService  string
		greenService string
		cluster      string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				e:            deploy.MockECSClient{},
				blueService:  "test-cluster",
				greenService: "test-cluster-green",
				cluster:      "test-cluster",
			},
			want:    "test-cluster",
			want1:   "test-cluster-green",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := determineBlueGreen(tt.args.e, tt.args.blueService, tt.args.greenService, tt.args.cluster)
			if (err != nil) != tt.wantErr {
				t.Errorf("determineBlueGreen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("determineBlueGreen() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("determineBlueGreen() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

*/
