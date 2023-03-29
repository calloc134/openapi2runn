package cmd

// 設定ファイルの設定を格納する構造体を定義
type config struct {
	AllowOverride bool `toml:"allowOverride"`
}

// パラメータを格納する構造体を定義
type paramSpec struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Example any
}

// メソッドを格納する構造体を定義
type baseSpec struct {
	Method string
	Params []paramSpec
	Body   []paramSpec
}

// パスを格納する構造体を定義
type pathSpec struct {
	DirName string
	Path    string
	Methods []baseSpec
}
