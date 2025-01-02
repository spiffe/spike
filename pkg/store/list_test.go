package store

import "testing"

func TestKV_List(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *KV
		want  []string
	}{
		{
			name: "empty_store",
			setup: func() *KV {
				return &KV{
					data: make(map[string]*Secret),
				}
			},
			want: []string{},
		},
		{
			name: "non_empty_store",
			setup: func() *KV {
				return &KV{
					data: map[string]*Secret{
						"test/path": {
							Metadata: Metadata{
								CurrentVersion: 1,
							},
							Versions: map[int]Version{
								1: {
									Data: map[string]string{
										"key": "value",
									},
									Version: 1,
								},
							},
						},
					},
				}
			},
			want: []string{"test/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kv := tt.setup()
			got := kv.List()

			if len(got) != len(tt.want) {
				t.Errorf("got %v want %v", got, tt.want)
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got %v want %v", got, tt.want)
				}
			}
		})
	}
}
