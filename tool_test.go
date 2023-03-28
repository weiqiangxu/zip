package zip

import "testing"

func TestTgzPacker_Pack(t *testing.T) {
	type args struct {
		sourceFullPath string
		targetFilePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				sourceFullPath: "/Users/Documents/sh",
				targetFilePath: "/Users/Documents/zz.tar.gz",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := NewTgzPacker()
			err := tp.Pack(tt.args.sourceFullPath, tt.args.targetFilePath)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTgzPacker_UnPack(t *testing.T) {
	type args struct {
		tarFileName string
		dstDir      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "um tar",
			args: args{
				tarFileName: "/Users/Documents/zz.tar.gz",
				dstDir:      "/Users/Documents/un_tar",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &TgzPacker{}
			if err := tp.UnPack(tt.args.tarFileName, tt.args.dstDir); (err != nil) != tt.wantErr {
				t.Errorf("UnPack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
